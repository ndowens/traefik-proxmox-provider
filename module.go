// Package caddy_proxmox_provider is a Caddy config provider that polls
// Proxmox VE and generates reverse-proxy routes from labels embedded in
// each LXC container's Notes field.
//
// Label format (one per line in the Notes field):
//
//	caddy.reverse_proxy myapp.example.com 192.168.1.10:8080
//	caddy.tls email@example.com
//
// Any line that does NOT start with "caddy." is ignored, so you can mix
// free-form notes with labels freely.
package caddy_proxmox_provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"

	"github.com/ndowens/caddy-proxmox-provider/internal/caddyconfig_builder"
	"github.com/ndowens/caddy-proxmox-provider/internal/proxmox"
)

func init() {
	caddy.RegisterModule(ProxmoxProvider{})
}

// ProxmoxProvider is a Caddy config loader that reads LXC container
// notes from a Proxmox VE cluster and generates Caddy configuration.
type ProxmoxProvider struct {
	// PollInterval controls how often Proxmox is polled (default: 30s).
	PollInterval caddy.Duration `json:"poll_interval,omitempty"`

	// APIEndpoint is the base URL of your Proxmox API, e.g. "https://pve.local:8006".
	APIEndpoint string `json:"api_endpoint"`

	// APITokenID is the full token identifier, e.g. "root@pam!caddy".
	APITokenID string `json:"api_token_id"`

	// APIToken is the secret token value.
	APIToken string `json:"api_token"`

	// ValidateSSL controls TLS certificate verification (default: true).
	ValidateSSL *bool `json:"validate_ssl,omitempty"`

	// LabelPrefix is the line prefix recognised as a label (default: "caddy.").
	LabelPrefix string `json:"label_prefix,omitempty"`

	logger *zap.Logger
	client *proxmox.Client
	cancel context.CancelFunc
	mu     sync.RWMutex
	config []byte // last generated Caddy JSON config
}

// CaddyModule returns the Caddy module information.
func (ProxmoxProvider) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.config_loaders.proxmox",
		New: func() caddy.Module { return new(ProxmoxProvider) },
	}
}

// Provision sets up the provider.
func (p *ProxmoxProvider) Provision(ctx caddy.Context) error {
	p.logger = ctx.Logger(p)

	if p.APIEndpoint == "" {
		return fmt.Errorf("api_endpoint is required")
	}
	if p.APITokenID == "" || p.APIToken == "" {
		return fmt.Errorf("api_token_id and api_token are required")
	}

	validateSSL := true
	if p.ValidateSSL != nil {
		validateSSL = *p.ValidateSSL
	}

	prefix := "caddy."
	if p.LabelPrefix != "" {
		prefix = p.LabelPrefix
	}

	interval := time.Duration(p.PollInterval)
	if interval == 0 {
		interval = 30 * time.Second
	}

	p.client = proxmox.NewClient(p.APIEndpoint, p.APITokenID, p.APIToken, validateSSL)

	pollCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	// Do an initial load synchronously so config is available immediately.
	if err := p.poll(prefix); err != nil {
		p.logger.Warn("initial Proxmox poll failed", zap.Error(err))
	}

	go p.pollLoop(pollCtx, interval, prefix)
	return nil
}

// LoadConfig implements caddy.ConfigLoader.
func (p *ProxmoxProvider) LoadConfig(_ caddy.Context) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.config == nil {
		return []byte("{}"), nil
	}
	return p.config, nil
}

// Cleanup stops the polling goroutine.
func (p *ProxmoxProvider) Cleanup() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *ProxmoxProvider) pollLoop(ctx context.Context, interval time.Duration, prefix string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.poll(prefix); err != nil {
				p.logger.Warn("Proxmox poll error", zap.Error(err))
			}
		}
	}
}

func (p *ProxmoxProvider) poll(prefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	containers, err := p.client.ListLXC(ctx)
	if err != nil {
		return fmt.Errorf("listing LXC containers: %w", err)
	}

	cfg, err := caddyconfig_builder.BuildConfig(containers, prefix, p.logger)
	if err != nil {
		return fmt.Errorf("building config: %w", err)
	}

	raw, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	p.mu.Lock()
	p.config = raw
	p.mu.Unlock()

	p.logger.Info("config updated from Proxmox", zap.Int("containers", len(containers)))
	return nil
}

// UnmarshalCaddyfile sets up the provider from Caddyfile tokens.
//
//	{
//	    config_loader proxmox {
//	        api_endpoint  https://pve.local:8006
//	        api_token_id  root@pam!caddy
//	        api_token     xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
//	        poll_interval 30s
//	        validate_ssl  true
//	        label_prefix  caddy.
//	    }
//	}
func (p *ProxmoxProvider) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "api_endpoint":
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.APIEndpoint = d.Val()
			case "api_token_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.APITokenID = d.Val()
			case "api_token":
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.APIToken = d.Val()
			case "poll_interval":
				if !d.NextArg() {
					return d.ArgErr()
				}
				dur, err := time.ParseDuration(d.Val())
				if err != nil {
					return d.Errf("invalid poll_interval: %v", err)
				}
				p.PollInterval = caddy.Duration(dur)
			case "validate_ssl":
				if !d.NextArg() {
					return d.ArgErr()
				}
				v := d.Val() == "true"
				p.ValidateSSL = &v
			case "label_prefix":
				if !d.NextArg() {
					return d.ArgErr()
				}
				p.LabelPrefix = d.Val()
			default:
				return d.Errf("unknown option: %s", d.Val())
			}
		}
	}
	return nil
}

// Interface guards
var (
	_ caddy.Module        = (*ProxmoxProvider)(nil)
	_ caddy.Provisioner   = (*ProxmoxProvider)(nil)
	_ caddy.CleanerUpper  = (*ProxmoxProvider)(nil)
	_ caddy.ConfigLoader  = (*ProxmoxProvider)(nil)
	_ caddyfile.Unmarshaler = (*ProxmoxProvider)(nil)
)

// Ensure caddyconfig is used (import side-effect)
var _ = caddyconfig.JSON
var _ = http.DefaultClient

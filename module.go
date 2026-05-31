// Package caddy_proxmox_provider is a Caddy app module that polls
// Proxmox VE and hot-reloads reverse-proxy routes from labels in each
// LXC container's Notes field.
//
// It registers itself as a Caddy app ("caddy.apps.proxmox") so it can
// be declared in the global JSON config or via the Caddyfile global
// options block using the "proxmox" directive.
package caddy_proxmox_provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"go.uber.org/zap"

	builder "github.com/ndowens/caddy-proxmox-provider/internal/caddyconfig_builder"
	"github.com/ndowens/caddy-proxmox-provider/internal/proxmox"
)

func init() {
	caddy.RegisterModule(ProxmoxApp{})
	httpcaddyfile.RegisterGlobalOption("proxmox", parseGlobalOption)
}

// ProxmoxApp is the Caddy app module. It runs as a background service
// inside Caddy, polling Proxmox and updating routes via caddy.Load().
type ProxmoxApp struct {
	// PollInterval controls how often Proxmox is polled (default: 30s).
	PollInterval caddy.Duration `json:"poll_interval,omitempty"`

	// APIEndpoint is the base URL, e.g. "https://pve.local:8006".
	APIEndpoint string `json:"api_endpoint"`

	// APITokenID is the full token identifier, e.g. "root@pam!caddy".
	APITokenID string `json:"api_token_id"`

	// APIToken is the secret token value.
	APIToken string `json:"api_token"`

	// ValidateSSL controls TLS certificate verification (default: true).
	ValidateSSL *bool `json:"validate_ssl,omitempty"`

	// LabelPrefix is the line prefix treated as a label (default: "caddy.").
	LabelPrefix string `json:"label_prefix,omitempty"`

	logger *zap.Logger
	client *proxmox.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// CaddyModule returns the Caddy module information.
func (ProxmoxApp) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "proxmox",
		New: func() caddy.Module { return new(ProxmoxApp) },
	}
}

// Provision sets up the app.
func (a *ProxmoxApp) Provision(ctx caddy.Context) error {
	a.logger = ctx.Logger(a)

	if a.APIEndpoint == "" {
		return fmt.Errorf("api_endpoint is required")
	}
	if a.APITokenID == "" || a.APIToken == "" {
		return fmt.Errorf("api_token_id and api_token are required")
	}

	validateSSL := true
	if a.ValidateSSL != nil {
		validateSSL = *a.ValidateSSL
	}
	if a.LabelPrefix == "" {
		a.LabelPrefix = "caddy."
	}
	if a.PollInterval == 0 {
		a.PollInterval = caddy.Duration(30 * time.Second)
	}

	a.client = proxmox.NewClient(a.APIEndpoint, a.APITokenID, a.APIToken, validateSSL)
	a.ctx, a.cancel = context.WithCancel(ctx)
	return nil
}

// Start begins polling Proxmox. Implements caddy.App.
func (a *ProxmoxApp) Start() error {
	// Initial poll — errors are non-fatal so Caddy still starts.
	if err := a.poll(); err != nil {
		a.logger.Warn("initial Proxmox poll failed", zap.Error(err))
	}

	go a.pollLoop()
	return nil
}

// Stop cancels the polling goroutine. Implements caddy.App.
func (a *ProxmoxApp) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}

func (a *ProxmoxApp) pollLoop() {
	ticker := time.NewTicker(time.Duration(a.PollInterval))
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if err := a.poll(); err != nil {
				a.logger.Warn("Proxmox poll error", zap.Error(err))
			}
		}
	}
}

func (a *ProxmoxApp) poll() error {
	ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	containers, err := a.client.ListLXC(ctx)
	if err != nil {
		return fmt.Errorf("listing LXC containers: %w", err)
	}

	cfg, err := builder.BuildConfig(containers, a.LabelPrefix, a.logger)
	if err != nil {
		return fmt.Errorf("building config: %w", err)
	}

	if err := a.applyConfig(cfg); err != nil {
		return fmt.Errorf("applying config: %w", err)
	}

	a.logger.Info("routes updated from Proxmox",
		zap.Int("containers", len(containers)),
	)
	return nil
}

// applyConfig pushes the generated HTTP server config to Caddy's admin API
// using a PATCH request, which merges it into the running config without
// disturbing other apps (TLS, logging, etc.).
func (a *ProxmoxApp) applyConfig(generated *builder.CaddyConfig) error {
	if generated.Apps.HTTP == nil {
		return nil
	}

	httpRaw, err := json.Marshal(generated.Apps.HTTP)
	if err != nil {
		return fmt.Errorf("marshalling HTTP app: %w", err)
	}

	// PATCH /config/apps/http replaces only the http app in the live config.
	adminAddr := caddy.DefaultAdminListen
	url := "http://" + adminAddr + "/config/apps/http"

	req, err := http.NewRequestWithContext(a.ctx, http.MethodPatch, url, bytes.NewReader(httpRaw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("admin API PATCH: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("admin API returned %s: %s", resp.Status, body)
	}
	return nil
}

// UnmarshalCaddyfile parses the global "proxmox" block.
func (a *ProxmoxApp) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "api_endpoint":
				if !d.NextArg() {
					return d.ArgErr()
				}
				a.APIEndpoint = d.Val()
			case "api_token_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				a.APITokenID = d.Val()
			case "api_token":
				if !d.NextArg() {
					return d.ArgErr()
				}
				a.APIToken = d.Val()
			case "poll_interval":
				if !d.NextArg() {
					return d.ArgErr()
				}
				dur, err := time.ParseDuration(d.Val())
				if err != nil {
					return d.Errf("invalid poll_interval: %v", err)
				}
				a.PollInterval = caddy.Duration(dur)
			case "validate_ssl":
				if !d.NextArg() {
					return d.ArgErr()
				}
				v := d.Val() == "true"
				a.ValidateSSL = &v
			case "label_prefix":
				if !d.NextArg() {
					return d.ArgErr()
				}
				a.LabelPrefix = d.Val()
			default:
				return d.Errf("unknown option: %s", d.Val())
			}
		}
	}
	return nil
}

// parseGlobalOption is the httpcaddyfile handler for the "proxmox" global block.
func parseGlobalOption(d *caddyfile.Dispenser, _ any) (any, error) {
	app := new(ProxmoxApp)
	if err := app.UnmarshalCaddyfile(d); err != nil {
		return nil, err
	}
	appJSON, err := json.Marshal(app)
	if err != nil {
		return nil, err
	}
	return httpcaddyfile.App{
		Name:  "proxmox",
		Value: appJSON,
	}, nil
}

// Interface guards
var (
	_ caddy.App                   = (*ProxmoxApp)(nil)
	_ caddy.Provisioner           = (*ProxmoxApp)(nil)
	_ caddyfile.Unmarshaler       = (*ProxmoxApp)(nil)
)

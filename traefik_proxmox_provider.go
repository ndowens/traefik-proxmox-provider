// Package traefik_proxmox_provider is a plugin to use a proxmox cluster as a provider for Traefik.
package traefik_proxmox_provider

import (
	"context"
	"encoding/json"

	"github.com/NX211/traefik-proxmox-provider/provider"
)

// Config the plugin configuration.
type Config struct {
	PollInterval   string `json:"pollInterval" yaml:"pollInterval" toml:"pollInterval"`
	ApiEndpoint    string `json:"apiEndpoint" yaml:"apiEndpoint" toml:"apiEndpoint"`
	ApiTokenId     string `json:"apiTokenId" yaml:"apiTokenId" toml:"apiTokenId"`
	ApiToken       string `json:"apiToken" yaml:"apiToken" toml:"apiToken"`
	ApiLogging     string `json:"apiLogging" yaml:"apiLogging" toml:"apiLogging"`
	ApiValidateSSL string `json:"apiValidateSSL" yaml:"apiValidateSSL" toml:"apiValidateSSL"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	cfg := provider.CreateConfig()
	return &Config{
		PollInterval:   cfg.PollInterval,
		ApiEndpoint:    cfg.ApiEndpoint,
		ApiTokenId:     cfg.ApiTokenId,
		ApiToken:       cfg.ApiToken,
		ApiLogging:     cfg.ApiLogging,
		ApiValidateSSL: cfg.ApiValidateSSL,
	}
}

// Provider a plugin.
type Provider struct {
	provider *provider.Provider
}

// New creates a new Provider plugin.
func New(ctx context.Context, config *Config, name string) (*Provider, error) {
	providerConfig := &provider.Config{
		PollInterval:   config.PollInterval,
		ApiEndpoint:    config.ApiEndpoint,
		ApiTokenId:     config.ApiTokenId,
		ApiToken:       config.ApiToken,
		ApiLogging:     config.ApiLogging,
		ApiValidateSSL: config.ApiValidateSSL,
	}

	innerProvider, err := provider.New(ctx, providerConfig, name)
	if err != nil {
		return nil, err
	}

	return &Provider{
		provider: innerProvider,
	}, nil
}

// Init initializes the provider.
func (p *Provider) Init() error {
	return p.provider.Init()
}

// Provide creates and sends dynamic configuration.
func (p *Provider) Provide(cfgChan chan<- json.Marshaler) error {
	return p.provider.Provide(cfgChan)
}

// Stop the provider.
func (p *Provider) Stop() error {
	return p.provider.Stop()
}

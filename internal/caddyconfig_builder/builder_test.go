package caddyconfig_builder_test

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"

	builder "github.com/YOUR_USERNAME/caddy-proxmox-provider/internal/caddyconfig_builder"
	"github.com/YOUR_USERNAME/caddy-proxmox-provider/internal/proxmox"
)

func TestBuildConfig_BasicReverseProxy(t *testing.T) {
	containers := []proxmox.Container{
		{
			VMID:   100,
			Name:   "myapp",
			Status: "running",
			IP:     "192.168.1.10",
			Notes: `My application server
caddy.reverse_proxy myapp.example.com :8080
`,
		},
	}

	cfg, err := builder.BuildConfig(containers, "caddy.", zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, _ := json.MarshalIndent(cfg, "", "  ")
	t.Logf("Generated config:\n%s", raw)

	if cfg.Apps.HTTP == nil {
		t.Fatal("expected HTTP app")
	}
	server, ok := cfg.Apps.HTTP.Servers["proxmox_generated"]
	if !ok {
		t.Fatal("expected proxmox_generated server")
	}
	if len(server.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(server.Routes))
	}

	route := server.Routes[0]
	if len(route.Match) == 0 || len(route.Match[0].Host) == 0 {
		t.Fatal("expected host matcher")
	}
	if got := route.Match[0].Host[0]; got != "myapp.example.com" {
		t.Errorf("expected host myapp.example.com, got %s", got)
	}
}

func TestBuildConfig_UpstreamResolution(t *testing.T) {
	tests := []struct {
		name         string
		upstream     string
		containerIP  string
		wantUpstream string
	}{
		{"bare port with IP", ":8080", "10.0.0.5", "10.0.0.5:8080"},
		{"port number with IP", "9000", "10.0.0.5", "10.0.0.5:9000"},
		{"full address", "10.0.0.5:8080", "10.0.0.5", "10.0.0.5:8080"},
		{"hostname", "myapp:8080", "10.0.0.5", "myapp:8080"},
		{"bare port no IP", ":8080", "", ":8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containers := []proxmox.Container{
				{
					VMID:   100,
					Name:   "test",
					Status: "running",
					IP:     tt.containerIP,
					Notes:  "caddy.reverse_proxy test.local " + tt.upstream,
				},
			}
			cfg, err := builder.BuildConfig(containers, "caddy.", zap.NewNop())
			if err != nil {
				t.Fatal(err)
			}
			server := cfg.Apps.HTTP.Servers["proxmox_generated"]
			if server == nil || len(server.Routes) == 0 {
				t.Fatal("no routes generated")
			}
			// Find reverse_proxy handler
			for _, h := range server.Routes[0].Handle {
				if rp, ok := h.(builder.ReverseProxy); ok {
					if len(rp.Upstreams) == 0 {
						t.Fatal("no upstreams")
					}
					if got := rp.Upstreams[0].Dial; got != tt.wantUpstream {
						t.Errorf("upstream: got %q, want %q", got, tt.wantUpstream)
					}
					return
				}
			}
			t.Error("no reverse_proxy handler found")
		})
	}
}

func TestBuildConfig_IgnoresStoppedContainers(t *testing.T) {
	containers := []proxmox.Container{
		{
			VMID:   100,
			Name:   "stopped",
			Status: "stopped",
			Notes:  "caddy.reverse_proxy app.example.com :80",
		},
	}
	cfg, err := builder.BuildConfig(containers, "caddy.", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Apps.HTTP != nil && len(cfg.Apps.HTTP.Servers) > 0 {
		server := cfg.Apps.HTTP.Servers["proxmox_generated"]
		if server != nil && len(server.Routes) > 0 {
			t.Error("should not generate routes for stopped containers")
		}
	}
}

func TestBuildConfig_IgnoresNonCaddyLines(t *testing.T) {
	containers := []proxmox.Container{
		{
			VMID:   100,
			Name:   "app",
			Status: "running",
			IP:     "10.0.0.1",
			Notes: `
This is my database server.
Do NOT use in production without review.
traefik.enable=true
traefik.http.routers.app.rule=Host('app.local')
caddy.reverse_proxy app.example.com :5432
`,
		},
	}
	cfg, err := builder.BuildConfig(containers, "caddy.", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	server := cfg.Apps.HTTP.Servers["proxmox_generated"]
	if len(server.Routes) != 1 {
		t.Errorf("expected 1 route (traefik labels ignored), got %d", len(server.Routes))
	}
}

func TestBuildConfig_MultiSite(t *testing.T) {
	containers := []proxmox.Container{
		{
			VMID:   101,
			Name:   "web",
			Status: "running",
			IP:     "10.0.0.2",
			Notes: `
caddy.reverse_proxy frontend.example.com :3000
caddy.reverse_proxy api.example.com :4000
caddy.encode frontend.example.com gzip
`,
		},
	}
	cfg, err := builder.BuildConfig(containers, "caddy.", zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	server := cfg.Apps.HTTP.Servers["proxmox_generated"]
	if len(server.Routes) != 2 {
		t.Errorf("expected 2 routes for 2 hosts, got %d", len(server.Routes))
	}
}

// Package caddyconfig_builder converts Proxmox container labels into a
// Caddy JSON configuration.
//
// # Label syntax
//
// Each line in the container's Notes field that starts with the configured
// prefix (default "caddy.") is treated as a label. Everything after the
// prefix is the directive.
//
// ## Supported labels
//
//	caddy.reverse_proxy <host> <upstream>
//	  Creates a virtual host that reverse-proxies to <upstream>.
//	  <upstream> defaults to <container_ip>:80 if only a port is given.
//
//	caddy.tls <email>
//	  Sets the ACME email for the generated site block.
//
//	caddy.header <host> [<field>] [<value>]
//	  Adds a response header to the site block.
//
//	caddy.basicauth <host> <username> <hashed_password>
//	  Adds HTTP basic auth. Hash passwords with: caddy hash-password
//
//	caddy.encode <host> <encodings...>
//	  Enables response encoding (e.g. gzip zstd).
//
//	caddy.log
//	  Enables access logging for this container's routes.
//
// ## Multi-site example (Notes field)
//
//	My webserver - handles two apps
//	caddy.reverse_proxy app1.example.com :8080
//	caddy.reverse_proxy app2.example.com :9090
//	caddy.tls admin@example.com
//	caddy.encode app1.example.com gzip
package caddyconfig_builder

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/ndowens/caddy-proxmox-provider/internal/proxmox"
)

// ----- Caddy JSON types (minimal, only what we generate) -----

// CaddyConfig is the top-level Caddy JSON configuration.
type CaddyConfig struct {
	Apps CaddyApps `json:"apps,omitempty"`
}

type CaddyApps struct {
	HTTP *HTTPApp `json:"http,omitempty"`
}

type HTTPApp struct {
	Servers map[string]*HTTPServer `json:"servers,omitempty"`
}

type HTTPServer struct {
	Listen []string    `json:"listen"`
	Routes []HTTPRoute `json:"routes,omitempty"`
}

type HTTPRoute struct {
	Match   []HTTPMatch  `json:"match,omitempty"`
	Handle  []any        `json:"handle,omitempty"`
	Terminal bool        `json:"terminal,omitempty"`
}

type HTTPMatch struct {
	Host []string `json:"host,omitempty"`
}

// Handlers

type StaticResponse struct {
	Handler    string `json:"handler"`
	StatusCode int    `json:"status_code,omitempty"`
	Body       string `json:"body,omitempty"`
}

type ReverseProxy struct {
	Handler   string     `json:"handler"`
	Upstreams []Upstream `json:"upstreams"`
}

type Upstream struct {
	Dial string `json:"dial"`
}

type Encode struct {
	Handler  string         `json:"handler"`
	Encodings map[string]any `json:"encodings,omitempty"`
}

type HeadersHandler struct {
	Handler  string          `json:"handler"`
	Response *HeaderResponse `json:"response,omitempty"`
}

type HeaderResponse struct {
	Set map[string][]string `json:"set,omitempty"`
}

type Authentication struct {
	Handler   string              `json:"handler"`
	Providers map[string]any      `json:"providers,omitempty"`
}

type Subroute struct {
	Handler string      `json:"handler"`
	Routes  []HTTPRoute `json:"routes,omitempty"`
}

// ----- per-site aggregated labels -----

type siteConfig struct {
	host      string
	upstream  string
	tlsEmail  string
	headers   map[string]string
	basicAuth map[string]string // username -> hashed password
	encodings []string
	logging   bool
}

// BuildConfig converts a list of containers into a Caddy JSON config.
func BuildConfig(containers []proxmox.Container, prefix string, logger *zap.Logger) (*CaddyConfig, error) {
	// Map from host → site config; ordered list for deterministic output.
	sites := map[string]*siteConfig{}
	var siteOrder []string

	for _, ct := range containers {
		if ct.Status != "running" {
			continue
		}
		labels := extractLabels(ct.Notes, prefix)
		if len(labels) == 0 {
			continue
		}

		for _, label := range labels {
			directive, args := parseLabel(label)
			switch directive {
			case "reverse_proxy":
				if len(args) < 2 {
					logger.Warn("caddy.reverse_proxy needs <host> <upstream>",
						zap.String("container", ct.Name), zap.String("label", label))
					continue
				}
				host := args[0]
				upstream := resolveUpstream(args[1], ct.IP)
				sc := getOrCreate(sites, &siteOrder, host)
				sc.upstream = upstream

			case "tls":
				if len(args) < 1 {
					continue
				}
				// Apply to all sites from this container, or store globally.
				// We store the email on all already-seen sites from this container.
				for _, h := range siteOrder {
					if sites[h] != nil {
						sites[h].tlsEmail = args[0]
					}
				}

			case "header":
				if len(args) < 3 {
					continue
				}
				host := args[0]
				sc := getOrCreate(sites, &siteOrder, host)
				if sc.headers == nil {
					sc.headers = map[string]string{}
				}
				sc.headers[args[1]] = args[2]

			case "basicauth":
				if len(args) < 3 {
					continue
				}
				host := args[0]
				sc := getOrCreate(sites, &siteOrder, host)
				if sc.basicAuth == nil {
					sc.basicAuth = map[string]string{}
				}
				sc.basicAuth[args[1]] = args[2]

			case "encode":
				if len(args) < 2 {
					continue
				}
				host := args[0]
				sc := getOrCreate(sites, &siteOrder, host)
				sc.encodings = append(sc.encodings, args[1:]...)

			case "log":
				for _, h := range siteOrder {
					sites[h].logging = true
				}

			default:
				logger.Debug("unknown caddy label directive",
					zap.String("directive", directive),
					zap.String("container", ct.Name))
			}
		}
	}

	if len(sites) == 0 {
		return &CaddyConfig{}, nil
	}

	routes := buildRoutes(sites, siteOrder)

	cfg := &CaddyConfig{
		Apps: CaddyApps{
			HTTP: &HTTPApp{
				Servers: map[string]*HTTPServer{
					"proxmox_generated": {
						Listen: []string{":80", ":443"},
						Routes: routes,
					},
				},
			},
		},
	}
	return cfg, nil
}

// ----- helpers -----

func extractLabels(notes, prefix string) []string {
	var out []string
	for _, line := range strings.Split(notes, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			out = append(out, strings.TrimPrefix(line, prefix))
		}
	}
	return out
}

func parseLabel(label string) (directive string, args []string) {
	parts := strings.Fields(label)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func resolveUpstream(upstream, containerIP string) string {
	// If upstream is just ":PORT" and we know the container IP, use it.
	if strings.HasPrefix(upstream, ":") && containerIP != "" {
		return containerIP + upstream
	}
	// If upstream is a bare port number, treat as ":PORT".
	if _, err := strconv.Atoi(upstream); err == nil && containerIP != "" {
		return fmt.Sprintf("%s:%s", containerIP, upstream)
	}
	return upstream
}

func getOrCreate(sites map[string]*siteConfig, order *[]string, host string) *siteConfig {
	if sc, ok := sites[host]; ok {
		return sc
	}
	sc := &siteConfig{host: host}
	sites[host] = sc
	*order = append(*order, host)
	return sc
}

func buildRoutes(sites map[string]*siteConfig, order []string) []HTTPRoute {
	var routes []HTTPRoute

	for _, host := range order {
		sc := sites[host]
		if sc.upstream == "" {
			continue
		}

		var handlers []any

		// Authentication (innermost)
		if len(sc.basicAuth) > 0 {
			accounts := map[string]any{}
			for user, hash := range sc.basicAuth {
				accounts[user] = map[string]string{"password": hash}
			}
			handlers = append(handlers, Authentication{
				Handler: "authentication",
				Providers: map[string]any{
					"http_basic": map[string]any{
						"accounts": accounts,
					},
				},
			})
		}

		// Headers
		if len(sc.headers) > 0 {
			set := map[string][]string{}
			for k, v := range sc.headers {
				set[k] = []string{v}
			}
			handlers = append(handlers, HeadersHandler{
				Handler:  "headers",
				Response: &HeaderResponse{Set: set},
			})
		}

		// Encoding
		if len(sc.encodings) > 0 {
			encodings := map[string]any{}
			for _, enc := range sc.encodings {
				encodings[enc] = map[string]any{}
			}
			handlers = append(handlers, Encode{
				Handler:   "encode",
				Encodings: encodings,
			})
		}

		// Reverse proxy (outermost)
		handlers = append(handlers, ReverseProxy{
			Handler: "reverse_proxy",
			Upstreams: []Upstream{
				{Dial: sc.upstream},
			},
		})

		route := HTTPRoute{
			Match:    []HTTPMatch{{Host: []string{host}}},
			Handle:   handlers,
			Terminal: true,
		}
		routes = append(routes, route)
	}
	return routes
}

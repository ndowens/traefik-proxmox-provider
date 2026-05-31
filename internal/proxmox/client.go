// Package proxmox provides a minimal Proxmox VE API client focused on
// listing LXC containers and reading their configuration (including Notes).
package proxmox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Container represents an LXC container returned by the Proxmox API.
type Container struct {
	// VMID is the numeric container ID.
	VMID int `json:"vmid"`
	// Name is the container hostname / name.
	Name string `json:"name"`
	// Status is "running", "stopped", etc.
	Status string `json:"status"`
	// Node is the cluster node that hosts this container.
	Node string `json:"node"`
	// Notes is the raw text from the container's Notes field.
	Notes string `json:"notes"`
	// IP is the primary IP address (populated from network config).
	IP string `json:"-"`
}

// Client is a thin HTTP client for the Proxmox REST API.
type Client struct {
	baseURL     string
	tokenID     string
	token       string
	httpClient  *http.Client
}

// NewClient creates a new Proxmox API client.
func NewClient(baseURL, tokenID, token string, validateSSL bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !validateSSL}, //nolint:gosec
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		tokenID:    tokenID,
		token:      token,
		httpClient: &http.Client{Transport: tr},
	}
}

// ListLXC returns all LXC containers across all nodes in the cluster,
// including their Notes and primary IP address.
func (c *Client) ListLXC(ctx context.Context) ([]Container, error) {
	nodes, err := c.listNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}

	var all []Container
	for _, node := range nodes {
		containers, err := c.listNodeLXC(ctx, node)
		if err != nil {
			// Non-fatal: log and continue with other nodes.
			continue
		}
		for i := range containers {
			containers[i].Node = node
			// Fetch detailed config to get Notes + network info.
			if err := c.enrichContainer(ctx, node, &containers[i]); err != nil {
				continue
			}
		}
		all = append(all, containers...)
	}
	return all, nil
}

// ---- private helpers ----

type apiResponse[T any] struct {
	Data T `json:"data"`
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	url := c.baseURL + "/api2/json" + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "PVEAPIToken="+c.tokenID+"="+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("proxmox API %s: %s", resp.Status, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

type nodeInfo struct {
	Node string `json:"node"`
}

func (c *Client) listNodes(ctx context.Context) ([]string, error) {
	var resp apiResponse[[]nodeInfo]
	if err := c.get(ctx, "/nodes", &resp); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(resp.Data))
	for _, n := range resp.Data {
		names = append(names, n.Node)
	}
	return names, nil
}

type lxcInfo struct {
	VMID   int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *Client) listNodeLXC(ctx context.Context, node string) ([]Container, error) {
	var resp apiResponse[[]lxcInfo]
	if err := c.get(ctx, "/nodes/"+node+"/lxc", &resp); err != nil {
		return nil, err
	}
	out := make([]Container, len(resp.Data))
	for i, lxc := range resp.Data {
		out[i] = Container{
			VMID:   lxc.VMID,
			Name:   lxc.Name,
			Status: lxc.Status,
		}
	}
	return out, nil
}

// lxcConfig is the subset of LXC config fields we care about.
type lxcConfig struct {
	Description string `json:"description"` // Proxmox stores Notes as "description"
	// Network interfaces are returned as "net0", "net1", etc.
	// We parse them dynamically below.
}

func (c *Client) enrichContainer(ctx context.Context, node string, ct *Container) error {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/config", node, ct.VMID)

	// Proxmox returns a flat JSON object; we need the raw map to parse net* keys.
	var resp apiResponse[map[string]json.RawMessage]
	if err := c.get(ctx, path, &resp); err != nil {
		return err
	}

	if raw, ok := resp.Data["description"]; ok {
		var notes string
		if err := json.Unmarshal(raw, &notes); err == nil {
			ct.Notes = notes
		}
	}

	// Extract primary IP from net0 (format: "name=eth0,bridge=vmbr0,ip=192.168.1.10/24,...")
	if raw, ok := resp.Data["net0"]; ok {
		var netStr string
		if err := json.Unmarshal(raw, &netStr); err == nil {
			ct.IP = parseNetIP(netStr)
		}
	}

	return nil
}

// parseNetIP extracts the IP address (without CIDR) from a Proxmox net string.
func parseNetIP(net string) string {
	for _, part := range strings.Split(net, ",") {
		if strings.HasPrefix(part, "ip=") {
			ip := strings.TrimPrefix(part, "ip=")
			// Strip CIDR notation
			if idx := strings.Index(ip, "/"); idx != -1 {
				ip = ip[:idx]
			}
			if ip != "dhcp" && ip != "" {
				return ip
			}
		}
	}
	return ""
}

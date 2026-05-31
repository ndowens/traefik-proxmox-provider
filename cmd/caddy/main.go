// Command caddy is a custom Caddy build that includes the Proxmox provider.
// Build with: go build -o caddy ./cmd/caddy
package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	// Import Caddy standard modules
	_ "github.com/caddyserver/caddy/v2/modules/standard"

	// Import this plugin
	_ "github.com/ndowens/caddy-proxmox-provider"
)

func main() {
	caddycmd.Main()
}

# caddy-proxmox-provider

A [Caddy](https://caddyserver.com/) config-loader plugin that polls a
[Proxmox VE](https://www.proxmox.com/) cluster and automatically generates
reverse-proxy routes from labels embedded in each LXC container's **Notes**
field.

Inspired by [traefik-proxmox-provider](https://github.com/NX211/traefik-proxmox-provider).

---

## How it works

1. Caddy starts and loads this plugin as its config source.
2. The plugin polls the Proxmox API on a configurable interval.
3. For every **running** LXC container it reads the Notes field.
4. Lines starting with `caddy.` are parsed as labels.
5. The plugin generates a Caddy JSON config and hot-reloads it with zero
   downtime.

---

## Proxmox Notes label syntax

Write one label per line in the container's **Notes** field.
All other lines are ignored, so you can mix free-form notes with labels.

```
My webserver — hosts the company intranet
Updated 2024-01-15

caddy.reverse_proxy intranet.example.com :8080
caddy.tls admin@example.com
caddy.encode intranet.example.com gzip zstd
```

### Supported labels

| Label | Args | Description |
|-------|------|-------------|
| `caddy.reverse_proxy` | `<host> <upstream>` | Create a reverse-proxy virtual host |
| `caddy.tls` | `<email>` | Set ACME email for all routes in this container |
| `caddy.header` | `<host> <field> <value>` | Add a response header |
| `caddy.basicauth` | `<host> <user> <bcrypt-hash>` | Enable HTTP basic auth |
| `caddy.encode` | `<host> <enc…>` | Enable response encoding (e.g. `gzip`, `zstd`) |
| `caddy.log` | _(none)_ | Enable access logging for this container's routes |

### Upstream shorthand

The `<upstream>` argument in `caddy.reverse_proxy` understands several
shorthand forms.  The container's primary IP (from `net0`) is used
automatically when only a port is given:

| Notes value | Container IP | Resolved upstream |
|-------------|-------------|-------------------|
| `:8080` | `192.168.1.10` | `192.168.1.10:8080` |
| `8080` | `192.168.1.10` | `192.168.1.10:8080` |
| `192.168.1.10:8080` | _(any)_ | `192.168.1.10:8080` |
| `myapp:8080` | _(any)_ | `myapp:8080` |

---

## Installation

### With xcaddy (recommended)

```bash
xcaddy build --with github.com/YOUR_USERNAME/caddy-proxmox-provider
```

### With Docker

```dockerfile
FROM caddy:2.8-builder AS builder
RUN xcaddy build --with github.com/YOUR_USERNAME/caddy-proxmox-provider

FROM caddy:2.8-alpine
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
```

Or use the provided `Dockerfile` in this repo.

---

## Configuration

### Caddyfile (global options block)

```caddyfile
{
    config_loader proxmox {
        api_endpoint  https://pve.example.com:8006
        api_token_id  root@pam!caddy
        api_token     {env.PROXMOX_API_TOKEN}
        poll_interval 30s
        validate_ssl  true
        label_prefix  caddy.
    }
}
```

### JSON

```json
{
  "config_loader": {
    "module": "proxmox",
    "api_endpoint": "https://pve.example.com:8006",
    "api_token_id": "root@pam!caddy",
    "api_token": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "poll_interval": "30s",
    "validate_ssl": true,
    "label_prefix": "caddy."
  }
}
```

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `api_endpoint` | string | _(required)_ | Proxmox API base URL |
| `api_token_id` | string | _(required)_ | API token ID, e.g. `root@pam!caddy` |
| `api_token` | string | _(required)_ | API token secret |
| `poll_interval` | duration | `30s` | How often to poll Proxmox |
| `validate_ssl` | bool | `true` | Verify Proxmox TLS certificate |
| `label_prefix` | string | `caddy.` | Line prefix treated as a label |

---

## Proxmox permissions

Create a dedicated role and API token with the minimum required permissions:

```bash
# Create role
pveum role add caddy-provider -privs "VM.Audit,Sys.Audit,Datastore.Audit"

# Create API token
pveum user token add root@pam caddy

# Assign role to token
pveum acl modify / -token 'root@pam!caddy' -role caddy-provider
```

> **Save the token secret** when it is first displayed — it will not be shown again.

---

## Development

```bash
# Run tests
go test ./...

# Build a local Caddy binary
go build -o caddy ./cmd/caddy

# Run with example config
./caddy run --config Caddyfile.example --adapter caddyfile
```

---

## License

MIT

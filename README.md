# Traefik Proxmox Provider

A Traefik provider that automatically configures routing based on Proxmox VE virtual machines and containers.

> **Fork note:** This is a fork of [NX211/traefik-proxmox-provider](https://github.com/NX211/traefik-proxmox-provider) with the following changes:
> - **`traefik.enable` defaults to true** — every VM/container is proxied unless you add `traefik.enable=false`
> - **TLS on by default** — no need to add `tls=true` labels; opt out with `tls=false`
> - **`websecure` entrypoint by default** — no need to add `entrypoints=websecure`
> - **Port in rule label** — write `Host(\`app.example.com\`):8080` instead of a separate `loadbalancer.server.port` label

## Features

- Automatically discovers Proxmox VE virtual machines and containers
- Configures routing based on VM/container notes field
- **TLS and `websecure` entrypoint enabled by default**
- **Port can be specified inline in the router rule**
- Configurable polling interval
- SSL validation options
- Full support for Traefik's routing, middleware, and TLS options

## Installation

1. Add the plugin to your Traefik static configuration:

```yaml
experimental:
  plugins:
    traefik-proxmox-provider:
      moduleName: github.com/NX211/traefik-proxmox-provider
      version: v0.8.1
```

2. Configure the provider in your Traefik dynamic configuration:

```yaml
providers:
  plugin:
    traefik-proxmox-provider:
      pollInterval: "30s"
      apiEndpoint: "https://proxmox.example.com"
      apiTokenId: "root@pam!traefik_prod"
      apiToken: "your-api-token"
      apiLogging: "info"
      apiValidateSSL: "true"
```

## Proxmox API Token Setup

```bash
# For Proxmox VE 8.x or earlier:
pveum role add traefik-provider -privs "VM.Audit,VM.Monitor,Sys.Audit,Datastore.Audit"

# For Proxmox VE 9.0+:
pveum role add traefik-provider -privs "VM.Audit,VM.GuestAgent.Audit,Sys.Audit,Datastore.Audit"

pveum user token add root@pam traefik_prod
pveum acl modify / -token 'root@pam!traefik_prod' -role traefik-provider
```

## VM/Container Labeling

Add Traefik labels to a VM or container's **Notes** field in Proxmox (one label per line).

### Minimal setup (new — no labels needed except the rule)

```
traefik.enable=true
traefik.http.routers.myapp.rule=Host(`myapp.example.com`):8080
```

This is equivalent to the old verbose form:

```
traefik.enable=true
traefik.http.routers.myapp.rule=Host(`myapp.example.com`)
traefik.http.routers.myapp.entrypoints=websecure
traefik.http.routers.myapp.tls=true
traefik.http.services.appservice.loadbalancer.server.port=8080
```

### Disabling TLS (plain HTTP)

```
traefik.enable=true
traefik.http.routers.myapp.rule=Host(`myapp.example.com`):8080
traefik.http.routers.myapp.tls=false
traefik.http.routers.myapp.entrypoints=web
```

### Full example with all options

```
traefik.enable=true
traefik.http.routers.myapp.rule=Host(`myapp.example.com`):8080
traefik.http.routers.myapp.tls.certresolver=myresolver
traefik.http.routers.myapp.middlewares=auth@file,compression
traefik.http.services.myapp.loadbalancer.healthcheck.path=/health
```

## Configuration Options

| Option | Type | Default | Description |
|---|---|---|---|
| `pollInterval` | string | `"30s"` | How often to poll the Proxmox API |
| `apiEndpoint` | string | — | URL of your Proxmox VE API |
| `apiTokenId` | string | — | API token ID (e.g. `root@pam!traefik_prod`) |
| `apiToken` | string | — | API token secret |
| `apiLogging` | string | `"info"` | Log level (`debug` or `info`) |
| `apiValidateSSL` | string | `"true"` | Whether to validate SSL certificates |

## Label Behaviour Changes vs Upstream

| Behaviour | Upstream | This fork |
|---|---|---|
| TLS | Off unless `tls=true` | **On by default**; set `tls=false` to disable |
| Entrypoint | None (Traefik default) | **`websecure`** unless overridden |
| Port | Separate `loadbalancer.server.port` label | **Inline `:port` suffix on rule** OR separate label |

## License

Apache-2.0 — see [LICENSE](LICENSE)

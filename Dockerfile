# syntax=docker/dockerfile:1

# ── Stage 1: build custom Caddy with the Proxmox provider ─────────────────────
ARG CADDY_VERSION=2.8.4

FROM caddy:${CADDY_VERSION}-builder AS builder

# Replace YOUR_USERNAME with your GitHub username (or any module path)
# to publish this module.  For local use, xcaddy can reference a local path.
RUN xcaddy build \
    --with github.com/YOUR_USERNAME/caddy-proxmox-provider

# ── Stage 2: minimal runtime image ────────────────────────────────────────────
FROM caddy:${CADDY_VERSION}-alpine

COPY --from=builder /usr/bin/caddy /usr/bin/caddy

# Caddy data dir (certificates etc.)
VOLUME ["/data", "/config"]

EXPOSE 80 443

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]

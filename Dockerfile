# Stage 1: builder
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git ca-certificates

WORKDIR /src

ARG TRAEFIK_REPO=https://github.com/traefik/traefik.git
ARG TRAEFIK_REF=v3.7.1
ARG PLUGIN_REPO=https://github.com/ndowens/traefik-proxmox-provider.git
ARG PLUGIN_REF=master

# Clone Traefik
RUN git clone --depth 1 --branch ${TRAEFIK_REF} ${TRAEFIK_REPO} traefik

# Clone plugin into workspace
RUN git clone --depth 1 --branch ${PLUGIN_REF} ${PLUGIN_REPO} traefik-proxmox-provider

WORKDIR /src/traefik

RUN sed -i 's|1.25.0|1.25|' go.mod

RUN go mod tidy

# Add replace directive so Traefik's go.mod resolves the local plugin copy
RUN printf '\nreplace github.com/ndowens/traefik-proxmox-provider => ../traefik-proxmox-provider\n' >> go.mod

# Add side-effect import to ensure plugin is linked
RUN cat > internal/local_plugins_imports.go <<'EOF'
// Code generated to import local plugin for build
package main

import (
    _ "github.com/ndowens/traefik-proxmox-provider"
)
EOF

# Build Traefik
RUN CGO_ENABLED=0 go build -trimpath -o /traefik ./cmd/traefik

# Stage 2: runtime image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /traefik /usr/bin/traefik
EXPOSE 80 443 8080
ENTRYPOINT ["/usr/bin/traefik"]

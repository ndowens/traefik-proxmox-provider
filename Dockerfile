# Stage 1: builder                                                                                                                                             
FROM golang:1.20-alpine AS builder                                                                                                                             
RUN apk add --no-cache git ca-certificates                                                                                                                     
                                                                                                                                                               
WORKDIR /src

# Get Traefik source
ARG TRAEFIK_REPO=https://github.com/traefik/traefik.git
ARG TRAEFIK_REF=v3.7.1
RUN git clone --depth 1 --branch ${TRAEFIK_REF} ${TRAEFIK_REPO} traefik

# Copy local plugin if using local path (uncomment the following line when using local plugin)
COPY ../traefik-proxmox-provider /src/traefik-proxmox-provider

WORKDIR /src/traefik
RUN sed -i 's,1.25.0,1.25,' go.mod
# If using a local plugin copy above, add a replace directive so Traefik's go.mod can find it.
# This injects a replace into the module before building.
# (Only needed for local plugin; when using published module, skip this.)
ARG LOCAL_PLUGIN_PATH=
RUN if [ -n "$LOCAL_PLUGIN_PATH" ]; then \
      printf 'replace github.com/ndowens/traefik-proxmox-provider => %s\n' "$LOCAL_PLUGIN_PATH" >> go.mod; \
    fi

# Ensure plugin import is referenced so it gets linked into the binary.
# Create a small file that imports the plugin package for side-effect registration.
RUN cat > internal/local_plugins_imports.go <<'EOF'
// Code generated for adding local plugin imports
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

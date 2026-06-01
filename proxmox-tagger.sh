#!/usr/bin/env bash
# proxmox-tagger.sh - Helper script to set Traefik labels on Proxmox VMs/containers
# Usage: ./proxmox-tagger.sh <vmid> <key> <value>
#   e.g. ./proxmox-tagger.sh 100 traefik.enable true

set -euo pipefail

VMID="${1:-}"
KEY="${2:-}"
VALUE="${3:-}"

if [[ -z "$VMID" || -z "$KEY" || -z "$VALUE" ]]; then
  echo "Usage: $0 <vmid> <key> <value>"
  echo "  e.g. $0 100 traefik.enable true"
  exit 1
fi

# Detect whether it's a VM or container
if qm status "$VMID" &>/dev/null; then
  TYPE="qm"
elif pct status "$VMID" &>/dev/null; then
  TYPE="pct"
else
  echo "ERROR: VM/container $VMID not found"
  exit 1
fi

# Read existing description
if [[ "$TYPE" == "qm" ]]; then
  CURRENT=$(qm config "$VMID" | grep '^description:' | sed 's/^description: //' || true)
else
  CURRENT=$(pct config "$VMID" | grep '^description:' | sed 's/^description: //' || true)
fi

# Remove existing value for this key, then append the new one
UPDATED=$(echo "$CURRENT" | grep -v "^${KEY}=" || true)
UPDATED="${UPDATED}
${KEY}=${VALUE}"
UPDATED=$(echo "$UPDATED" | sed '/^$/d')

if [[ "$TYPE" == "qm" ]]; then
  qm set "$VMID" --description "$UPDATED"
else
  pct set "$VMID" --description "$UPDATED"
fi

echo "Set ${KEY}=${VALUE} on $TYPE $VMID"

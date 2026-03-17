#!/bin/bash
set -e

# Read input JSON from Terraform external data source (stdin)
INPUT=$(cat /dev/stdin)

# Parse JSON fields using grep/sed (no jq dependency)
SRC_DIR=$(echo "$INPUT"    | grep -o '"src_dir":"[^"]*"'    | cut -d'"' -f4)
BINARY_OUT=$(echo "$INPUT" | grep -o '"binary_out":"[^"]*"' | cut -d'"' -f4)
SRC_HASH=$(echo "$INPUT"   | grep -o '"src_hash":"[^"]*"'   | cut -d'"' -f4)

HASH_FILE="${BINARY_OUT}.srchash"
CURRENT_HASH=$(cat "$HASH_FILE" 2>/dev/null || echo "")

if [ "$CURRENT_HASH" != "$SRC_HASH" ] || [ ! -f "$BINARY_OUT" ]; then
    mkdir -p "$(dirname "$BINARY_OUT")" >&2
    cd "$SRC_DIR" >&2
    go mod tidy >&2
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
        -o "$BINARY_OUT" \
        -mod=readonly -trimpath \
        -ldflags="-s -w" \
        . >&2
    printf '%s' "$SRC_HASH" > "$HASH_FILE"
fi

printf '{"hash":"%s"}' "$SRC_HASH"

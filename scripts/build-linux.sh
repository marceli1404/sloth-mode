#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p linux/bin
(cd connector && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ../linux/bin/sloth-mode-connector-linux-amd64 .)
(cd connector && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ../linux/bin/sloth-mode-connector-linux-arm64 .)
echo "Built Linux amd64 and arm64 connectors. Run ./linux/install.sh to install for the current user."

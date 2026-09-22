#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

go build -o ../bin/alter-bridge ./cmd/alter-bridge
echo "built bin/alter-bridge"

#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COMPOSE_FILE="docker-compose.server.yml"

if ! command -v docker >/dev/null 2>&1; then
  echo "[ERROR] docker not found. Please install Docker first."
  exit 1
fi

if docker compose version >/dev/null 2>&1; then
  COMPOSE="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE="docker-compose"
else
  echo "[ERROR] docker compose not found. Please install Docker Compose first."
  exit 1
fi

mkdir -p nginx/ssl

if [[ ! -f nginx/ssl/cert.pem || ! -f nginx/ssl/key.pem ]]; then
  CN="$(hostname -f 2>/dev/null || hostname)"
  echo "[INFO] generating self-signed TLS cert for CN=$CN ..."
  openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout nginx/ssl/key.pem \
    -out nginx/ssl/cert.pem \
    -subj "/C=CN/O=Meeting System/CN=${CN}"

  # Some configs/tools expect a chain file; for self-signed cert we reuse cert.pem.
  cp -f nginx/ssl/cert.pem nginx/ssl/chain.pem
fi

echo "[INFO] starting services..."
$COMPOSE -f "$COMPOSE_FILE" up -d --build

echo "[INFO] service status:"
$COMPOSE -f "$COMPOSE_FILE" ps

echo
echo "[OK] Web UI:"
echo "  https://<server-ip>/"
echo "  http://<server-ip>:8800/ (no camera/mic on most browsers)"

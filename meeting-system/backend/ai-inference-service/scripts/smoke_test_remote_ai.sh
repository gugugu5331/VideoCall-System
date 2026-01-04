#!/usr/bin/env bash
#
# Quick smoke test against ai-inference-service gateway.
# Usage:
#   HOST=http://localhost:8085 FILE=/root/VideoCall-System/test.wav ./smoke_test_remote_ai.sh

set -euo pipefail

HOST="${HOST:-http://localhost:8085}"
FILE="${FILE:-/root/VideoCall-System/test.wav}"
FORMAT="${FORMAT:-wav}"
RATE="${RATE:-16000}"

if [[ ! -f "$FILE" ]]; then
  echo "Audio file not found: $FILE" >&2
  exit 1
fi

echo "Using host: $HOST"
echo "Audio file: $FILE"

BASE64_AUDIO="$(base64 "$FILE" | tr -d '\n')"

post() {
  local path="$1"
  local payload="$2"
  echo -e "\n=== POST $path ==="
  curl -s -X POST "$HOST$path" \
    -H "Content-Type: application/json" \
    -d "$payload" | jq .
}

post "/api/v1/ai/asr" "{\"audio_data\":\"$BASE64_AUDIO\",\"format\":\"$FORMAT\",\"sample_rate\":$RATE}"
post "/api/v1/ai/emotion" "{\"audio_data\":\"$BASE64_AUDIO\",\"format\":\"$FORMAT\",\"sample_rate\":$RATE}"
post "/api/v1/ai/synthesis" "{\"audio_data\":\"$BASE64_AUDIO\",\"format\":\"$FORMAT\",\"sample_rate\":$RATE}"

echo -e "\nDone."

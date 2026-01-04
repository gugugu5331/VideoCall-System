#!/usr/bin/env bash
set -euo pipefail

# End-to-end PCM streaming test for ai-inference-service gRPC.
# Usage:
#   ./e2e_stream_pcm.sh [host] [port] [pcm_path]
# Env:
#   SAMPLE_RATE=16000 CHANNELS=1 CHUNK_MS=200 TASKS="speech_recognition,emotion_detection,synthesis_detection"

HOST="${1:-localhost}"
PORT="${2:-9085}"
PCM_PATH="${3:-}"
FORMAT="${FORMAT:-pcm}"
SAMPLE_RATE="${SAMPLE_RATE:-16000}"
CHANNELS="${CHANNELS:-1}"
CHUNK_MS="${CHUNK_MS:-200}"
TASKS="${TASKS:-speech_recognition,emotion_detection,synthesis_detection}"
STREAM_ID="${STREAM_ID:-e2e-$(date +%s)}"

if ! command -v grpcurl >/dev/null 2>&1; then
  echo "grpcurl not found. Please install grpcurl and retry." >&2
  exit 1
fi

cleanup_pcm=""
if [[ -z "${PCM_PATH}" ]]; then
  if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 not found. Provide a PCM16 file path as the third argument." >&2
    exit 1
  fi
  PCM_PATH="$(mktemp)"
  cleanup_pcm="${PCM_PATH}"
  SAMPLE_RATE="${SAMPLE_RATE}" CHANNELS="${CHANNELS}" python3 - <<'PY' > "${PCM_PATH}"
import math
import os
import struct
import sys

sample_rate = int(os.environ.get("SAMPLE_RATE", "16000"))
channels = int(os.environ.get("CHANNELS", "1"))
duration_sec = 1.0
freq = 440.0
amp = 0.2

total = int(sample_rate * duration_sec)
for i in range(total):
    value = int(amp * 32767 * math.sin(2 * math.pi * freq * i / sample_rate))
    frame = struct.pack("<h", value)
    sys.stdout.buffer.write(frame * channels)
PY
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 not found. Install python3 to generate streaming JSON." >&2
  exit 1
fi

PCM_PATH="${PCM_PATH}" FORMAT="${FORMAT}" SAMPLE_RATE="${SAMPLE_RATE}" CHANNELS="${CHANNELS}" \
CHUNK_MS="${CHUNK_MS}" TASKS="${TASKS}" STREAM_ID="${STREAM_ID}" \
python3 - <<'PY' | grpcurl -plaintext -d @ "${HOST}:${PORT}" grpc.AIService/StreamAudioProcessing
import base64
import json
import os
import sys

pcm_path = os.environ["PCM_PATH"]
fmt = os.environ.get("FORMAT", "pcm")
sample_rate = int(os.environ.get("SAMPLE_RATE", "16000"))
channels = int(os.environ.get("CHANNELS", "1"))
chunk_ms = int(os.environ.get("CHUNK_MS", "200"))
tasks = [t.strip() for t in os.environ.get("TASKS", "").split(",") if t.strip()]
stream_id = os.environ.get("STREAM_ID", "e2e-stream")

with open(pcm_path, "rb") as f:
    data = f.read()

bytes_per_ms = int(sample_rate * channels * 2 / 1000)
chunk_size = max(bytes_per_ms * chunk_ms, 1) if bytes_per_ms > 0 else len(data)

if not data:
    sys.exit("PCM file is empty")

seq = 0
for offset in range(0, len(data), chunk_size):
    chunk = data[offset:offset + chunk_size]
    msg = {
        "data": base64.b64encode(chunk).decode("ascii"),
        "sequence": seq,
        "stream_id": stream_id,
        "format": fmt,
        "sample_rate": sample_rate,
        "channels": channels,
        "tasks": tasks,
        "is_final": offset + chunk_size >= len(data),
    }
    print(json.dumps(msg))
    seq += 1
PY

if [[ -n "${cleanup_pcm}" ]]; then
  rm -f "${cleanup_pcm}"
fi

#!/usr/bin/env bash
set -euo pipefail

# One-key deploy for GPU AI nodes (SSH key auth required).
#
# Usage:
#   MODEL_DIR=/models AI_HTTP_PORT=8800 UNIT_MANAGER_PORT=8801 ./deploy_gpu_ai_nodes.sh
#
# Optional:
#   REMOTE_DIR=/root/VideoCall-System SSH_USER=root ./deploy_gpu_ai_nodes.sh
#   GPU_AI_NODES="host1:port host2:port" ./deploy_gpu_ai_nodes.sh
#   GPU_AI_NODES_FILE=./nodes.txt ./deploy_gpu_ai_nodes.sh
#
# Note:
# - This script intentionally does NOT support password-based SSH to avoid leaking secrets.

SSH_USER="${SSH_USER:-root}"
REMOTE_DIR="${REMOTE_DIR:-/root/VideoCall-System}"
MODEL_DIR="${MODEL_DIR:-/models}"
AI_HTTP_PORT="${AI_HTTP_PORT:-8800}"
UNIT_MANAGER_PORT="${UNIT_MANAGER_PORT:-8801}"

GPU_AI_NODES="${GPU_AI_NODES:-}"
GPU_AI_NODES_FILE="${GPU_AI_NODES_FILE:-}"

NODES=()
if [[ -n "${GPU_AI_NODES_FILE}" ]]; then
  if [[ ! -f "${GPU_AI_NODES_FILE}" ]]; then
    echo "GPU_AI_NODES_FILE not found: ${GPU_AI_NODES_FILE}" >&2
    exit 1
  fi
  while IFS= read -r line; do
    line="${line%%#*}"
    line="$(echo "${line}" | xargs)"
    [[ -z "${line}" ]] && continue
    NODES+=("${line}")
  done < "${GPU_AI_NODES_FILE}"
elif [[ -n "${GPU_AI_NODES}" ]]; then
  read -r -a NODES <<< "${GPU_AI_NODES}"
else
  echo "Please set GPU_AI_NODES (\"host1:port host2:port\") or GPU_AI_NODES_FILE" >&2
  exit 1
fi

SSH_BASE_OPTS=(
  -o BatchMode=yes
  -o StrictHostKeyChecking=no
  -o UserKnownHostsFile=/dev/null
  -o ConnectTimeout=10
)

for node in "${NODES[@]}"; do
  host="${node%:*}"
  port="${node##*:}"
  echo "==> Deploying to ${host}:${port}"

  ssh "${SSH_BASE_OPTS[@]}" -p "${port}" "${SSH_USER}@${host}" \
    "cd '${REMOTE_DIR}/meeting-system' && \
     MODEL_DIR='${MODEL_DIR}' AI_HTTP_PORT='${AI_HTTP_PORT}' UNIT_MANAGER_PORT='${UNIT_MANAGER_PORT}' \
     docker compose -f deployment/gpu-ai/docker-compose.gpu-ai.yml up -d --build"
done

echo "Done."

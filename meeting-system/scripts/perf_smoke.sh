#!/usr/bin/env bash
# 集成网关、信令、业务压测的巡检脚本，可用于 CI 或定期 cron 任务

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESULTS_ROOT="${RESULTS_DIR:-"$ROOT_DIR/perf-results"}"
STAMP="$(date +%Y%m%d-%H%M%S)"
RESULTS_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RESULTS_DIR"

log() {
    echo "[$(date +'%F %T')] $*"
}

require_cmd() {
    for cmd in "$@"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            echo "missing command: $cmd" >&2
            exit 1
        fi
    done
}

require_cmd go docker awk

if [[ -f "$ROOT_DIR/.env" ]]; then
    # shellcheck disable=SC1090
    source "$ROOT_DIR/.env"
fi

SIGNALING_URL="${SIGNALING_URL:-ws://localhost:8081/ws/signaling}"
SIGNALING_SECRET="${SIGNALING_SECRET:-${JWT_SECRET:-test-secret}}"
SIGNALING_MEETING_ID="${SIGNALING_MEETING_ID:-1}"

log "results will be stored under $RESULTS_DIR"

log "1) gateway perf smoke"
(
    cd "$ROOT_DIR/nginx/scripts"
    ./test-gateway.sh --performance
) | tee "$RESULTS_DIR/gateway-perf.log"

log "2) signaling quick stress"
(
    cd "$ROOT_DIR/backend/signaling-service"
    ./run_stress_test.sh --url "$SIGNALING_URL" --secret "$SIGNALING_SECRET" --meeting "$SIGNALING_MEETING_ID" --quick
) | tee "$RESULTS_DIR/signaling-quick.log"

log "3) business HTTP stress suite"
(
    cd "$ROOT_DIR/backend/stress-test"
    STRESS_CONCURRENT_USERS="${STRESS_CONCURRENT_USERS:-10,50}" \
    PEAK_USERS="${PEAK_USERS:-100}" \
    STABILITY_USERS="${STABILITY_USERS:-50}" \
    STABILITY_DURATION="${STABILITY_DURATION:-30s}" \
    TEST_DURATION="${TEST_DURATION:-10s}" \
    REQUEST_TIMEOUT="${REQUEST_TIMEOUT:-5s}" \
    go run . | tee "$RESULTS_DIR/business-stress.log"
)

log "done. collected logs:"
ls -1 "$RESULTS_DIR"

#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
WORKERS="${WORKERS:-12}"
PIDS=()

if [[ ! "$WORKERS" =~ ^[1-9][0-9]*$ ]] || (( WORKERS > 64 )); then
  printf 'WORKERS must be between 1 and 64\n' >&2
  exit 1
fi

set_mode() {
  curl --silent --show-error --fail --connect-timeout 2 --max-time 5 \
    --request POST "$BASE_URL/internal/sre-demo/$1" --output /dev/null
}

cleanup() {
  local pid
  for pid in "${PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  for pid in "${PIDS[@]}"; do
    wait "$pid" 2>/dev/null || true
  done
  set_mode normal || true
}

traffic() {
  local deadline="$1"
  while (( SECONDS < deadline )); do
    if ! curl --silent --show-error --connect-timeout 2 --max-time 3 \
      --output /dev/null "$BASE_URL/api/articles"; then
      printf 'Request failed at transport level\n' >&2
    fi
    sleep 0.2
  done
}

phase() {
  local mode="$1" duration="$2" worker deadline
  set_mode "$mode"
  printf '%(%H:%M:%S)T mode=%s duration=%ss workers=%s\n' -1 "$mode" "$duration" "$WORKERS"
  deadline=$((SECONDS + duration))
  PIDS=()
  for ((worker = 0; worker < WORKERS; worker++)); do
    traffic "$deadline" &
    PIDS+=("$!")
  done
  for worker in "${PIDS[@]}"; do
    wait "$worker"
  done
  PIDS=()
}

set_mode normal
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

case "${1:-demo}" in
  demo)
    phase normal 45
    phase error 30
    phase slow 30
    phase normal 60
    ;;
  normal|error|slow)
    DURATION="${2:-60}"
    if [[ ! "$DURATION" =~ ^[1-9][0-9]*$ ]]; then
      printf 'Duration must be a positive integer in seconds\n' >&2
      exit 1
    fi
    phase "$1" "$DURATION"
    ;;
  *)
    printf 'Usage: bash scripts/sre-demo.sh [demo|normal|error|slow] [seconds]\n' >&2
    exit 1
    ;;
esac

#!/usr/bin/env bash
set -e

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

BACKEND_PID=""
cleanup() {
  if [[ -n "$BACKEND_PID" ]] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    echo "A encerrar o backend (PID $BACKEND_PID)..."
    kill "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
  fi
  exit 0
}
trap cleanup SIGINT SIGTERM EXIT

echo "A iniciar o backend (API :8080)..."
if [[ -x "./dockscope" ]]; then
  ./dockscope &
else
  go run ./cmd/dockscope &
fi
BACKEND_PID=$!

echo "A aguardar o backend..."
if command -v curl &>/dev/null; then
  for i in {1..15}; do
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health 2>/dev/null | grep -q 200; then
      echo "Backend pronto."
      break
    fi
    [[ $i -eq 15 ]] && echo "Backend não respondeu a tempo." && exit 1
    sleep 1
  done
else
  sleep 2
fi

echo "A iniciar o frontend (Vite)..."
cd web && npm run dev

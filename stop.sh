#!/bin/bash
# Stop all dev services on ports 8080 and 5173
# Usage: bash stop.sh

set -o pipefail
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

PASS=0
FAIL=0

kill_port() {
  local port=$1
  local pids
  local killed=0
  local skipped=0

  echo -n "  Port $port ... "

  # Find PIDs listening on this port
  pids=$(netstat -ano 2>/dev/null | grep ":$port " | grep LISTENING | awk '{print $NF}' | sort -u)

  if [ -z "$pids" ]; then
    echo -e "${YELLOW}CLEAN (no process)${NC}"
    return 0
  fi

  for pid in $pids; do
    # Check if PID is valid
    if ! kill -0 "$pid" 2>/dev/null && ! tasklist //FI "PID eq $pid" 2>/dev/null | grep -q "$pid"; then
      echo -e "${YELLOW}SKIP (PID $pid already dead)${NC}"
      ((skipped++))
      continue
    fi

    # Try taskkill first (Windows), fallback to kill (Unix)
    if taskkill //F //PID "$pid" 2>/dev/null; then
      ((killed++))
      ((PASS++))
    elif kill -9 "$pid" 2>/dev/null; then
      ((killed++))
      ((PASS++))
    else
      echo -e "${RED}FAIL (PID $pid: permission denied or not found)${NC}"
      ((FAIL++))
      return 1
    fi
  done

  if [ $killed -gt 0 ]; then
    echo -e "${GREEN}OK (killed $killed PID(s))${NC}"
  elif [ $skipped -gt 0 ]; then
    echo -e "${YELLOW}CLEAN ($skipped already dead)${NC}"
  fi
}

echo ""
echo "=== Stopping Lab Management Services ==="

# Kill by port
kill_port 8080
PORT8080_RC=$?

kill_port 5173
PORT5173_RC=$?

# Fallback: if port-based kill failed, try by image name
if [ $PORT8080_RC -ne 0 ] || [ $PORT5173_RC -ne 0 ]; then
  echo ""
  echo "=== Fallback: trying image name ==="
  for img in "server.exe" "main.exe" "go.exe"; do
    echo -n "  $img ... "
    output=$(taskkill //F //IM "$img" 2>&1)
    if [ $? -eq 0 ]; then
      echo -e "${GREEN}OK${NC}"
      ((PASS++))
    elif echo "$output" | grep -qi "not found\|no such\|û���ҵ�"; then
      echo -e "${YELLOW}skipped (not running)${NC}"
    else
      echo -e "${RED}FAIL (reason: $output)${NC}"
      ((FAIL++))
    fi
  done
fi

echo ""
echo "=== Summary ==="
echo -e "  ${GREEN}Passed:  $PASS${NC}"
if [ $FAIL -gt 0 ]; then
  echo -e "  ${RED}Failed:  $FAIL${NC}"
fi

# Final verification
echo ""
echo -n "Verification: "
REMAINING=$(netstat -ano 2>/dev/null | grep -E ":(8080|5173) " | grep LISTENING)
if [ -z "$REMAINING" ]; then
  echo -e "${GREEN}Ports 8080 and 5173 are clean${NC}"
  exit 0
else
  echo -e "${RED}Processes still running!${NC}"
  echo "$REMAINING" | while read line; do
    echo -e "  ${RED}$line${NC}"
  done
  exit 1
fi

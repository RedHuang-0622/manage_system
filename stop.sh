#!/bin/bash
# Stop all dev services on ports 8080 and 5173
for port in 8080 5173; do
  pids=$(netstat -ano 2>/dev/null | grep ":$port " | grep LISTENING | awk '{print $NF}' | sort -u)
  for pid in $pids; do
    taskkill //F //PID $pid 2>/dev/null && echo "  Killed PID $pid (port $port)"
  done
done
echo "[stop] All done."

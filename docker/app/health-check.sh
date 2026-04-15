#!/bin/bash

HEALTH_ENDPOINT="http://localhost:8080/api/v1/health"
CHECK_INTERVAL=5  # seconds

while true; do
  # Check the health endpoint
  if ! curl -s --fail $HEALTH_ENDPOINT > /dev/null; then
    echo "Health check failed. Restarting the server process..."

    pkill server_process || true
    /usr/bin/app/server_process &
  fi
  sleep $CHECK_INTERVAL
done

#!/bin/bash

# Set number of parallel requests
NUM_REQUESTS=20

# Forwarder URL
URL=http://localhost:9443

# Array to store errors
ERRORS=()
j=$1
# Launch requests
for i in $(seq 1 $NUM_REQUESTS); do
  response=$(curl -s -w "%{http_code}" $URL -o /dev/null)
  sleep 0.5
  http_code=$(tail -n1 <<< "$response")
  if [ $http_code -ne 200 ]; then
    ERRORS+=("Request $j--$i failed with HTTP code $http_code")
  fi
done

# Print errors
if [ ${#ERRORS[@]} -gt 0 ]; then
  for err in "${ERRORS[@]}"; do
    echo $err
  done
fi
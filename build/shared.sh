#!/bin/sh

set -o errexit
set -o nounset
if set -o | grep -q "pipefail"; then
  set -o pipefail
fi

RETRIES=10
DBNAME=$(basename "$IOTSTORE_DATABASE_URL" | awk -F'[?]' '{print $1}')

until psql -c '\q' "$DBNAME" || [ "$RETRIES" -eq 0 ]; do
  echo "Waiting for postgres server, $((RETRIES--)) remaining attempts"
  sleep 1
done
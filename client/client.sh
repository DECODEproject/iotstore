#!/usr/bin/env bash

set -euo pipefail

# This script attempts to invoke the three RPC methods exposed by the datastore
# via Curl just for quick local sanity checking. Requires base64, curl and jq
# tools installed and available on the current $PATH to run.

datastore_base="http://localhost:8080/twirp/decode.iot.datastore.Datastore/"

function write_data {
  data=$(echo "$2" | base64)

  echo "--> write some data for alice for public key $1"
  curl --request "POST" \
       --location "${datastore_base}WriteData" \
       --header "Content-Type: application/json" \
       --silent \
       --data "{\"public_key\":\"${1}\",\"data\":\"${data}\"}" \
       | jq "."
}

function read_data {
  start_time=$(date -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)
  echo "$start_time"

  echo "--> read data for public_key $1"
  curl --request "POST" \
       --location "${datastore_base}ReadData" \
       --header "Content-Type: application/json" \
       --silent \
       --data "{\"public_key\":\"${1}\", \"page_size\":3, \"start_time\": \"${start_time}\"}" \
       | jq "."
}

for i in {1..10}; do
  write_data abc123 "hello${i}"
done

read_data abc123

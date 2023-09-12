#!/usr/bin/env bash

echo "* Packing payload..."

# Get directory relative to this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

tar -cvpf ${DIR}/payload.tar -C ${DIR}/payload .
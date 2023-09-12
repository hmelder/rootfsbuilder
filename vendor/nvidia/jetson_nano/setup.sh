#!/usr/bin/env sh

echo "* Packing payload..."

# Get directory relative to this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

tar -cpf ${DIR}/payload.tar -C ${DIR}/payload .
#!/bin/sh

cd /

export SERVER_HOST="${SERVER_HOST:-127.0.0.1}"
export SERVER_PORT=${SERVER_PORT:-10110}

export SCHEDULER_INTERVAL=${SCHEDULER_INTERVAL:-5}
export SCHEDULER_AGENT_CONTACT_DELAY=${SCHEDULER_AGENT_CONTACT_DELAY:-30}

export AUTH_TYPE="${AUTH_TYPE:-jwt}"

export ETCD_ADDRESS="${ETCD_ADDRESS:-http://127.0.0.1:2379}"

IP_ADDRESSES_FILE="ip_addresses"
if [ -f "${IP_ADDRESSES_FILE}" ]; then
    FILE_IP_ADDRESSES=$(cat ${IP_ADDRESSES_FILE})
fi

export IP_ADDRESSES="${IP_ADDRESSES:-$FILE_IP_ADDRESSES}"

envsubst < chinchilla.toml.tmpl > chinchilla.toml

exec $@

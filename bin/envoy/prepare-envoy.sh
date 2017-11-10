#!/usr/bin/env bash

set -o errexit
set -o pipefail

if [[ -z "${ENVOY_EGRESS_PORT-}" ]] || [[ -z "${REDIRECT_EGRESS_CIDR_BLOCK-}" ]] || [[ -z "${ENVOY_CONFIG_DIR-}" ]] || [[ -z "${ENVOY_ADMIN_PORT-}" ]] || [[ -z "${ENVOY_XDS_API_HOST-}" ]] || [[ -z "${ENVOY_XDS_API_PORT-}" ]]; then
  echo "ENVOY_EGRESS_PORT, REDIRECT_EGRESS_CIDR_BLOCK, ENVOY_CONFIG_DIR, ENVOY_ADMIN_PORT, ENVOY_XDS_API_HOST, ENVOY_XDS_API_PORT are required"
  exit 1
fi

# Redirect outgoing traffic going to an IP address in REDIRECT_EGRESS_CIDR_BLOCK to ENVOY_EGRESS_PORT
iptables -t nat -A OUTPUT -p tcp -d ${REDIRECT_EGRESS_CIDR_BLOCK} -j REDIRECT --to-port ${ENVOY_EGRESS_PORT} -m comment --comment "lattice redirect to envoy"

mkdir -p "${ENVOY_CONFIG_DIR}"
ENVOY_XDS_API=${ENVOY_XDS_API_HOST}:${ENVOY_XDS_API_PORT}
cat <<EOF >> ${ENVOY_CONFIG_DIR}/config.json
{
  "listeners": [],
  "lds": {
    "cluster": "xds-api",
    "refresh_delay_ms": 10000
  },
  "admin": {
    "access_log_path": "/dev/null",
    "address": "tcp://0.0.0.0:${ENVOY_ADMIN_PORT}"
  },
  "cluster_manager": {
    "clusters": [
      {
        "name": "xds-api",
        "connect_timeout_ms": 250,
        "type": "static",
        "lb_type": "round_robin",
        "hosts": [
          {
            "url": "tcp://${ENVOY_XDS_API}"
          }
        ]
      }
    ],
    "cds": {
      "cluster": {
        "name": "xds-api-cds",
        "connect_timeout_ms": 250,
        "type": "static",
        "lb_type": "round_robin",
        "hosts": [
          {
            "url": "tcp://${ENVOY_XDS_API}"
          }
        ]
      },
      "refresh_delay_ms": 10000
    },
    "sds": {
      "cluster": {
        "name": "xds-api-sds",
        "connect_timeout_ms": 250,
        "type": "static",
        "lb_type": "round_robin",
        "hosts": [
          {
            "url": "tcp://${ENVOY_XDS_API}"
          }
        ]
      },
      "refresh_delay_ms": 10000
    }
  }
}
EOF

exit 0

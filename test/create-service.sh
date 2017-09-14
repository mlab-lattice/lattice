#!/usr/bin/env bash

cat <<EOF | kubectl create -f -
apiVersion: lattice.mlab.com/v1
kind: Service
metadata:
  name: ${1}
  namespace: default
spec:
  path: /a/b/c,
  buildName: ${2}
  definition:
    resources:
      min_instances: 1
    components:
    - name: http
      ports:
      - name: http
        port: 9999
        protocol: http
      exec:
        command:
        - node
        - lib/PrivateHelloService.js
        - -p
        - "9999"
      health_check:
        http:
          path: /status
          port: http
EOF

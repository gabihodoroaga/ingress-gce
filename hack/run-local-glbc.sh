#!/bin/bash

# Run glbc. First run `run-local.sh` to set things up.
#
# Files touched: /tmp/kubectl-proxy.log /tmp/glbc.log

# Take in consideration if the valriable is already defined
#GOOGLE_APPLICATION_CREDENTIALS="${HOME}/.config/gcloud/application_default_credentials.json"

if [ ! -r ${GOOGLE_APPLICATION_CREDENTIALS} ]; then
    echo "You must login your application default credentials"
    echo "$ gcloud auth application-default login"
    exit 1
fi

GCECONF=${GCECONF:-${PWD}/../tmp/gce.conf}
GLBC=${GLBC:-./glbc}
PORT=${PORT:-7127}
V=${V:-4}

echo "GCECONF=${GCECONF} GLBC=${GLBC} PORT=${PORT} V=${V}"

if [ ! -x "${GLBC}" ]; then
    echo "ERROR: No ${GLBC} executable found" >&2
    exit 1
fi

echo "$(date) start" >> ${PWD}/../tmp/kubectl-proxy.log
kubectl proxy --port="${PORT}" \
    >> ${PWD}/../tmp/kubectl-proxy.log &

PROXY_PID=$!
cleanup() {
    echo "Killing proxy (pid=${PROXY_PID})"
    kill ${PROXY_PID}
}
trap cleanup EXIT

kubectl apply -f docs/deploy/resources/default-http-backend.yaml

sleep 2 # Wait for proxy to start up
${GLBC} \
    --v=${V} \
    --apiserver-host=http://localhost:${PORT} \
    --running-in-cluster=false \
    --logtostderr \
    --config-file-path=${GCECONF} \
    --enable-backendconfig-healthcheck \
    2>&1 | tee -a ${PWD}/../tmp/glbc.log

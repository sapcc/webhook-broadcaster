#!/bin/sh

/webhook-broadcaster -concourse-url "${CONCOURSE_URL}" \
    -auth-user "${CONCOURSE_USER}" \
    -auth-password "${CONCOURSE_PASSWORD}" \
    -refresh-interval "${CONCOURSE_REFRESH_INTERVAL:-5m}" \
    -listen-addr "${CONCOURSE_LISTEN_ADDRESS:-:8080}" \
    -webhook-concurrency "${CONCOURSE_WEBHOOK_CONCURRENCY:-20}"

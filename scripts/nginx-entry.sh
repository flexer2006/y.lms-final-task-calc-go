#!/bin/sh
# shellcheck disable=SC2016
envsubst '${SSL_CERT_PATH} ${SSL_KEY_PATH}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

exec "$@"
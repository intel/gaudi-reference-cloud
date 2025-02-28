FROM nginx:1.26.1-alpine-slim

ENV PROXY_CACHE_VALID_OK "1h"
ENV PROXY_CACHE_MAX_SIZE "10g"
ENV PROXY_CACHE_INACTIVE "60m"
ENV PROXY_CACHE_VALID_OK "1h"
ENV PROXY_CACHE_VALID_NOTFOUND "1m"
ENV PROXY_CACHE_VALID_FORBIDDEN "30s"
ENV CORS_ENABLED 0

ENV DIRECTORY_LISTING_PATH_PREFIX=""
ENV STRIP_LEADING_DIRECTORY_PATH=""
ENV PREFIX_LEADING_DIRECTORY_PATH=""

COPY common/etc /etc
COPY common/docker-entrypoint.sh /docker-entrypoint.sh
COPY common/docker-entrypoint.d /docker-entrypoint.d/
COPY oss/etc /etc

RUN set -eux && \
    apk add nginx bash && \
    apk update && apk upgrade --available && \
    apk add --no-cache curl libedit && \
    printf "%s%s%s%s\n" \
    "@nginx " \
    "http://nginx.org/packages/mainline/alpine/v" \
    `egrep -o '^[0-9]+\.[0-9]+' /etc/alpine-release` \
    "/main" | tee -a /etc/apk/repositories && \
    curl -o /tmp/nginx_signing.rsa.pub https://nginx.org/keys/nginx_signing.rsa.pub && \
    mv /tmp/nginx_signing.rsa.pub /etc/apk/keys/ && \
    apk add nginx-module-xslt@nginx nginx-module-njs@nginx && \
    mkdir -p /var/cache/nginx/s3_proxy && \
    chown nginx:nginx /var/cache/nginx/s3_proxy && \
    chmod -R -v +x /docker-entrypoint.sh /docker-entrypoint.d/*.sh

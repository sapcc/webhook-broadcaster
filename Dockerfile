FROM alpine

RUN apk add --no-cache curl
ADD bin/linux/webhook-broadcaster /usr/local/bin/

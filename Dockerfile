FROM alpine

CMD apk add --no-cache curl
ADD bin/linux/webhook-ingester /usr/local/bin/

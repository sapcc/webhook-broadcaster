FROM alpine
LABEL source_repository="https://github.com/sapcc/webhook-broadcaster"

RUN apk add --no-cache curl
ADD bin/linux/webhook-broadcaster /usr/local/bin/
ADD start.sh /

ENTRYPOINT [ "/start.sh" ]


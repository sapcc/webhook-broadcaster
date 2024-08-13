FROM golang:1.23 AS builder

WORKDIR /go/src/github.com/sapcc/webhook-broadcaster
ADD go.mod go.sum ./
RUN go mod download
ADD cache/main.go .
RUN CGO_ENABLED=0 go build -v -o /dev/null
ADD . .
RUN go test -v .
RUN CGO_ENABLED=0 go build -v -o /webhook-broadcaster

FROM alpine:latest
LABEL maintainer="jan.knipper@sap.com"
LABEL source_repository="https://github.com/sapcc/webhook-broadcaster"

RUN apk -U upgrade && apk add --no-cache ca-certificates

COPY --from=builder /webhook-broadcaster /webhook-broadcaster
ADD start.sh /start.sh

ENTRYPOINT ["/start.sh"]

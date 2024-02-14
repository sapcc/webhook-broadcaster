FROM golang:1.22 AS builder

WORKDIR /go/src/github.com/sapcc/webhook-broadcaster
ADD go.mod go.sum ./
RUN go mod download
ADD cache/main.go .
RUN CGO_ENABLED=0 go build -v -o /dev/null
ADD . .
RUN go test -v .
RUN CGO_ENABLED=0 go build -v -o /webhook-broadcaster

RUN apt update -qqq && \
    apt install -yqqq ca-certificates && \
    update-ca-certificates

FROM gcr.io/distroless/static-debian12
LABEL maintainer="jan.knipper@sap.com"
LABEL source_repository="https://github.com/sapcc/webhook-broadcaster"

COPY --from=builder /webhook-broadcaster /webhook-broadcaster
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
ADD start.sh /start.sh

ENTRYPOINT ["/start.sh"]

FROM eu.gcr.io/optiopay/go-tester:1.12.1 as builder

WORKDIR /go/src/github.com/optiopay/webhook-broadcaster

RUN go get -u github.com/golang/dep/cmd/dep

COPY . .

RUN dep ensure \
    && go test -v \
    && CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .

FROM eu.gcr.io/optiopay/alpine:3.9

RUN apk add --no-cache curl
COPY --from=builder /go/src/github.com/optiopay/webhook-broadcaster/webhook-broadcaster /usr/local/bin/webhook-broadcaster
ADD start.sh /

RUN chmod +x /usr/local/bin/webhook-broadcaster

ENTRYPOINT [ "/start.sh" ]

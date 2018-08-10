PKG:=github.com/sapcc/webhook-ingester
IMAGE:=sapcc/webhook-ingester
VERSION:=0.1
build:
	go build -o bin/webhook-ingester $(PKG)

docker:
	go test -v
	GOOS=linux go build -o bin/linux/webhook-ingester $(PKG)
	docker build -t $(IMAGE):$(VERSION) .

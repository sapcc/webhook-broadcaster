PKG:=github.com/sapcc/webhook-broadcaster
IMAGE:=sapcc/concourse-webhook-broadcaster
VERSION:=0.1
build:
	go build -o bin/webhook-broadcaster $(PKG)

docker:
	go test -v
	GOOS=linux go build -o bin/linux/webhook-broadcaster $(PKG)
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)

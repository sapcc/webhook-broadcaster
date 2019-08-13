.PHONY: build docker push clean docker-clean

PKG:=github.com/sapcc/webhook-broadcaster
IMAGE:=sapcc/concourse-webhook-broadcaster
VERSION:=0.6.3

build:
	go build -v -o bin/webhook-broadcaster $(PKG)

docker:
	go test -v
	GOOS=linux CGO_ENABLED=0 go build -o bin/linux/webhook-broadcaster $(PKG)
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)

clean:
	rm -rf bin

docker-clean:
	docker rmi $(IMAGE):$(VERSION)

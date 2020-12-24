IMAGE:=sapcc/concourse-webhook-broadcaster
VERSION:=0.6.2
build:
	go build -v -o bin/webhook-broadcaster .

docker:
	go test -v
	GOOS=linux CGO_ENABLED=0 go build -o bin/linux/webhook-broadcaster $(PKG)
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)

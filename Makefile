TARGET = webhook-broadcaster
GOTARGET = github.com/sapcc/$(TARGET)
REGISTRY ?= keppel.eu-de-1.cloud.sap/ccloud/concourse-webhook-broadcaster
VERSION ?= 0.10.0
IMAGE = $(REGISTRY)/$(BIN)
DOCKER ?= docker

all: container

test:
	go test .

container:
	$(DOCKER) build --network=host -t $(REGISTRY):latest -t $(REGISTRY):$(VERSION) .

push:
	$(DOCKER) push $(REGISTRY):latest
	$(DOCKER) push $(REGISTRY):$(VERSION)

.PHONY: all test container push

clean:
	rm -f $(TARGET)
	$(DOCKER) rmi $(REGISTRY):latest
	$(DOCKER) rmi $(REGISTRY):$(VERSION)

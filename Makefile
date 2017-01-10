.PHONY: build image clean push install

IMAGE := csphere/qingcloud-docker-network
BUILDER_IMAGE := $(IMAGE):build
TARGET_IMAGE := $(IMAGE):$(shell cat VERSION)
BIN_NAME := qingcloud-docker-network
GIT_COMMIT := $(shell git rev-parse --short HEAD)

default: build

image: Dockerfile.dist build
	docker build -t $(TARGET_IMAGE) -t $(IMAGE) --force-rm -f $< .

build: Dockerfile.build
	[ -d bin ] || mkdir bin
	docker build -t $(BUILDER_IMAGE) --build-arg GIT_COMMIT=$(GIT_COMMIT) --force-rm -f $< .
	docker run --rm -v $(shell pwd)/bin:/data $(BUILDER_IMAGE) cp /bin/$(BIN_NAME) /data

clean:
	rm bin/$(BIN_NAME)

push:
	docker push $(TARGET_IMAGE)
	docker push $(IMAGE):latest


install: build
	cp bin/$(BIN_NAME) /bin/
	cp contrib/init/systemd/qingcloud-docker-network.service /etc/systemd/system/
	cp contrib/etc/qingcloud-docker-network.conf /etc/
	systemctl daemon-reload
	systemctl enable $(BIN_NAME).service

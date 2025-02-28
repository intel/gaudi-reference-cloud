# Makefile

SERVICE_IMAGE_NAME = local-ovs

build:
	docker build -t $(SERVICE_IMAGE_NAME) .

clean:
	docker rmi -f $(SERVICE_IMAGE_NAME) || true

.PHONY: build clean





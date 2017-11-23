IMAGE=docker.jw4.us/import

ifeq ($(REVISION),)
	DIRTY=$(shell git diff-index --quiet HEAD || echo "-dirty")
	REVISION=$(shell git rev-parse --short HEAD)$(DIRTY)
endif

all: image

clean:
	-rm ./import
	go clean .

image:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags="-s -w" -o import .
	docker build -t $(IMAGE):latest -t $(IMAGE):$(REVISION) .

push: clean image
	docker push $(IMAGE):$(REVISION)
	docker push $(IMAGE):latest


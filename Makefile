export PATH := $(PATH):`go env GOPATH`/bin
export GO111MODULE=on
LDFLAGS := -s -w
VERSION := $(shell git tag --sort=creatordate | grep -E '[0-9]' | tail -1)
DOCKER_BASE := apiiro/public-images/network-broker
AGENT_TAG := $(shell echo ${VERSION} | grep -o '^v[0-9.]*')

all: env fmt build

build: frps frpc

env:
	@go version

# compile assets into binary file
file:
	rm -rf ./assets/frps/static/*
	rm -rf ./assets/frpc/static/*
	cp -rf ./web/frps/dist/* ./assets/frps/static
	cp -rf ./web/frpc/dist/* ./assets/frpc/static

fmt:
	go fmt ./...

fmt-more:
	gofumpt -l -w .

gci:
	gci write -s standard -s default -s "prefix(github.com/fatedier/frp/)" ./

vet:
	go vet ./...

frps:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags frps -o bin/frps ./cmd/frps

frpc:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags frpc -o bin/frpc ./cmd/frpc

frpc-dockerfile:
	@echo "Tag $(AGENT_TAG)"

	docker buildx build --platform linux/amd64 \
  		--push --pull \
  		-t us-docker.pkg.dev/$(DOCKER_BASE)/frpc:latest \
  		-t us-docker.pkg.dev/$(DOCKER_BASE)/frpc:$(AGENT_TAG) \
  		-f dockerfiles/Dockerfile-for-frpc \
  		.

test: gotest

gotest:
	go test -v --cover ./assets/...
	go test -v --cover ./cmd/...
	go test -v --cover ./client/...
	go test -v --cover ./server/...
	go test -v --cover ./pkg/...

e2e:
	./hack/run-e2e.sh

e2e-trace:
	DEBUG=true LOG_LEVEL=trace ./hack/run-e2e.sh

e2e-compatibility-last-frpc:
	if [ ! -d "./lastversion" ]; then \
		TARGET_DIRNAME=lastversion ./hack/download.sh; \
	fi
	FRPC_PATH="`pwd`/lastversion/frpc" ./hack/run-e2e.sh
	rm -r ./lastversion

e2e-compatibility-last-frps:
	if [ ! -d "./lastversion" ]; then \
		TARGET_DIRNAME=lastversion ./hack/download.sh; \
	fi
	FRPS_PATH="`pwd`/lastversion/frps" ./hack/run-e2e.sh
	rm -r ./lastversion

alltest: vet gotest e2e
	
clean:
	rm -f ./bin/frpc
	rm -f ./bin/frps
	rm -rf ./lastversion

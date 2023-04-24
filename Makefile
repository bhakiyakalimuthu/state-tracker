all: clean build test test-race lint gofumpt docker-image docker-run
APP_NAME := state-tracker

GOPATH := $(if $(GOPATH),$(GOPATH),~/go)
VERSION := $(shell git describe --tags --always)

clean:
	rm -rf ${APP_NAME} build/

build:
	go build -trimpath -ldflags "-X main._BuildVersion=${VERSION}" -v -o ${APP_NAME} cmd/main.go

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	gofmt -d -s .
	gofumpt -d -extra .
	go vet ./...
	staticcheck ./...
	golangci-lint run

gofumpt:
	gofumpt -l -w -extra .

docker-image-server:
	DOCKER_BUILDKIT=1 docker build --platform linux/amd64 --progress=plain  --build-arg VERSION=${VERSION} -f dockerfiles/server/Dockerfile . -t ${APP_NAME}-server-${VERSION}:${VERSION}

osx-docker-image-server:
	DOCKER_BUILDKIT=1 docker build --platform linux/arm64  --progress=plain  --build-arg APP_NAME=${APP_NAME}-server --build-arg VERSION=${VERSION} -f dockerfiles/server/Dockerfile . -t ${APP_NAME}-server-${VERSION}:latest

docker-run-server:
	docker run --env-file=.env.example -p 9090:9090 ${APP_NAME}-server-${VERSION}

docker-image-client:
	DOCKER_BUILDKIT=1 docker build --platform linux/amd64 --progress=plain  --build-arg VERSION=${VERSION} -f dockerfiles/client/Dockerfile . -t ${APP_NAME}-client-${VERSION}:${VERSION}

osx-docker-image-client:
	DOCKER_BUILDKIT=1 docker build --platform linux/arm64  --progress=plain  --build-arg APP_NAME=${APP_NAME}-client --build-arg VERSION=${VERSION} -f dockerfiles/client/Dockerfile . -t ${APP_NAME}-client-${VERSION}:latest

docker-run-client:
	docker run  --env-file=.env.example  ${APP_NAME}-client-${VERSION}

ifndef GO
	GO := go
endif
GO_BUILD := CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GO} build -o
DOCKER_BUILD := DOCKER_BUILDKIT=1 docker build
OSPD_POD := $(shell kubectl get pods -o wide | awk '/openvas/{print $$1}')

all: build-docker

build-cmds:
	- mkdir bin || true
	 ${GO_BUILD} bin/run-feature-tests cmd/ft/main.go
	
build-docker: build-cmds
	${DOCKER_BUILD} -t greenbone/scanner-lab-feature-tests -f Dockerfile .

run:
	go run ./cmd/ft/main.go

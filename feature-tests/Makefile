ifndef GO
	GO := go
endif
GO_BUILD := CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ${GO} build -o
DOCKER_BUILD := DOCKER_BUILDKIT=1 docker build

all: build-docker

build-cmds:
	- mkdir bin || true
	 ${GO_BUILD} bin/run-feature-tests cmd/ft/main.go
	
build-docker: build-cmds
	${DOCKER_BUILD} -t greenbone/scanner-lab-feature-tests -f Dockerfile .

# only needed when k3s is not using docker backend
# https://rancher.com/docs/k3s/latest/en/advanced/#using-docker-as-the-container-runtime
import-to-k3s: build-docker
	docker save greenbone/scanner-lab-feature-tests | sudo k3s ctr images import -
	sudo systemctl restart k3s.service
	kubectl rollout restart deployment/openvas

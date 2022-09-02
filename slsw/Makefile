ifndef GITHUB_REPOSITORY_OWNER
	GITHUB_REPOSITORY_OWNER := greenbone
endif

ifndef SL_C_REGISTRY
	SL_C_REGISTRY := ghcr.io/${GITHUB_REPOSITORY_OWNER}
endif

IMAGE_NAME := ${SL_C_REGISTRY}/scanner-lab-simulate-website
IMAGE_TAG := latest
BUILD_IMAGE := DOCKER_BUILDKIT=1 docker build --build-arg VERSION=${IMAGE_TAG}
TMP_EXPORT_PATH := /tmp/$(subst /,_,${IMAGE_NAME})_${IMAGE_TAG}.tar

GO_BUILD := CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o

build-slsw:
	mkdir -p bin
	 ${GO_BUILD} bin/simuate-single-website cmd/slsw/main.go

build: build-slsw
	${BUILD_IMAGE} -f Dockerfile -t ${IMAGE_NAME}:${IMAGE_TAG} .

push: build
	docker push ${IMAGE_NAME}:${IMAGE_TAG}

import-into-k3s: build
	docker save --output ${TMP_EXPORT_PATH} ${IMAGE_NAME}:${IMAGE_TAG}
	sudo k3s ctr images import ${TMP_EXPORT_PATH}
	rm ${TMP_EXPORT_PATH}

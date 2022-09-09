ifndef GITHUB_REPOSITORY_OWNER
	GITHUB_REPOSITORY_OWNER := greenbone
endif

ifndef SL_C_REGISTRY
	SL_C_REGISTRY := ghcr.io/${GITHUB_REPOSITORY_OWNER}
endif

IMAGE_NAME := ${SL_C_REGISTRY}/scanner-lab-slackware
IMAGE_TAG := latest
BUILD_IMAGE := DOCKER_BUILDKIT=1 docker build --build-arg VERSION=${IMAGE_TAG}
TMP_EXPORT_PATH := /tmp/$(subst /,_,${IMAGE_NAME})_${IMAGE_TAG}.tar
SW_PACKAGES_DIR := /var/log/packages/
SW_PACKAGES_CMD := ls ${SW_PACKAGES_DIR} | sort

define run_product_version = 
$(shell cat version.txt)
endef

define run_version =
$(subst Slackware ,,$(call run_product_version))
endef

generate-host-ssh-keys:
	ssh-keygen -t rsa -f ${IMAGE_TAG}_rsa_key -N ""
	ssh-keygen -t ecdsa -f ${IMAGE_TAG}_ecdsa_key -N ""
	ssh-keygen -t dsa -f ${IMAGE_TAG}_dsa_key -N ""
	ssh-keygen -t ed25519 -f ${IMAGE_TAG}_ed25519_key -N ""

packages.lst:
	[ ! -e ${SW_PACKAGES_DIR} ] && docker run --rm ${IMAGE_NAME}:${IMAGE_TAG} sh -c '${SW_PACKAGES_CMD}' > packages.lst || ${SW_PACKAGES_CMD} > packages.lst

version.txt:
	[ ! -e /etc/slackware-version ] && docker run --rm ${IMAGE_NAME}:${IMAGE_TAG} sh -c 'cat /etc/slackware-version' > version.txt || cp /etc/slackware-version version.txt

create-products: version.txt packages.lst
	mkdir products || true
	sh create_products.sh > products/slackware_$(call run_version).notus

build:
	${BUILD_IMAGE} -f Dockerfile -t ${IMAGE_NAME}:${IMAGE_TAG} .

push: build
	docker push ${IMAGE_NAME}:${IMAGE_TAG}

import-into-k3s: build
	docker save --output ${TMP_EXPORT_PATH} ${IMAGE_NAME}:${IMAGE_TAG}
	sudo k3s ctr images import ${TMP_EXPORT_PATH}
	rm ${TMP_EXPORT_PATH}

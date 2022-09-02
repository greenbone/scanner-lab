ifndef GITHUB_REPOSITORY_OWNER
	GITHUB_REPOSITORY_OWNER := greenbone
endif

ifndef SL_C_REGISTRY
	SL_C_REGISTRY := ghcr.io/${GITHUB_REPOSITORY_OWNER}
endif

DEFAULT_SL_C_REGISTRY := ghcr.io/greenbone
 
RSYNC := rsync -ltvrP --delete --exclude private/ --perms --chmod=Fugo+r,Fug+w,Dugo-s,Dugo+rx,Dug+w
RSYNC_BASE := rsync://feed.community.greenbone.net:
VERSION := 22.04

NASL_TARGET_DEFAULT := /var/lib/openvas/plugins
nasl_target := ${INSTALL_PREFIX}${NASL_TARGET_DEFAULT}

NOTUS_TARGET_DEFAULT := /var/lib/notus
notus_target := ${INSTALL_PREFIX}${NOTUS_TARGET_DEFAULT}

SC_TARGET_DEFAULT := /var/lib/gvm/data-objects/gvmd/22.04/scan-configs
sc_target := ${INSTALL_PREFIX}${SC_TARGET_DEFAULT}

PVD := openvas-persistent-volumes-deployment.yaml
PVD_LOCAL := openvas-persistent-volumes-deployment-local.yaml

ifeq ($(wildcard ${PVD_LOCAL}),)
	pvd := ${PVD}
else
	pvd := ${PVD_LOCAL}
endif

ifeq ($(wildcard slackware-deployment-local.yaml),)
	slackware-depl := slackware-deployment.yaml 
else
	slackware-depl := slackware-deployment-local.yaml 
endif

ifeq ($(wildcard slsw-deployment-local.yaml),)
	slsw-depl := slsw-deployment.yaml 
else
	slsw-depl := slsw-deployment-local.yaml 
endif

all: deploy

create-local-volume-deployment:
	sed 's|${NASL_TARGET_DEFAULT}|${nasl_target}|' ${PVD} > ${PVD_LOCAL}
	sed -i 's|${NOTUS_TARGET_DEFAULT}|${notus_target}|' ${PVD_LOCAL}
	sed -i 's|${SC_TARGET_DEFAULT}|${sc_target}|' ${PVD_LOCAL}

# changes the deployment files for repository build images
# this is done so that forks can simply change the address
# and use their published image instead of officials
create-local-sl-deployment:
	sed 's|${DEFAULT_SL_C_REGISTRY}|${SL_C_REGISTRY}|' slackware-deployment.yaml > slackware-deployment-local.yaml
	sed 's|${DEFAULT_SL_C_REGISTRY}|${SL_C_REGISTRY}|' slsw-deployment.yaml > slsw-deployment-local.yaml


check-persistent-volume-paths:
	@ grep "${nasl_target}" ${pvd} > /dev/null || (printf "\e[31m${nasl_target} not configured in ${pvd}\e[0m\n" && false)
	@ grep "${notus_target}" ${pvd} > /dev/null || (printf "\e[31m${notus_target} not configured in ${pvd}\e[0m\n" && false)
	@ grep "${sc_target}" ${pvd} > /dev/null || (printf "\e[31m${sc_target} not configured in ${pvd}\e[0m\n" && false)

check-feed-dirs: check-persistent-volume-paths
	@ [ -d ${nasl_target} ] || (printf "\e[31m${nasl_target} missing\e[0m\n" && false)
	@ [ -d ${notus_target} ] || (printf "\e[31m${notus_target} missing\e[0m\n" && false)
	@ [ -d ${sc_target} ] || (printf "\e[31m${sc_target} missing\e[0m\n" && false)
	@ test -w ${nasl_target} || (printf "\e[31m${USER} cannot write into ${nasl_target}\e[0m\n" && false) 
	@ test -w ${notus_target} || (printf "\e[31m${USER} cannot write into ${notus_target}\e[0m\n" && false) 
	@ test -w ${sc_target} || (printf "\e[31m${USER} cannot write into ${sc_target}\e[0m\n" && false) 

update-feed: check-feed-dirs
	${RSYNC} ${RSYNC_BASE}/community/vulnerability-feed/${VERSION}/vt-data/nasl/ ${nasl_target}
	${RSYNC} ${RSYNC_BASE}/community/vulnerability-feed/${VERSION}/vt-data/notus/ ${notus_target}
	${RSYNC} ${RSYNC_BASE}/community/data-feed/${VERSION}/scan-configs/ ${sc_target}

deploy-persistent-volumes: check-persistent-volume-paths
	kubectl apply -f ${pvd}

deploy-openvas: deploy-persistent-volumes
	kubectl apply -f openvas-deployment.yaml

deploy-victim:
	kubectl apply -f victim-deployment.yaml

deploy-slsw:
	kubectl apply -f ${slsw-deply}

deploy-slackware:
	kubectl apply -f ${slackware-depl}

deploy: deploy-openvas deploy-victim deploy-slackware deploy-slsw

delete-openvas:
	kubectl delete deployment/openvas

delete-persistent-volumes: delete-openvas
	kubectl delete -f ${pvd}

delete-victim:
	kubectl delete deployment/victim

delete-slsw:
	kubectl delete deployment/slsw

delete-slackware:
	kubectl delete deployment/slackware

delete: delete-persistent-volumes delete-victim delete-slackware delete-slsw

update-openvas:
	kubectl rollout restart deployment/openvas

update-victim:
	kubectl rollout restart deployment/victim

update-slsw:
	kubectl rollout restart deployment/slsw

update-slackware:
	kubectl rollout restart deployment/slackware

update: update-openvas update-victim update-slackware update-slsw

build:
	$(MAKE) -C slsw build
	$(MAKE) -C slackware build

push:
	$(MAKE) -C slsw push
	$(MAKE) -C slackware push

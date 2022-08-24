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

create-local-volume-deployment:
	sed 's|${NASL_TARGET_DEFAULT}|${nasl_target}|' ${PVD} > ${PVD_LOCAL}
	sed -i 's|${NOTUS_TARGET_DEFAULT}|${notus_target}|' ${PVD_LOCAL}
	sed -i 's|${SC_TARGET_DEFAULT}|${sc_target}|' ${PVD_LOCAL}

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
	kubectl apply -f slsw-deployment.yaml

deploy: deploy-openvas deploy-victim deploy-slsw

delete-openvas:
	kubectl delete deployment/openvas

delete-persistent-volumes: delete-openvas
	kubectl delete -f ${pvd}

delete-victim:
	kubectl delete deployment/victim

delete-slsw:
	kubectl delete deployment/slsw

delete: delete-persistent-volumes delete-victim delete-slsw

update-openvas:
	kubectl rollout restart deployment/openvas

update-victim:
	kubectl rollout restart deployment/victim

update-slsw:
	kubectl rollout restart deployment/slsw

update: update-openvas update-victim update-slsw

RSYNC := rsync -ltvrP --delete --exclude private/ --perms --chmod=Fugo+r,Fug+w,Dugo-s,Dugo+rx,Dug+w
RSYNC_BASE := rsync://feed.community.greenbone.net:
VERSION := 22.04
NASL_TARGET := /var/lib/openvas/plugins/
NOTUS_TARGET := /var/lib/notus/
SC_TARGET := /var/lib/gvm/data-objects/gvmd/22.04/scan-configs/

check-feed-dirs:
	@ [ -d ${NASL_TARGET} ] || (printf "\e[31m${NASL_TARGET} missing\e[0m\n" && false)
	@ [ -d ${NOTUS_TARGET} ] || (printf "\e[31m${NOTUS_TARGET} missing\e[0m\n" && false)
	@ [ -d ${SC_TARGET} ] || (printf "\e[31m${SC_TARGET} missing\e[0m\n" && false)
	@ test -w ${NASL_TARGET} || (printf "\e[31m${USER} cannot write into ${NASL_TARGET}\e[0m\n" && false) 
	@ test -w ${NOTUS_TARGET} || (printf "\e[31m${USER} cannot write into ${NOTUS_TARGET}\e[0m\n" && false) 
	@ test -w ${SC_TARGET} || (printf "\e[31m${USER} cannot write into ${SC_TARGET}\e[0m\n" && false) 

update-feed: check-feed-dirs
	${RSYNC} ${RSYNC_BASE}/community/vulnerability-feed/${VERSION}/vt-data/nasl/ ${NASL_TARGET}
	${RSYNC} ${RSYNC_BASE}/community/vulnerability-feed/${VERSION}/vt-data/notus/ ${NOTUS_TARGET}
	${RSYNC} ${RSYNC_BASE}/community/data-feed/${VERSION}/scan-configs/ ${SC_TARGET}

deploy-openvas:
	kubectl apply -f openvas.yaml

deploy-victim:
	kubectl apply -f victim.yaml

deploy-slsw:
	kubectl apply -f slsw.yaml

deploy: deploy-openvas deploy-victim deploy-slsw

delete-openvas:
	kubectl delete deployment/openvas

delete-victim:
	kubectl delete deployment/victim

delete-slsw:
	kubectl delete deployment/slsw

delete: delete-openvas delete-victim delete-slsw

update-openvas:
	kubectl rollout restart deployment/openvas

update-victim:
	kubectl rollout restart deployment/victim

update-slsw:
	kubectl rollout restart deployment/slsw

update: update-openvas update-victim update-slsw

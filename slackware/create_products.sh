#!/bin/sh
PRODUCT_NAME=$(cat version.txt)
OID=$(grep oid advisories/slackware.notus | sed 's/.*"\(\([0-9]\+\.\)\+\)\([0-9]\+\).*/\1\3/')
LST=$(cat packages.lst)
printf '{ "version": "1.0", "package_type": "slack", "product_name": "'
printf "$PRODUCT_NAME"
printf '", "advisories": [ { "oid": "'
printf "$OID"
printf '", "fixed_packages": ['
for pkg in $LST; do
  printf '{ "name": null, "full_version": null, "full_name": "'
  printf "$pkg"
  printf '", "specifier": ">"},'
done
printf '{ "name": null, "full_version": null, "full_name": null, "specifier": ">="}'
printf ']}]}\n'

# scanner-lab

Is a repository to simplify testing openvas in a contained environment by using Kubernetes.

This is in an early stage and will be adapted over time.

## Installation

On a newly created environment you need to have

- make
- rsync
- this repository

on your machine.

Requirements:

- `/var/lib/openvas/plugins/`
- `/var/lib/notus/`
- `/var/lib/gvm/data-objects/gvmd/22.04/scan-configs/`

must exist and writeable by the user so that `make update-feed` can succeed.

You can verify it by running `make check-feed-dirs`. If there is no output and no error code this is correctly setup.

### Install k3s

Although k3s is just a single binary it is useful to have a systemd integration for that they prepared a script which you can download via:

```
  curl -Lo install_k3s.sh https://get.k3s.io
```

review and execute it.

The script should install:
- `/usr/local/bin/k3s`
- `/usr/local/bin/kubectl` - kubernetes client (symlinked to k3s)
- `/usr/local/bin/crictl` - CRI client (symlinked to k3s)
- `/usr/local/bin/k3s-killall.sh` - to kill k3s
- `/usr/local/bin/k3s-uninstall.sh` - to uninstall

Additionally it should create
- `/etc/systemd/system/k3s.service`
and enabling it per default.

To allow user execution set a `KUBECONFIG` variable:


```
export KUBECONFIG=~/.kube/config
```

if you already have running pods you can copy the configuration like:

```
mkdir -p ~/.kube
sudo k3s kubectl config view --raw > "$KUBECONFIG"
```

Further resources:
- https://rancher.com/docs/k3s/latest/en/installation/installation-requirements/
- https://rancher.com/docs/k3s/latest/en/quick-start/


### Apply deployments

```
make update-feed
make deploy
```

### Remove deployments

```
make delete
```

### Update

```
make update-feed
make update
```

### Scale
```
kubectl scale deployments/victim --replicas=100
kubectl scale deployments/slsw --replicas=100
```

## Useful commands


### Use own paths instead of defaults

If you follow the standard setup

- $YOUR_PATH/var/lib/openvas/plugins
- $YOUR_PATH/var/lib/notus
- $YOUR_PATH/var/lib/gvm/data-objects/gvmd/22.04/scan-configs

but just in a different path you can also set `INSTALL_PREFIX` either via environment or make variable instead of overriding each feed variable before executing `create-local-volume-deployment`.

If you want to use different source paths than set you can create a own `openvas-persistent-volumes-deployment-local.yaml` by executing:

```
make \
  nasl_target=$YOUR_NASL_PATH \
  notus_target=$YOUR_NOTUS_PATH \
  sc_target=$YOUR_SCAN_CONFIG_PATH \
  create-local-volume-deployment
```

Be aware that when you want to run `make update-feed` you need to apply the same values as you did when creating `openvas-persistent-volumes-deployment-local.yaml`
If you change the `INSTALL_PREFIX` then you have to delete the persistent volume and openvas and deploy afterwards:

```
make delete-persistant-volumes
make deploy-openvas
```

### start a scan

```
kubectl exec -ti deployment/openvas -c ospd -- bash
ospd-scans \
  -a localhost:4242 \
  --cert-path /var/lib/gvm/CA/cacert.pem \
  --certkey-path /var/lib/gvm/private/CA/serverkey.pem \
  --host 10.42.0.0/24 \
  --policies "Discovery,Full and fast" \
  --cmd start-finish
```

### openvas logs
```
kubectl exec -ti deployment/openvas -c ospd -- tail -f /var/log/gvm/openvas.log 
```

## Usage

To use the exposed TCP socket to OSPD you have to get the IP-Address of openvas:

```
kubectl get pods -l app=openvas -o wide
```

and the certificate and key file:
```
cd feature-tests
make fetch-certs
```

afterwards you can connect to it via:

```
echo "<get_version/>" | gnutls-cli \
  --port=4242 \
  --insecure \
  --x509certfile=/tmp/ca.pem \
  --x509keyfile=/tmp/key.pem \
  $(kubectl get pods -o wide | awk '/openvas/{print $6}')
```

### run feature tests

```
cd ./feature-tests
make run
```



# eulabeia-lab

Is a lab to test openvas in a larger environment.

## Installation

On a newly created environment you need to have
- curl
- victim.yaml
- openvas.yaml

on your machine.

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
kubectl apply -f openvas.yaml
kubectl apply -f victim.yaml
kubectl apply -f slsw.yaml
```

### Remove deployments

```
kubectl delete deployments/openvas
kubectl delete deployments/victim
kubectl delete deployments/slsw
```

### Update

```
kubectl rollout restart -f openvas.yaml
kubectl rollout restart -f victim.yaml
kubectl rollout restart -f slsw.yaml
```

### Scale
```
kubectl scale deployments/victim --replicas=100
kubectl scale deployments/slsw --replicas=100
```

## Useful commands

### start a scan

```
kubectl exec -ti deployment/openvas -c ospd -- bash
OSPD_SOCKET=/run/ospd/ospd-openvas.sock ospd-scans -host 10.42.0.0/24 -policies "Discovery,Full and fast" -cmd start-finish
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

afterwards you can connect to it via:

```
echo -e "<get_vts/>" | nc 10.42.0.81 4242
```

### run feature tests

TBD

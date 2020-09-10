## Install Panoptes-Stream on Kubernetes with Consul
--------------

### Minikube

#### Installing [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
#### Installing Consul


```console
# helm repo add hashicorp https://helm.releases.hashicorp.com
# helm install -f consul-values.yaml hashicorp hashicorp/consul
```
#### Installing Panoptes-Stream
```console
# helm install panoptes panoptes-stream
```
#### Verify by kubectl
```console
# kubectl get pods
```
![minikube dashboard](imgs/minikube_kubectl.png)
#### Verify by Consul UI
```console
# minikube service hashicorp-consul-ui
```
![minikube dashboard](imgs/minikube_consul.png)
#### Verify by Minikube dashboard
```console
# minikube dashboard
```
![minikube dashboard](imgs/minikube_dashboard.png)
#### Clean-up
```console
# helm uninstall hashicorp 
# helm uninstall panoptes 
```
### Install Panoptes-Stream on Kubernetes


### Minikube

#### Installing [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
#### Installing Consul
```
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install -f consul-values.yaml hashicorp hashicorp/consul
```
#### Installing Panoptes-Stream
```
helm install panoptes panoptes-stream
```
#### Verify by Consul UI
```
minikube service hashicorp-consul-ui
```
#### Verify by Minikube dashboard
```
minikube dashboard
```
#### Clean-up
```
helm uninstall hashicorp 
helm uninstall panoptes 
```

### Kubernetes Cluster

#### Installing Consul
Open consul-values.yaml and set replicas and bootstrapExpect to 3 or 5 
```
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install -f consul-values.yaml hashicorp hashicorp/consul
```
#### Installing Panoptes-Stream
Open panoptes-stream/values.yaml and set affinity: true (recommended) and edit replicaCount to any number that you need.
```
helm install panoptes panoptes-stream
```
#### Viewing the Consul UI
```
kubectl port-forward service/hashicorp-consul-server 8500:8500
```
#### Clean-up
```
helm uninstall hashicorp 
helm uninstall panoptes 
```
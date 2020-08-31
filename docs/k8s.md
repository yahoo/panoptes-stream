### Panoptes-Stream Kubernetes


#### Minikube

- Install (Minikube)[https://kubernetes.io/docs/tasks/tools/install-minikube/] 
- Install Consul
```
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install -f values.yaml hashicorp hashicorp/consul
```
- Install Panoptes-Stream
```
helm install panoptes panoptes-stream
```
- Verify by Consul UI
```
minikube service hashicorp-consul-ui
```
- Verify by Minikube dashboard
```
minikube dashboard
```
- Clean-up
```
helm uninstall hashicorp 
helm uninstall panoptes 
```
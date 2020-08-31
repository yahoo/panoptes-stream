### Panoptes-Stream Helm Chart

#### Prerequisites
- Helm 2.10+ or Helm 3.0+
- Kubernetes 1.9+

#### Usage
Detailed installation instructions for Panoptes on Kubernetes are found (here)[/docs/k8s.md].

- Add the HashiCorp Consul Helm Repository and Install it (dependency)
```
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install -f values.yaml hashicorp hashicorp/consul
```
- Once the Consul pods are running then install Panoptes
```
helm install panoptes panoptes-stream
```


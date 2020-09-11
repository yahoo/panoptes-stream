## Helm Chart
---

#### Reference
- namespace (string: default) - The default kubernetes namespace.
- replicaCount (integer: 3) - Number of replicas for the Deployment / StatefulSet.
- initializingShards: (integer: 1) - The minimum requirement available shards nodes. 
- serviceDiscovery: (string: k8s) - The service discovery mode: k8s / consul. 
- configuration: (string: k8s) - The configuration management: k8s (configmap) / consul.
- affinity: (boolean: false) - It allows / doesn't allow to schedule Panoptes pods on the same node.


#### Deploy in panoptes namespace

- Create panoptes namespace
```console
kubectl create ns panoptes
```
- Change the namespace to panoptes at [template/values.yaml](/k8s/panoptes-stream/templates/values.yaml)
```yaml
namespace: panoptes
```
- Helm install
```console
helm install panoptes panoptes-stream -n panoptes
```

### Kubernetes Cluster
--------------

#### Panoptes-Stream on Kubernetes with Consul

##### Create a GKE cluster
You can skip this step if you already have a k8s cluster or you want to create a cluster somewhere else. the below command create a 3x nodes k8s cluster at GKE US Central with e2-medium machine type (2 vCPU and 4 GB memory)

```
gcloud container clusters create panoptes-demo -z us-central1-b -m e2-medium
```

##### Installing Consul
Open [consul-values.yaml](https://github.com/yahoo/panoptes-stream/blob/master/scripts/k8s/consul-values.yaml) and set replicas and bootstrapExpect to 3 or 5 
```console
# helm repo add hashicorp https://helm.releases.hashicorp.com
# helm install -f consul-values.yaml hashicorp hashicorp/consul
```
##### Installing Panoptes-Stream
Open [panoptes-stream/values.yaml](https://github.com/yahoo/panoptes-stream/blob/master/helm/panoptes-stream/values.yaml) and set as below:

```
affinity: true
serviceDiscovery: consul
configuration: consul
```
Then run the below command once you're at helm directory:
```console
# helm install panoptes-stream panoptes-stream
```
You can verify it by kubectl command as below
```
# kubectl get pods -o wide
NAME                               READY   STATUS      RESTARTS   AGE   IP          NODE                                           NOMINATED NODE   READINESS GATES
hashicorp-consul-cx4fr             1/1     Running     0          88s   10.32.2.7   gke-panoptes-demo-default-pool-6f4c0fd2-kqvm   <none>           <none>
hashicorp-consul-n95bl             1/1     Running     0          88s   10.32.0.3   gke-panoptes-demo-default-pool-6f4c0fd2-gfrc   <none>           <none>
hashicorp-consul-rtxb8             1/1     Running     0          88s   10.32.1.6   gke-panoptes-demo-default-pool-6f4c0fd2-xd98   <none>           <none>
hashicorp-consul-server-0          1/1     Running     0          88s   10.32.0.4   gke-panoptes-demo-default-pool-6f4c0fd2-gfrc   <none>           <none>
hashicorp-consul-server-1          1/1     Running     0          88s   10.32.1.7   gke-panoptes-demo-default-pool-6f4c0fd2-xd98   <none>           <none>
hashicorp-consul-server-2          1/1     Running     0          88s   10.32.2.8   gke-panoptes-demo-default-pool-6f4c0fd2-kqvm   <none>           <none>
panoptes-job-frd2b                 0/1     Completed   0          23s   10.32.0.5   gke-panoptes-demo-default-pool-6f4c0fd2-gfrc   <none>           <none>
panoptes-stream-84969fbcf7-gqgqg   1/1     Running     0          19s   10.32.1.8   gke-panoptes-demo-default-pool-6f4c0fd2-xd98   <none>           <none>
panoptes-stream-84969fbcf7-s4hh8   1/1     Running     0          19s   10.32.0.6   gke-panoptes-demo-default-pool-6f4c0fd2-gfrc   <none>           <none>
panoptes-stream-84969fbcf7-s7vvs   1/1     Running     0          19s   10.32.2.9   gke-panoptes-demo-default-pool-6f4c0fd2-kqvm   <none>           <none>
```
##### Viewing the Consul UI
```console
# kubectl port-forward service/hashicorp-consul-server 8500:8500
```
##### Clean-up
```console
# helm uninstall hashicorp 
# helm uninstall panoptes-stream 
```
In case you need to delete the GKE cluster, execute the below command
```
# gcloud container clusters delete panoptes-demo -z us-central1-b
```
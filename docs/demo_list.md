### Panoptes Streaming Demo
---
You can try four different Panoptes streaming demonstrations at your laptop. they use Panoptes gNMI simulator to simulate Juniper routers so you can see exactly what's happening at the real production. The demo will take a minute or two to execute and the results will be accessible at grafana dashboards.


You need to install [docker](https://docs.docker.com/get-docker/), in case you don’t have it already. 


![panoptes consul demo](imgs/demo_shards_etcd.png)

- [Single node](demo.md)
- [Cluster with Consul](demo_consul_shards.md)
- [Cluster with etcd](demo_etcd_shards.md)
- [Cluster with Pseudo](demo_pseudo_shards.md)
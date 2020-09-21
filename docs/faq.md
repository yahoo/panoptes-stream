## FAQ
---
* What's the minimum requirement to run a single Panoptes instance?
    >Panoptes needs at least 0.1 GB RAM and 1x vCPU. You can run it with a single binary without any dependency and configure it through following options: YAML file, etcd or consul.
* Can Panoptes cluster work without etcd or consul?
    >Yes, you can install it without etcd or consul. Panoptes uses Pseudo build-in health checking service instead and you need to configure node's addresses for each Panoptes daemon.
* Does Panoptes require to reload when the   YAML configuration updates?
    >No, Panoptes detects the changes immediately and apply the new configuration after a few seconds.      
* What does routing metrics means?
    >Panoptes can route the metrics / sensors to specific producer with given topic or route to a specific database with a given table. for example you can send the BGP metrics to Kafka with topic: BGP and send the interface counters to Kafka with topic: Interface also at the same time you can route VLAN metrics to InfluxDB with measurement vlan.
* How to monitor Panoptes?
    >Panoptes provides local monitoring information and exposes them through http endpoint with Prometheus format.         
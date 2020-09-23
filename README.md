![panoptes logo](/docs/imgs/panoptes_streaming_logo.png)

## Panoptes Streaming



Panoptes Streaming is a cloud native distributed streaming network telemetry. It can be installed as a single binary without any dependencies or clustered nodes to collect network telemetry through gRPC and produce or ingest them to the given destinations. Panoptes can grow horizontally through auto sharding of devices to achieve scalability and fault tolerance by itself or by using a 3rd party service discovery also it runs on the Kubernetes as a cloud native application. you can check the dockerized [demonstrations](/docs/demo_list.md) on your laptop quickly and see how it works in real-time through pre-provisioned grafana dashboards and device simulator.

### Features
- Supports gNMI, Juniper JTI and Cisco MDT.
- Routes sensors to producers and databases. 
- Fault tolerance and scalability through auto sharding.
- Dynamic configuration management.
- Guaranteed telemetry delivery.
- Plugin and cloud friendly architecture.

![panoptes steaming](/docs/imgs/diagram.png)
### Documentation
- [Getting Started](/docs/getting_started.md)
- [Architecture](/docs/architecture.md)
- [Demo](/docs/demo_list.md)
- [FAQ](/docs/faq.md)

###### Sample dashboard
![demo grafana](/docs/imgs/grafana.png)

### License
Code is licensed under the Apache License, Version 2.0 (the "License").
Content is licensed under the CC BY 4.0 license. Terms available at https://creativecommons.org/licenses/by/4.0/.

### Contribute
Welcomes any kind of contribution, please follow the next steps:

- Fork the project on github.com.
- Create a new branch.
- Commit changes to the new branch.
- Send a pull request.
![panoptes logo](/docs/imgs/panoptes_streaming_logo.png)

## Panoptes Streaming

[![Github Actions](https://github.com/yahoo/panoptes-stream/workflows/panoptes-stream/badge.svg)](https://github.com/yahoo/panoptes-stream/actions?query=workflow%3Apanoptes-stream) [![Go report](https://goreportcard.com/badge/github.com/yahoo/panoptes-stream)](https://goreportcard.com/report/github.com/yahoo/panoptes-stream)  [![Coverage Status](https://coveralls.io/repos/github/yahoo/panoptes-stream/badge.svg?branch=master&service=github)](https://coveralls.io/github/yahoo/panoptes-stream) [![PkgGoDev](https://pkg.go.dev/badge/github.com/yahoo/panoptes-stream?tab=doc)](https://pkg.go.dev/github.com/yahoo/panoptes-stream?tab=doc)

Panoptes Streaming is a cloud native distributed streaming network telemetry. It can be installed as a single binary or clustered nodes to collect network telemetry through gRPC and produce or ingest them to given destinations. Panoptes can grow horizontally by auto sharding of devices to achieve scalability and availability. It runs on Kubernetes as a cloud native application and its helm chart supports multiple use cases of Panoptes on Kubernetes. you can check out the dockerized [demonstrations](/docs/demo_list.md) on your laptop quickly and see how it works in real-production through pre-provisioned grafana dashboards and gNMI-simulator.

### Features
- Supports gNMI, Juniper JTI and Cisco MDT.
- Routes sensors to producers and databases. 
- Availability and scalability by auto sharding.
- Dynamic configuration management.
- Guaranteed telemetry delivery.
- Plugin and cloud friendly architecture.

![panoptes steaming](/docs/imgs/diagram.png)
### Documentation
- [Getting Started](/docs/getting_started.md)
- [Architecture](/docs/architecture.md)
- [Demo](/docs/demo_list.md)

###### Sample grafana dashboard
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

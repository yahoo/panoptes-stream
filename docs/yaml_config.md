
### .yaml Basic gNMI Configuration with influxdb

#### Use Case
- Trying the panoptes
- Small to medium networks 


#### .yaml format (dial-in mode)

Here's an basic example .yaml file that shows the required fields and object spec for collecting a sensor path from a device and ingest the data to influxdb every 10 seconds.

```yaml
devices:
  - host: 192.168.59.3
    port: 50051
    sensors:
      - sensor1

sensors:
  sensor1:
    path: /interfaces/interface/state/counters
    mode: sample
    sampleInterval: 10
    service: arista.gnmi
    output: influxdb1::ifcounters

databases:
  influxdb1:
    service: influxdb
    configFile: etc/influxdb.yaml
```

#### etc/influxdb.yaml

```yaml
server: http://localhost:8086
bucket: mybucket
```

#### execute panoptes binary
```
# panoptes -config etc/config.yaml
```
#### execute panoptes docker

#### execute panoptes docker-compose




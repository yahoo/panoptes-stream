## .yaml Basic gNMI Configuration (dial-in mode)
Here's a basic example with minimum requirements .yaml file that shows the required fields for collecting a sensor path from a device and showing the result on the console stdout every 10 seconds.

#### Use Case
- Trying the panoptes


#### Create .yaml file 
- Open a text editor and paste the below yaml and save it as config.yaml.
- Edit IP address and Port. you can add username and password if needed.
- Change the service name to other vendors (cisco.gnmi or juniper.gnmi) if needed.

```yaml
devices:
  - host: 192.168.55.3
    port: 50051
    sensors:
      - sensor1

sensors:
  sensor1:
    path: /interfaces/interface/state/counters
    mode: sample
    sampleInterval: 10
    service: arista.gnmi
    output: console::stdout
```

#### Execute panoptes by binary
```
panoptes -config config.yaml
```
#### Execute panoptes by docker run
```
docker run -d --name panoptes -v $PWD:/etc/panoptes panoptes-stream -config /etc/panoptes/config.yaml 
docker logs -f panoptes
```

#### Result is stream of the metrics like below sample every 10 seconds:
```json
{
  "key": "out-discards",
  "labels": {
    "name": "Ethernet2"
  },
  "prefix": "/interfaces/interface/state/counters",
  "system_id": "192.168.59.3",
  "timestamp": 1596848835935721428,
  "value": 0
}{
  "key": "out-errors",
  "labels": {
    "name": "Ethernet2"
  },
  "prefix": "/interfaces/interface/state/counters",
  "system_id": "192.168.59.3",
  "timestamp": 1596848835935724719,
  "value": 0
}{
  "key": "out-octets",
  "labels": {
    "name": "Ethernet2"
  },
  "prefix": "/interfaces/interface/state/counters",
  "system_id": "192.168.59.3",
  "timestamp": 1596918079559461774,
  "value": 4695711345
}
```

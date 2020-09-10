## Panoptes configuration with yaml
------------

This document will show you how to write Panoptes configuration using [YAML](https://en.wikipedia.org/wiki/YAML).   

### Configuration specs
The Panoptes configuration categories as follows at YAML configuration:

- [Devices](#devices)
- [Sensors](#sensors)
- [Producers](#producers)
- [Databases](#databases)
- [Global](#global)

The below picture shows how the configurations link together.
 
![panoptes config](imgs/link-config.png)

#### Devices 
The devices are defined as an array of devices including device information and the sensors / paths that we need to subscribe. 

Example device configuration:

```yaml
devices:
  - host: 192.168.55.3
    port: 50051
    username: admin
    password: admin
    sensors:
      - sensor1
```
You can see all available device config keys at [configuration reference](config_reference.md#devices).

#### Sensors 
The sensors are defined as a list of sensors. you can assign them to one or more devices under devices configuration.

```yaml
sensors:
  sensor1:
    path: /interfaces/interface/state/counters
    mode: sample
    sampleInterval: 10
    service: juniper.gnmi
    output: console::stdout
```

You can see all available sensor config keys at [configuration reference](config_reference.md#sensors). 

#### Producers
The producers are defined as a list of producers. you can assign them to one or more sensors under sensors configuration.

```yaml
producers:
  kafka1:
    service: kafka
    config:
      brokers:
        - 192.168.55.10:9092
      topics:
        - interfaces
        - bgp
```

#### Databases
The databases are defined as a list of databases. you can assign them to one or more sensors under sensors configuration.

```yaml
databases:
  influxdb1:
    service: influxdb
    config:
      server: http://influxdb:8086
      bucket: mybucket
```

You can see all available sensor config keys at [configuration reference](config_reference.md#databases).

#### Global


- [Status](#status) 
- [Discovery](#discovery)
- [Dialout](#dialout)
- [Shard](#shard)
- [Logger](#logger)
- [DeviceOptions](#deviceoptions)


![panoptes global config](imgs/global-config.png)

#### Status

```yaml
status:
  addr: "0.0.0.0:8081"
```  

#### Discovery

```yaml
discovery:
  service: "consul"
  config:
    address: "127.0.0.1:8500"
```

#### Shards

```yaml
shards:
  enabled: true
  numberOfNodes: 2
  InitializingShards: 
```  

#### Logger

```yaml
logger:
  level: debug
  outputPaths:
    - /var/log/panoptes
  errorOutputPaths:
    - stderr
```    

#### Dialout

```yaml
dialout:
  defaultOutput: console::stdout,
  services: 
  	cisco.mdt: 
      addr: 0.0.0.0:50055
``` 

#### Device Options

```yaml
deviceOptions:
  tlsConfig:
    Enabled: false
  username: "admin"
  password: "admin"
```  


#### Watcher and vim editor
If you edit the yaml config with vim editor during running Panoptes, please add the following commands at your ~/.vimrc to prevent stopping Panoptes watcher. 
```
set nobackup
set nowritebackup
``` 
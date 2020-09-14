## Panoptes configuration with etcd
------------

This document will show you how to configure Panoptes with [etcd](http://etcd.io) key value store.

Panoptes can find etcd address and other configuration from one of the following options:
- Yaml configuration file
- Default configuration once you set a dash as argument. 
- Through environment variables.

```console
panoptes -etcd config.yaml
panoptes -etcd -
```

sample .yaml file
```yaml
endpoints: 
  - 192.168.55.5:2379
  - 192.168.55.6:2379 
prefix: panoptes/config/
```

You can set environment variables with following format: PANOPTES_CONFIG_ETCD_{{key}}
For instance: ```PANOPTES_CONFIG_ETCD_ADDRESS=192.168.55.5:2379,192.168.55.6:2379```

### Configuration specs
The Panoptes configuration categories as follows at etcd key value store:
- [Devices](#devices)
- [Sensors](#sensors)
- [Producers](#producers)
- [Databases](#databases)
- [Global](#global)

The below picture shows how the configurations link together.
 
![panoptes config](imgs/link-config.png)

#### Devices 
The devices is a folder which included devices as key value. key can be any name and value is in json format.
Example device configuration:

Key: core1.lax  
Value: 
```json
{ 
    "host": "core1.lax", 
    "port": 50051, 
    "sensors" : ["sensor1"],
    "username" : "demo",
    "password" : "demo"
}
```

You can see all available device config keys at [configuration reference](config_reference.md#devices). 

#### Sensors 
The sensors is a folder which included sensors as key value. key is the sensor name and value is in json format.
Example sensor configuration:

Key: sensor1  
Value:
```json
{ 
    "service": "cisco.gnmi", 
    "path": "/interfaces/interface/state/counters", 
    "mode": "sample", 
    "sampleInterval": 10, 
    "output":"console::stdout" 
}
```

You can see all available sensor config keys at [configuration reference](config_reference.md#sensors). 

#### Producers
The producers is a folder which included producers configurations as key value. key is the producer name and value is in json format.  

Example [Kafka](https://kafka.apache.org/) producer configuration:

Key: kafka1   
Value:
```json
{
    "service": "kafka",
    "config" : {
        "brokers": ["127.0.0.1:9092"],
         "topics":["interface","bgp"],
        "batchSize" : 1000
    }
}
```
The key and topics will assign to the sensor's output like: "output": "kafka1::interface" or "output": "kafka1::bgp"
Kafka output syntax: KEY::TOPIC 

You can see all available producers config keys at [configuration reference](config_reference.md#producers). 

#### Databases
The databases is a folder which included databases as key value. key is the database name and value is in json format.

Example [Influxdb database](https://www.influxdata.com/) configuration:

Key: influxdb1   
Value:
```json
{
    "service": "influxdb",
    "config": { 
    "server": "http://localhost:8086", 
    "bucket":"mybucket"
    } 
}
```
The key and a measurement name related to sensor will assign to the sensor's output like: "output": "influxdb1::ifcounters" or "output": "influxdb1::bgp"
Influxdb output syntax: KEY::MEASUREMENT

You can see all available databases config keys at [configuration reference](config_reference.md#database). 

#### Global
The global is a key not a folder and it's quite different than other configuration categories as follows:

- [Status](#status) 
- [Discovery](#discovery)
- [Dialout](#dialout)
- [Shard](#shard)
- [Logger](#logger)
- [DeviceOptions](#deviceoptions)

They are all optional and depends on what you need. for instance if you want to have sharding, you need to enable and configure Status and Discovery or if you want to change the logging level, you can configure logger.

![panoptes global config](imgs/global-config.png)

Key: global  
Value:
```json
{
  "status": {},
  "shard": {},
  "discovery": {},
  "dialout": {},
  "logger": {},
  "deviceOptions": {},
}
```

#### Status
Panoptes has built-in self monitoring and healthcheck that they expose through HTTP or HTTPS. the Panoptes metrics are readable by Prometheus server.

Example status configuration:

```json
"status": {
    "addr": "0.0.0.0:8081",
    "tlsConfig": {
    "enabled": true,
    "certFile": "/demo/certs/panoptes.crt",
    "keyFile": "/demo/certs/panoptes.key",
   }
}
```

#### Discovery
Panoptes can register itself to etcd discovery service and it is required once you enabled the Sharding feature. 

Example discovery configuration:

```json
"discovery": {
   "service": "etcd",
   "config": { 
     "endpoints": ["127.0.0.1:2379"],
   } 
 }
```

Detailed instructions for Panoptes service discovery are found [here](discovery.md)

#### Dialout
Panoptes supports gRPC [dial-out](glossary.md#dialout) mode for Cisco MDT at the moment. you can configure it to listen on a specific address and port that should be reachable from your dial-out mode devices. 

Example Dial-Out mode configuration:

```json
"dialout": {
    "services": {
  		"cisco.mdt": {
     		"addr": "0.0.0.0:50055"
    	}
    }
 } 
```

#### Shards
By enabling sharding, Panoptes's nodes try to auto sharding of network devices and take over if one or more nodes have been failed. if you need details information please read [Sharding Deep Dive](shards.md)

Example Shard configuration:

```json
"shards" : {
   "enabled": true,
   "numberOfNodes": 3,
   "InitializingShards" : 1
 }
```

#### Logger
The level of logging to show and output destination after the Panoptes has started. the output can be file or console (stdout/stdin).

Example Logging configuration:

```json
"logger": {
    "level":"debug", 
    "outputPaths": ["stdout"], 
    "errorOutputPaths":["stderr"]
  }
```

#### Device Options
The device options are shared configuration between all of the devices. the device overlapped configuration is priority, it means a device configured options are preferred to global device options.

Example Device Options configuration:

```json
"deviceOptions": {
    "username": "demo",
    "password": "demo",
    "tlsConfig": { 
      "enabled": false, 
      "certFile": "/demo/certs/panoptes.crt", 
      "keyFile": "/demo/certs/panoptes.key", 
      "insecureSkipVerify": true 
    }
}
```

#### Initializing etcd

```
export ETCDCTL_API=3
etcdctl put panoptes/config/devices/
etcdctl put panoptes/config/database/
etcdctl put panoptes/config/producers/
etcdctl put panoptes/config/sensors/
etcdctl put panoptes/config/global {}
```

#### Backup and Restore etcd
```
ETCDCTL_API=3 etcdctl snapshot save etcd_kv_snapshot
```
```
ETCDCTL_API=3 etcdctl snapshot --data-dir=/etcd restore etcd_kv_snapshot
```

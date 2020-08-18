## Panoptes configuration with Consul.
This document will show you how to configure Panoptes with [Consul](http://consul.io) key value store. 

### Configuration specs
The Panoptes configuration categories as follows at Consul key value store:
- [Devices](#devices)
- [Sensors](#sensors)
- [Producers](#producers)
- [Databases](#databases)
- [Global](#global)

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
The producers is a folder which included producers as key value. key is the producer name and value is in json format.
Example producer configuration:

Key: kafka1
Value:
```json
{
    "service": "kafka",
    "config" : {
        "brokers": ["127.0.0.1:9092"],
        "batchSize" : 1000, 
        "topics":["interface","bgp"]
    }
}
```

You can see all available producers config keys at [configuration reference](config_reference.md#producers). 

#### Databases
The databases is a folder which included databases as key value. key is the database name and value is in json format.
Example database configuration:

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

You can see all available databases config keys at [configuration reference](config_reference.md#database). 

#### Global
The global is a key not a folder and it's quite different than other configuration categories as follows:

- [Status](#status) 
- [Discovery](#discovery)
- [Dialout](#dialout)
- [Shard](#shard)
- [Logger](#logger)
- [DeviceOptions](#deviceoptions)


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
Panoptes can register itself to Consul discovery service and it required once you enabled Sharding feature. 

Example discovery configuration:

```json
"discovery": {
   "service": "consul",
   "config": { 
     "address": "127.0.0.1:8500",
     "healthcheckURL": "http://127.0.0.1:8081/healthcheck"
   } 
 }
```

#### Dialout
Panoptes supports gRPC [dial-out](glossary#dialout) mode for Cisco MDT at the moment. you can configure it to listen on a specific address and port that should be reachable from your dial-out mode devices. 

```json
"dialout": {
    "services": {
  		"cisco.mdt": {
     		"addr": "0.0.0.0:50055"
    	}
    }
 } 
```

#### Shard


```json
"shard" : {
   "enabled": true,
   "numberOfNodes": 3,
   "InitializingShard" : 1
 }
```

#### Logger

```json
"logger": {
    "level":"debug", 
    "outputPaths": ["stdout"], 
    "errorOutputPaths":["stderr"]
  }
```

#### Device Options

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


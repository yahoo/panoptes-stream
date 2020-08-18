## Panoptes-Stream with Consul and Grafana 
In the below senario, we're going to setup Panoptes with [Consul](https://www.consul.io/) as configuration management and service discovery. 

### Configuration specs
The Panoptes configuration splitted to below categories at Consul key value store:

#### Devices 
The devices is a folder which included devices as key value. key can be any name and value is in json format.
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

You can see all available device config parameters at [configuration reference](config_reference.md#devices). 

#### Sensors 
The sensors is a folder which included sensors as key value. key is the sensor name and value is in json format.
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

You can see all available sensor config parameters at [configuration reference](config_reference.md#sensors). 

#### Producers
The producers is a folder which included producers as key value. key is the producer name and value is in json format.
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

You can see all available producers config parameters at [configuration reference](config_reference.md#producers). 

#### Databases
The databases is a folder which included databases as key value. key is the database name and value is in json format.
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

You can see all available databases config parameters at [configuration reference](config_reference.md#database). 

#### Global




## schema

- folder: config/producers/
key: producer name
value: json object
example:
``` 
key: config/producers/kafka1 
value: {"service": "kafka", "config" : {"brokers": "127.0.0.1:9092", "batchSize" : 100}}
```

- folder: config/devices/
key: device name
value: json object
example:
``` 
key: config/devices/core1.lax 
value: {"host": "core1.lax", "port": 50051, "sensors" : ["sensor1", "sensor2"]}
```

- folder: config/sensors/
key: sensor name
value: json object
example: 
```
key: config/sensors/sensor1 
value: {"service": "juniper.jti", "path": "/interfaces/interface", "mode": "sample", "interval": 5}
```

- folder: config/global/
key: global variable
value: global value

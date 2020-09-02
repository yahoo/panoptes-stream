
### .yaml Configuration

#### Use Case
- Trying the panoptes
- Small to medium networks 


#### .yaml format (dial-in mode)

Here's an basic example .yaml file that shows the required fields and object spec for collecting a sensor path from a device and print them out on the console every 10 seconds.

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
    service: juniper.gnmi
    output: console::stdout
```

#### execute panoptes binary
```
# panoptes -config etc/config.yaml
```
#### execute panoptes docker
```
# docker run -d --name panoptes -v $PWD:/etc/panoptes panoptes-stream -config /etc/panoptes/config.yaml 
```


#### Add influxdb
```yaml
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

#### Add Kafka
```yaml
producers:
  kafka1:
    service: kafka
    configFile: etc/kafka.yaml
```    

#### etc/kafka.yaml
```yaml
brokers:
  - 127.0.0.1:9092
batchSize: 100
topics:
  - interface
  - bgp
```

#### Customize Logging 

```yaml
logger:
  level: debug
  outputPaths:
    - /var/log/panoptes
  errorOutputPaths:
    - stderr
```    

#### Enable Discovery

```yaml
discovery:
  service: "consul"
  configFile: etc/consul.yaml
```

#### Global Device Options

```yaml
deviceOptions:
  tlsConfig:
    Enabled: false
  username: "admin"
  password: "admin"
```  

#### Status
```yaml
status:
  addr: "0.0.0.0:8081"
```  


#### Watcher and vim editor
If you edit the yaml config with vim editor during running Panoptes, please add the following commands at your ~/.vimrc to prevent stopping Panoptes watcher. 
```
set nobackup
set nowritebackup
``` 




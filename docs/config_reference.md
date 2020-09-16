## Configuration Keys
--------

#### Device 

| key          | description                                             |
|--------------|---------------------------------------------------------|
|host          | IP address or FQDN; it support IPv4 and IPv6.           |
|port          | The telemetry port that configured at device.           | 
|username      | username if authentication is enabled at device.        |
|password      | password if authentication is enabled at device.        |
|tlsConfig     | [TLS configuration](/docs/config_tls.md) parameters.|


#### Sensor  

| key              | description                                                                                             |
|------------------|---------------------------------------------------------------------------------------------------------|
|service           |telemetry name based on the vendor. current supported [services](#telemetry-services).                             |
|output            |the output can be a producer or a database that you already configured.                                  |
|path              |The sensor path describes a YANG path or a subset of data definitions in a YANG model with a container.  |
|mode              |streaming subscription mode: sample or on_change.                                                        |
|sampleInterval    |the data in sample mode must be sent once per sample interval in seconds.                                |
|suppressRedundant |once it enabled the unchanged data in on_change mode sends every heartbeatInterval (vendor must support).|
|heartbeatInterval |specifies the maximum allowable silent period in seconds (vendor must support).                          |
|subscription      |a subscription binds one or more sensor paths (Cisco).                                                   |
|disabled          |disable the sensor.                                                                                      |


#### Producers
| key               | description                                          |
|-------------------|------------------------------------------------------|
| service           | producer name: kafka or nsq               |
| config            |  depends on the producer|


##### Kafka

| key               | description                                          |
|-------------------|------------------------------------------------------|
| brokers           |                |
| topics            |  |
| batchSize
| batchTimeout
| maxAttempts
| queueCapacity
| keepAlive
| compression
| ioTimeout
| tlsConfig



#### Databases
| key               | description                                          |
|-------------------|------------------------------------------------------|
| service           | database name: influxdb               |
| config            |  depends on the database|


##### InfluxDB

| key               | description                                          |
|-------------------|------------------------------------------------------|
| server           |                |
| bucket            |  |
| org
| token
| batchSize
| maxRetries
| timeout


#### Telemetry Services  

| service          | description                                       |
|------------------|---------------------------------------------------|
|cisco.gnmi        | Cisco gNMI plugin                                 |
|cisco.mdt         | Cisco Model-Driven Telemetry plugin               |
|juniper.gnmi      | Juniper gNMI                                      |
|juniper.jti       | Juniper Junos Telemetry Interface plugin          |
|arista.gnmi       | Arista gNMI                                       |


#### Status

| key               | description                                       |
|-------------------|---------------------------------------------------|
|disabled           | disable the status (including healthcheck)        |
|addr               | status ip address and port (ip:port)              |
|tlsConfig          | [TLS configuration](/docs/config_tls.md) parameters.     |

#### Shards

| key               | description                                       |
|-------------------|---------------------------------------------------|
|enabled            |enable shard (sharding of network devices)         |
|initializingShards |minimum number of available nodes required to start|
|numberOfNodes      |maximum number of nodes                            |

#### Discovery
| key               | description                                          |
|-------------------|------------------------------------------------------|
| service           | service discovery name: consul or etcd               |
| config            |  |

#### Pseudo
| key               | description                                          |
|-------------------|------------------------------------------------------|
|instances
|probe
|path
|timeout
|interval
|maxRetry
|tlsConfig 


#### Dialout
| key               | description                                           |
|-------------------|-------------------------------------------------------|
|services           |dial-out service configuration                         |
|defaultOutput      |default output                                         |
|tlsConfig          |[TLS configuration](/docs/config_tls.md) parameters.|


#### Device Options
| key               | description                                           |
|-------------------|-------------------------------------------------------|
|username           |username if authentication is enabled at device.       |
|password           |password if authentication is enabled at device.       |
|tlsConfig          |[TLS configuration](/docs/config_tls.md) parameters.|

#### Global
| key               | description                                          |
|-------------------|------------------------------------------------------| 
|watcherDisabled    |disable watcher (not recommended)                     |
|bufferSize         |shared buffer between telemetries                     |
|outputBufferSize   |output buffer (per producer or database)              |

#### TLS   

| key               | description                                       |
|-------------------|---------------------------------------------------|
|enabled| enable or disable the TLS configuration.|
|caFile| if caFile is empty, Panoptes uses the host's root CA set.|
|certFile| the certficate file contain PEM encoded data.
|keyFile| the private key file contain PEM encoded data.
|insecureSkipVerify|it controls whether a client verifies the server's certificate chain and host name.|

---

#### Dialout cisco.mdt

| key               | description                                       |
|-------------------|-|
|addr| server ip address and port (ip:port)|
|workers| number of workers|

## Configuration Keys
--------

#### Device 

| key          | description                                             |
|--------------|---------------------------------------------------------|
|host          | IP address or FQDN; it support IPv4 and IPv6.           |
|port          | the telemetry port that configured at device.           | 
|username      | username if authentication is enabled at device.        |
|password      | password if authentication is enabled at device.        |
|timeout       | timeout for dialing a gRPC connection (unit is second).  |
|tlsConfig     | [TLS configuration](/docs/config_tls.md) parameters.|


#### Sensor  

| key              | description                                                                                             |
|------------------|---------------------------------------------------------------------------------------------------------|
|service           |telemetry name based on the vendor. current supported [services](#telemetry-services).                   |
|output            |the output can be a producer or a database that you already configured.                                  |
|path              |The sensor path describes a YANG path or a subset of data definitions in a YANG model with a container.  |
|mode              |streaming subscription mode: sample or on_change.                                                        |
|sampleInterval    |the data in sample mode must be sent once per sample interval in seconds.                                |
|suppressRedundant |once it enabled the unchanged data sends every heartbeatInterval in on_change mode (vendor must support).|
|heartbeatInterval |specifies the maximum allowable silent period in seconds (vendor must support).                          |
|subscription      |a subscription binds one or more sensor paths (Cisco).                                                   |
|disabled          |disable the sensor.                                                                                      |


#### Producer
| key               | description                                          |
|-------------------|------------------------------------------------------|
| service           | producer name: kafka or nsq               |
| config            |  depends on the producer|


##### Kafka

| key               | description                                          |
|-------------------|------------------------------------------------------|
| brokers           |list of brokers |
| topics            |list of topics|
| batchSize         |size of batch|
| batchTimeout      |flush at least every batchTimeout|
| maxAttempts       |limit of how many attempts will be made before delivering the error|
| keepAlive         |keep-alive period for an active network|
| compression       |compression codec to compress Kafka messages (gzip, snappy, lz4)|
| tlsConfig         |[TLS configuration](/docs/config_tls.md) parameters.|


##### NSQ
| key               | description                                          |
|-------------------|------------------------------------------------------|
| addr              |
| topics            |list of topics|
| batchSize         |size of batch|
| batchTimeout      |flush at least every batchTimeout|


#### Database
| key               | description                                          |
|-------------------|------------------------------------------------------|
| service           | database name: influxdb               |
| config            | depends on the database|


##### InfluxDB

| key               | description                                          |
|-------------------|------------------------------------------------------|
| server            |server url               |
| bucket            |name of the location where time series data is stored|
| org|organization name|
| token|authentication token
| batchSize|size of batch
| maxRetries|maximum count of retry attempts of failed writes
| timeout|HTTP request timeout|


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
|minimumShards      |minimum number of available nodes required         |
|numberOfNodes      |maximum number of nodes                            |

#### Discovery
| key               | description                                          |
|-------------------|------------------------------------------------------|
| service           | service discovery name: consul, etcd or pseudo       |
| config            | depends on the [discovery](discovery.md) |

#### Pseudo
| key               | description                                          |
|-------------------|------------------------------------------------------|
|instances|list of instance's IP addresses|
|probe| http or https
|path| healthcheck url path
|timeout| prob timeout
|interval|healthcheck time interval
|maxRetry| healthcheck maximum number of retries|
|tlsConfig|[TLS configuration](/docs/config_tls.md) parameters. 


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
|timeout            |timeout for dialing a gRPC connection (unit is second).|
|tlsConfig          |[TLS configuration](/docs/config_tls.md) parameters.   |

#### Global
| key               | description                                          |
|-------------------|------------------------------------------------------| 
|watcherDisabled    |disable watcher and switch to sighup mode             |
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

#### Dialout cisco.mdt

| key               | description                                       |
|-------------------|-|
|addr| server ip address and port (ip:port)|
|workers| number of workers|

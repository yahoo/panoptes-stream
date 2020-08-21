
### Device 

| key          | description                                             |
|--------------|---------------------------------------------------------|
|host          | IP address or FQDN; it support IPv4 and IPv6.           |
|port          | The telemetry port that configured at device.           | 
|username      | username if authentication is enabled at device.        |
|password      | password if authentication is enabled at device.        |
|tlsConfig     | TLS configuration parameters. check [TLS config](#tls). |


### Sensor  

| key              | description                                                                                             |
|------------------|---------------------------------------------------------------------------------------------------------|
|service           |telemetry name based on the vendor. currect supported [services](#services).                             |
|output            |the output can be a producer or a database that you already configured.                                  |
|path              |The sensor path describes a YANG path or a subset of data definitions in a YANG model with a container.  |
|mode              |streaming subscription mode: sample or on_change.                                                        |
|sampleInterval    |the data in sample mode must be sent once per sample interval in seconds.                                |
|suppressRedundant |once it enabled the unchanged data in on_change mode sends every heartbeatInterval (vendor must support).|
|heartbeatInterval |specifies the maximum allowable silent period in seconds (vendor must support).                          |
|subscription      |a subscription binds one or more sensor paths (Cisco).                                                   |
|disabled          |disable the sensor.                                                                                      |


#### Telemetry Services  

| service          | description                                       |
|------------------|---------------------------------------------------|
|cisco.gnmi        | Cisco gNMI plugin                                 |
|cisco.mdt         | Cisco Model-Driven Telemetry plugin               |
|juniper.gnmi      | Juniper gNMI                                      |
|juniper.jti       | Juniper Junos Telemetry Interface plugin          |
|arista.gnmi       | Arista gNMI                                       |


#### TLS   

| key               | description                                       |
|-------------------|---------------------------------------------------|
|enabled            | enable the TLS                                    |    
|certFile           | certificate file                                  |
|keyFile            | private key file                                  |
|caFile             | certificate authority certification               |
|insecureSkipVerify |it controls whether a client verifies the server's certificate chain and host name |

#### Status keys

| key               | description                                       |
|-------------------|---------------------------------------------------|
|disabled           | disable the status (including healthcheck)        |
|addr               | status ip address and port (ip:port)              |
|tlsConfig          | TLS configuration                                 |

#### Shard keys

| key               | description                                       |
|-------------------|---------------------------------------------------|
|enabled            | 
|initializingShard  |
|numberOfNodes      |

#### Discovery
| key               | description                                          |
|-------------------|------------------------------------------------------|
| service           | service discovery name: consul or etcd               |
| configFile        | it's valid once configuration management is yaml     |
| config            | it's valid once configuration management is not yaml |

#### Dialout
| key               | description                                          |
|-------------------|------------------------------------------------------|
|services           |
|defaultOutput      |
|tlsConfig          |


#### Device Options
| key               | description                                          |
|-------------------|------------------------------------------------------|
|username           |
|password           |
|tlsConfig          |
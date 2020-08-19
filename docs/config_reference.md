
### Device 

| key          | description                                             |
|--------------|---------------------------------------------------------|
|host          | IP address or FQDN; it support IPv4 and IPv6            |
|port          | The telemetry port that configured at device            | 
|username      | username if authentication is enabled at device         |
|password      | password if authentication is enabled at device         |
|tlsConfig     | TLS configuration parameters. check [TLS config](#tls)  |


### Sensor  

| key              | description                                                                                            |
|------------------|--------------------------------------------------------------------------------------------------------|
|Service           |telemetry name based on the vendor. currect supported [services](#services).                            |
|Output            |the output can be a producer or a database that you already configured.                                 |
|Path              |The sensor path describes a YANG path or a subset of data definitions in a YANG model with a container  |
|Mode              |streaming subscription mode: sample or on_change                                                        |
|SampleInterval    |the data in sample mode must be sent once per sample interval in seconds                                | 
|SuppressRedundant |once it enabled the unchanged data in on_change mode sends every heartbeatInterval (vendor must support)|
|HeartbeatInterval |specifies the maximum allowable silent period in seconds (vendor must support)                          |
|Subscription      |a subscription binds one or more sensor paths (Cisco)                                                   |
|Disabled          |disable the sensor                                                                                      |


#### Telemetry Services  

| service          | description                                       |
|------------------|---------------------------------------------------|
|cisco.gnmi        | Cisco gNMI plugin                                 |
|cisco.mdt         | Cisco Model-Driven Telemetry plugin               |
|juniper.gnmi      | Juniper gNMI                                      |
|juniper.jti       | Juniper Junos Telemetry Interface plugin          |
|arista.gnmi       | Arista gNMI                                       |


#### TLS (client)  

| key               | description                                       |
|-------------------|---------------------------------------------------|
|Enabled            | enable the TLS                                    |    
|CertFile           | certificate file                                  |
|KeyFile            | private key file                                  |
|CAFile             | certificate authority certification               |
|InsecureSkipVerify |it controls whether a client verifies the server's certificate chain and host name |

#### Status keys

#### Shard keys

#### Discovery

#### Dialout

#### Device Options

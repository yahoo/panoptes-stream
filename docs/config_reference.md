
### Device 

| key          | description                                             |
|--------------|---------------------------------------------------------|
|host          | IP address or FQDN; it support IPv4 and IPv6            |
|port          | The telemetry port that configured at device            | 
|username      | username if authentication is enabled at device         |
|password      | password if authentication is enabled at device         |
|tlsConfig     | TLS configuration parameters. check [TLS config](#tls)  |


### Sensor  

| key              | description                                                                 |
|------------------|-----------------------------------------------------------------------------|
|Service           |telemetry name based on the vendor. currect supported [services](#services). |
|Output            |the output can be a producer or a database that you already configured.      |
|Path              |
|Mode              |
|SampleInterval    |
|HeartbeatInterval |
|SuppressRedundant |
|Subscription      |
|Disabled          |


#### Telemetry Services  

| service          | description                                       |
|------------------|---------------------------------------------------|
|cisco.gnmi        | 
|cisco.mdt         |
|juniper.gnmi      |
|juniper.jti       |
|arista.gnmi       |


#### TLS   

| key               | description                                       |
|-------------------|---------------------------------------------------|
|Enabled            |    
|CertFile           | 
|KeyFile            |
|CAFile             |
|InsecureSkipVerify |

#### Status keys

#### Shard keys

#### Discovery

#### Dialout

#### Device Options

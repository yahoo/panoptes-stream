## Service Discovery Reference
---
The Panoptes supports the following service discoveries:

- Consul
- etcd
- Kubernetes / k8s
- Pseudo

The service discovery is required to enable shards / cluster service and you need to select one of them depending on the size of your nodes and your infrastructure. These four options cover almost all the use cases. The Pseudo is the last option for those who don't want to use another 3rd party software like Consul for service discovery but they need to enable shards. Pseudo is a simple health checking system without registration, instead it gets the node's information through configuration. It is recommended for small shards / cluster size.


### Consul

|| |
|-|-|
|Address| The address and port of the Consul agent        |
|Prefix| The distributed locking prefix at Consul|
|HealthcheckURL| The health check url | 
|TLSConfig | [TLS configuration](/docs/config_tls.md).|

Configuration type: consul or etcd
Key: global
Value:
```json
{
  "discovery": { 
    "service": "consul", 
    "config": { 
      "address": "127.0.0.1:8500", 
      "healthcheckURL": "http://127.0.0.1:8081/healthcheck"
    } 
  }
}
```
Configuration type: yaml
```yaml
discovery:
  service: consul
    config:
      address: 127.0.0.1:8500
      healthcheckURL: http://127.0.0.1:8081/healthcheck 

```

### etcd

|| |
|-|-|
|Endpoints| The addresses and ports of the Consul agent        | 
|Prefix| The distributed locking prefix at Consul|
|TLSConfig | [TLS configuration](/docs/config_tls.md).|

Configuration type: consul or etcd
Configuration type: consul or etcd
Key: global
Value:
```json
{
  "discovery": { 
    "service": "consul", 
    "config": { 
      "endpoints": ["127.0.0.1:2379"],
      "healthcheckURL": "http://127.0.0.1:8081/healthcheck"
    } 
  }
}
```

Configuration type: yaml
```yaml
discovery:
  service: consul
  config:
    endpoints:
      - 127.0.0.1:2379
    healthcheckURL: "http://127.0.0.1:8081/healthcheck"  


```

### Kubernetes / k8s

|| |
|-|-|
|Namespace|The namespace that the Panoptes's pods  deployed|

Configuration type: yaml / configmap
```yaml
discovery:
  service: k8s
    config:
       namespace: panoptes 
```

### Pseudo

|| |
|-|-|
|Instances| The addresses and ports of the Panoptes nodes| 
|Probe|The liveness probe to know Panoptes's node health|     
|Path | The path to access on the HTTP/HTTPS server|      
|Timeout|the probe timeout in second|   
|Interval| the probe interval in second|  
|TLSConfig| [TLS configuration](/docs/config_tls.md).|


Configuration type: consul or etcd
Key: global
Value:
```json
"discovery": {
    "service": "pseudo",
    "config": {
        "instances" : ["panoptes-node1:8081","panoptes-node1:8081"],
        "probe": "http",
        "path": "/healthcheck"
    }
}
```

Configuration type: yaml
```yaml
discovery:
  service: pseudo
  config:
    instances:
      - panoptes-node1:8081
      - panoptes-node2:8081
    probe: http
    path: /healthcheck
```

## TLS configuration

|key                 | description                                                                         |
|--------------------|-------------------------------------------------------------------------------------|
|enabled             | Enable or disable the TLS configuration.                                            |
|caFile              | If caFile is empty, Panoptes uses the host's root CA set.                           |
|certFile            | The certificate file contain PEM encoded data.                                      |
|keyFile             | The private key file contain PEM encoded data.                                      |
|insecureSkipVerify  | It controls whether a client verifies the server's certificate chain and host name. |


JSON
```json
"tlsConfig": { 
    "enabled": true,
    "caFile": "/etc/panoptes/certs/ca.pem", 
    "certFile": "/etc/panoptes/cert.pem", 
    "keyFile": "/etc/panoptes/key.pem", 
    "insecureSkipVerify": true 
}
```  

YAML
```yaml
tlsConfig:
    enabled": true 
    caFile: /etc/panoptes/certs/ca.pem
    certFile": /etc/panoptes/cert.pem
    keyFile": /etc/panoptes/key.pem 
    insecureSkipVerify": true 
```       

##### Generate self-signed TLS Certificates by [cfssl](https://github.com/cloudflare/cfssl) and [cfssljson](https://github.com/cloudflare/cfssl)

```
cfssl gencert -initca ca-csr.json | cfssljson -bare ca
```

```
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -profile=default \
  panoptes-csr.json | cfssljson -bare panoptes
```

<details><summary>click here to see the ca-csr.json</summary>
<p>

```json
{
  "hosts": [
    "localhost"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Los Angeles",
      "O": "Panoptes",
      "OU": "CA",
      "ST": "California"
    }
  ]
}
```
</p>
</details>


<details><summary>click here to see the ca-config.json</summary>
<p>

```json
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "default": {
        "usages": ["signing", "key encipherment", "server auth", "client auth"],
        "expiry": "8760h"
      }
    }
  }
}
```
</p>
</details>


<details><summary>click here to see the panoptes-csr.json</summary>
<p>

```json
{
  "CN": "localhost",
  "hosts": [
    "localhost",
    "127.0.0.1"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Los Angeles",
      "O": "Panoptes Labs",
      "OU": "Panoptes",
      "ST": "California"
    }
  ]
}
```
</p>
</details>



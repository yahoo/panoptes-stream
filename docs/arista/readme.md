### Arista gNMI

#### configuration with TLS 

```
management api gnmi
   transport grpc panoptes
      ssl profile telemetry
      port 50051
   provider eos-native
```

```
management security
   ssl profile telemetry
      certificate telemetry.crt key telemetry.key
```

#### configuration without TLS

```
management api gnmi
   transport grpc panoptes
      port 50051
   provider eos-native
```

for more information: https://eos.arista.com/openconfig-4-20-2-1f-release-notes/





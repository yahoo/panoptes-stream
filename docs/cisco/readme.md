### Cisco gNMI and model-driven telemetry

#### MDT - Dialin (without TLS)

```
Router# configure
Router (config)# grpc
Router (config-grpc)# port 50051
Router (config-grpc)# no-tls
```

```
Router(config)#telemetry model-driven
Router(config-model-driven)#sensor-group SubGroup1
Router(config-model-driven-snsr-grp)# sensor-path openconfig-interfaces:interfaces/interface
Router(config-model-driven-snsr-grp)# commit
```

```
Router(config)#telemetry model-driven  
Router(config-model-driven)#subscription Sub1  
Router(config-model-driven-subs)#sensor-group-id SubGroup1 sample-interval 30000  
Router(config-mdt-subscription)# commit
```

#### validation show commands
```
show grpc
show telemetry model-driven subscription Sub1
show run telemetry model-driven
```

#### MDT - Dialout (without TLS)

```
Router(config)#telemetry model-driven
Router(config-model-driven)#destination-group panoptes
Router(config-model-driven-dest)#address family ipv4 192.168.0.1 port 57500
Router(config-model-driven-dest-addr)#encoding self-describing-gpb
Router(config-model-driven-dest-addr)#protocol grpc no-tls
Router(config-model-driven-dest-addr)#commit
```

```
Router(config)#telemetry model-driven
Router(config-model-driven)#sensor-group SubGroup1
Router(config-model-driven-snsr-grp)# sensor-path openconfig-interfaces:interfaces/interface
Router(config-model-driven-snsr-grp)# commit
```

```
Router(config)#telemetry model-driven  
Router(config-model-driven)#subscription Sub1  
Router(config-model-driven-subs)#sensor-group-id SubGroup1 sample-interval 30000  
Router(config-model-driven-subs)#destination-id panoptes 
Router(config-mdt-subscription)# commit
```

for more information: https://www.cisco.com/c/en/us/td/docs/iosxr/asr9000/telemetry/b-telemetry-cg-asr9000-61x/b-telemetry-cg-asr9000-61x_chapter_010.html
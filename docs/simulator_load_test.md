
### Binary

#### Start consul
```
consul agent -dev
```

#### initialize 100 fake devices 
```
sudo ./simulator_config.sh add 100
```

#### run simulator
```
simulator
```

#### run panoptes
```
panoptes -consul -
```

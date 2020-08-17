### .yaml spec

#### Required Fields
In the .yaml file for the Panoptes configuration, you want to create, you'll need to set values for the following fields:

- devices
    * host - host ip address or FQDN of the device  
    * port - port of the device
    * sensors - list of the sensors

- sensors (depends on the sensor mode)
    * path
    * mode
    * service
    * output

- databases:
    * service
    * configFile
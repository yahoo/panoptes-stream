### Single node with InfluxDB demonstration
---

![panoptes demo](imgs/demo.png)

This is a demo with Panoptes gNMI simulator. you can see the simulated data same as real devices in the grafana dashboard.

##### Checkout GitHub
```
git checkout github.com/yahoo/panoptes-stream
cd panoptes-stream
```
##### Start the containers 
```console
# docker-compose up -d
```
![panoptes demo](imgs/demo_dc_shell.png)
browse http://localhost:3000 and login to grafana by user: panoptes pass: panoptes and select Panoptes Demo dashboard

![panoptes demo](imgs/demo_dc_grafana.png)

##### Clean up
```console
# docker-compose down
```

 <span style="color:purple">All demonstrations</span>
Please check out the [demo page](demo_list.md) to see all of the demonstrations for different scenarios. 



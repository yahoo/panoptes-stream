#!/bin/sh

### usage:
### add number_of_simulate-devices 
### del number_of_simulate-devices

add_data()
{
    for i in $(eval echo {1..$1})
    do
   value=$(cat <<EOF
      {
         "host": "simulate-device$i",
         "port": 50051,
         "service" : "juniper.gnmi",
         "sensors":["simulate-sensor"]
      }
EOF
    )
    
    echo "${value}"
    
    echo `consul kv put panoptes/config/devices/simulate-device$i "$value"`
    sed -i '' "/simulate-device$i/d" /etc/hosts
    echo "127.0.0.1 simulate-device$i" >> /etc/hosts
done

   value=$(cat <<EOF
      {
         "service" : "juniper.gnmi",
         "path": "/interfaces/interface/state/counters/",
         "mode": "sample",
         "sampleInterval": 10,
         "output":"console::stdout"
      }
EOF
    )
    echo `consul kv put panoptes/config/sensors/simulate-sensor "$value"`

}

del_data()
{
    for i in $(eval echo {1..$1})
    do
        echo `consul kv delete panoptes/config/devices/simulate-device$i`
        sed -i '' "/simulate-device$i/d" /etc/hosts
    done

    echo `consul kv delete panoptes/config/sensors/simulate-sensor`
}

if [ "$1" == 'add' ]
then
    add_data $2
elif [ "$1" == 'del' ]
then
    del_data $2
fi


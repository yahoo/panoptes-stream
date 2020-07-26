package yaml

import (
	"testing"

	yml "gopkg.in/yaml.v3"
)

var yamlContent = `
devices: 
  - host: "192.168.59.3"
    password: admin
    port: 50051
    sensors: 
      - sensor1
sensors: 
  sensor1: 
    mode: on_change
    output: "console::stdout"
    path: /interfaces/interface/state
    service: juniper.gnmi
`

func TestConfigDevices(t *testing.T) {
	yamlCfg := &yamlConfig{}
	yml.Unmarshal([]byte(yamlContent), yamlCfg)
	devices := configDevices(yamlCfg)
	if len(devices) < 1 {
		t.Error("expect to have a device but return ", len(devices))
	}
}

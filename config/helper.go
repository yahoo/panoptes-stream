package config

import "encoding/json"

type DeviceTemplate struct {
	DeviceConfig `yaml:",inline"`

	Sensors []string
}

func ConvDeviceTemplate(d DeviceTemplate) Device {
	device := Device{}
	b, _ := json.Marshal(&d)
	json.Unmarshal(b, &device)
	device.Sensors = nil
	return device
}

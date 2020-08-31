//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package discovery

// Discovery represents discovery interface
type Discovery interface {
	Register() error
	Deregister() error
	GetInstances() ([]Instance, error)
	Watch(chan<- struct{})
}

// Instance represents discovered node information
type Instance struct {
	ID      string
	Meta    map[string]string
	Address string
	Status  string
}

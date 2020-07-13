package discovery

type Discovery interface {
	Register() error
	Deregister() error
	GetInstances() ([]Instance, error)
	Watch(chan<- struct{})
}

type Instance struct {
	ID      string
	Meta    map[string]string
	Address string
	Status  string
}

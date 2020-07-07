package discovery

type Discovery interface {
	Register()
	Deregister()
	GetInstances() []Instance
	Watch(string, chan<- struct{})
}

type Instance struct {
	ID      string
	Meta    map[string]string
	Address string
	Status  string
}

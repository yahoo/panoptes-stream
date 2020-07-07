package discovery

type Discovery interface {
	Register()
	Deregister()
	GetInstances() []Instance
}

type Instance struct {
	ID      string
	Meta    map[string]string
	Address string
	Status  string
}

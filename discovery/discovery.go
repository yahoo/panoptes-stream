package discovery

type Discovery interface {
	Register()
	Deregister()
}

package ext

const (
	InterlockExtNameLabel  = "interlock.ext.name"
	InterlockHostnameLabel = "interlock.hostname"
	InterlockDomainLabel   = "interlock.domain"
)

type LoadBalancer interface {
	Reload() error
	Update() error
}

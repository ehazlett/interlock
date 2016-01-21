package ext

const (
	InterlockExtNameLabel = "interlock.ext.name"
)

type LoadBalancer interface {
	Reload() error
	Update() error
}

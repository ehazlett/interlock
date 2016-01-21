package ext

const (
	InterlockExtNameLabel           = "interlock.ext.name"
	InterlockHostnameLabel          = "interlock.hostname"
	InterlockDomainLabel            = "interlock.domain"
	InterlockSSLLabel               = "interlock.ssl"
	InterlockSSLOnlyLabel           = "interlock.sslonly"
	InterlockSSLBackendLabel        = "interlock.sslbackend"
	InterlockSSLCertLabel           = "interlock.sslcert"
	InterlockSSLCertKeyLabel        = "interlock.sslcertkey"
	InterlockPortLabel              = "interlock.port"
	InterlockWebsocketEndpointLabel = "interlock.websocket_endpoint"
	InterlockAliasDomainLabel       = "interlock.alias_domain"
)

type LoadBalancer interface {
	Reload() error
	Update() error
}

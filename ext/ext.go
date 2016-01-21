package ext

const (
	InterlockExtNameLabel             = "interlock.ext.name"
	InterlockHostnameLabel            = "interlock.hostname"
	InterlockDomainLabel              = "interlock.domain"
	InterlockSSLLabel                 = "interlock.ssl"
	InterlockSSLOnlyLabel             = "interlock.ssl_only"
	InterlockSSLBackendLabel          = "interlock.ssl_backend"
	InterlockSSLBackendTLSVerifyLabel = "interlock.ssl_backend_tls_verify"
	InterlockSSLCertLabel             = "interlock.ssl_cert"
	InterlockSSLCertKeyLabel          = "interlock.ssl_cert_key"
	InterlockPortLabel                = "interlock.port"
	InterlockWebsocketEndpointLabel   = "interlock.websocket_endpoint"
	InterlockAliasDomainLabel         = "interlock.alias_domain"
	InterlockHealthCheckLabel         = "interlock.health_check"
	InterlockHealthCheckIntervalLabel = "interlock.health_check_interval"
	InterlockBalanceAlgorithmLabel    = "interlock.balance_algorithm"
	InterlockBackendOptionLabel       = "interlock.backend_option"
)

type LoadBalancer interface {
	Reload() error
	Update() error
}

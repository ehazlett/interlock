package ext

import (
	"github.com/samalba/dockerclient"
)

const (
	InterlockExtNameLabel             = "interlock.ext.name"               // common
	InterlockHostnameLabel            = "interlock.hostname"               // haproxy, nginx
	InterlockDomainLabel              = "interlock.domain"                 // haproxy, nginx
	InterlockSSLLabel                 = "interlock.ssl"                    // nginx
	InterlockSSLOnlyLabel             = "interlock.ssl_only"               // haproxy, nginx
	InterlockSSLBackendLabel          = "interlock.ssl_backend"            // haproxy, nginx
	InterlockSSLBackendTLSVerifyLabel = "interlock.ssl_backend_tls_verify" // haproxy, nginx
	InterlockSSLCertLabel             = "interlock.ssl_cert"               // nginx
	InterlockSSLCertKeyLabel          = "interlock.ssl_cert_key"           // nginx
	InterlockPortLabel                = "interlock.port"                   // haproxy, nginx
	InterlockWebsocketEndpointLabel   = "interlock.websocket_endpoint"     // nginx
	InterlockAliasDomainLabel         = "interlock.alias_domain"           // haproxy, nginx
	InterlockHealthCheckLabel         = "interlock.health_check"           // haproxy
	InterlockHealthCheckIntervalLabel = "interlock.health_check_interval"  //haproxy
	InterlockBalanceAlgorithmLabel    = "interlock.balance_algorithm"      // haproxy
	InterlockBackendOptionLabel       = "interlock.backend_option"         // haproxy
)

type Extension interface {
	HandleEvent(event *dockerclient.Event) error
	Update() error
	Reload() error
}

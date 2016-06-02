package ext

import (
	"github.com/samalba/dockerclient"
)

const (
	InterlockAppLabel                 = "interlock.app"                    // internal
	InterlockExtNameLabel             = "interlock.ext.name"               // common
	InterlockExtServiceNameLabel      = "interlock.ext.service_name"       // common
	InterlockHostnameLabel            = "interlock.hostname"               // haproxy, nginx
	InterlockNetworkLabel             = "interlock.network"                // common
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
	InterlockHealthCheckIntervalLabel = "interlock.health_check_interval"  // haproxy
	InterlockBalanceAlgorithmLabel    = "interlock.balance_algorithm"      // haproxy
	InterlockBackendOptionLabel       = "interlock.backend_option"         // haproxy
	InterlockIPHashLabel              = "interlock.ip_hash"                // nginx
	InterlockContextRootLabel         = "interlock.context_root"           // haproxy, nginx
	InterlockContextRootRewriteLabel  = "interlock.context_root_rewrite"   // haproxy, nginx
)

type Extension interface {
	Name() string
	HandleEvent(event *dockerclient.Event) error
}

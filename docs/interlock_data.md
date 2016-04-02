# Interlock Override Data
Interlock supports overriding and customizing specific configuration values
at container runtime.  This enables customization like specifying additional
alias domains or SSL certificates.

These are the options available to customize as well as their extension
compatibility.

|Label|Extensions Supported|
|----|----|
|`interlock.ext.name`               | internal |
|`interlock.hostname`               | haproxy, nginx|
|`interlock.domain`                 | haproxy, nginx|
|`interlock.ssl`                    | nginx|
|`interlock.ssl_only`               | haproxy, nginx|
|`interlock.ssl_backend`            | haproxy, nginx|
|`interlock.ssl_backend_tls_verify` | haproxy, nginx|
|`interlock.ssl_cert`               | nginx|
|`interlock.ssl_cert_key`           | nginx|
|`interlock.port`                   | haproxy, nginx|
|`interlock.context_root`           | haproxy, nginx|
|`interlock.websocket_endpoint`     | nginx|
|`interlock.alias_domain`           | haproxy, nginx|
|`interlock.health_check`           | haproxy|
|`interlock.health_check_interval`  | haproxy|
|`interlock.balance_algorithm`      | haproxy|
|`interlock.backend_option`         | haproxy|

# Port
If an upstream container uses multiple ports you can select the port for 
the proxy to use by specifying the following label: `interlock.port=8080`.
This will cause the proxy container to use port `8080` when sending requests
to the upstream containers.

# Alias Domains
You can specify alias domains to enable the same set of upstream containers
to serve multiple domains.  To specify an alias domain, specify a label such as
`interlock.alias_domain=foo.com`.  You can specify multiple by using the
following syntax: `interlock.alias_domain.0=foo.local`.

# Context Root
Interlock supports specifying a context root instead of using a hostname.
Specify a label such as `interlock.context_root=/myapp`.  The upstreams
will be configured to serve under the context instead of the hostname and
domain.  The proxy will also rewrite requests so they appear from the root.

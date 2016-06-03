# Interlock Override Data
Interlock supports overriding and customizing specific configuration values
at container runtime.  This enables customization like specifying additional
alias domains or SSL certificates.

These are the options available to customize as well as their extension
compatibility.

|Label|Extensions Supported|Description|
|----|----|-----|
|`interlock.ext.name`               | internal | |
|`interlock.hostname`               | haproxy, nginx| service hostname |
|`interlock.domain`                 | haproxy, nginx| service domain |
|`interlock.network`                | haproxy, nginx| docker network to join and use (overlay) |
|`interlock.ssl`                    | nginx| enable ssl |
|`interlock.ssl_only`               | haproxy, nginx| add a redirect to the ssl service |
|`interlock.ssl_backend`            | haproxy, nginx| use ssl for the service backend |
|`interlock.ssl_backend_tls_verify` | haproxy, nginx| verify tls for the service backend |
|`interlock.ssl_cert`               | nginx| name of the ssl certificate |
|`interlock.ssl_cert_key`           | nginx| name of the ssl key |
|`interlock.port`                   | haproxy, nginx| container port to use as the upstream |
|`interlock.context_root`           | haproxy, nginx| context path to use for upstreams |
|`interlock.context_root_rewrite`   | haproxy, nginx| rewrite requests before sending to upstream |
|`interlock.websocket_endpoint`     | nginx| endpoint to use for websocket support |
|`interlock.alias_domain`           | haproxy, nginx| one or more alias domains  for the upstream (i.e. www.example.com and example.com) |
|`interlock.health_check`           | haproxy| haproxy health check for backend |
|`interlock.health_check_interval`  | haproxy| interval to use for backend health check|
|`interlock.balance_algorithm`      | haproxy| load balancing algorithm to use in haproxy|
|`interlock.backend_option`         | haproxy| one or more backend options as specified by haproxy|

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

By default no rewrite rules will be added.  You can enable rewrites to be added
by adding the label `interlock.context_root_rewrite=true`.  This will cause
requests to be rewritten before being sent to the application.  For example,
if you use a context of `/myapp` and you have rewrite enabled, requests to
`/myapp/foo` will be rewritten as `/foo`.

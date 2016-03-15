# Interlock Override Data
Interlock supports overriding and customizing specific configuration values
at container runtime.  This enables customization like specifying additional
alias domains or SSL certificates.

These are the options available to customize as well as their extension
compatibility.

|Label|Extensions Supported|
|----|----|
|`interlock.ext.name`               | common|
|`interlock.hostname`               | haproxy, nginx|
|`interlock.domain`                 | haproxy, nginx|
|`interlock.ssl`                    | nginx|
|`interlock.ssl_only`               | haproxy, nginx|
|`interlock.ssl_backend`            | haproxy, nginx|
|`interlock.ssl_backend_tls_verify` | haproxy, nginx|
|`interlock.ssl_cert`               | nginx|
|`interlock.ssl_cert_key`           | nginx|
|`interlock.port`                   | haproxy, nginx|
|`interlock.websocket_endpoint`     | nginx|
|`interlock.alias_domain`           | haproxy, nginx|
|`interlock.health_check`           | haproxy|
|`interlock.health_check_interval`  | haproxy|
|`interlock.balance_algorithm`      | haproxy|
|`interlock.backend_option`         | haproxy|
|`interlock.protocol`               | haproxy|


# Interlock
Dynamic, event-driven Docker plugin system using [Swarm](https://github.com/docker/swarm).

# Usage

`docker run -p 80:8080 -d ehazlett/interlock -swarm tcp://1.2.3.4:2375`

If you're not running Swarm, you'll also need to tell Interlock the IP address of the backends, with `-proxy-backend-override-address <ip-addr>`.

If you want SSL support, enter a path to the cert (probably want a mounted volume) and then expose 443:

`docker run -p 80:8080 -p 443:8443 -d -v /etc/ssl:/ssl ehazlett/interlock -swarm tcp://1.2.3.4:2375 -ssl-cert=/etc/ssl/cert.pem`

* You should then be able to access `http://<your-host-ip>/haproxy?stats` to see the proxy stats.
* Add some CNAMEs or /etc/host entries for your IP.  Interlock uses the `hostname` in the container config to add backends to the proxy.

# Commandline options

* `swarm`: url to swarm (default: tcp://127.0.0.1:2375)
* `config`: path to config file
* `proxy-conf-path`: path to proxy file (will be generated and created)
* `proxy-pid-path`: path to proxy pid file
* `proxy-backend-override-address`: force proxy to use this address in all backends
* `syslog`: address to syslog
* `proxy-port`: proxy listen port. Default: 8080
* `ssl-cert`: path to single ssl certificate or directory (for SNI). This enables SSL in proxy configuration
* `ssl-port`: ssl listen port (must have cert above). Default: 8443
* `ssl-opts`: string of SSL options (eg. ciphers or ssl, tls versions)
* `tlscacert`: TLS ca certificate to use with swarm (optional)
* `tlscert`: TLS certificate to use with swarm (optional)
* `tlskey`: TLS key to use with swarm (options)
* `version`: show version and exit

Example for SNI (multidomain) https:

```
docker run -it -p 80:8080 -p 443:8443 -d -v /etc/ssl:/etc/ssl ehazlett/interlock \
    -ssl-cert /etc/ssl/default.crt \
    -ssl-opts 'crt /etc/ssl no-sslv3 ciphers EECDH+ECDSA+AESGCM:EECDH+aRSA+AESGCM:EECDH+ECDSA+SHA384:EECDH+ECDSA+SHA256:EECDH+aRSA+SHA384:EECDH+aRSA+SHA256:EECDH+aRSA+RC4:EECDH:EDH+aRSA:RC4:!aNULL:!eNULL:!LOW:!3DES:!MD5:!EXP:!PSK:!SRP:!DSS'
```

In this example HAProxy will use concatinated certificates from /etc/ssl/<hostname>.crt for SNI requests, falling back to /etc/ssl/default.crt.  It also specifies secure openssl ciphers and disables SSLv3 support (POODLE attack vulnerability)



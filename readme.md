# Interlock
Dynamic, event-driven HAProxy using [Swarm](https://github.com/docker/swarm).  Interlock configures backends for HAProxy by listening to Docker events (start/stop, etc).

Note: Interlock requires HAProxy 1.5+

# Usage
To get started, generate a config (save this as `/tmp/controller.conf`):

Replace `1.2.3.4` in `addr` listed in the engines section with your IP for your Docker host.  You must enable TCP in the Docker daemon (see the [Docker Docs](http://docs.docker.com/reference/commandline/cli/) for details).

Note: To enable SSL, enter a valid path for the certificate.

```
{
    "swarm_url": "tcp://127.0.0.1:2375"
    "port": 8080,
    "connect_timeout": 5000,
    "server_timeout": 10000,
    "client_timeout": 10000,
    "max_conn": 2048,
    "sysload_addr": "",
    "proxy_config_path": "/tmp/proxy.conf",
    "stats_user": "stats",
    "stats_password": "stats",
    "pid_path": "/tmp/proxy.pid",
    "ssl_cert": "/path/to/cert.pem",
    "ssl_port": 8443,
    "ssl_opts": "no-sslv3 ciphers EECDH+ECDSA+AESGCM:EECDH+aRSA+AESGCM:EECDH+ECDSA+SHA384:EECDH+ECDSA+SHA256:EECDH+aRSA+SHA384:EECDH+aRSA+SHA256:EECDH+aRSA+RC4:EECDH:EDH+aRSA:RC4:!aNULL:!eNULL:!LOW:!3DES:!MD5:!EXP:!PSK:!SRP:!DSS",
}
```

* Pull the Interlock image from the Docker Hub: `docker pull ehazlett/interlock`
* Then start the interlock container:

`docker run -p 80:8080 -d -v /tmp/controller.conf:/etc/interlock/controller.conf ehazlett/interlock -config /etc/interlock/controller.conf`

If you want SSL support, enter a path to the cert (probably want a mounted volume) and then expose 443:

`docker run -p 80:8080 -p 443:8443 -d -v /tmp/controller.conf:/etc/interlock/controller.conf -v /etc/ssl:/ssl ehazlett/interlock -config /etc/interlock/controller.conf`

* You should then be able to access `http://<your-host-ip>/haproxy?stats` to see the proxy stats.
* Add some CNAMEs or /etc/host entries for your IP.  Interlock uses the `hostname` in the container config to add backends to the proxy.

# Commandline options

* `swarm-url` - url to swarm (default: tcp://127.0.0.1:2375)
* `config` - path to config file
* `proxy-conf-path` - path to proxy file (will be generated and created)
* `proxy-pid-path` - path to proxy pid file
* `syslog` - address to syslog
* `proxy-port` - proxy listen port. Default: 8080
* `ssl-cert` - path to single ssl certificate or directory (for SNI). This enables SSL in proxy configuration
* `ssl-port` - ssl listen port (must have cert above). Default: 8443
* `ssl-opts` - string of SSL options (eg. ciphers or ssl, tls versions)

Example for SNI (multidomain) https:

`docker run -it -p 80:8080 -d ehazlett/interlock -shipyard-url <your-shipyard-url> -shipyard-service-key <your-shipyard-service-key> -ssl-cert /etc/ssl/default.crt -ssl-opts 'crt /etc/ssl no-sslv3 ciphers EECDH+ECDSA+AESGCM:EECDH+aRSA+AESGCM:EECDH+ECDSA+SHA384:EECDH+ECDSA+SHA256:EECDH+aRSA+SHA384:EECDH+aRSA+SHA256:EECDH+aRSA+RC4:EECDH:EDH+aRSA:RC4:!aNULL:!eNULL:!LOW:!3DES:!MD5:!EXP:!PSK:!SRP:!DSS'`

In this example HAProxy will use concatinated certificates from /etc/ssl/<hostname>.crt for SNI requests, falling back to /etc/ssl/default.crt, also we specify secure openssl ciphers and disable SSLv3 support (which vunerable to POODLE attack)


# Optional Data
There is also the ability to send configuration data when running containers.  This allows for customization of the backend configuration in HAProxy.  To use this, specify the options as a JSON payload in the environment variable `INTERLOCK_DATA` when launching a container.  For example:

## Data Fields

* `hostname`: override the container hostname -- this is the combined with the domain to create the endpoint
* `domain`: override the container domain
* `alias_domains`: specify a list of alias domains to add (`{"alias_domains": ["foo.com", "bar.com"]}`)
* `port`: specify which container port to use for backend (`{"port": 8080}`)
* `warm`: connect to the container before adding to the backend (`{"warm": true}`)
* `check`: specify a custom check for the backend (`{"check": "httpchk GET /"}`)
* `backend_options`: specify a list of additional options for the backend (`{"backend_options": ["forceclose", "http-no-delay"]}`) -- see http://cbonte.github.io/haproxy-dconv/configuration-1.5.html#4.1
* `check_interval`: specify the interval (in ms) when to run the health check (`{"check_interval": 10000}`)  default: 5000
* `ssl_only`: configure redirect to SSL for backend (`{"ssl_only": true}`)

For example:

`docker run -it -P -d --hostname www.example.com -e INTERLOCK_DATA='{"alias_domains": ["foo.com"], "port": 8080, "warm": true}' ehazlett/go-demo`

This will create a backend to access the container at "www.example.com" and an alias domain `foo.com`, use the port that was allocated for the container port "8080" and make a GET request to the backend container before adding.

# Monitoring
You can use `/haproxy?monitor` to check the status of HAProxy.

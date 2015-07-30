# HAProxy
The HAProxy plugin adds an event driven load balancer and reverse proxy for
Docker.  It automatically adds containers that are running in a Swarm to
HAProxy.

This works by listening on the event stream for containers being created. When a
container is created, the plugin will create an HAProxy backend if the container
has a hostname and at least one exposed port.  The plugin will take care of
adding multiple containers using the same hostname to the proper HAProxy
backend.

 > Note: Interlock HAProxy plugin requires HAProxy 1.5+

# Configuration
The following configuration is available through environment variables:

- `HAPROXY_PROXY_CONFIG_PATH`: HAProxy generated config file path (default: `/proxy.conf`)
- `HAPROXY_PROXY_BACKEND_OVERRIDE_ADDRESS`: Manually set the proxy backend address -- this is needed if not using Swarm (i.e. only Docker)
- `HAPROXY_PORT`: Port to serve (default: `8080`)
- `HAPROXY_PID_PATH`: HAProxy pid path
- `HAPROXY_MAX_CONN`: Max connections (default: `2048`)
- `HAPROXY_CONNECT_TIMEOUT`: Connection timeout (default: `5000`)
- `HAPROXY_SERVER_TIMEOUT`: Server connection timeout (default: `10000`)
- `HAPROXY_CLIENT_TIMEOUT`: Client connection timeout (default: `10000`)
- `HAPROXY_STATS_USER`: HAProxy admin username (default: `stats`)
- `HAPROXY_STATS_PASSWORD`: HAProxy admin password (default: `interlock`)
- `HAPROXY_SSL_PORT`: HAProxy SSL port (default: `8443`)
- `HAPROXY_SSL_CERT`: Path to SSL certificate for HAProxy
- `HAPROXY_SSL_OPTS`: SSL options for HAProxy

> Note: environment variables are optional.  There are sensible defaults provided.

# Usage

An example run of an Interlock container using the `haproxy` plugin is as follows:

`docker run -p 80:8080 -d ehazlett/interlock --swarm-url tcp://1.2.3.4:2375 --plugin haproxy start`

Some things to note about this running container:

- You should then be able to access `http://<your-host-ip>/haproxy?stats` to see
  the proxy stats.  As noted in the above "Configuration" section, the default
  username for this page is `stats` and the default password is `interlock`.
- Interlock logs to STDOUT, so you can view the Interlock logs with `docker logs
  <container-name>`.
- If you wish to see the configuration file which Interlock has generated for
  HAProxy, it is located at `/proxy.conf` by default.  `docker exec -it
  <interlock-container-name> cat /proxy.conf` is an example of a command which will
  print it.
- You can add some CNAMEs or `/etc/host` entries for your IP.  Interlock will
  use the `hostname` value specified in the container config to add backends to
  the proxy.

## Machine

If you used Machine to create your Swarm, you can use this command to start interlock:

`docker run -p 80:8080 -d -v /etc/docker:/etc/docker ehazlett/interlock --swarm-url $DOCKER_HOST --swarm-tls-ca-cert=/etc/docker/ca.pem --swarm-tls-cert=/etc/docker/server.pem --swarm-tls-key=/etc/docker/server-key.pem --plugin haproxy start`

## SSL

If you want SSL support, enter a path to the cert (probably want a mounted volume) and then expose 443:

`docker run -p 80:8080 -p 443:8443 -d -v /etc/ssl:/ssl -e HAPROXY_SSL_CERT=/ssl/cert.pem ehazlett/interlock --swarm-url tcp://1.2.3.4:2375 --plugin haproxy start`

Example for SNI (multidomain) https:

```
docker run -it -p 80:8080 -p 443:8443 -d -v /etc/ssl:/etc/ssl -e HAPROXY_SSL_CERT=/etc/ssl/default.crt \
    -e HAPROXY_SSL_OPTIONS='crt /etc/ssl no-sslv3 ciphers EECDH+ECDSA+AESGCM:EECDH+aRSA+AESGCM:EECDH+ECDSA+SHA384:EECDH+ECDSA+SHA256:EECDH+aRSA+SHA384:EECDH+aRSA+SHA256:EECDH+aRSA+RC4:EECDH:EDH+aRSA:RC4:!aNULL:!eNULL:!LOW:!3DES:!MD5:!EXP:!PSK:!SRP:!DSS' ehazlett/interlock --swarm-url tcp://1.2.3.4:2375 -p haproxy start
```

In this example HAProxy will use concatinated certificates from /etc/ssl/<hostname>.crt for SNI requests, falling back to /etc/ssl/default.crt.  It also specifies secure openssl ciphers and disables SSLv3 support (POODLE attack vulnerability)

# Running a Container
- `docker run -P --hostname test.local ehazlett/docker-demo`

This will create the container and make it available via `http://test.local`.
Note: you will have to create an entry in your local hosts file to resolve.

# Interlock Data
The HAProxy plugin can use additional data from a container's `INTERLOCK_DATA` 
environment variable.  This must be specified as a JSON payload in the variable.
The following options are available:

- `hostname`: override the container hostname -- this is the combined with the domain to create the endpoint
- `domain`: override the container domain
- `alias_domains`: specify a list of alias domains to add (`{"alias_domains": ["foo.com", "bar.com"]}`)
- `port`: specify which container port to use for backend (`{"port": 8080}`)
- `check`: specify a custom check for the backend (`{"check": "httpchk GET /"}`)
- `backend_options`: specify a list of additional options for the backend (`{"backend_options": ["forceclose", "http-no-delay"]}`) -- see http://cbonte.github.io/haproxy-dconv/configuration-1.5.html#4.1 -- these only work for the keywords that begin with `option`.
- `backend_params`: specify a list of additional lines for the backend section of the config (`{"backend_params": ["http-request set-header X-Foo Bar"]}`) -- see http://cbonte.github.io/haproxy-dconv/configuration-1.5.html#4.1 -- this works for anything that can be specified in a 'backend'.
- `check_interval`: specify the interval (in ms) when to run the health check (`{"check_interval": 10000}`)  default: 5000
- `ssl_only`: configure redirect to SSL for backend (`{"ssl_only": true}`)
- `balance_algorithm`: haproxy balancing algorithm (default: `roundrobin`) http://cbonte.github.io/haproxy-dconv/configuration-1.5.html#balance

For example:

```
docker run -ti \
    -P \
    -d \
    --hostname www.example.com \
    -e INTERLOCK_DATA='{"alias_domains": ["foo.com"], "port": 8080, "warm": true}' \
    ehazlett/go-demo
```

This will create a backend to access the container at "www.example.com" and an alias domain `foo.com`, use the port that was allocated for the container port "8080" and make a GET request to the backend container before adding.

# Monitoring
You can use `/haproxy?monitor` to check the status of HAProxy.

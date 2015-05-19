# nginx
This is an [nginx](http://nginx.org) plugin.  It will dynamically build
backends based upon container events (start adds backends, stop/remove removes
backends).

 > Note: Interlock nginx plugin requires nginx v1.6+

# Configuration
The following configuration is available through environment variables:

- `NGINX_PROXY_CONFIG_PATH`: nginx generated config file path (default: `/etc/nginx/nginx.conf`)
- `NGINX_PROXY_BACKEND_OVERRIDE_ADDRESS`: Manually set the proxy backend address -- this is needed if not using Swarm (i.e. only Docker)
- `NGINX_PORT`: Port to serve (default: `80`)
- `NGINX_PID_PATH`: nginx pid path (default: `/nginx.pid`)
- `NGINX_MAX_CONN`: Max connections (default: `1024`)
- `NGINX_MAX_PROCESSES`: Max connections (default: `2`)
- `NGINX_RLIMIT_NOFILE`: Max number of open files (default: `65535`)
- `NGINX_PROXY_CONNECT_TIMEOUT`: proxy connect timeout in seconds (default: `600`)
- `NGINX_PROXY_READ_TIMEOUT`: proxy read timeout in seconds (default: `600`)
- `NGINX_PROXY_SEND_TIMEOUT`: proxy send timeout in seconds (default: `600`)
- `NGINX_SEND_TIMEOUT`: send timeout in seconds (default: `600`)
- `NGINX_SSL_PORT`: SSL port (default: `443`)
- `NGINX_SSL_CERT_DIR`: Path to root directory for SSL certificates
- `NGINX_SSL_CIPHERS`: List of SSL ciphers (default: `HIGH:!aNULL:!MD5`)
- `NGINX_SSL_PROTOCOLS`: List of SSL protocols (default: `SSLv3 TLSv1 TLSv1.1 TLSv1.2`)
- `NGINX_USER`: User to run nginx (default: `www-data`)

> Note: environment variables are optional.  There are sensible defaults provided.

# Usage

An example run of an Interlock container using the `nginx` plugin is as follows:

`docker run -p 80:80 -d ehazlett/interlock --swarm-url tcp://1.2.3.4:2375 --plugin nginx start`

If you want SSL support, enter a path to the cert (probably want a mounted volume) and then expose 443:

> Note: the SSL certificate must exist in the directory specified.  The paths are joined so you only need to specify the certificate name -- not the full path.

`docker run -p 80:80 -p 443:443 -d -v /etc/ssl:/ssl -e NGINX_SSL_CERT_DIR=/ssl ehazlett/interlock --swarm-url tcp://1.2.3.4:2375 --plugin nginx start`

Then run a container using `INTERLOCK_DATA` to specify the certificate name to use:

`docker run --rm -P --hostname foo.local -e INTERLOCK_DATA='{"ssl":true,"ssl_certificate":"evanhazlett.com.pem","ssl_certificate_key":"evanhazlett.com.key"}' ehazlett/docker-demo`

# Interlock Data
The HAProxy plugin can use additional data from a container's `INTERLOCK_DATA` 
environment variable.  This must be specified as a JSON payload in the variable.
The following options are available:

- `hostname`: override the container hostname -- this is the combined with the domain to create the endpoint
- `domain`: override the container domain
- `alias_domains`: specify a list of alias domains to add (`{"alias_domains": ["foo.com", "bar.com"]}`)
- `port`: specify which container port to use for backend (`{"port": 8080}`)
- `ssl`: configure SSL for backend (`{"ssl": true}`)
- `ssl_only`: configure redirect to SSL for backend (`{"ssl_only": true}`)
- `websocket_endpoints`: list of endpoints to proxy websockets (`{"websocket_endpoints": ["/exec"]}`)

For example:

```
docker run -ti \
    -P \
    -d \
    --hostname www.example.com \
    -e INTERLOCK_DATA='{"alias_domains": ["foo.com"], "port": 8080}' \
    ehazlett/go-demo
```

This will create a backend to access the container at "www.example.com" and an alias domain `foo.com` and use the port that was allocated for the container port "8080".

# Monitoring
You can use `/nginx_status` to check the status of Nginx.

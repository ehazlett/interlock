# HAProxy
The HAProxy plugin adds an event driven load balancer and reverse proxy for
Docker.  This works by listening on the event stream.  When an event is received,
the plugin will create a backend using the hostname and exposed port from the
container.  The plugin will take care of adding multiple containers using
the same hostname to the proper HAProxy backend.

Note: Interlock requires HAProxy 1.5+

# Example
- `docker run -P --hostname test.local ehazlett/docker-demo`

This will create the container and make it available via `http://test.local`.
Note: you will have to create an entry in your local hosts file to resolve.

# Extra Data
The HAProxy plugin can use additional data from a container's `INTERLOCK_DATA` 
environment variable.  This must be specified as a JSON payload in the variable.
The following options are available:

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

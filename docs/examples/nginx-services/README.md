# Interlock with Docker Services
Starting in Docker 1.12 there is support for Services.  In this example, we
will use Services to run Interlock and the Proxy.

Note: if you have not initialized the Swarm, make sure to init it with `docker swarm init`

# Nginx
This will create a global Nginx service that will run on every node and
publish port 80:

```
docker service create \
    --name interlock-nginx \
    --publish 80:80 \
    --mode global \
    --label interlock.ext.name=nginx \
    nginx \
        nginx -g "daemon off;" -c /etc/nginx/nginx.conf
```

You should now be able to visit `http://<node-ip>` and see the Nginx welcome
page.

# Interlock
We will now start a global Interlock service that will run on every node:

First, we need an Interlock configuration.  Create a file called `config.toml`
with the content:

```
ListenAddr = ":8080"
DockerURL = "unix:///var/run/docker.sock"
PollInterval = "2s"

[[Extensions]]
Name = "nginx"
ConfigPath = "/etc/nginx/nginx.conf"
PidPath = "/var/run/nginx.pid"
TemplatePath = ""
MaxConn = 1024
Port = 80
```

Now create the Interlock service:

```
docker service create \
    --mode global \
    --name interlock \
    --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock,writable=true \
    --env INTERLOCK_CONFIG="$(cat config.toml)" \
    ehazlett/interlock:latest -D run
```

This will load the `config.toml` as an environment variable.  Also note that
we are using the Docker socket and do not need to have any other access to
Docker.  If you need to update the configuration, simply update the config
file and run
`docker service update --env INTERLOCK_CONFIG="$(cat config.toml)" interlock`.

# Demo
We will now start a demo service.  We will use labels in the service for
Interlock to configure the upstream:

```
docker service create \
    --name demo \
    --env SHOW_VERSION=1 \
    --label interlock.hostname=demo \
    --label interlock.domain=local \
    --label interlock.port=8080 \
    ehazlett/docker-demo:latest
```

To test locally, create an entry in `/etc/hosts` to the Node IP for
`demo.local`.  An example entry is:

```
10.10.0.150     demo.local
```

You should now be able to access `http://demo.local` and see the demo.

You can also now scale the service and see the new tasks show up in the demo:

```
docker service scale demo=6
```

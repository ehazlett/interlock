# Getting Started
Interlock runs as a [Docker](https://www.docker.com) container.  It is distributed as a Docker image and is released on the [Docker Hub](https://hub.docker.com).

Simply run `docker pull ehazlett/interlock:latest` to get the latest Interlock
image.

# Interlock Options
```
NAME:
   interlock - an event driven extension system for docker

USAGE:
   interlock [global options] command [command options] [arguments...]
   
VERSION:
   1.0.0
   
AUTHOR(S):
   @ehazlett <ejhazlett@gmail.com> 
   
COMMANDS:
   spec     generate a configuration file
   run      run interlock
   help, h  Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --debug, -D      Enable debug logging
   --help, -h       show help
   --version, -v    print the version
   
```

# Configuration
Interlock uses a configuration store (file/kv) to set options for Interlock and the
extensions.  See [Configuration](configuration.md) for full details.

# Extensions
Interlock has an extension system that interacts with
other Docker containers.  For this example, we will use
[Nginx](https://www.nginx.com).  See [Extensions](extensions.md) for
more details on extensions.

Interlock works by listening on the Docker event stream for container events
such as `create`, `start`, `stop`, `kill`, etc.  When an event is received,
Interlock will reconfigure the extension and then reload it.  In the Nginx
extension this will re-configure the Nginx configuration file and then signal
the Nginx container causing a reload.  This happens in milliseconds with zero
downtime.

# Quickstart
This will get a quick Interlock + Nginx load balancer.

Note: It is recommended to use [Swarm](https://www.docker.com/products/docker-swarm) as Interlock will handle updating the configuration with the proper
Swarm node.  If you do not use Swarm, you will need to set the `BackendOverrideAddress` option to a resolvable IP address so Nginx knows which node to route the request.  In the example below, we will use the Docker socket for simplicity.

## Nginx
Create a volume for the Nginx config:

`docker volume create --name nginx`

Create the Interlock config `config.toml`:

```
ListenAddr = ":8080"
DockerURL = "unix:///var/run/docker.sock"

[[Extensions]]
  Name = "nginx"
  ConfigPath = "/etc/conf/nginx.conf"
  PidPath = "/etc/conf/nginx.pid"
  BackendOverrideAddress = "172.17.0.1"
```

Start Interlock:

```
docker run \
    -P \
    -d \
    -ti \
    -v nginx:/etc/conf \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v $(pwd)/config.toml:/etc/config.toml
    ehazlett/interlock
    -D run -c /etc/config.toml
```

Start the Nginx container:

```
docker run \
    -ti \
    --net=host \
    --label interlock.ext.name=nginx \
    -v nginx:/etc/conf \
    nginx \
    nginx -g "daemon off;" -c /etc/conf/nginx.conf
```

You can now start some containers with exposed ports to see Interlock add them to Nginx and reload:

`docker run -d -ti -P --hostname foo.local nginx`

```
INFO[0000] interlock 1.0.0
DEBU[0000] docker client: url=unix:///var/run/docker.sock 
DEBU[0000] loading extension: name=nginx configpath=/etc/conf/nginx.conf 
DEBU[0000] updating load balancers                      
INFO[0000] configuration updated                         ext=nginx
DEBU[0077] container start: id=bc9c4f2b9a8697406377191ade8a187c80ef37d9cb59391e1e14608e97495605 image=nginx 
DEBU[0077] container start: id=bc9c4f2b9a8697406377191ade8a187c80ef37d9cb59391e1e14608e97495605 image=nginx 
DEBU[0078] updating load balancers                      
DEBU[0078] websocket endpoints: []                       ext=nginx
DEBU[0078] alias domains: []                             ext=nginx
INFO[0078] foo.local: upstream=172.17.0.1:32778          ext=nginx
INFO[0078] configuration updated                         ext=nginx
INFO[0078] restarted proxy container: id=cfb04b4d050d name=/distracted_goldstine  ext=nginx
```

For detailed configuration, continue to [Configuration](configuration.md).

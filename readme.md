# Interlock
Dynamic, event-driven Docker plugin system using [Swarm](https://github.com/docker/swarm).

# Usage
Run `docker run ehazlett/interlock list-plugins` to show available plugins.

Example:

`docker run -P ehazlett/interlock -s tcp://1.2.3.4:2375 --plugin example start`

To run with `boot2docker` with TLS (in production, you would want to give interlock its own 
certificate, but sharing certificates is a shortcut that shouldn't cause a terrible security 
hole in development):

```
docker run -P -v /var/lib/boot2docker/tls:/etc/ssl/docker ehazlett/interlock \
  -s tcp://172.17.42.1:2376 \
  --swarm-tls-ca-cert /etc/ssl/docker/ca.pem \
  --swarm-tls-cert /etc/ssl/docker/cert.pem \
  --swarm-tls-key /etc/ssl/docker/key.pem \
  --plugin example start
```

# Commandline options

- `--swarm-url`: url to swarm (default: tcp://127.0.0.1:2375)
- `--swarm-tls-ca-cert`: TLS CA certificate to use with swarm (optional)
- `--swarm-tls-cert`: TLS certificate to use with swarm (optional)
- `--swarm-tls-key`: TLS certificate key to use with swarm (options)
- `--plugin`: enable plugin
- `--debug`: enable debug output
- `--version`: show version and exit

# Plugins
See the [Plugins](https://github.com/ehazlett/interlock/tree/master/plugins)
directory for available plugins and their corresponding readme.md for usage.

| Name | Description |
|-----|-----|
| [Example](https://github.com/ehazlett/interlock/tree/master/plugins/example) | Example Plugin for Reference|
| [HAProxy](https://github.com/ehazlett/interlock/tree/master/plugins/haproxy) | [HAProxy](http://www.haproxy.org/) Load Balancer |
| [Nginx](https://github.com/ehazlett/interlock/tree/master/plugins/nginx) | [Nginx](http://nginx.org) Load Balancer |
| [Stats](https://github.com/ehazlett/interlock/tree/master/plugins/stats) | Container stat forwarding to [Carbon](http://graphite.wikidot.com/carbon) |

# To build

Building interlock is easy with the script:

`script/build $VERSION`

# License
Licensed under the Apache License, Version 2.0. See LICENSE for full license text.

# Interlock
Dynamic, event-driven Docker plugin system using [Swarm](https://github.com/docker/swarm).

# Usage
Run `docker run ehazlett/interlock list-plugins` to show available plugins.

Example:

`docker run -P ehazlett/interlock -s tcp://1.2.3.4:2375 --plugin example start`

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

# Interlock
Interlock is an event driven extension system for [Docker](https://www.docker.com).
It uses the [Docker Event](https://docs.docker.com/engine/reference/api/docker_remote_api/#docker-events) stream to notify "extensions".  Currently the supported extensions
are HAProxy and Nginx.  This provides a dynamic load balancer and reverse proxy
utilizing either HAProxy or Nginx.

# Getting Started
To get started with Interlock, see [Getting Started](getting_started.md)

# Examples
There are examples using [Docker Compose](https://docs.docker.com/compose/)
for both Nginx and HAProxy in the [examples](examples) directory.

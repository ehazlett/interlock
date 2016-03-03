# Extensions
Extensions provide backend functionality for Interlock.  These can be just
about anything (metrics, autostart, autoscale, etc).  Interlock currently
ships with support for two load balancing extensions.  Interlock also uses
external extension containers instead of bundling in a single image.  This
keeps Interlock lightweight as well as providing the ability to specify your
own container image if desired.  By default, it is recommended to use official
Docker images from the Docker Hub for each extension.

# Load Balancing
The following load balancing extensions are supported:

- [HAProxy](../extensions/haproxy.md)
- [Nginx](../extensions/nginx.md)

<<<<<<< HEAD
Interlock will re-configure HAProxy upon a container event (start, stop, kill, remove, etc)
and trigger a reload on the HAProxy container or containers.

To start an HAProxy container that Interlock will manage, simply add the following
label upon start: `interlock.ext.name=haproxy`.  For example:

`docker run -p 80:80 --label interlock.ext.name=haproxy haproxy`

Interlock will restart all containers with that label whenever the HAProxy config
is updated.

Note: If you run Interlock as a privileged container and on the same host
as the HAProxy container, Interlock will attempt to drop SYN packets upon
reload to force clients to resend requests to drop as few packets as possible.
If not, a normal container restart will be performed and connections will be
dropped.  See [here](http://marc.info/?l=haproxy&m=133262017329084&w=2) for
details.  If you want to make sure to drop as few packets as possible, try
the Nginx proxy container as it handles connection queueing automatically
and this manual queue is not necessary.

## Nginx
[Nginx](http://www.haproxy.org/) is an HTTP and reverse proxy server, 
a mail proxy server, and a generic TCP proxy server.
It is recommended to use the official [Docker Hub Nginx Image](https://hub.docker.com/_/nginx/).

Interlock will re-configure Nginx upon a container event (start, stop, kill, remove, etc)
and trigger a reload on the Nginx container or containers.

To start an Nginx container that Interlock will manage, simply add the following
label upon start: `interlock.ext.name=nginx`.  For example:

`docker run -p 80:80 --label interlock.ext.name=nginx nginx`

Interlock will reload all containers with that label whenever the Nginx config
is updated.  Interlock sends a `SIGHUP` to the container.  This will cause
Nginx to reload the configuration without connection interruption.

# Metrics
- [Beacon](../extensions/beacon.md)

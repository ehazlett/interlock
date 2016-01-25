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

## HAProxy
[HAProxy](http://www.haproxy.org/) is a high performance TCP/HTTP load balancer.
It is recommended to use the official [Docker Hub HAProxy Image](https://hub.docker.com/_/haproxy/).

Interlock will re-configure HAProxy upon a container event (start, stop, kill, remove, etc)
and trigger a reload on the HAProxy container.

To start an HAProxy container that Interlock will manage, simply add the following
label upon start: `interlock.ext.name=haproxy`.  For example:

`docker run -p 80:80 --label interlock.ext.name=haproxy haproxy`

Interlock will restart all containers with that label whenever the HAProxy config
is updated.

Note: To do a [proper HAProxy restart](http://engineeringblog.yelp.com/2015/04/true-zero-downtime-haproxy-reloads.html) requires host level privileges and is not portable.  Therefore, you
might notice a few dropped connections under high load.  If you
need zero downtime, try the Nginx extension.  It's reloading mechanism will 
handle the connection queueing automatically.

## Nginx
[Nginx](http://www.haproxy.org/) is a high performance TCP/HTTP load balancer.
It is recommended to use the official [Docker Hub HAProxy Image](https://hub.docker.com/_/haproxy/).

Interlock will re-configure HAProxy upon a container event (start, stop, kill, remove, etc)
and trigger a reload on the HAProxy container.

To start an HAProxy container that Interlock will manage, simply add the following
label upon start: `interlock.ext.name=haproxy`.  For example:

`docker run -p 80:80 --label interlock.ext.name=haproxy haproxy`

Interlock will reload all containers with that label whenever the HAProxy config
is updated.  Interlock sends a `SIGHUP` to the container.  This will cause
Nginx to reload the configuration without connection interruption.

+++
date = "2016-04-14T09:42:58-04:00"
next = "/extensions/nginx/"
prev = "/extensions/"
title = "HAProxy"
toc = false
weight = 5

+++

[HAProxy](http://www.haproxy.org/) is a high performance TCP/HTTP load balancer.

Interlock will automatically reconfigure the load balancer whenever an event
is received such as a container start or stop.  This means you can start new
containers or scale existing services without interacting or updating with
an upstream load balancer.

{{% notice note %}}
It is recommended to use the official [Docker Hub HAProxy Image](https://hub.docker.com/_/haproxy/).
{{% /notice %}}

To start an HAProxy container that Interlock will manage, simply add the following
label upon start: `interlock.ext.name=haproxy`.

For example:

```bash
docker run -p 80:80 --label interlock.ext.name=haproxy haproxy
```

Interlock will restart all containers with that label whenever the HAProxy config
is updated.

{{% notice note %}}
If you run Interlock as a privileged container and on the same host
as the HAProxy container, Interlock will attempt to drop SYN packets upon
reload to force clients to resend requests to drop as few packets as possible.
If not, a normal container restart will be performed and connections will be
dropped.  See [here](http://marc.info/?l=haproxy&m=133262017329084&w=2) for
details.  If you want to make sure to drop as few packets as possible, try
the Nginx proxy container as it handles connection queueing automatically
and this manual queue is not necessary.
{{% /notice %}}

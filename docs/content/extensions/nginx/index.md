+++
date = "2016-04-14T09:43:01-04:00"
next = ""
prev = "/extensions/haproxy/"
title = "Nginx"
toc = false
weight = 10

+++

[Nginx](https://www.nginx.com/) is an HTTP and reverse proxy server, 
a mail proxy server, and a generic TCP proxy server.

Interlock will automatically reconfigure the load balancer whenever an event
is received such as a container start or stop.  This means you can start new
containers or scale existing services without interacting or updating with
an upstream load balancer.

{{% notice note %}}
It is recommended to use the official [Docker Hub Nginx Image](https://hub.docker.com/_/nginx/).
{{% /notice %}}

Interlock will re-configure Nginx upon a container event (start, stop, kill, remove, etc)
and trigger a reload on the Nginx container or containers.

To start an Nginx container that Interlock will manage, simply add the following
label upon start: `interlock.ext.name=nginx`.  

For example:

`docker run -p 80:80 --label interlock.ext.name=nginx nginx`

Interlock will reload all containers with that label whenever the HAProxy config
is updated.  Interlock sends a `SIGHUP` to the container.  This will cause
Nginx to reload the configuration without connection interruption.

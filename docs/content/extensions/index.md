+++
chapter = true
date = "2016-04-14T09:41:59-04:00"
icon = "<b></b>"
next = "/extensions/haproxy/"
prev = ""
title = "Extensions"
weight = 0

+++

### Interlock

# Extensions
Extensions provide backend functionality for Interlock.  These can be just
about anything (metrics, autostart, autoscale, etc).  Interlock currently
ships with support for two load balancing extensions.  Interlock also uses
external extension containers instead of bundling in a single image.  This
keeps Interlock lightweight as well as providing the ability to specify your
own container image if desired.  

{{% notice note %}}
It is recommended to use official
Docker images from the Docker Hub for each extension.
{{% /notice %}}

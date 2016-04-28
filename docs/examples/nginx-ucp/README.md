# Interlock + Docker Universal Control Plane
This example shows Interlock with Docker Universal Control Plane.

Start with the [Docker Swarm](https://docs.docker.com/swarm/install-w-machine/)
evaluation tutorial.  Once you have a working Swarm cluster continue below.

Note: this uses [Docker Compose](http://docs.docker.com/compose).  Please make
sure you have the latest version installed.

# Setup
To make this example portable, we use an environment variable to configure
Interlock to your Swarm cluster.  Run the following to set it up (replace
`<IP-of-your-UCP-controller>` with the IP or DNS name of your controller and
`<swarm-port>` with your Swarm port (default: 2376 or custom if using Docker
Machine):

`export SWARM_HOST=tcp://<IP-of-your-UCP-controller>:<swarm-port>`

# Start Nginx

`docker-compose up -d nginx`

Run `docker ps` to confirm the proxy container is running:

```
CONTAINER ID        IMAGE               COMMAND                  CREATED             STATUS              PORTS                                NAMES
61edbed29237        nginx:latest        "nginx -g 'daemon off"   12 seconds ago      Up 6 seconds        192.168.122.99:80->80/tcp, 443/tcp   ucp-00/nginxucp_nginx_1
```

Make note of the node it is running on; in this case `ucp-00`.  Make sure to
add a local `/etc/host` entry for `test.local` to point to the IP of the
node that this proxy container is running on.  Note: that in Docker Machine
this might be different as UCP can use a local VM network that might not
be routable to your host.

# Start Interlock

`docker-compose up -d interlock`

Run `docker ps` to confirm Interlock is running:

```
CONTAINER ID        IMAGE                      COMMAND                  CREATED             STATUS              PORTS                                NAMES
099319ef7611        ehazlett/interlock:1.1.0   "/bin/interlock -D ru"   15 seconds ago      Up 8 seconds        192.168.122.170:32769->8080/tcp      ucp-02/nginxucp_interlock_1
61edbed29237        nginx:latest               "nginx -g 'daemon off"   2 minutes ago       Up 2 minutes        192.168.122.99:80->80/tcp, 443/tcp   ucp-00/nginxucp_nginx_1
```

# Start Example App

`docker-compose up -d app`

Run `docker ps` to confirm container is running:

```
CONTAINER ID        IMAGE                         COMMAND                  CREATED              STATUS              PORTS                                NAMES
9029efe43da8        ehazlett/docker-demo:latest   "/bin/docker-demo -li"   About a minute ago   Up About a minute   192.168.122.96:32769->8080/tcp       ucp-01/nginxucp_app_1
099319ef7611        ehazlett/interlock:1.1.0      "/bin/interlock -D ru"   2 minutes ago        Up 2 minutes        192.168.122.170:32769->8080/tcp      ucp-02/nginxucp_interlock_1
61edbed29237        nginx:latest                  "nginx -g 'daemon off"   5 minutes ago        Up 5 minutes        192.168.122.99:80->80/tcp, 443/tcp   ucp-00/nginxucp_nginx_1
```

Open a browser to http://test.local and you should see the demo app:

![Screenshot](screenshot.png?raw=true)

Try scaling to see Interlock add the new containers:

`docker-compose scale app=4`

You should now be able to refresh the page a few times to see the hostname
change as it is balancing between upstreams.

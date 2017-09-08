# Interlock + Docker Swarm + Stack deploy
This example shows a Interlock in a Swarm cluster deployed using docker stack deploy.

Start with the [Docker Swarm](https://docs.docker.com/swarm/install-w-machine/)
evaluation tutorial.  Once you have a working Swarm cluster continue below.

Note: you need a manager and a worker node to run this example

Note: this uses [Docker Compose](http://docs.docker.com/compose).  Please make
sure you have the latest version installed.

# Setup
To make this example portable, we use an environment variable to configure
Interlock to your Swarm cluster.  Run the following to set it up:

`docker-machine env manager`

export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://192.168.99.100:2376"
export DOCKER_CERT_PATH="/Users/jccote/.docker/machine/machines/manager"
export DOCKER_MACHINE_NAME="manager"
# Run this command to configure your shell:
# eval $(docker-machine env manager)

# build an interlock image using a custom tag
# this builds the image into the DOCKER_HOST specified above that is the manager node
`make -e TAG=mytag image`

# you can verify the image was created in the manager node
`docker image ls`
REPOSITORY           TAG                 IMAGE ID            CREATED             SIZE
ehazlett/interlock   mytag               7b20d1a87b1e        15 seconds ago      23.2MB
<none>               <none>              0da48c200507        34 seconds ago      634MB
nginx                <none>              0b5dec81616c        44 hours ago        108MB
alpine               latest              7328f6f8b418        2 months ago        3.97MB
golang               1.6-alpine          1ea38172de32        8 months ago        283MB

# generate a stack file using docker-compose
`docker-compose -f ./docs/examples/nginx-swarm-stack-machine/docker-compose.yml config > stack.yml`

# deploy the stack using docker stack deploy and give your stack a name
`docker stack deploy -c stack.yml mystack`

# you should now have the following service running
`docker service ls`
ID                  NAME                MODE                REPLICAS            IMAGE                         PORTS
6jbqsojcwrbb        mystack_app         replicated          2/2                 ehazlett/docker-demo:latest   *:0->8080/tcp
kbeckpeyqbob        mystack_nginx       replicated          1/1                 nginx:latest                  *:80->80/tcp
ykdsht0davud        mystack_interlock   replicated          1/1                 ehazlett/interlock:jcc

Once up you can check the logs to ensure Interlock is detecting:

`docker logs mystack_interlock.1.5s9qd89crem17f384o2zt2kv2`


You can also verify that the nginx routes are created properly:
`docker exec -it mystack_nginx.1.5s9qd89crem17f384o2zt2kv2 /bin/bash -c "cat /etc/nginx/nginx.conf"`

 upstream ctx___web {
        zone ctx___web_backend 64k;
	server 10.0.0.8:8080;
	server 10.0.0.5:8080;

    }

    server {
        listen 7070;
        server_name _;


	location /web {

	    proxy_pass http://ctx___web;
	}


The sample web applications should be available at
http://192.168.99.100:7070/web

# Interlock + Docker Swarm
This example shows a Interlock in a Swarm cluster.

Start with the [Docker Swarm](https://docs.docker.com/swarm/install-w-machine/)
evaluation tutorial.  Once you have a working Swarm cluster continue below.

Note: this uses [Docker Compose](http://docs.docker.com/compose).  Please make
sure you have the latest version installed.

# Setup
To make this example portable, we use an environment variable to configure
Interlock to your Swarm cluster.  Run the following to set it up:

`export SWARM_HOST=tcp://$(docker-machine ip manager):3376`

# Start Interlock

`docker-compose up -d interlock`

# Start Nginx

`docker-compose up -d nginx`

# Start Example App

`docker-compose up -d app`

Once up you can check the logs to ensure Interlock is detecting:

`docker-compose logs`

Try scaling to see Interlock add the new containers:

`docker-compose scale app=4`

You should now be able to put a host entry for the node the `nginx` container
is running on (use `docker ps` to find the node) to `test.local` 
and access via your browser.

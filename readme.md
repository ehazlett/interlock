# Interlock
Dynamic, event-driven HAProxy using [Citadel](https://github.com/citadel/citadel).  Interlock configures backends for HAProxy by listening to Docker events (start/stop, etc).

# Usage
To get started, generate a config (save this as `/tmp/controller.conf`):

Replace `1.2.3.4` in `addr` listed in the engines section with your IP for your Docker host.  You must enable TCP in the Docker daemon (see the [Docker Docs](http://docs.docker.com/reference/commandline/cli/) for details).

```
{
    "port": 8080,
    "connect_timeout": 5000,
    "server_timeout": 10000,
    "client_timeout": 10000,
    "max_conn": 2048,
    "sysload_addr": "",
    "proxy_config_path": "/tmp/proxy.conf",
    "stats_user": "stats",
    "stats_password": "stats",
    "pid_path": "/tmp/proxy.pid",
    "engines": [
        {
            "engine": {
                "id": "local",
                "addr": "http://1.2.3.4:4243",
                "cpus": 1.0,
                "memory": 1024,
                "labels": []
            },
            "ssl_cert": "",
            "ssl_key": "",
            "ca_cert": ""
        }
    ]
}
```

* Pull the Interlock image from the Docker Hub: `docker pull ehazlett/interlock`
* Then start the interlock container:

`docker run -p 80:8080 -d -v /tmp/controller.conf:/etc/interlock/controller.conf ehazlett/interlock -config /etc/interlock/controller.conf`

* You should then be able to access `http://<your-host-ip>/haproxy?stats` to see the proxy stats.
* Add some CNAMEs or /etc/host entries for your IP.  Interlock uses the `hostname` in the container config to add backends to the proxy.

# Optional Data
There is also the ability to send configuration data.  This allows for custom specification of ports (by default interlock will use the first exposed port), DNS aliases (extra domain names that reference a backend) and the ability to "warm" each container before adding.  To do this, specify the options as a JSON payload in the environment variable `INTERLOCK_DATA`.  For example:

`docker run -it -P -d -e INTERLOCK_DATA='{"alias_domains": ["foo.com", "example.com"], "port": 8080, "warm": true}' ehazlett/go-demo`

This will create two alias domains `foo.com` and `example.com`, use the port that was allocated for the container port "8080" and make a GET request to the backend container before adding.

# Monitoring
You can use `/haproxy?monitor` to check the status of HAProxy.

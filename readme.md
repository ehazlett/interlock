# Interlock
Dynamic, event-driven HAProxy using [Citadel](https://github.com/citadel/citadel)

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

Pull the Interlock image from the Docker Hub: `docker pull ehazlett/interlock`

Then start the interlock container:

`docker run -p 80:8080 -d -v /tmp/interlock:/etc/interlock/controller.conf` ehazlett/interlock -config /etc/interlock/controller.conf`

You should then be able to access `http://<your-host-ip>/haproxy?stats` to see the proxy stats.

Add some CNAMEs or /etc/host entries for your IP.

Interlock uses the `hostname` in the container config to add backends to the proxy.

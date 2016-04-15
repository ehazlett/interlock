+++
date = "2016-04-14T09:19:54-04:00"
next = ""
prev = ""
title = "Getting Started"
toc = true
weight = 5

+++

Interlock uses TOML for configuration.  This can be loaded in a few different
ways.

# File
As Interlock runs in a [Container](https://www.docker.com) there are a few
different ways to use a file for configuration.  The easiest way is to bind
mount an external file into a known location in the container.  For example,
if you create the file at `/data/config.toml` you can use it by running:

```bash
docker run \
    -ti \
    -d \
    -P \
    --restart=always \
    -v /data/config.toml:/config.toml \
    ehazlett/interlock \
    run -c /config.toml
```

# Environment Variable
You can also use the environment variable `INTERLOCK_CONFIG` to load the
configuration for Interlock.  The content of the variable must be the 
configuration as text.  You then expose this environment variable to the 
container upon launch.  For example:

```bash
export INTERLOCK_CONFIG='
ListenAddr = ":8080"
DockerURL = "unix:///var/run/docker.sock"
TLSCACert = ""
TLSCert = ""
TLSKey = ""
AllowInsecure = false
EnableMetrics = true'
```

You can then run Interlock:

```bash
docker run \
    -ti \
    -d \
    -P \
    -e INTERLOCK_CONFIG \
    --restart=always \
    ehazlett/interlock \
    run
```

# Key/Value Store
Interlock supports etcd and consul [libkv](https://github.com/docker/libkv)
key-value store backends.  This can be used to store the configuration instead
of the file.  This is useful when deploying several instances of Interlock
for HA and scaling.

To configure Interlock to use a kv store, use the `--discovery` option.  You
will need to have the configuration loaded in the KV store.

Interlock will read the key `/interlock/v1/config` for the configuration.  You
can use `curl` to load the config.  For example:

```bash
curl https://1.2.3.4:4001/v2/keys/interlock/v1/config -XPUT -d value='ListenAddr = ":8080"
DockerURL = "tcp://127.0.0.1:2376"

[[Extensions]]
  Name = "haproxy"
  ConfigPath = "/usr/local/etc/haproxy/haproxy.cfg"
  PidPath = "/usr/local/etc/haproxy/haproxy.pid"
  MaxConn = 1024
  Port = 80'
  ```

You can then start Interlock and direct it at the KV store:

```bash
docker run \
    -ti \
    -d \
    -P \
    ehazlett/interlock \
    run --discovery etcd://1.2.3.4:4001
```

# Configuration

Interlock uses a configuration store to configure options and extensions. The configuration store can be one of:

* File
* Environment variable
* Key value store

## File configuration

```
ListenAddr = ":8080"
DockerURL = "unix:///var/run/docker.sock"
TLSCACert = ""
TLSCert = ""
TLSKey = ""
AllowInsecure = false
EnableMetrics = true
PollInterval = ""

[[Extensions]]
  Name = "nginx"
  ConfigPath = "/etc/nginx/nginx.conf"
  PidPath = "/var/run/nginx.pid"
  TemplatePath = "/etc/interlock/nginx.conf.template"
  BackendOverrideAddress = ""
  MaxConn = 1024
  Port = 80
  SSLCertPath = ""
  SSLPort = 0
  User = "www-data"
  WorkerProcesses = 2
  RLimitNoFile = 65535
  ProxyConnectTimeout = 600
  ProxySendTimeout = 600
  ProxyReadTimeout = 600
  SendTimeout = 600
  SSLCiphers = "HIGH:!aNULL:!MD5"
  SSLProtocols = "SSLv3 TLSv1 TLSv1.1 TLSv1.2"
  NginxPlusEnabled = false
```

# Event Stream vs. Polling
Different infrastructure requires different strategy.  Interlock can trigger
updates based upon two methods: event stream and polling.  In some
environments the event stream can be interrupted.  Interlock has the ability
to reconnect if there is a failure but if the events simply stop sending
you can get stuck.  Interlock also offers the ability to poll.  Interlock
will poll Docker at the given interval (must be greater than two (2) seconds)
for changes.  If changes are detected, Interlock will trigger an update.

To enable the event stream, simply omit the `PollInterval` or set the value
to `""`.  If you set an interval, Interlock will switch to use polling.

# Environment variable configuration

You can also put the config as text in the environment variable 
`INTERLOCK_CONFIG`.  If you pass command flags they will override the
environment data.

# Key value store configuration

Interlock supports etcd and consul [libkv](https://github.com/docker/libkv)
key-value store backends.  This can be used to store the configuration instead
of the file.  This is useful when deploying several instances of Interlock
for HA and scaling.

To configure Interlock to use a kv store, use the `--discovery` option.  You
will need to have the configuration loaded in the KV store.

Interlock will read the key `/interlock/v1/config` for the configuration.  You
can use `curl` to load the config.  Here is an example:

```
curl https://1.2.3.4:4001/v2/keys/interlock/v1/config -XPUT -d value='ListenAddr = ":8080"
DockerURL = "tcp://127.0.0.1:2376"

[[Extensions]]
  Name = "haproxy"
  ConfigPath = "/usr/local/etc/haproxy/haproxy.cfg"
  PidPath = "/var/run/haproxy.pid"
  TemplatePath = "/usr/local/etc/interlock/haproxy.cfg.template"
  MaxConn = 1024
  Port = 80'
```

You can then start Interlock and point it at the KV store:

`docker run -ti -d --net=host ehazlett/interlock run --discovery etcd://1.2.3.4:4001`

# Reference

The following table lists all options, their type and the extensions in which
they are compatible:

|Option|Type|Extensions Supported|
|----|----|----|
|Name                   | string | extension name |
|ConfigPath             | string | config file path |
|PidPath                | string | haproxy, nginx |
|TemplatePath           | string | haproxy, nginx |
|BackendOverrideAddress | string | haproxy, nginx |
|ConnectTimeout         | int    | haproxy |
|ServerTimeout          | int    | haproxy |
|ClientTimeout          | int    | haproxy |
|MaxConn                | int    | haproxy, nginx |
|Port                   | int    | haproxy, nginx |
|SyslogAddr             | string | haproxy |
|AdminUser              | string | haproxy |
|AdminPass              | string | haproxy |
|SSLCertPath            | string | haproxy, nginx |
|SSLCert                | string | haproxy |
|SSLPort                | int    | haproxy, nginx |
|SSLOpts                | string | haproxy |
|SSLServerVerify        | string | haproxy |
|SSLDefaultDHParam      | int    | haproxy |
|NginxPlusEnabled       | bool   | nginx |
|User                   | string | nginx |
|WorkerProcesses        | int    | nginx |
|RLimitNoFile           | int    | nginx |
|ProxyConnectTimeout    | int    | nginx |
|ProxySendTimeout       | int    | nginx |
|ProxyReadTimeout       | int    | nginx |
|SendTimeout            | int    | nginx |
|SSLCiphers             | string | nginx |
|SSLProtocols           | string | nginx |
|StatInterval           | int    | beacon |

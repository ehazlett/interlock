# Configuration
Interlock uses a configuration store (file/kv) to configure options and extensions.

# Example Configuration
```
ListenAddr = ":8080"
DockerURL = "unix:///var/run/docker.sock"
TLSCACert = ""
TLSCert = ""
TLSKey = ""
AllowInsecure = false
EnableMetrics = true

[[Extensions]]
  Name = "nginx"
  ConfigPath = "/etc/conf/nginx.conf"
  PidPath = "/etc/conf/nginx.pid"
  BackendOverrideAddress = ""
  ConnectTimeout = 5000
  ServerTimeout = 10000
  ClientTimeout = 10000
  MaxConn = 1024
  Port = 80
  SyslogAddr = ""
  NginxPlusEnabled = false
  AdminUser = "admin"
  AdminPass = ""
  SSLCertPath = ""
  SSLCert = ""
  SSLPort = 0
  SSLOpts = ""
  User = "www-data"
  WorkerProcesses = 2
  RLimitNoFile = 65535
  ProxyConnectTimeout = 600
  ProxySendTimeout = 600
  ProxyReadTimeout = 600
  SendTimeout = 600
  SSLCiphers = "HIGH:!aNULL:!MD5"
  SSLProtocols = "SSLv3 TLSv1 TLSv1.1 TLSv1.2"
```

# Reference
The following table lists all options, their type and the extensions in which
they are compatible:

|Option|Type|Extensions Supported|
|----|----|----|
|Name                   | string | extension name |
|ConfigPath             | string | config file path |
|PidPath                | string | haproxy, nginx |
|BackendOverrideAddress | string | haproxy, nginx |
|ConnectTimeout         | int    | haproxy |
|ServerTimeout          | int    | haproxy |
|ClientTimeout          | int    | haproxy |
|MaxConn                | int    | haproxy, nginx |
|Port                   | int    | haproxy, nginx |
|SyslogAddr             | string | haproxy |
|NginxPlusEnabled       | bool   | nginx |
|AdminUser              | string | haproxy |
|AdminPass              | string | haproxy |
|SSLCertPath            | string | haproxy, nginx |
|SSLCert                | string | haproxy |
|SSLPort                | int    | haproxy, nginx |
|SSLOpts                | string | haproxy |
|SSLServerVerify        | string | haproxy |
|SSLDefaultDHParam      | int    | haproxy |
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

# Key Value Storage Support
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
  PidPath = "/usr/local/etc/haproxy/haproxy.pid"
  MaxConn = 1024
  Port = 80'
```

You can then start Interlock and point it at the KV store:

`docker run -ti -d --net=host ehazlett/interlock run --discovery etcd://1.2.3.4:4001`


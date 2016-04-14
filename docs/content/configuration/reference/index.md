+++
date = "2016-04-14T09:39:18-04:00"
next = ""
prev = ""
title = "Reference"
toc = false
weight = 100

+++

Use the following for reference for configuring Interlock and the extensions.

The following table lists all options, their type and the extensions in which
they are compatible:

|Option|Type|Extensions Supported|
|----|----|----|
|ConfigPath             | string | internal |
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

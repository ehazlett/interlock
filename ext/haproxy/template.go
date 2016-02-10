package haproxy

const (
	haproxyConfTemplate = `# managed by interlock
global
    {{ if .Config.SyslogAddr }}log {{ .Config.SyslogAddr }} local0
    log-send-hostname{{ end }}
    maxconn {{ .Config.MaxConn }}
    pidfile {{ .Config.PidPath }}
    ssl-server-verify {{ .Config.SSLServerVerify }}
    tune.ssl.default-dh-param {{ .Config.SSLDefaultDHParam }}

defaults
    mode http
    retries 3
    option redispatch
    option httplog
    option dontlognull
    option http-server-close
    option forwardfor
    timeout connect {{ .Config.ConnectTimeout }}
    timeout client {{ .Config.ClientTimeout }}
    timeout server {{ .Config.ServerTimeout }}

frontend http-default
    bind *:{{ .Config.Port }}
    {{ if .Config.SSLCert }}bind *:{{ .Config.SSLPort }} ssl crt {{ .Config.SSLCert }} {{ .Config.SSLOpts }}{{ end }}
    monitor-uri /haproxy?monitor
    {{ if .Config.AdminUser }}stats realm Stats
    stats auth {{ .Config.AdminUser }}:{{ .Config.AdminPass}}{{ end }}
    stats enable
    stats uri /haproxy?stats
    stats refresh 5s
    {{ range $host := .Hosts }}acl is_{{ $host.Name }} hdr_beg(host) {{ $host.Domain }}
    use_backend {{ $host.Name }} if is_{{ $host.Name }}
    {{ end }}
{{ range $host := .Hosts }}backend {{ $host.Name }}
    http-response add-header X-Request-Start %Ts.%ms
    balance {{ $host.BalanceAlgorithm }}
    {{ range $option := $host.BackendOptions }}option {{ $option }}
    {{ end }}
    {{ if $host.Check }}option {{ $host.Check }}{{ end }}
    {{ if $host.SSLOnly }}redirect scheme https if !{ ssl_fc  }{{ end }}
    {{ range $i,$up := $host.Upstreams }}server {{ $up.Container }} {{ $up.Addr }} check inter {{ $up.CheckInterval }}{{ if $host.SSLBackend }} ssl sni req.hdr(Host) verify {{ $host.SSLBackendTLSVerify }}{{ end }}
    {{ end }}
{{ end }}`
)

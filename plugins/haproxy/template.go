package haproxy

const (
	haproxyTmpl = `# managed by interlock
global
    {{ if .PluginConfig.SyslogAddr }}log {{ .PluginConfig.SyslogAddr }} local0
    log-send-hostname{{ end }}
    maxconn {{ .PluginConfig.MaxConn }}
    pidfile {{ .PluginConfig.PidPath }}

defaults
    mode http
    retries 3
    option redispatch
    option httplog
    option dontlognull
    option http-server-close
    option forwardfor
    timeout connect {{ .PluginConfig.ConnectTimeout }}
    timeout client {{ .PluginConfig.ClientTimeout }}
    timeout server {{ .PluginConfig.ServerTimeout }}

frontend http-default
    bind *:{{ .PluginConfig.Port }}
    {{ if .PluginConfig.SSLCert }}bind *:{{ .PluginConfig.SSLPort }} ssl crt {{ .PluginConfig.SSLCert }} {{ .PluginConfig.SSLOpts }}{{ end }}
    monitor-uri /haproxy?monitor
    {{ if .PluginConfig.StatsUser }}stats realm Stats
    stats auth {{ .PluginConfig.StatsUser }}:{{ .PluginConfig.StatsPassword }}{{ end }}
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
    {{ range $i,$up := $host.Upstreams }}server {{ $up.Container }} {{ $up.Addr }} check inter {{ $up.CheckInterval }}{{ if $host.SSLBackend }} ssl verify {{ $host.SSLBackendTLSVerify }}{{ end }}
    {{ end }}
{{ end }}`
)

package haproxy

const (
	haproxyConfTemplate = `# managed by interlock
global
	log 127.0.0.1 local0
	log 127.0.0.1 local1 notice
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
    {{ range $host := .Hosts }}{{ if ne $host.ContextRoot.Path "" }}acl url{{ $host.ContextRoot.Name }} url_beg -i {{ $host.ContextRoot.Path }}
    use_backend {{ $host.Name }} if url{{$host.ContextRoot.Name}}{{ end }}
    acl is_{{ $host.Name }} hdr_dom(host) {{ $host.Domain }}
    use_backend {{ $host.Name }} if is_{{ $host.Name }}
    {{ end }}

{{ range $host := .Hosts }}{{ if ne $host.ContextRoot.Path "" }}
    acl missing_slash path_reg ^{{ $host.ContextRoot.Path }}[^/]*$
    redirect code 301 prefix / drop-query append-slash if missing_slash
    {{ if $host.ContextRootRewrite }}acl ctx{{ $host.ContextRoot.Name }} path_beg -i {{ $host.ContextRoot.Path }}/
    reqrep ^([^\ ]*)\ {{ $host.ContextRoot.Path }}/(.*)     \1\ /\2 {{ end }}{{ end }}
    backend {{ $host.Name }}
    http-response add-header X-Request-Start %Ts.%ms
    http-request set-header X-Forwarded-Port %[dst_port]
    http-request add-header X-Forwarded-Proto https if { ssl_fc }
    balance {{ $host.BalanceAlgorithm }}
    {{ range $option := $host.BackendOptions }}option {{ $option }}
    {{ end }}
    {{ if $host.Check }}option {{ $host.Check }}{{ end }}
    {{ if $host.SSLOnly }}redirect scheme https code 301 if !{ ssl_fc }{{ end }}
	{{ if $host.SSLOnly }}http-response set-header Strict-Transport-Security "max-age=16000000; includeSubDomains; preload;"{{ end }}
    {{ range $i,$up := $host.Upstreams }}server {{ $up.Container }} {{ $up.Addr }} check inter {{ $up.CheckInterval }}{{ if $host.SSLBackend }} ssl verify {{ $host.SSLBackendTLSVerify }} sni req.hdr(Host){{ end }}
    {{ end }}
{{ end }}
`
)

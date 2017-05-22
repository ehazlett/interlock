package nginx

var nginxConfTemplate = `# managed by interlock
user  {{ .Config.User }};
worker_processes  {{ .Config.WorkerProcesses }};
worker_rlimit_nofile {{ .Config.RLimitNoFile }};

error_log  /var/log/error.log warn;
pid        {{ .Config.PidPath }};


events {
    worker_connections  {{ .Config.MaxConn }};
}


http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    server_names_hash_bucket_size 128;
    client_max_body_size 2048M;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile        on;
    #tcp_nopush     on;

    keepalive_timeout  65;

    # If we receive X-Forwarded-Proto, pass it through; otherwise, pass along the
    # scheme used to connect to this server
    map $http_x_forwarded_proto $proxy_x_forwarded_proto {
      default $http_x_forwarded_proto;
      ''      $scheme;
    }

    #gzip  on;
    proxy_connect_timeout {{ .Config.ProxyConnectTimeout }};
    proxy_send_timeout {{ .Config.ProxySendTimeout }};
    proxy_read_timeout {{ .Config.ProxyReadTimeout }};
    proxy_set_header        X-Real-IP         $remote_addr;
    proxy_set_header        X-Forwarded-For   $proxy_add_x_forwarded_for;
    proxy_set_header        X-Forwarded-Proto $proxy_x_forwarded_proto;
    proxy_set_header        Host              $http_host;
    send_timeout {{ .Config.SendTimeout }};

    # ssl
    ssl_prefer_server_ciphers on;
    ssl_ciphers {{ .Config.SSLCiphers }};
    ssl_protocols {{ .Config.SSLProtocols }};
    {{ if .Config.DHParam}}ssl_dhparam {{ .Config.DHParamPath }};{{ end }}

    map $http_upgrade $connection_upgrade {
        default upgrade;
        ''      close;
    }

    # default host return 503
    server {
            listen {{ .Config.Port }};
            server_name _;

            location / {
                return 503;
            }

            location /nginx_status {
                stub_status on;
                access_log off;
            }
    }

    {{ range $host := .Hosts }}
    upstream {{ $host.Upstream.Name }} {
        {{ if $host.IPHash }}ip_hash; {{else}}zone {{ $host.Upstream.Name }}_backend 64k;{{ end }}

        {{ range $up := $host.Upstream.Servers }}server {{ $up.Addr }};
        {{ end }}
    }
    {{ range $k, $ctxroot := $host.ContextRoots }}
    upstream ctx{{ $k }} {
        {{ if $host.IPHash }}ip_hash; {{else}}zone ctx{{ $ctxroot.Name }}_backend 64k;{{ end }}
	{{ range $d := $ctxroot.Upstreams }}server {{ $d }};
	{{ end }}
    } {{ end }}

    server {
        listen {{ $host.Port }};
        server_name{{ range $name := $host.ServerNames }} {{ $name }}{{ end }};

	{{ range $ctxroot := $host.ContextRoots }}
	location {{ $ctxroot.Path }} {
	    {{ if $ctxroot.Rewrite }}rewrite ^([^.]*[^/])$ $1/ permanent;
	    rewrite  ^{{ $ctxroot.Path }}/(.*)  /$1 break;{{ end }}
	    proxy_pass http://ctx{{ $ctxroot.Name }};
	}
	{{ end }}

        {{ if $host.SSLOnly }}return 302 https://$server_name$request_uri;{{ else }}
        location / {
            {{ if $host.SSLBackend }}proxy_pass https://{{ $host.Upstream.Name }};{{ else }}proxy_pass http://{{ $host.Upstream.Name }};{{ end }}
        }

        {{ range $ws := $host.WebsocketEndpoints }}
        location {{ $ws }} {
            {{ if $host.SSLBackend }}proxy_pass https://{{ $host.Upstream.Name }};{{ else }}proxy_pass http://{{ $host.Upstream.Name }};{{ end }}
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
        }

        location /nginx_status {
            stub_status on;
            access_log off;
        }

        {{ end }}
    }
    {{ if $host.SSL }}
    server {
        listen {{ $host.SSLPort }};
        ssl on;
        ssl_certificate {{ $host.SSLCert }};
        ssl_certificate_key {{ $host.SSLCertKey }};
        server_name{{ range $name := $host.ServerNames }} {{ $name }}{{ end }};

        location / {
            {{ if $host.SSLBackend }}proxy_pass https://{{ $host.Upstream.Name }};{{ else }}proxy_pass http://{{ $host.Upstream.Name }};{{ end }}
        }

        {{ range $ws := $host.WebsocketEndpoints }}
        location {{ $ws }} {
            {{ if $host.SSLBackend }}proxy_pass https://{{ $host.Upstream.Name }};{{ else }}proxy_pass http://{{ $host.Upstream.Name }};{{ end }}
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
        }

        location /nginx_status {
            stub_status on;
            access_log off;
        }
        {{ end }}
    }
    {{ end }}

    {{ end }} {{/* end context root */}}
    {{ end }} {{/* end host range */}}

    include {{ .Config.ConfigBasePath }}/conf.d/*.conf;
}
`

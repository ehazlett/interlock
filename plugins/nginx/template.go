package nginx

var nginxConfTemplate = `# managed by interlock
user  {{ .User }};
worker_processes  {{ .MaxProcesses }};
worker_rlimit_nofile {{ .RLimitNoFile }};

error_log  /var/log/error.log warn;
pid        {{ .PidPath }};


events {
    worker_connections  {{ .MaxConnections }};
}


http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    server_names_hash_bucket_size 128;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile        on;
    #tcp_nopush     on;

    keepalive_timeout  65;

    #gzip  on;
    proxy_connect_timeout {{ .ProxyConnectTimeout }};
    proxy_send_timeout {{ .ProxySendTimeout }};
    proxy_read_timeout {{ .ProxyReadTimeout }};
    send_timeout {{ .SendTimeout }};

    # ssl
    ssl_ciphers {{ .SSLCiphers }};
    ssl_protocols {{ .SSLProtocols }};

    map $http_upgrade $connection_upgrade {
        default upgrade;
        ''      close;
    }

    # default host return 503
    server {
            listen {{ .Port }};
            return 503;
    }

    {{ range $host := .Hosts }}
    upstream {{ $host.Upstream.Name }} {
        {{ range $up := $host.Upstream.Servers }}server {{ $up.Addr }};
        {{ end }}
    }
    server {
        listen {{ $host.Port }};
        server_name{{ range $name := $host.ServerNames }} {{ $name }}{{ end }};
        {{ if $host.SSLOnly }}return 301 https://$server_name$request_uri;{{ else }}
        location / {
            proxy_pass http://{{ $host.Upstream.Name }};
        }

        {{ range $ws := $host.WebsocketEndpoints }}
        location {{ $ws }} {
            proxy_pass http://{{ $host.Upstream.Name }};
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
        }
        {{ end }}
        {{ end }}
    }
    {{ if $host.SSL }}
    server {
        listen {{ .SSLPort }};
        ssl on;
        ssl_certificate {{ $host.SSLCert }};
        ssl_certificate_key {{ $host.SSLCertKey }};
        server_name{{ range $name := $host.ServerNames }} {{ $name }}{{ end }};

        location / {
            proxy_pass http://{{ $host.Upstream.Name }};
        }

        {{ range $ws := $host.WebsocketEndpoints }}
        location {{ $ws }} {
            proxy_pass http://{{ $host.Upstream.Name }};
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
        }
        {{ end }}
    }
    {{ end }}
    {{ end }}

    include /etc/nginx/conf.d/*.conf;
}
`

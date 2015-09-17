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
    client_max_body_size 2048M;

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
    proxy_set_header        X-Real-IP       $remote_addr;
    proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
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
            {{ if ne .Port 8080 }}
            listen 8080;
            {{ end }}
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
        {{ range $up := $host.Upstream.Servers }}server {{ $up.Addr }};
        {{ end }}
    }
    server {
        listen {{ $host.Port }};
        {{ if ne $host.Port 8080 }}
        listen 8080;
        {{ end }}

        server_name{{ range $name := $host.ServerNames }} {{ $name }}{{ end }};
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
        {{ end }}
    }
    {{ if $host.SSL }}
    server {
        listen {{ .SSLPort }};
        {{ if ne .SSLPort 8443 }}
        listen 8443;
        {{ end }}
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
    {{ end }}

    include /etc/nginx/conf.d/*.conf;
}
`

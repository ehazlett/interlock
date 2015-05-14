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

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile        on;
    #tcp_nopush     on;

    keepalive_timeout  65;

    #gzip  on;

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
        listen {{ $host.ListenPort }};
        server_name{{ range $name := $host.ServerNames }} {{ $name }}{{ end }};

        location / {
            proxy_pass http://{{ $host.Upstream.Name }};
        }
    }
    {{ end }}
}
`

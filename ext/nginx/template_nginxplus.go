package nginx

var nginxPlusConfTemplate = `# managed by interlock
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
    ssl_ciphers {{ .Config.SSLCiphers }};
    ssl_protocols {{ .Config.SSLProtocols }};

    map $http_upgrade $connection_upgrade {
        default upgrade;
        ''      close;
    }

    # default host return 503
    server {
            listen {{ .Config.Port }};
            server_name _;

	    root /usr/share/nginx/html;

    	    location = / {
    	        return 301 /status.html;
    	    }

    	    location = /status.html { }

    	    location /status {
    	        status;
    	    }
    }

    {{ range $host := .Hosts }}
    upstream {{ $host.Upstream.Name }} {
        zone {{ $host.Upstream.Name }}_backend 64k;

        {{ range $up := $host.Upstream.Servers }}server {{ $up.Addr }};
        {{ end }}
    }
    server {
        listen {{ $host.Port }};

        server_name{{ range $name := $host.ServerNames }} {{ $name }}{{ end }};
        {{ if $host.SSLOnly }}return 302 https://$server_name$request_uri;{{ else }}
        location / {
            {{ if $host.SSLBackend }}proxy_pass https://{{ $host.Upstream.Name }};{{ else }}proxy_pass http://{{ $host.Upstream.Name }};{{ end }}
        }

        status_zone {{ $host.Upstream.Name }}_backend;

        {{ range $ws := $host.WebsocketEndpoints }}
        location {{ $ws }} {
            {{ if $host.SSLBackend }}proxy_pass https://{{ $host.Upstream.Name }};{{ else }}proxy_pass http://{{ $host.Upstream.Name }};{{ end }}
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
        }

    	location /status {
    	    status;
    	}

        {{ end }}
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
    {{ end }}

    include {{ .Config.ConfigBasePath }}/conf.d/*.conf;
}
`

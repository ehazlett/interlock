package config

import (
	"testing"
)

var (
	sampleConfig = `
ListenAddr = ":8080"
DockerURL = "unix:///var/run/docker.sock"
`
)

func TestParseConfig(t *testing.T) {
	cfg, err := ParseConfig(sampleConfig)
	if err != nil {
		t.Fatalf("error parsing config: %s", err)
	}

	if cfg.ListenAddr != ":8080" {
		t.Fatalf("expected listen addr :8080; received %s", cfg.ListenAddr)
	}

	if cfg.DockerURL != "unix:///var/run/docker.sock" {
		t.Fatalf("expected docker url unix:///var/run/docker.sock; received %s", cfg.DockerURL)
	}
}

func TestSetConfigDefaults(t *testing.T) {
	cfg := &ExtensionConfig{
		Name: "nginx",
	}

	if err := SetConfigDefaults(cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.MaxConn != 1024 {
		t.Fatalf("expected default max connections of 1024; received %d", cfg.MaxConn)
	}

	if cfg.Port != 80 {
		t.Fatalf("expected default port of 80; received %d", cfg.Port)
	}
}

func TestSetNginxConfigDefaults(t *testing.T) {
	cfg := &ExtensionConfig{
		Name: "nginx",
	}

	if err := SetConfigDefaults(cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.User != "www-data" {
		t.Fatalf("expected default user of www-data; received %d", cfg.User)
	}

	if cfg.WorkerProcesses != 2 {
		t.Fatalf("expected default worker processes of 2; received %d", cfg.WorkerProcesses)
	}

	if cfg.RLimitNoFile != 65535 {
		t.Fatalf("expected default rlimit no file of 65535; received %d", cfg.RLimitNoFile)
	}

	if cfg.ProxyConnectTimeout != 600 {
		t.Fatalf("expected default proxy connect timeout of 600; received %d", cfg.ProxyConnectTimeout)
	}

	if cfg.ProxySendTimeout != 600 {
		t.Fatalf("expected default proxy send timeout of 600; received %d", cfg.ProxySendTimeout)
	}

	if cfg.ProxyReadTimeout != 600 {
		t.Fatalf("expected default proxy read timeout of 600; received %d", cfg.ProxyReadTimeout)
	}

	if cfg.SendTimeout != 600 {
		t.Fatalf("expected default send timeout of 600; received %d", cfg.SendTimeout)
	}

	if cfg.SSLCiphers != "HIGH:!aNULL:!MD5" {
		t.Fatalf("expected default SSL ciphers of HIGH:!aNULL:!MD5; received %d", cfg.SSLCiphers)
	}

	if cfg.SSLProtocols != "SSLv3 TLSv1 TLSv1.1 TLSv1.2" {
		t.Fatalf("expected default SSL protocols of SSLv3 TLSv1 TLSv1.1 TLSv1.2; received %d", cfg.SSLProtocols)
	}
}

func TestSetHAProxyConfigDefaults(t *testing.T) {
	cfg := &ExtensionConfig{
		Name: "haproxy",
	}

	if err := SetConfigDefaults(cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.ConnectTimeout != 5000 {
		t.Fatalf("expected default connect timeout of 5000; received %d", cfg.ConnectTimeout)
	}

	if cfg.ServerTimeout != 10000 {
		t.Fatalf("expected default server timeout of 10000; received %d", cfg.ServerTimeout)
	}

	if cfg.ClientTimeout != 10000 {
		t.Fatalf("expected default client timeout of 10000; received %d", cfg.ClientTimeout)
	}

	if cfg.AdminUser != "admin" {
		t.Fatalf("expected default admin user of admin; received %d", cfg.AdminUser)
	}

	if cfg.AdminPass != "" {
		t.Fatalf("expected default admin password of \"\"; received %d", cfg.AdminPass)
	}

	if cfg.SSLDefaultDHParam != 1024 {
		t.Fatalf("expected default SSL default DH param of 1024; received %d", cfg.SSLDefaultDHParam)
	}

	if cfg.SSLServerVerify != "required" {
		t.Fatalf("expected default SSL server verify of required; received %d", cfg.SSLServerVerify)
	}
}

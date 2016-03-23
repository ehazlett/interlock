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
		Name: "haproxy",
	}

	if err := SetConfigDefaults(cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.ConnectTimeout != 5000 {
		t.Fatalf("expected default connect timeout of 5000; received %d", cfg.ConnectTimeout)
	}

	if cfg.MaxConn != 1024 {
		t.Fatalf("expected default max connections of 1024; received %d", cfg.MaxConn)
	}
}

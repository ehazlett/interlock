package utils

import (
	"testing"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func TestSSLEnabled(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockSSLLabel: "1",
		},
	}

	if !SSLEnabled(cfg) {
		t.Fatal("expected ssl enabled")
	}
}

func TestSSLEnabledNoLabel(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	if SSLEnabled(cfg) {
		t.Fatal("expected ssl disabled")
	}
}

func TestSSLOnly(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockSSLOnlyLabel: "1",
		},
	}

	if !SSLOnly(cfg) {
		t.Fatal("expected ssl only")
	}
}

func TestSSLOnlyNoLabel(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	if SSLOnly(cfg) {
		t.Fatal("expected not ssl only")
	}
}

func TestSSLBackend(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockSSLBackendLabel: "1",
		},
	}

	if !SSLBackend(cfg) {
		t.Fatal("expected ssl backend")
	}
}

func TestSSLBackendNoLabel(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	if SSLBackend(cfg) {
		t.Fatal("expected no ssl backend")
	}
}

func TestSSLCertName(t *testing.T) {
	testCert := "cert.pem"

	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockSSLCertLabel: testCert,
		},
	}

	if SSLCertName(cfg) != testCert {
		t.Fatalf("expected ssl cert %s", testCert)
	}
}

func TestSSLCertNameNoLabel(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	if SSLCertName(cfg) != "" {
		t.Fatal("expected no ssl cert")
	}
}

func TestSSLCertKey(t *testing.T) {
	testKey := "key.pem"

	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{
			ext.InterlockSSLCertKeyLabel: testKey,
		},
	}

	if SSLCertKey(cfg) != testKey {
		t.Fatalf("expected ssl key %s", testKey)
	}
}

func TestSSLCertKeyNoLabel(t *testing.T) {
	cfg := &dockerclient.ContainerConfig{
		Labels: map[string]string{},
	}

	if SSLCertKey(cfg) != "" {
		t.Fatal("expected no ssl key")
	}
}

package utils

import (
	"testing"

	ctypes "github.com/docker/engine-api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func TestSSLEnabled(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockSSLLabel: "1",
		},
	}

	if !SSLEnabled(cfg.Labels) {
		t.Fatal("expected ssl enabled")
	}
}

func TestSSLEnabledNoLabel(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	if SSLEnabled(cfg.Labels) {
		t.Fatal("expected ssl disabled")
	}
}

func TestSSLOnly(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockSSLOnlyLabel: "1",
		},
	}

	if !SSLOnly(cfg.Labels) {
		t.Fatal("expected ssl only")
	}
}

func TestSSLOnlyNoLabel(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	if SSLOnly(cfg.Labels) {
		t.Fatal("expected not ssl only")
	}
}

func TestSSLBackend(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockSSLBackendLabel: "1",
		},
	}

	if !SSLBackend(cfg.Labels) {
		t.Fatal("expected ssl backend")
	}
}

func TestSSLBackendNoLabel(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	if SSLBackend(cfg.Labels) {
		t.Fatal("expected no ssl backend")
	}
}

func TestSSLCertName(t *testing.T) {
	testCert := "cert.pem"

	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockSSLCertLabel: testCert,
		},
	}

	if SSLCertName(cfg.Labels) != testCert {
		t.Fatalf("expected ssl cert %s", testCert)
	}
}

func TestSSLCertNameNoLabel(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	if SSLCertName(cfg.Labels) != "" {
		t.Fatal("expected no ssl cert")
	}
}

func TestSSLCertKey(t *testing.T) {
	testKey := "key.pem"

	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockSSLCertKeyLabel: testKey,
		},
	}

	if SSLCertKey(cfg.Labels) != testKey {
		t.Fatalf("expected ssl key %s", testKey)
	}
}

func TestSSLCertKeyNoLabel(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	if SSLCertKey(cfg.Labels) != "" {
		t.Fatal("expected no ssl key")
	}
}

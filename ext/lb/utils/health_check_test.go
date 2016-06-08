package utils

import (
	"fmt"
	"testing"

	ctypes "github.com/docker/engine-api/types/container"
	"github.com/ehazlett/interlock/ext"
)

func TestHealthCheck(t *testing.T) {
	testCheck := "get /"

	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockHealthCheckLabel: testCheck,
		},
	}

	chk := HealthCheck(cfg)

	if chk != testCheck {
		t.Fatalf("expected %s; received %s", testCheck, chk)
	}
}

func TestHealthCheckNoLabel(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	chk := HealthCheck(cfg)

	if chk != "" {
		t.Fatalf("expected no health check; received %s", chk)
	}
}

func TestHealthCheckInterval(t *testing.T) {
	testInterval := 1000

	cfg := &ctypes.Config{
		Labels: map[string]string{
			ext.InterlockHealthCheckIntervalLabel: fmt.Sprintf("%d", testInterval),
		},
	}

	i, err := HealthCheckInterval(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if i != testInterval {
		t.Fatalf("expected %s; received %s", testInterval, i)
	}
}

func TestHealthCheckIntervalNoLabel(t *testing.T) {
	cfg := &ctypes.Config{
		Labels: map[string]string{},
	}

	i, err := HealthCheckInterval(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if i != DefaultHealthCheckInterval {
		t.Fatalf("expected %s; received %s", DefaultHealthCheckInterval, i)
	}
}

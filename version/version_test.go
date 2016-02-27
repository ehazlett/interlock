package version

import (
	"testing"
)

func TestFullVersion(t *testing.T) {
	version := FullVersion()

	expected := Version + " (" + GitCommit + ")"

	if version != expected {
		t.Fatalf("invalid version returned: %s", version)
	}
}

package version

var (
	Version = "0.3.3"

	// GitCommit will be overwritten automatically by the build system
	GitCommit = "HEAD"
)

func FullVersion() string {
	return Version + " (" + GitCommit + ")"
}

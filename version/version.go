package version

var (
	VERSION = "0.2.3"

	// GITCOMMIT will be overwritten automatically by the build system
	GITCOMMIT = "HEAD"

	FULL_VERSION = VERSION + " (" + GITCOMMIT + ")"
)

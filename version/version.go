package version

var (
	VERSION = "0.2.7"

	// GITCOMMIT will be overwritten automatically by the build system
	GITCOMMIT = "HEAD"

	FULL_VERSION = VERSION + " (" + GITCOMMIT + ")"
)

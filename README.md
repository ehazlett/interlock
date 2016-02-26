# Interlock
Dynamic, event-driven extension system using [Swarm](https://github.com/docker/swarm).  Extensions include HAProxy and Nginx for dynamic load balancing.

NOTE: this is a major refactor from the previous version of Interlock.  The
latest tag (v0.3.3) is the legacy version.  That is kept for legacy purposes.
It is strongly recommended to use the latest release as legacy is no longer
maintained.

# Getting Started
To get started with Interlock view the [Documentation](docs).

# Building
To build a local copy of Interlock, you must have the following:

- Go 1.5+
- Use the Go vendor experiment

You can use the `Makefile` to build the binary.  For example:

`make build`

This will build the binary in `cmd/interlock/interlock`.

There is also a Docker image target in the makefile.  You can build it with
`make image`.

# License
Licensed under the Apache License, Version 2.0. See LICENSE for full license text.

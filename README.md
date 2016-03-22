# Interlock [![Build Status](https://travis-ci.org/ehazlett/interlock.svg?branch=master)](https://travis-ci.org/ehazlett/interlock)
Dynamic, event-driven extension system using [Swarm](https://github.com/docker/swarm).  Extensions include HAProxy and Nginx for dynamic load balancing.

NOTE: this is a major refactor from the previous version of Interlock.  See the
[Release Notes](https://github.com/ehazlett/interlock/releases/tag/v1.0.0) for v1.0 for changes.

The latest tag (v0.3.3) is the legacy version.  The `latest` tag will be
tagged as the new version after a couple of releases to allow for migration
from legacy users.  It is strongly recommended to use the latest release 
as legacy is no longer maintained.

The recommended release is `ehazlett/interlock:1.0`

# Quickstart
For a quick start with Compose, see the [Swarm Example](docs/examples/nginx-swarm-machine).

# Documentation
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

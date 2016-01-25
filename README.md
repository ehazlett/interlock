# Interlock
Dynamic, event-driven service load balancer using [Swarm](https://github.com/docker/swarm).

# Getting Started
To get started with Interlock view the [Documentation](docs/index.md).

# Building
To build a local copy of Interlock, you must have the following:

- Go 1.5+
- Use the Go vendor experiment

You can use the `Makefile` to build the binary.  For example:

`make build`

This will build the binary in `cmd/interlock/interlock`.

There is also a Docker image target in the makefile.  You can build it with
`make image`.

You can also use Docker to build in a container if you do not want to worry
about the host Go setup.  To build in a container run:

`make build-container`

This will build the executable and place in `cmd/interlock/interlock`.  Note: this
executable will be built for Linux so you will either need to build a container
afterword or be using Linux as your host OS to use.

# License
Licensed under the Apache License, Version 2.0. See LICENSE for full license text.

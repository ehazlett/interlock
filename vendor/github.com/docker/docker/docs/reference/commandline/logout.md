<!--[metadata]>
+++
title = "logout"
description = "The logout command description and usage"
keywords = ["logout, docker, registry"]
[menu.main]
parent = "smn_cli"
+++
<![end-metadata]-->

# logout

<<<<<<< HEAD
    Usage: docker logout [SERVER]

    Log out from a Docker registry, if no server is
	specified "https://index.docker.io/v1/" is the default.

      --help          Print usage
=======
```markdown
Usage:  docker logout [SERVER]

Log out from a Docker registry.
If no server is specified, the default is defined by the daemon.

Options:
      --help   Print usage
```
>>>>>>> 12a5469... start on swarm services; move to glade

For example:

    $ docker logout localhost:8080

<!--[metadata]>
+++
title = "unpause"
description = "The unpause command description and usage"
keywords = ["cgroups, suspend, container"]
[menu.main]
parent = "smn_cli"
+++
<![end-metadata]-->

# unpause

<<<<<<< HEAD
    Usage: docker unpause [OPTIONS] CONTAINER [CONTAINER...]

    Unpause all processes within a container

      --help          Print usage
=======
```markdown
Usage:  docker unpause CONTAINER [CONTAINER...]

Unpause all processes within one or more containers

Options:
      --help   Print usage
```
>>>>>>> 12a5469... start on swarm services; move to glade

The `docker unpause` command uses the cgroups freezer to un-suspend all
processes in a container.

See the
[cgroups freezer documentation](https://www.kernel.org/doc/Documentation/cgroup-v1/freezer-subsystem.txt)
for further details.

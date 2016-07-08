<!--[metadata]>
+++
title = "volume rm"
description = "the volume rm command description and usage"
keywords = ["volume, rm"]
[menu.main]
parent = "smn_cli"
+++
<![end-metadata]-->

# volume rm

<<<<<<< HEAD
    Usage: docker volume rm [OPTIONS] VOLUME [VOLUME...]

    Remove a volume

      --help             Print usage
=======
```markdown
Usage:  docker volume rm VOLUME [VOLUME]...

Remove a volume

Aliases:
  rm, remove

Options:
      --help   Print usage
```
>>>>>>> 12a5469... start on swarm services; move to glade

Removes one or more volumes. You cannot remove a volume that is in use by a container.

    $ docker volume rm hello
    hello

## Related information

* [volume create](volume_create.md)
* [volume inspect](volume_inspect.md)
* [volume ls](volume_ls.md)
* [Understand Data Volumes](../../tutorials/dockervolumes.md)

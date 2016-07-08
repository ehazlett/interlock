<!--[metadata]>
+++
title = "stop"
description = "The stop command description and usage"
keywords = ["stop, SIGKILL, SIGTERM"]
[menu.main]
parent = "smn_cli"
+++
<![end-metadata]-->

# stop

<<<<<<< HEAD
    Usage: docker stop [OPTIONS] CONTAINER [CONTAINER...]

    Stop a container by sending SIGTERM and then SIGKILL after a
    grace period

      --help             Print usage
      -t, --time=10      Seconds to wait for stop before killing it
=======
```markdown
Usage:  docker stop [OPTIONS] CONTAINER [CONTAINER...]

Stop one or more running containers

Options:
      --help       Print usage
  -t, --time int   Seconds to wait for stop before killing it (default 10)
```
>>>>>>> 12a5469... start on swarm services; move to glade

The main process inside the container will receive `SIGTERM`, and after a grace
period, `SIGKILL`.

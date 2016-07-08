<!--[metadata]>
+++
title = "node accept"
description = "The node accept command description and usage"
keywords = ["node, accept"]
[menu.main]
parent = "smn_cli"
+++
<![end-metadata]-->

# node accept

<<<<<<< HEAD
    Usage:  docker node accept NODE [NODE...]

    Accept a node in the swarm
=======
```markdown
Usage:  docker node accept NODE [NODE...]

Accept a node in the swarm

Options:
      --help   Print usage
```
>>>>>>> 12a5469... start on swarm services; move to glade

Accept a node into the swarm. This command targets a docker engine that is a manager in the swarm cluster.


```bash
$ docker node accept <node name>
```

## Related information

* [node promote](node_promote.md)
* [node demote](node_demote.md)

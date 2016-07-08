<!--[metadata]>
+++
title = "node promote"
description = "The node promote command description and usage"
keywords = ["node, promote"]
[menu.main]
parent = "smn_cli"
+++
<![end-metadata]-->

# node promote

<<<<<<< HEAD
    Usage:  docker node promote NODE [NODE...]

    Promote a node as manager in the swarm
=======
```markdown
Usage:  docker node promote NODE [NODE...]

Promote a node to a manager in the swarm

Options:
      --help   Print usage
```
>>>>>>> 12a5469... start on swarm services; move to glade

Promotes a node that is pending a promotion to manager. This command targets a docker engine that is a manager in the swarm cluster.


```bash
$ docker node promote <node name>
```

## Related information

* [node accept](node_accept.md)
* [node demote](node_demote.md)

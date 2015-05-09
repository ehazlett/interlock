# External
External is a plugin to allow for read-only external script integration.  A
JSON representation of the events are passed as `stdin` to each script.

# Configuration
The following configuration is available through environment variables:

- `EXTERNAL_PATHS`: Comma separated list of paths to be notified when an event occurs

# Example

Example script in bash:

```bash
#!/bin/bash

read INPUT
echo $INPUT >> /tmp/echo.log
```

This log anything on `stdin` to `/tmp/echo.log`.

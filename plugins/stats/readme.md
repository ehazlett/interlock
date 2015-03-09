# Stats
This plugin reports stats as reported from the Docker stats API.

# Configuration
The following configuration is available through environment variables:

- `STATS_CARBON_ADDRESS`: Carbon receiver address (i.e. `1.2.3.4:2003`)
- `STATS_PREFIX`: Stat prefix (default: `docker.stats`)
- `STATS_IMAGE_NAME_FILTER`: Regex to match against container image name to gather stats (default: `.*` - all containers)

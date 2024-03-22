# minimal-server-monitoring
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Tag](https://img.shields.io/github/v/tag/mcarbonne/minimal-server-monitoring)](https://github.com/mcarbonne/minimal-server-monitoring/tags)
[![Stars](https://img.shields.io/github/stars/mcarbonne/minimal-server-monitoring.svg)](https://github.com/mcarbonne/minimal-server-monitoring)
[![Go Report Card](https://goreportcard.com/badge/github.com/mcarbonne/minimal-server-monitoring)](https://goreportcard.com/report/github.com/mcarbonne/minimal-server-monitoring)

This tool lets you monitor a typical home server running applications in containers and receive alerts on your smartphone. It is designed to be light and simple (no database, no GUI, a single configuration file).


## Features
- run in a container (tested with both docker and podman)
- send notifications to any supported services by [shoutrrr](https://containrrr.dev/shoutrrr/v0.8/services/overview/)
- alert when a container is restarting forever
- alert when a container isn't started
- alert when a target is unreachable (ping)
- alert when available disk space is low
- alert when systemd service is failed
- notify when a container image is updated (provide an alternative to [watchtower](https://containrrr.dev/watchtower/) if you are running podman with podman-auto-update)

## Versioning and packaging
This tool follows [semantic versioning](https://semver.org/).

Pre-built images are available on github packages:
- `ghcr.io/mcarbonne/minimal-server-monitoring:main` (`main` branch)
- `ghcr.io/mcarbonne/minimal-server-monitoring:latest`: latest tagged version
- `ghcr.io/mcarbonne/minimal-server-monitoring:x.x.x`
- `ghcr.io/mcarbonne/minimal-server-monitoring:x.x`
- `ghcr.io/mcarbonne/minimal-server-monitoring:x`

For automatic updates ([watchtower](https://github.com/containrrr/watchtower), [podman-auto-update](https://docs.podman.io/en/latest/markdown/podman-auto-update.1.html)...), using the lastest major tag available (`ghcr.io/mcarbonne/minimal-server-monitoring:1`) is recommanded to avoid breaking changes.

## Minimal configuration
### Bare minimum (container monitoring only, and alerts with shoutrrr)
```
docker run -e MACHINENAME=$(hostname) -e SHOUTRRR=XXXXXXX -v .../cache.json:/app/cache.json -v /var/run/docker.sock:/var/run/docker.sock:ro \
--name minimal-server-monitoring -d ghcr.io/mcarbonne/minimal-server-monitoring:1
```

### Custom config.json
```
docker run \
-v .../config.json:/app/config.json:ro \
-v .../cache.json:/app/cache.json \
-v /var/run/docker.sock:/var/run/docker.sock:ro \
-v /run/systemd:/run/systemd:ro \
--name minimal-server-monitoring -d ghcr.io/mcarbonne/minimal-server-monitoring:1
```

- `-v .../config.json:/app/config.json:ro`: override default configuration file with your settings. Default configuration file is available [here](docker_config.json). Have a look at [example_config.json](example_config.json) for an exhaustive lists of available parameters.
- `-v .../cache.json:/app/cache.json`: persist the cache
- `-v /var/run/docker.sock:/var/run/docker.sock:ro`: give access to the host docker daemon (required for container provider). Use `/run/podman/podman.sock:/var/run/docker.sock:ro` if you are using podman.
- `-v /run/systemd:/run/systemd:ro`: give access to the host systemd (required for systemd provider)

## Internal
```mermaid
flowchart TD
subgraph Scraping
    Storage
      Sc(Schedule scrapers)
      Sc-..->S1 & S2 & S3
      S1("`**Scraper n°1**
      - provider: container
      - scrape_interval: 15s`")
      S2("`**Scraper n°2**
      - provider: ping
      - scrape_interval: 30s`")
      S3(...)
    S1 & S2 & S3 -->SC
    SC{{Collect ScrapeResult}}
    Storage[(Storage)]
    S1 & S2 & S3<-.->Storage
end

SC--"- states\n- messages"-->AlertCenter

subgraph AlertCenter
    AC{{"Generate notifications"}}
    AC--notifications-->F
    F{{Filtering}}
    F--filtered notifications-->G
    G{{Grouping}}
end
G--filtered and grouped notifications-->Notifier
subgraph Notifier
    C{{Send notifications}}
    N1(Shoutrrr)
    N2(...)
    C-->N1
    C-->N2
end
```

### Scraping
Schedule configured scrapers.
Each scraper may emit multiple states and multiple messages.
On contrary to some other monitoring tools, decisions are taken in scrapers (i.e. is metric healthy).

Multiple instances of a given provider may be allowed (depending on provider).

A **State** metric is the combination of a metricId, a state (boolean) and a message.
Example: `metricId: "container_XXXX_state", isHealthy: false, message: "XXXX isn't running"`

A **Message** metric is the combination of a metricId and a message.
Example: `metricId: "container_XXXX_updated", message: "container XXXX was updated ...."`

Providers can persist data using **Storage**, a simple key-value database.

The following providers are implemented :

#### container
- no parameters
- only one instance allowed
- messages (for every running containers):
  - when a container image is updated
- states (for every running containers):
  - container status (check if started)
  - container restart (check if restarting forever)

#### ping
|parameter|description|required|default value|
|-----|-----------|--------|-------------|
|targets|list of ip addresses/hostnames to ping|yes|-|
|retry_count|how many times to retry if ping failed|no|3|

- provide one state: is target reachable.
- multiple instances allowed

#### filesystemusage
|parameter|description|required|default value|
|-----|-----------|--------|-------------|
|mountpoints|list of mount points to check|yes|-|
|threshold_percent|minimum threshold (percentage) of available disk space|no|20|

- provide one state per mountpoint
- multiple instances allowed

#### systemd
- no parameters
- only one instance allowed
- states (for every services):
  - service active state (`ActiveState != failed`)

### AlertCenter
AlertCenter is here to:
- emit notifications from scrape result
- avoid beeing flooded with notifications (filtering + grouping)

#### Generate notifications
If a state is marked as failed `unhealthy_threshold` time in a row, a notification is sent (metric XX failed).
If a state is marked as OK `healthy_threshold` time in a row, a notification is sent (metric XX OK).

Messages are forwared as notifications (no processing at this step).

#### Filtering
Avoid sending too many notifications for a given `metricId`.
Each `metricId` is allowed to send at most 5 messages every 30 minutes.

#### Grouping
When processing a notification, wait up to 15 seconds to group at most 10 notifications.

### Notifier
Send all notifications to all configured notifiers.
Multiple instances of each type are allowed.

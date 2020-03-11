# uva-rp1

Research Project 1

## 30: Automated planning and adaptation of Named Data Networks in cloud environments

[notes](notes)

## Repo

### Directories

- api: contains the GRPC definition for communications between components
- cmd: binary build entrypoint
- dash: cli to scrape status from nodes
- demo: Additional docker build directories for various components
- deploy: Canonical docker build directories and docker-compose test setups
- fileserve: python server/client to serve a pid from filesystem or retrieve from arxiv
- ndn-traffic-generator: used to have a patched version with higher logging resolution
- nfdstat: shared library for interacting with NFD through cli
- notes: assorted notes
- primary: Load Balancer
- secondary: Slave Cache Server
- trafficwrap: test helper used to count stats for ndn-traffic-generator
- watcher: Discovery Server

### api

Defines 2 sets of server-clients: watcher-primary and primary-secondary

`service Reflector` defines the server implemented by **watcher**.
**primary** registers with a watcher with `Primaries`,
sending its own identity and in return getting updates of all known primaries,
keepalive is through keeping the stream open.
**watcher** opens a 2 way stream with another **watcher** through `Gossip`,
_limitation: must be configured one way, no duplicate neighbor detection implemented_

`service Info` defines the server implemented by **primary**.
`Channels` returns the scheme/host/port of the NFD instance it controls,
_limitation: will only return 1_.
`Routes` returns updates of locally routable prefixes.
`SecondaryStatus` waits for **secondary** to connect, then sends requests for NFD status.
`PrimaryStatus` returns the local status with as well as all connected secondaries.
`Register` waits for **secondary** to connect, then sends back control messages (remote primaries the secondary should use as upstreams)

### Primary

Connect to watcher to recv locations of other primaries,
distribute primaries to secondaries (sticky distribution).
Handle connections from secondary,
evenly distributing upstream primaries.

- distributor: finds difference between new and current routes, distributes them to secondaries
- info: channels, routes, status handlers
- localsec: control the load balancer, special case for 0 secondaries
- primary: entrypoint
- register: connects to watcher, receives uodates and triggers route distribution to secondaries
- scraper: scrapes current NFD status

### Secondary

Connect to primary,
receive control directives (list of upstream primaries to connect to),
connect to upstream primaries to get the channel / route updates,
add / remove connections and routes from NFD as necesary.

- control: connect to primary and apply control directives
- secondary: entrypoint
- status: return status

### Watcher

Register primaries,
connect to other watchers,
sends out known primaries to all connected primaries and watchers.

- notify: sends updates to all connected clients
- prom: unused code _originally for prometheus_
- reflector: handle client connections, gossip with other watchers
- watcher: entrypoint

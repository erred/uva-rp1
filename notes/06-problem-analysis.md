# Problem Analysis

## Conceptual

### Algorithm

#### Problem

Given a set of data repositories, an existing (public cloud) network, and usage statistics, design a placement of routers (caches) and links

#### Observations

Data ingest into public clouds is (usually) free

Data egress from public clouds has a fixed per unit cost per region

Data transfer between regions has a fixed per unit cost per region

It is never cheaper to route between regions before egressing from the cloud

#### Solution

A bipartite graph of producers (data repositories) and routers (caches) is optimal for cost.

Placing a producer outside the cloud may be cheaper depending on network costs.

Producers should be placed in regions with lower inter-region costs

Router locations can be selected for cost or latency

## Technical

### Monitoring and Management

#### Problem

Create a web app to visualize usage statistics and the network graph

Create a monitoring solution to feed data into web app

#### Observation

Writing web apps is hard

NFD has a status api and a nfd-status-http server

#### Solution

Utilize Grafana for web app visualization (see grafana-diagram plugin)

Utilize Prometheus or StatsD+Graphite to collect and store data (write adapter from nfd-status-http to common stats format)

### Router / NFD

#### Problem

How to scale routers

#### Observations

NFD uses an in memory ContentStore (cache)

Naive load balancing (ex round robin) between routers results in ineffective cache utilization (effective cache size = cache size of a single router)

#### Solutions

- Ignore problem
- Scale based on machine size
- Name aware sticky (uncached) load balancer for partitioning requests between caches
- Implement shared content store between routers (ex remote KV store, database)

### Algorithm

#### Problem

What does the algorithm actually do

**_How to input regions/networks and data repositories_**

#### Observations

Network layout / connectivity is a non problem

#### Solution (partial)

Decide on machine size (cache efficiency based on usage / cost)

Decide on router location (latency based on usage / cost)

### Auto Deployment

#### Problem

How to automate deployment of NFD

**_Does this need human input (how automated is the process)?_**

#### Observations

Need reusable deployment format

Depends on router scaling choice

Routers may need unique configuration

#### Solutions (partial)

Package in a docker container (AWS ECS?)

Write discovery hub / router config getter (current official solution is manual nfdc face setup)

(NLSR depends on existing faces / connections)

(HTTP / DNS / multicast based producer discovery, for end hosts to find gateways)

Generate config per router (try to minimize)

### Other Information

- Need fresh data: have consumers set **MustBeFresh** flag in Interest to bypass cache

<!-- Docs to Markdown version 1.0β17 -->

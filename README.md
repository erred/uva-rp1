# uva-rp1

Research Project 1

## 30: Automated planning and adaptation of Named Data Networks in cloud environments

[notes](notes)

## Containers

### ndn-collector

contains:

- [graphite-project/docker-graphite-statsd](https://github.com/graphite-project/docker-graphite-statsd)
- [grafana/grafana](https://github.com/grafana/grafana)

### ndn-server

contains:

- ndn-base
- ndn-sidecar

### ndn-router

TODO: implement

### ndn-client

TODO: implement

## Build Containers

not intended to run directly

### ndn-sidecar

contains a NFD monitoring tool for sending StatsD metrics

### ndn-base

contains:

- [ndn-cxx](https://github.com/named-data/ndn-cxx)
- [ndn-tools](https://github.com/named-data/ndn-tools)
- [named-data/NFD](https://github.com/named-data/NFD)
- [named-data/NLSR](https://github.com/named-data/NLSR),
- [named-data/ndn-traffic-generator](https://github.com/named-data/ndn-traffic-generator)

# uva-rp1

Research Project 1 Notes

HackMD: [proposal](https://hackmd.io/fWNmSIkLTHyEzYbGh7GWdA)

## docker containers

### ndn-collector

contains:

- [graphite-project/docker-graphite-statsd](https://github.com/graphite-project/docker-graphite-statsd)
- [grafana/grafana](https://github.com/grafana/grafana)

### ndn-server

contains:

- ndn-base
- ndn-sidecar

### ndn-sidecar

contains a NFD monitoring tool for sending StatsD metrics

### ndn-base

contains:

- [ndn-cxx](https://github.com/named-data/ndn-cxx)
- [ndn-tools](https://github.com/named-data/ndn-tools)
- [named-data/NFD](https://github.com/named-data/NFD)
- [named-data/NLSR](https://github.com/named-data/NLSR),
- [named-data/ndn-traffic-generator](https://github.com/named-data/ndn-traffic-generator)

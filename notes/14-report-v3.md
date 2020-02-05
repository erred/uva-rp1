# report

draft v3

Canonical: [github]

[github]: https://github.com/seankhliao/uva-rp1/blob/master/notes/14-report-v3.md

## Automated Deployment of Named Data Networks

TODO: shove in cloud/federated/...

### Abstract

### Introduction

### Related Work

- general icn work
- icn deployents overview
- cablelabs underlay project
- coordinated caching

### Named Data Networking

- hierarchical naming / routing
- data association by fiat
- pki
- obj permanence
- overlay
- in memory
- prefixtree, single threaded
- go-ndn

### Proof Of Concept

- considerdations: minimal disruption
- control plane separation
- scaling up
- scaling out
- network architecture

### Discussion

As the name NDN suggests, the names are a big part of the design.
The human-readable, hierarchical naming scheme has several implications,
first of which is routing.

- pki: network control
- caching consistency
- names / dedupe
- scaling up
- scaling out
- overlay v underlay
- project / technical maturity
- research qs

### Future Work

The rapid pace of evolution of the NDN project
and its underlying protocol makes it difficult
for software to stay up to date and relevent.
Some degree of standardization or
client libraries at higher abstraction levels
would definitely help.

An implementation of a scalable gossip protocol
over NDN would also be interesting,
particularly if applied to node discovery
as most test networks appear to use a manually
configured core.

### Conclusion

### References

https://envri.eu/research-infrastructures-fair/#1571658120513-37deeabc-258a

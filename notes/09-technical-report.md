# Technical Report

January 2020

## Context

Evaluation of technologies for deploying a decentralised
Content Distribution Network (CDN) in the cloud.

### Requirements

- Transparent caching
- Data integrity
- Multiple / deduplicated names
- Network policies?

## Comments

The need for the same data to be referenced under multiple names,
leads me to think data should instead be referenced by its content
(hash) and a human-friendly translation layer be provided to map
names to hashes. This has the advantage that for the protocol and
applications, the same name will always reference the same data,
vastly simplifying the protocol. This also provides an algorthmic
guarantee against link rot.

IPFS appears to be the most production ready,
having tools to serve large (641TiB) datasets
with clustering, replication, etc...

## Prospective Technologies

### IPFS

InterPlanetary FileSystem

[github: IPFS](https://github.com/ipfs)

- Idea: Immutable, content addressed, distributed filesystem
- Lead developer: Protocol Labs
- Details:
  - each directory / file / block gets own hash/name
  - access data directly through hash
    or direcory-hash/original/file/name.txt
  - distributed hash table / network deduplication
  - mutable IPNS hash referencing immutable IPFS content
  - pretty names through DNSLink /ipns/your.domain/file/name.txt
- Peer / network discovery:
  - gossip through bootstrap peer list
- Reference implementation: Go
- Future development:
  - Blockchain / Trillian based IPNS for historical lookup

### Dat

[github: Dat Project](https://github.com/datproject)

- Idea: Updateable bittorrent with cryptographic keys
- Lead developer: Dat Foundation
- Details:
  - publish data under: dat://write-key-pub/your/file.txt
  - users with a corresponding read key can connect
    to a swarm and retrieve data
    similar to how bittorrent works.
  - built in versioning
- Peer / network discovery:
  - Current: Local broadcast / global DNS
- Reference implementation: javascript
- Future development:
  - multiple writer
  - hyperswarm: replace discovery mechanism

### Academic Torrents

[github: Academic Torrents](https://github.com/AcademicTorrents)

- Idea: Bittorrent for academic data
- Lead Developer: NSF / U.Mass Boston
- Details:
  - Bittorrent
- Peer / network discovery:
  - Bittorrent
  - RSS feed for selective mirroring
- Reference implementation: mixed
- Future development: ?

### ICN / NDN

#### NDN

Named Data Networking

[github: Named Data Networking](https://github.com/named-data)

- Idea: Routing on names with inline caching
- Lead developer: NIST / UCLA
- Details:
  - publish signed data under /some/heirarchical/prefix/file.txt
  - users have their requests routed by name prefix
  - signed data allows for caching
- Peer / network discovery:
  - broadcast / dns local gateway discovery
- Reference implementation: C++17
- Future development: ?

#### hICN

Fast Data Project Hybrid Information Centric Networking

[github: FDio](https://github.com/fdio)

- Idea: IPv6 addressing / routing / forwarding for data
- Lead developer: Cisco
- Details:
  - IPv6, caching mandatory
  - forwarding works similarly to NDN
  - names: 64b routable prefix + 64b data identifier + 32b suffix
- Peer / network discovery:
  - IPv6
- Reference implementation: C++11
- Future development: ?

### Related Projects

#### IPLD

InterPlanetary Linked Data

[github: IPLD](https://github.com/ipld)

- Idea: Semantic Web / Linked Data for hash based, content addressable data networks
  - ex: IPFS, Git, Blockchain

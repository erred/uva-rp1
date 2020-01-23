# NDN Technical Review

January 2020

## Short Version

Experimental technology with evolving protocol.
Tiny ecosystem constantly broken by updates in core projects.
Do not recommend.

## Medium Version

NDN as a protocol was designed as a Layer 3/4 protocol
that could replace IP/UDP/TCP networks
(focus on how data could be reused in routing).

However current implementations use it as an overlay
(mainly over IP/UDP/TCP) for its caching properties.
This is seriously limited by that the protocol and tooling,
which are not designed to be first class CDNs.

## Long Version

### Concept

Name based routing with _inline_ caching.

### Core Libraries

- CCNx superseded by CICN in process of being integrated with NDN-CXX.
- Current development is in `named-data/ndn-cxx` (C++14 with Experimental Extentions).

### Ecosystem

#### NFD

Experimental, modular.
Single threaded, event driven architecture.
No hooks into state changes (ready, online, ...).

Protocol places no limit on packet size,
NFD places a _practical limit_ (hard limit in code) of 8800 bytes.

Caching is limited to in memory cache,
internal implementation tied to shared data structure with FIB / PIB.
Max memory usage is 550MiB in default config
(8800 bytes x 65536 cache size).
Caching observability is limited to cache size, entries, and hits.
No good way to look at cache contents or individual items
(except implementing your own caching strategy?).

#### NLSR

Assumes static set of neighbors,
no peer discovery.

#### Client libraries

Python x2 and Javascript,
built on top of ndn-cxx,
constantly broken from protocol update.

Lacks features available in C++ libraries
needed to implement basic applications.
Most assume a local NFD node with a unix / tcp connection

#### Client tooling

Lacks basic tooling that isn't broken, ex working file server.

Tooling from `named-data` is written in C++ and works,
but focus heavily on traffic generation.

Third party mainly written in Python but either lack basic functionality
or are broken by protocol updates.

There are a lot of papers on NDN for vehicular mobility.

#### Other Implementations

- `go-ndn/nfd`: pure go reimplementation with a focus on operability.
  Last updated 3 years ago, protocol no longer compatible.

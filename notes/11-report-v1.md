# report

draft v1

Canonical: [github](https://github.com/seankhliao/uva-rp1/blob/master/notes/11-report-v1.md)

## Automated planning and adaptation of Named Data Networks in Cloud environments

TODO: come up with appropriate title

- Evaluation of Named Data Networks as Content Distribution Networks in Cloud Environments

### Abstract

TODO: write abstract

This research will focus primarily on the automation for planning and scaling a NDN.

### Introduction

#### Background

Research clouds / SURFsara... ?

Named Data Networking (NDN) is a particular implementation of
Information Centric Networking (ICN).
A clean-slate internet architecture,
it directly addresses content,
while integrating several other features,
such as security, multipath, mobility, and in-network caching.
It is this last point which is of particular interest,
specifically its application in building a Content Distribution Network (CDN).

Specifically this paper will be looking at:

- Theoretical properties that influence CDN design
- Technical properties and maturity of current projects

Additionally, this paper will also include:

- Proof of concept for automated network configuration
- Comparison with other related technologies

### Related Work

- https://zenodo.org/record/3521549
  Planning and Scaling a Named Data Network with Persistent Identifier Interoperability

- Network Architectures for Next Generation Internet Content Distribution
  https://www.omicsonline.org/open-access/network-architectures-for-next-generation-internet-content-distribution-2167-0919.1000e104.pdf

- https://lib.uva.nl/permalink/31UKB_UAM1_INST/c6hauk/ieee_s8821346
  Information-Centric Networking: Research and Standardization Status

- A Reality Check for Content Centric Networking
  http://diegoperino.com/publications/icn06c-perino.pdf

- https://arxiv.org/abs/1810.07229
  Optimal Cache Allocation for Named Data Caching under Network-Wide Capacity Constraint

- https://lib.uva.nl/permalink/31UKB_UAM1_INST/c6hauk/springer_jour10.1007%2Fs11235-018-0433-5
  COD: caching on demand in information-centric networking
- https://arxiv.org/abs/1606.06339
  Optimal Storage Aware Caching Policies for Content-Centric Clouds
- https://arxiv.org/abs/1606.07630
  Caching Strategies for Information Centric Networking: Opportunities and Challenges
- https://arxiv.org/abs/1612.00352
  Performance Evaluation of Caching Policies in NDN - an ICN Architecture
- https://lib.uva.nl/permalink/31UKB_UAM1_INST/c6hauk/ieee_s6883812
  CRCache: Exploiting the correlation between content popularity and network topology information for ICN caching

### NDN as CDN

#### Project Landscape

There are several active projects that fall under the ICN umbrella,
each with a slightly different focus.
The 2 major ones are Community ICN (CICN),
a Fast Data Project (FD.io) under Linux Foundation Networking (LFN),
and Named Data Networking (NDN),
a National Science Foundation Future Internet Architecture
funded project currently led by UCLA.
The remainder of this paper will focus primarily on
NDN and its implementation.

#### Theoretical properties

NDN works by giving every piece of content its own name.

intro, data + req/res

naming: data

naming: discoberability

naming: mutability

naming: heirarchy and order

data: mutability, cache invalidation

#### Technical properties

overlay

cache, scaling up

forwarding

route propagation, network autoconfig, scaling out

#### Automation and Proof of Concept

implementation

experiments

#### Other Technologies

[go-ndn][@gondn] [^gondn] was a 2013 clean implementation of NFD in Go
by a student at UCLA.
This implementation was more minimalist
and performance oriented in design,
performing an order of magnitude faster than the reference implementation.
It could also utilize a persistent, on-disk cache.
Currently not compatible with the reference implementation
dues to changes in the protocol.

The CICN [^cicn] project shares roots with the NDN project,
but is based on a later protocol design with different priorities.
It focuses on being a layer 3 protocol,
leaving features such as content discovery
to be implemented in higher level protocols.
In CICN, the returned data must have an exact match,
as opposed to the optional prefix match for NDN.
Additionally, in contrast with the permanant validity of NDN data packets,
CICN data packets have a lifetime after which they expire.

Bittorrent is a popular protocol distributing large datasets,
due to both its resiliency to network failures
and ability to distribute load between peers.
[Academic Torrents][@academictorrents] [^academictorrents] is a project in the research space
that provides a central point for discovery and collaboration.

InterPlanetary FileSystem (IPFS) [^ipfs] is a content addressed protocol.

### Discusssion

### Conclusion

### Future Work

### References

[@gondn]: https://medium.com/@tailinchu/nfd-vs-go-nfd-d9da283e5d7b
[@academictorrents]: http://academictorrents.com

[^gondn]: https://github.com/go-ndn
[^cicn]: https://wiki.fd.io/view/Cicn
[^academictorrents]: http://academictorrents.com/
[^ipf]: https://ipfs.io/

### Appendix

ICNRQ: https://trac.ietf.org/trac/irtf/wiki/icnrg

draft-irtf-icnrg-terminology-08

draft-irtf-icnrg-flic-02

draft-irtf-icnrg-nrsarch-considerations-03
draft-irtf-icnrg-nrs-requirements-03

draft-irtf-icnrg-deployment-guidelines-07

#### Timeline

- 2006 van Jacobson talk at google: https://www.youtube.com/watch?v=oCZMoY3q2uM
- 2007 CCNx launched at PARC
- 2009 CCNx announced: Networking Named Content
- 2010 renamed NDN as NSF Future Internet Architecture
  funds project with 10 institutions, including PARC
- 2012 PARC splits from NDN due to patents/licensing
- 2013 PARC produces CCNx 1.0
- 2017 Cisco buys CCNx 1.0, development continues under fd.io/cicn
- 2019 Cisco announces hICN

https://icnrg.github.io/draft-icnrg-harmonization/draft-icnrg-harmonization-00.html

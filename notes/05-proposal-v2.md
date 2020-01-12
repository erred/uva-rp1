# RP1/30 Proposal: Network Capacity Planning for Named Data Networks

## Introductuion

- how are NDNs different from traditional networks
  - data centric vs endpoint centric
  - signed data allows caching at all layers, like plaintext HTTP caching
  - special planning needed to leverage caching efficiently in coordination with network design
  - like current CDNs (especially network operated ones) but with greater potential (not limited by control over HTTPS)

## Question

How to plan network (capacity) for Named Data Networks (NDN)

- given known data set / unknown access patterns
- Cache aware network planning
- +geographic distribution
- - cost awareness
- +continuous monitoring + recalculation of optimization / min cost
- +autodeployment

#### In scope

- network design / layout / architecture
  - heirarchical / flat
  - fully / sparse / randomly connected
- cache size / placement
  - more / less / big / small
- multilayer network planning?
  - harder without specific constraints in mind
- demand modelling (timescale)
  - fixed / random / growth
- optimize for
  - deployed bandwidth vs effective bandwidth (QoS?)

#### Out of scope

- anything with linear / integer programming (hopefully)
- spare capacity placement
- coordinated caching
- proactive caching

## Method

TODO: how

## Considerations

TODO: other things

#### Resources

Possible access to a public cloud like Amazon Web Service or Microsogt Azure.

#### Ethics

We do not foresee any ethical concerns.

## Results

TODO: expected deliverables / value

## Previous Research:

#### Networks

- The effect of Web caching on network planning
  - https://doi.org/10.1016/S0140-3664(99)00131-0
  - 1999, 50% cache hit reduces network bandwidth by over half
- Content delivery and caching from a network provider’s perspective
  - https://doi.org/10.1016/j.comnet.2011.07.026
  - 2011, coordinating CDN, P2P, and network providers for higher efficiency
- Tradeoffs in CDN designs for throughput oriented traffic
  - https://doi.org/10.1145/2413176.2413194
  - 2012, CDN design cache points vs network optimization
- Design principles of an operator-owned highly distributed content delivery network
  - https://doi.org/10.1109/MCOM.2013.6495772
  - 2013, network owned CDN planning
- Coded Caching and Storage Planning in Heterogeneous Networks
  - https://doi.org/10.1109/WCNC.2017.7925857
  - 2017, cache and backhaul relation
- QoS-Aware Virtual SDN Network Planning:
  - https://doi.org/10.23919/INM.2017.7987350
  - 2017, network topology and placement of SDN controllers

#### Maybe

- Link Capacity Planning for Fault Tolerant Operation in Hybrid SDN/OSPF Networks
  - https://doi.org/10.1109/GLOCOM.2016.7841957
  - 2016, partial SDN deployment in partitioned OSPF networks reduce spare overhead
- Capacity planning for the Google backbone network
  - https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/45385.pdf
  - cross layer optimization
- Big data caching for networking: moving from cloud to edge:
  - https://doi.org/10.1109/MCOM.2016.7565185
  - 2016, proactive data caching
- Hierarchical Web Caching Systems: Modeling, Design and Experimental Results
  - https://doi.org/10.1109/JSAC.2002.801752
  - 2002, multilayered caching in internet
- Hierarchical placement and network design problems
  - https://doi.org/10.1109/SFCS.2000.892328
  - 2000, heirarchal cache placementJust4Faded
  -
- Probabilistic in-network caching for information-centric networks
  - https://doi.org/10.1145/2342488.2342501
  - 2012, probabilistic caching reduces cache evictions
- Reliability in Layered Networks With Random Link Failures
  - https://doi.org/10.1109/TNET.2011.2143425
  - 2011, estimating failure effects on multilayer networks
- nCDN: CDN enhanced with NDN
  - https://doi.org/10.1109/INFCOMW.2014.6849272
  - 2014, how to use NDNs as CDN
- Planning and learning algorithms for routing in Disruption-Tolerant Networks:
  - https://doi.org/10.1109/MILCOM.2008.4753336
  - 2008, routing algorithms for delay/disruption tolerance
- FAVE: A fast and efficient network Flow AVailability Estimation method with bounded relative error
  - https://doi.org/10.1109/INFOCOM.2019.8737445
  - 2019, network flow capacity estimation through topology+failures

#### Not Relevent

- Cooperative Caching and Transmission Design in Cluster-Centric Small Cell Networks
  - https://doi.org/10.1109/TWC.2017.2682240
  - 2017, coordinated cluster caching
- Cooperative Caching in Wireless P2P Networks: Design, Implementation, and Evaluation
  - https://doi.org/10.1109/TPDS.2009.50
  - 2009, coordinated caching
- Analyzing Peer-To-Peer Traffic Across Large Networks:
  - https://doi.org/10.1109/TNET.2004.826277
  - 2004, monitoring current p2p networks, not really relevent
- Optimal Network Capacity Planning: A Shortest-Path Scheme:
  - https://doi.org/10.1287/opre.23.4.810
  - 1974, use Dijkstra
- Near Optimal Spare Capacity Planning In a Mesh Restorable Network:
  - https://doi.org/10.1109/GLOCOM.1991.188711
  - 1991, spare link placement through Integer/Linear Programming
- Spare Capacity Planning for Survivable Mesh Networks:
  - https://doi.org/10.1007/3-540-45551-5_80
  - 2001, genetic algorithm for link restoration (or path restoration?)
- Cycle-oriented distributed preconfiguration: ring-like speed with mesh-like capacity for self-planning network restoration:
  - https://doi.org/10.1109/ICC.1998.682929
  - 1998, spare link placement and path (re)selection algorithm
- Network planning with random demand:
  - https://doi.org/10.1007/BF02110042
  - 1994, (stochastic) linear programming with random p2p demand
- Rerouting flows when links fail:
  - https://doi.org/10.4230/LIPIcs.ICALP.2017.89
  - 2017, linear programming in rerouting for max flow (vs min cost)
- Bandwidth Modeling and Estimation in Peer to Peer Networks:
  - https://doi.org/10.5121/ijcnc.2010.2306
  - 2010, mathematical model for p2p networks

#### Databases (not really relevent)

- Distributed database network architecture:
  - https://doi.org/10.1016/0140-3664(81)90174-2
  - 1981, too old
- Distributed Database Systems in High-Speed Wide-Area Networks:
  - https://doi.org/10.1109/49.221208
  - 1993, database protocol design utilizing high speed networks
- Designing a distributed database on a local area network: a methodology and decision support system:
  - https://doi.org/10.1016/S0950-5849(99)00056-7
  - 2000, file/workload allocation over LAN

#### Patents (no new info)

- Network capacity planning:
  - https://patents.google.com/patent/US20070067296A1/en
- Network capacity planning:
  - https://patents.google.com/patent/US8296424B2/en
- Network capacity planning:
  - https://patents.google.com/patent/US9065730B2/en
- Multi-layer system capacity planning:
  - https://patents.google.com/patent/US10263705B1/en
- Network capacity planning for multiple instances of an application:
  - https://patents.google.com/patent/US20130046887A1/en
- Network capacity planning based on buffers occupancy monitoring:
  - https://patents.google.com/patent/US6690646B1/en
- Apparatus and method for network capacity evaluation and planning:
  - https://patents.google.com/patent/US6209033B1/en
- Intelligent caching and network management based on location and resource anticipation:
  - https://patents.google.com/patent/US20020198991A1/en

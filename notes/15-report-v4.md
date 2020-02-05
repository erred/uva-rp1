# Automated Deployment and Scaling of Named Data Networks in Cloud Environments

## Abstract

Named Data Networking (NDN) is a clean slate internet architecture
designed from the ground up to put data distribution front and center.

## Introduction

What is motivating problem (cloud)

What is NDN

What is actual problem

What is in scope

- q1
- q2

While caching and routing are fundamental parts of NDN,
this research will be limited to using the strategies available
with the reference implementation of the NDN router.

## Related Work

Internet Draft deployment considerations, cablelabs, ndn testbed

NDN performance and scalability

Caching / Routing strategies

## Named Data Networks

Cloud

Data Model

Scaling Up

Growing Network

Route Distribution

## Proof of Concept

Architecural choices

Scaling choices

Results

## Discussion

### Data Model

Data model, doi, net control

### Efficiency

As caching is an integral part of the network,
its efficiency is understandably of high concern.
The use of advanced caching strategies was left out of scope,
selecting a different one will, affect the following analysis.
The load balancing configuration used in the proof of concept
allows for a theoretical effective cache size
equal to all the caching servers combined.
However, this requires insight into cache hit rates
of individual pieces of content to rebalance routes between caches.
Currently, this would only be possible by either implementing a custom
caching strategy or by reconstructing cache state based on the debug logs.
Instead, the selected solution is a naive partitioning
based on the available upstream load balancing groups,
which leaves it vulnerable to both hot caches
(where all the requests are for a single group)
and reduced efficiency when content is available from multiple routes,
though this is somewhat mitigated by the choice of routing strategy.

With the load balancing configuration above,
there are 2 places where routing strategies can take affect:
how a load balancer routes to its caches, and how a cache routes to its upstreams.
Between a cache and its upstream servers, the choice was made to use
Adaptive Smoothed RTT-based Forwarding Strategy (ASF).
This was to compensate for the lack of configurability in link cost,
instead opting to probe upstreams regularly based on response times.
This opens the possibility of discovering longer paths through intermediary
load balancing groups should the direct path to an upstream fail,
simply by having the proposed intermediaries serve a shared prefix.
Between the load balancer and its caches,
Access Router Strategy was utilized as it could adapt to the rebalancing of caches
without intervention while having better efficiency than pure multicast.
While they are not the most efficient of strategies,
they were selected to retain flexibility during reconfiguations of the network.

Maintaining a separate TCP control channel for each connection
is perhaps not the most efficient of designs in terms of resource usage,
but in exchange it allows for fast reaction times.
Routes can be pushed directly to downstreams as soon as they have changed,
While faults will break the connection, triggering immediate updates,
as opposed to the alternative of waiting for downstreams to discover
the loss either through timeouts or keepalives.

### Fault Tolerance

On a data connection between 2 NDN nodes,
losses can be addressed either in the transport protocol or at the NDN level.
Using TCP as the transport protocol would hide any losses from the NDN protocol
but introduces its own head of line blocking problem similar to what HTTP/2
faces as NDN multiplexes multiple requests over a single TCP stream.
Using UDP exposes the losses to the NDN protocol,
and it would appear that the current implementation
simply retransmits the entire NDN packet, leading to poor performance.
There is an opportunity here to improve the retransmission protocol,
perhaps to something similar to QUIC,
though as of the time of writing, no such efforts are known to exist.

Within the shown proof of concept, there are 3 different types of nodes,
each of which can suffer from faults.
The loss of a discovery server would preclude the addition of any new
load balancing groups to the network,
but the existing network would continue to function normally,
including the propagation of routes and the disconection of a load balancing group.
Losing a caching server within a load balancing group would obviously decrease
the effective cache size of the group,
but would otherwise be handled in the same way as an intentional disconection,
triggering a redistribution of routes between the remaining caching servers.
The loss of a load balancer is however much more disruptive,
as it would disconnect any content hosts and clients from the network.
In this case, the recommended solution would be
to leverage the fact that NDN is not host-centric,
and maintain connections to several load balancing groups,
effectively multi-homing the content.

## Conclusion

## Future Work

## References

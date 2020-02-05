# Automated Deployment and Scaling of Named Data Networks in Cloud Environments

## Abstract

Named Data Networking (NDN) is a clean slate internet architecture
designed from the ground up to put data distribution front and center.
This research looks at some of the potential operational challenges
and solutions in applying NDN to a federated research cloud environment.
Specifically, a load-balanced solution is proposed to mitigate
performance limitations of the NDN router.

## Introduction

The European Open Science Cloud (EOSC) is a
'federated ecosystem of research data infrastructures'.
One such set of research infrastructure is
the ENVironmental Research Infrastructures (ENVRI) community,
itself a federated ecosystem of data infrastructure.
Within these communities, there is an increase in use of
Persistent Identifiers (PID) for naming their Digital Objects (DO).
The need to resolve and access this content across the network
in a way consistent with the Global Digital Object Cloud (DOC) vision,
appeared like a good match with the capabilities of
Named Data Networking (NDN).
Thus inspired this research into the deployment and scaling
of NDN networks for use in cloud environments.

Named Data Networking (NDN) is an implementation of
Information Centric Networking (ICN),
(also known as Content Centric Networking (CCN)),
clean-slate internet architectures that move away from
traditional host-centric IP-based networking,
to one that treats data objects as a first class citizen of the network.
This shift in perspective enables a new set of capabilities,
such as providing data independence from where they are served
while retaining authenticity, and by extension,
widespread use of caching at the network layer.

To apply NDN to solve content distribution problems,
it is necessary to operate such a network.
Based on a survey of the available tools within the ecosystem,
it would appear that this is still largely a manual process.
This research will therefore focus on challenges
and potential solutions to automating the deployment and
scaling of a NDN network with a primary usecase of content distribution.

Specifically:

- What is current landscape in deploying a NDN network?
- What are the challenges in deploying a scalable network?
- What are the feasible solutions to the challenges outlined above?
- How would this apply to operating a federated research cloud?

While caching and forwarding are fundamental parts of NDN,
this research will be limited to using the strategies available
with the reference implementation of the NDN router.

## Related Work

The ICN Research Group (ICNRG) has an Internet Draft on deployment guidelines,
which lists the different possible configurations and known trial deployments.
Amongst the common concerns were scalability,
as trials have been limited to less than 1000 users.

Multiple routing strategies have been proposed,
OSPFN is a modification of OSPF to distrubte NDN names over IP networks,
NLSR is a NDN native implementation of link state routing,
DCR implements distance based routing,
and CRoS-NDN uses a centralised controller to distribute routes based on global state.

## Named Data Networking

NDN utilizes a stateful forwarding model

A network can scale both its individual nodes and the number of nodes.
Scaling an individual node is not as straightforward as it may first appear.
The first problem comes from the cache, which in NFD is in-memory,
combined with the fact that there is simple way to dump and restore cache contents.
Another problem is the fact that the NDN forwarding model is stateful,
so it becomes even more challenging if the scaling is not to disrupt any client connections.

The general solutions are to replace the existing node
or to add new nodes either in front or behind.
Replacing the existing node loses both the cache and forwarding state.
Placing in a new node in front still breaks client connections,
but allows it copy the cache contents
from the old node as they are requested by clients.
Placing a new node behind keeps the current forwarding state and cache,
but may result in duplication of cache contents.
An alternate solution would be to load balance requests between multiple nodes,
this requires a NDN protocol aware load balancer
that can send similar requests to the same node
to effectively utilize the multiple caches without duplication of content.

Scaling the number of nodes within the network can itself be split into 2 parts,
establishing the connections over which to communicate,
and distributing routes
Growing Network

Route Distribution

## Proof of Concept

To solve the issue of scaling up a node,
the load-balancing model described above was implemented.
Practically, this meant running a controller on the load balancing node
responsible for collecting the available routes and distributing them
to caching nodes.
An additional controller was required for the caching nodes to register
with the load balancers, open connections to upstreams, and register routes.
Together they form a load balancing group
which will be the unit of scaling within the network.

Due to time constraints,
the decision was made to implement the central discovery server model
for connecting new load balancing groups to the network.
Load balancers connect to the discovery server, advertising itself
and getting back all the other known load balancers,
which it then distributes to its caching servers to connect to.
Caching servers will directly connect to these load balancers,
treating them as upstream servers.

Running

Architecural choices

Scaling choices

Results

## Discussion

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
though this is somewhat mitigated by the choice of forwarding strategy.

With the load balancing configuration above,
there are 2 places where forwarding strategies can take affect:
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

The selected routing strategy is simplistic,
caching nodes will connect directly to all its upstreams to get the available routes.
It relies heavily on the ASF forwarding strategy to actually select the best path.

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

### Resolving Digital Objects

On the current internet, we rely on centralised infrastructure to
resolve PIDs to data they represent. With an NDN overlay,
the resolvers could instead be replaced by an algorithmic translation
from PIDs to NDN names which could be natively routed
to retrieve the data from within the overlay network.

Data model, doi, net control

The hierarchical naming schem of NDN provides both opportunities and challenges
when applied to content distribution within federated clouds.
While the free flow of data is highly valued by users,
network operators are loth to give up complete control.
In this case, both the data names and the identities used to sign the data
can be used to apply access controls at the network level.

## Conclusion

## Future Work

The choice of a central discovery server was suboptimal in regards to fault tolerance.
Some future implementation may wish to have nodes gossip with each other
to discover new neighbours, utilizing either broadcast, mulitcast, or a bootstrap list
for finding an initial connection.
The proof of concept intentionally utilizes a simple routing scheme
to enforce the direction of data flow,
but once the underlying connections have been established it could
instead hand off routing to some more advanced implementation,
such as NLSR or CRoS-NDN.

## References

https://www.ietf.org/id/draft-irtf-icnrg-deployment-guidelines-07.txt
https://ec.europa.eu/info/sites/info/files/research_and_innovation/knowledge_publications_tools_and_data/documents/ec_rtd_factsheet-open-science_2019.pdf
https://link.springer.com/article/10.1186/s13174-019-0119-6
https://www.rd-alliance.org/group/data-fabric-ig/wiki/global-digital-object-cloud
http://www.envri-fair.eu/

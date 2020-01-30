# report

draft v2

Canonical: [github]

[github]: https://github.com/seankhliao/uva-rp1/blob/master/notes/12-report-v2.md

## Named Data Networks as a Federated Content Distribution Network

## Abstract

Content Distribution Networks...

Named Data Networking (NDN) is...

federated (mulit operator)

This research looks at the conceptual and technical properties of NDN
as applied to planning and operating a federated CDN.

proof of concept...

### Introduction

Research clouds / SUFsara...

In the current internet architecture,
There are at least 2 layers of routing to reach any piece of content,
once across the network to reach the host,
and once in the host to reach the content.
Furthermore, this data can change between diffrerent acceses.
Named Data Networking (NDN) proposes that the 2 layers should be merged,
and that content should not change between accesses.
This, combined with data signing,
should allow for content to become a first class citizen of the network,
addressable and cacheable independent of any particular host.

It is the the last point, cacheability across the network,
that is of interest to building a CDN.
This paper will specifically look at:

- How design elements of NDN interact with CDN design and operation
- Technical implementation of current projects

Additionally, this paper will also include:

- Proof of concept in automated deployment
- General comparison with related technologies.

### Related Work

NDN is a specific implementation of Information Centric Networking (ICN),
there are a few other implementations,
most notable Content Centric Networking (CCN),
together undergoing standarization [@standarization] organized through
Information-Centric Networking Research Group (ICNRG) [@icnrg].
The deployment considerations internet draft [@deployment] lists
the many ways of deploying an ICN network
along with experience reports from existing trials.
CableLabs have demonstrated interest in incremental deployment of CCN
in their CDN, though there are no results as yet [@cablelabs].

Caching is a major selling point for NDN,
as such there is an abundance of research on new strategies
as well as comparisons between them all [@cachestrat].
There is also research in optimal cache placement
across a network [@optimalcache].
Routing is another challenge for NDN,
in part due to the expanded address space,
and there is no lack of research in said direction [@diffuse], [@bloom].

### Named Data Networking

Named Data Networking (NDN) is based on the CCNx 0.x protocol,
which is similar but incompatible with the CCNx 1.0 protocol
used by the Content Centric Networking (CCN) /
Community Information Centric Networking (CICN) project.

#### Design Elements

flat v heirarchal

name resolution v flood v short path

overlay

##### Security

The delegation of address space is done through hierarchically signed keys.
This underpins NDN's ability to validate and cache data.
Operationally, this introduces extra complexity,
especially in a federated system with changing participants
and no single root of trust.
In practice,
we may end up with a similar situation
as current internet certificate authorities.

##### Names and Data

Data is considered valid for a name by the signature of an authorized key,
nothing else ties the data to its name.
This leaves open the possibility that the data served under a name changes.
Combined with NDN's decision not to limit data lifetimes,
this may result in inconsistent results across the network.

#### Technical Implementation

NDN Forwarding Daemon (NFD) is NDN's reference implementation of a router,
and only easily available / tested router.

##### Cache

NFD's cache is completely in-memory.
Combined with practical limitations on packet sizes,
there exists a clear limit on the memory consumption of a router.

##### Single Node Scaling

NFD's single-threaded, event driven architecture,
mean it does not scale with more compute cores.

##### Scaling Out

There are 2 parts to scaling out:
discovering neighbours in the network,
discovering the routes they serve

#### Proof of Concept

Here we designed and implemented a proof of concept
in requiring minimal configuration suitable for

Dynamic adjustment of single node capacity.
NFD's single-threaded architecture leaves little room
for scaling compute-bound bottlenecks.
Instead we can adjust cache capacity.
The in-memory cache and the wish to not disrupt client connections
limits

#### Other Technologies

### Discussion

The deep integration of PKI within the protocol
presents both opportunities and challenges for the network operator.
The challenge comes from needing to deploy and maintain
an appropriate set of root certificates
that cover all traffic that wishes to transit through the network.
This same point allows for a form of identity-based traffic filtering,
something that may be useful in relation with the various data protection laws.

Without a way of enforcing that the data served under a name
remains constant through its entire lifecycle,
combined with the decision to make data permanently valid,
NDN leaves open the possibility of inconsistent results across the network.
The World Wide Web also intended for data to be permanently available,
its current state is partly as a result of not having a way of enforcing that.
Combined with the availability of the cache bypass flag,
the decision to permanently store data may backfire
if application developers decide that caching problems are too hard to deal with.

Scaling a single NDN node is currently not an easy task.
The single-threaded, event-driven architecture limits the utility of more cores.
While the connectionless design of the forwarding table
should in theory allow for greater scalability,
this could be negated if clients elect to use TCP connections for better performance.
The in-memory cache also complicates scaling,
as current computer architectures do not allow for easy adjustment of RAM at runtime.
The proof of concept shown above demontrates that it is certainly possible
to load balance between a cluster to dynamically adjust memory usage
at some cost to cache efficiency,
while other, more performance oriented, implementations of NFD such as go-ndn/nfd
show that these are not insurmountable problems.

Expanding a NDN network involves setting up the initial connections
and propagating the available routes.

proof of concept

other:

### Future Work

### Conclusion

### References

[@standarization]: https://sci-hub.se/10.1109/ACCESS.2019.2938586
[@icnrg]: https://irtf.org/icnrg
[@deployment]: https://tools.ietf.org/id/draft-irtf-icnrg-deployment-guidelines-07.txt
[@cablelabs]: https://www.cablelabs.com/wp-content/uploads/2016/02/Content-Delivery-with-Content-Centric-Networking-Feb-2016.pdf
[@optimalcache]: https://arxiv.org/abs/1810.07229
[@cachestrat]: https://arxiv.org/abs/1606.07630
[@diffuse]: https://arxiv.org/pdf/1804.02752.pdf
[@bloom]: https://arxiv.org/pdf/1702.00340.pdf

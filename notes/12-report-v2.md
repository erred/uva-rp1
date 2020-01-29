# report

draft v2

Canonical: [github](https://github.com/seankhliao/uva-rp1/blob/master/notes/12-report-v2.md)

## Named Data Networks as a Distributed Content Distribution Network

## Abstract

Content Distribution Networks...

Named Data Networking (NDN) is...

This research looks at the conceptual and technical properties of NDN
as applied to planning and operating a distributed CDN.

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
and there is no lack of research in said direction [@diffue], [@bloom].

### References

[@standarization]: https://sci-hub.se/10.1109/ACCESS.2019.2938586
[@icnrg]: https://irtf.org/icnrg
[@deployment]: https://tools.ietf.org/id/draft-irtf-icnrg-deployment-guidelines-07.txt
[@cablelabs]: https://www.cablelabs.com/wp-content/uploads/2016/02/Content-Delivery-with-Content-Centric-Networking-Feb-2016.pdf
[@optimalcache]: https://arxiv.org/abs/1810.07229
[@cachestrat]: https://arxiv.org/abs/1606.07630
[@diffue]: https://arxiv.org/pdf/1804.02752.pdf
[@bloom]: https://arxiv.org/pdf/1702.00340.pdf

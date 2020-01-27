# Presentation

draft v1

Canonical version: [github](https://github.com/seankhliao/uva-rp1/blob/master/notes/10-presentation-v1.md)

## Intro

### Use Case

big data distribution through the cloud

caching.

### Current Tech

Mirrors: not transparent to the user

CDN: centralised, http based

Bittorrent: hard to manage?

### Next Generation Internet

start from scratch, hindsight is 20/20

Buzzwords: scalable, resilient, mobility, efficient

### How does it work

everything has a name: /some/prefix/tp/some/path/to/your/data

1 request (Interest) -> 1 response (Data)

caching: you always know what should be in the response

## Using it

Routing thing as a CDN...

### ndn-cxx

actively developed:
c++14 library, with prototype router, link state routing,
traffic generation tools

### Scale Up

Increase the capacity of a node

- Replace current node: breaks client connections
- Put node in front: breaks client connections
- Put node behind: double cachng? not scalable
- Load balance: yes

### Scale Out

Increase the size of the network

- One big node: single point of failure, non local cache
- Chain, ring: high cache efficiency, not scalable
- Tree: high cache efficiency, long cross tree traversal
- Mesh: easy!

### Automation

Management, Control, Data plane

Old: Boss, sysadmin (you), router

Current: partial automation

Future: Fully automated, Google Zero Touch Network

Native P2P: autodiscovery

### In Practice

problems, demo?

## Conclusion

Ready? no

Should you use it? no

## Ref

- Network Architectures for Next Generation Internet Content Distribution
  https://www.omicsonline.org/open-access/network-architectures-for-next-generation-internet-content-distribution-2167-0919.1000e104.pdf
- A Reality Check for Content Centric Networking
  http://diegoperino.com/publications/icn06c-perino.pdf

# Notes

## Core Concept

- Host Centric Networking

  - Current Internet architecture
  - Addresses (IP) for hosts and locations (URL)
  - Security at the pipe level IPSec / TLS

- Information Centric Networking

  - Single expanded namespace covering layers 3-7
  - Security at the content level
  - Network inline caching

### Problem statement

Inefficient and mutable data distribution

## Content Distribution

- HTTP Signed Exchange SXG
  - Location based names from DNS / URLs
  - on IP TCP/UDP HTTP{1,2,3}
  - signed Web Packages (bundled data)
  - separates data delivery (CDNs, messaging, etc) from data provenance
- Content Distribution Networks CDN
  - Location based names from DNS / URLs
  - on IP TCP/UDP HTTP{1,2,3}
  - heirarchical, centralized
  - peer discovery through DNS + IP anycast
- InterPlanetary File System IPFS
  - Content Addressable Storage
  - Content hash based names
  - Mutable links through DNSLink (DNS TXT) or IPNS (pib-priv key)
  - on IP TCP
  - distributed, peer 2 peer
  - peer discovery through a boostrapped DHT
- DAT
  - Content addressable storage
  - pubkey based author / "host" + standard paths
  - like IPFS + IPNS
  - on IP TCP
  - peer discovery through discovery server (future Hyperswarm?)
- Bittorrent
  - Hash based names
  - on IP UDP
  - distributed, peer 2 peer
  - peer discovery through bootstrapped DHT or trackers
- Community Information Centric Networking CICN
  - FDio/cicn, obseletes CCNx
  - merge with NDN ???
- Hybrid Information Centric Networking hICN
  - FDio/hicn
  - on IPv6 socket extensions
  - Transport layer: stream and datagram options
  - Overlay ICN network
  - peer discovery through routing protocol (name based)
- Named Data Networks NDN
  - replaces multitiered IP / port / DNS / URL / paths with single expanded namespace
  - on layer 2
  - transport layer concerns kicked to application level
  - Ppeer discovery through routing protocol (name based)

## Questions

- How does it deal with large data packets? MTUs?
- Resources that shouldn't be cached? GDPR?

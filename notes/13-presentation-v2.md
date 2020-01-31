# Presentation

draft v2, [github]

[github]: https://github.com/seankhliao/uva-rp1/blob/master/notes/13-presentation-v2.md

## title

TODO: think of title

## Abstract

> Some academics arrive to tell us that (once again) they have Fixed the Internet,
> and (once again) it runs on top of the current actually-working internet,
> and (once again) if you sign up you can communicate
> with as many as twelve other computers.

n-gate.com, in reference to SCION

## Introduction

### Motivation

#### No caching

- Data needs to move
- Inefficient and costly
- same data, moving in the same direction

#### Build a CDN

- works
- centralised

#### Bittorrent

- works
- no control

#### Federated research clouds

- reuse existing infra
- no central point of control

#### Requirements

- overlay on existing infra

### Named Data Networking

- Future Internet Architecure

#### Host / endpoint model

- 2 layers of routing

#### Ignore the host

- route directly to data
- hashtag #serverless

#### expand address space with names

- request response

#### tell everyone not to change data

- caching!

### Scaling up

- still need to scale a node / connections

#### replace

- disruptive

#### new cache in front

- disruptive

#### use any/multicast

- multicast doesn't work well across Internet
- broadcast needs networks

#### new cache behind

- chain of caches
- inefficient duplication

#### load balance caches

- yes

### scale out

grouwing your network

#### broadcast / multicast

- doesn't work
- needs network

#### bootstrap

- works, more complicated

#### central discovery servers

- works, easier

### automation

#### in or out of band

pics

#### ssh example

pics

#### data v control plane

- separataion!
- easier to implement
- rapidly evolving protocol, client libraries are broken

#### final arch

arch diagram

### conclusion

#### does it work

- yes
- prerecorded demo

#### could it be better

- yes
- stability, standardization

#### use it now

- no
- stability, standardization

#### use it in the future

- maybe?
- read report

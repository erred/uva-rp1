## Automated planning and adaptation of Named Data Networks in Cloud environments

Sean Liao

Bob Siemerink

## Introduction

The past year, Named Data Networks (NDN) attracted lots of research interests [1], and demonstrated great potential for improving the data distribution for scientific applications. However, it is still very complex to optimally plan for a NDN data specific infrastructures, where the data repositories and user communities are highly distributed. We are motivated to solve the challenges of optimal NDN infrastructure planning, and help the operators deliver a system that meets the desired requirements.

The research focuses specifically on cloud, since cloud is the future of the IT infrastructure and provides a good basis for automation. The good basis for automation is important, since automation is the status quo of infrastructure provisioning. By providing an optimized and automated solution, operators will be able to more easily deliver a NDN infrastructure, which meets the requirements of their organization and users.

## Research question

- How to plan the capacity and topology of an NDN overlay network in Cloud environment, and to effectively adapt Named Data Networks routers at runtime ?
  - Planning algorithms:
  - Automated deployment: e.g., describe the topology, and parallelise the deployment, if they are across different data centers?
  - Dynamic diagnose and adapt NDN based on performance?
  - How to model demand
  - How to model cost

## Scope

### In scope

- Researching modeling of demand and cost
- Creating a tool to advise operators and auto configure NDN routers

### Out of scope

- Spare capacity planning

## Approach and method

To answer the research question literature review will be executed, So that information about modeling demand and cost is gathered. This information will be used to model an algorithm to aid in NDN operation.

## Requirements

To test the tool and findings, access to a public cloud is required.

## Planning

<table>
  <tr>
   <td>Week 1
   </td>
   <td>Write proposal, explore NDN and Start literature research
   </td>
  </tr>
  <tr>
   <td>Week 2
   </td>
   <td>Algorithm and auto deploy
   </td>
  </tr>
  <tr>
   <td>Week 3
   </td>
   <td>management/GUI
   </td>
  </tr>
  <tr>
   <td>Week 4
   </td>
   <td>Report writing / finish up 
   </td>
  </tr>
  <tr>
   <td>Week 5
   </td>
   <td>Presentation preparation 
   </td>
  </tr>
</table>

## Products

- Research paper
- Planning/auto configuration software

## Ethical considerations

No ethical problems are foreseen

## References

##### **Networks**

- The effect of Web caching on network planning
  - [https://doi.org/10.1016/S0140-3664(99)00131-0](<https://doi.org/10.1016/S0140-3664(99)00131-0>)
- Content delivery and caching from a network provider’s perspective
  - [https://doi.org/10.1016/j.comnet.2011.07.026](https://doi.org/10.1016/j.comnet.2011.07.026)
- Tradeoffs in CDN designs for throughput oriented traffic
  - [https://doi.org/10.1145/2413176.2413194](https://doi.org/10.1145/2413176.2413194)
- Design principles of an operator-owned highly distributed content delivery network
  - [https://doi.org/10.1109/MCOM.2013.6495772](https://doi.org/10.1109/MCOM.2013.6495772)
- Coded Caching and Storage Planning in Heterogeneous Networks
  - [https://doi.org/10.1109/WCNC.2017.7925857](https://doi.org/10.1109/WCNC.2017.7925857)
- QoS-Aware Virtual SDN Network Planning:
  - [https://doi.org/10.23919/INM.2017.7987350](https://doi.org/10.23919/INM.2017.7987350)

##### **Maybe**

- Link Capacity Planning for Fault Tolerant Operation in Hybrid SDN/OSPF Networks
  - [https://doi.org/10.1109/GLOCOM.2016.7841957](https://doi.org/10.1109/GLOCOM.2016.7841957)
- Capacity planning for the Google backbone network
  - [https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/45385.pdf](https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/45385.pdf)
- Big data caching for networking: moving from cloud to edge:
  - [https://doi.org/10.1109/MCOM.2016.7565185](https://doi.org/10.1109/MCOM.2016.7565185)
- Hierarchical Web Caching Systems: Modeling, Design and Experimental Results
  - [https://doi.org/10.1109/JSAC.2002.801752](https://doi.org/10.1109/JSAC.2002.801752)
- Hierarchical placement and network design problems
  - [https://doi.org/10.1109/SFCS.2000.892328](https://doi.org/10.1109/SFCS.2000.892328)
- Probabilistic in-network caching for information-centric networks
  - [https://doi.org/10.1145/2342488.2342501](https://doi.org/10.1145/2342488.2342501)
- nCDN: CDN enhanced with NDN
  - [https://doi.org/10.1109/INFCOMW.2014.6849272](https://doi.org/10.1109/INFCOMW.2014.6849272)
- Planning and learning algorithms for routing in Disruption-Tolerant Networks:
  - [https://doi.org/10.1109/MILCOM.2008.4753336](https://doi.org/10.1109/MILCOM.2008.4753336)
- FAVE: A fast and efficient network Flow AVailability Estimation method with bounded relative error
  - [https://doi.org/10.1109/INFOCOM.2019.8737445](https://doi.org/10.1109/INFOCOM.2019.8737445)

**NDN**

- [1] Named Data Networking
  - https://dl.acm.org/doi/pdf/10.1145/2656877.2656887

<!-- Docs to Markdown version 1.0β17 -->

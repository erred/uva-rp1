PREFIX := seankhliao

.PHONY: all
all: nlsr ndn-traffic-generator

.PHONY: ndn-cxx nfd nlsr ndn-traffic-generator
ndn-cxx:
	docker build -f ndn-cxx/Dockerfile -t ${PREFIX}/ndn-cxx ndn-cxx

nfd: ndn-cxx
	docker build -f nfd/Dockerfile -t ${PREFIX}/nfd nfd

nlsr: nfd
	docker build -f nlsr/Dockerfile -t ${PREFIX}/nlsr nlsr

ndn-traffic-generator: nfd
	docker build -f ndn-traffic-generator/Dockerfile -t ${PREFIX}/ndn-traffic-generator ndn-traffic-generator

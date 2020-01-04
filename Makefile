PREFIX := seankhliao

.PHONY: ndn-cxx nfd
ndn-cxx:
	docker build -f ndn-cxx/Dockerfile -t ${PREFIX}/ndn-cxx ndn-cxx

nfd:
	docker build -f nfd/Dockerfile -t ${PREFIX}/nfd nfd

PREFIX := seankhliao

.PHONY: all
all: ndn-collector ndn-server

.PHONY: ndn-collector
ndn-collector:
	docker build -f ndn-collector/Dockerfile -t ${PREFIX}/ndn-collector ndn-collector

.PHONY: ndn-server
ndn-server: ndn-base ndn-sidecar
	docker build -f ndn-server/Dockerfile -t ${PREFIX}/ndn-server ndn-server

.PHONY: ndn-sidecar
ndn-sidecar:
	docker build -f ndn-sidecar/Dockerfile -t ${PREFIX}/ndn-sidecar ndn-sidecar

.PHONY: ndn-base
ndn-base:
	docker build -f ndn-base/Dockerfile -t ${PREFIX}/ndn-base ndn-base

PREFIX := seankhliao

.PHONY: all
all: ndn-collector ndn-server ndn-router ndn-client

.PHONY: push
push:
	docker push ${PREFIX}/ndn-base
	docker push ${PREFIX}/ndn-collector
	docker push ${PREFIX}/ndn-sidecar
	docker push ${PREFIX}/ndn-server
	docker push ${PREFIX}/ndn-router
	docker push ${PREFIX}/ndn-client

.PHONY: ndn-collector
ndn-collector:
	docker build -f ndn-collector/Dockerfile -t ${PREFIX}/ndn-collector ndn-collector

.PHONY: ndn-server
ndn-server: ndn-base ndn-sidecar
	docker build -f ndn-server/Dockerfile -t ${PREFIX}/ndn-server ndn-server

.PHONY: ndn-router
ndn-router: ndn-base ndn-sidecar
	docker build -f ndn-router/Dockerfile -t ${PREFIX}/ndn-router ndn-router

.PHONY: ndn-client
ndn-client: ndn-base
	docker build -f ndn-client/Dockerfile -t ${PREFIX}/ndn-client ndn-client

.PHONY: ndn-sidecar
ndn-sidecar:
	docker build -f ndn-sidecar/Dockerfile -t ${PREFIX}/ndn-sidecar ndn-sidecar

.PHONY: ndn-base
ndn-base:
	docker build -f ndn-base/Dockerfile -t ${PREFIX}/ndn-base ndn-base

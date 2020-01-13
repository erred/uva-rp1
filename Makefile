PREFIX := seankhliao
DOCKER_CMD = push \
		  pull
COMPOSE_CMD = up \
		  down
IMGS = ndn-server \
	   ndn-router \
	   ndn-client \
	   ndn-grafana \
	   ndn-prometheus

.PHONY: all build
all: build

build: ndn-server ndn-router ndn-client ndn-grafana ndn-prometheus

.PHONY: $(COMPOSE_CMD)
$(COMPOSE_CMD):
	docker-compose $@

.PHONY: $(DOCKER_CMD)
$(DOCKER_CMD):
	for img in $(IMGS); do docker $@ ${PREFIX}/$${img} ; done

.PHONY: ndn-server
ndn-server: ndn-base ndn-sidecar
	docker build -t ${PREFIX}/$@ $@

.PHONY: ndn-router
ndn-router: ndn-base ndn-sidecar
	docker build -t ${PREFIX}/$@ $@

.PHONY: ndn-client
ndn-client: ndn-base
	docker build -t ${PREFIX}/$@ $@

.PHONY: ndn-sidecar
ndn-sidecar:
	docker build -t ${PREFIX}/$@ $@

.PHONY: ndn-base
ndn-base:
	docker build -t ${PREFIX}/$@ $@

.PHONY: ndn-grafana
ndn-grafana:
	docker build -t ${PREFIX}/$@ $@

.PHONY: ndn-prometheus
ndn-prometheus:
	docker build -t ${PREFIX}/$@ $@

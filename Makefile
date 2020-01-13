PREFIX := seankhliao
DOCKER_CMD = push \
		  pull
COMPOSE_CMD = up \
		  down
IMGS = ndn-collector \
	   ndn-server \
	   ndn-router \
	   ndn-client

.PHONY: all
all: ndn-collector ndn-server ndn-router ndn-client

.PHONY: $(COMPOSE_CMD)
$(COMPOSE_CMD):
	docker-compose $@

.PHONY: $(DOCKER_CMD)
$(DOCKER_CMD):
	for img in $(IMGS); do docker $@ ${PREFIX}/$${img} ; done

.PHONY: ndn-collector
ndn-collector:
	docker build  -t ${PREFIX}/$@ $@

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

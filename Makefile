.PHONY: all
all: primary secondary watcher grafana traffic



api/primary.pb.go: api/primary.proto
	go generate --tags=tools ./api/



.PHONY: go
go: api/primary.pb.go
	docker build -t seankhliao/ndn-mesh .
.PHONY: nfd
nfd:
	docker build -t seankhliao/ndn-nfd deploy/nfd
.PHONY: grafana
grafana:
	docker build -t seankhliao/ndn-grafana deploy/grafana


.PHONY: traffic
traffic:
	docker build -t seankhliao/ndn-traffic deploy/traffic
.PHONY: primary
primary: go nfd
	docker build -t seankhliao/ndn-primary deploy/primary
.PHONY: secondary
secondary: go nfd
	docker build -t seankhliao/ndn-secondary deploy/secondary
.PHONY: watcher
watcher: go nfd
	docker build -t seankhliao/ndn-watcher deploy/watcher

.PHONY: push
push:
	docker push seankhliao/ndn-nfd
	docker push seankhliao/ndn-mesh
	docker push seankhliao/ndn-grafana
	docker push seankhliao/ndn-traffic
	docker push seankhliao/ndn-primary
	docker push seankhliao/ndn-secondary
	docker push seankhliao/ndn-watcher
.PHONY: pull
pull:
	docker pull seankhliao/ndn-nfd
	docker pull seankhliao/ndn-mesh
	docker pull seankhliao/ndn-grafana
	docker pull seankhliao/ndn-traffic
	docker push seankhliao/ndn-primary
	docker pull seankhliao/ndn-secondary
	docker pull seankhliao/ndn-watcher

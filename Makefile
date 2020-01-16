.PHONY: all
all: secondary watcher grafana



api/primary.pb.go: api/primary.proto
	go generate --tags=tools ./api/



.PHONY: go
go: api/primary.pb.go
	docker build -t seankhliao/ndn-mesh .
.PHONY: nfd
nfd:
	docker build -t seankhliao/ndn-base deploy/nfd
.PHONY: grafana
grafana:
	docker build -t seankhliao/ndn-grafana deploy/grafana



.PHONY: secondary
secondary: go nfd
	docker build -t seankhliao/ndn-secondary deploy/secondary
.PHONY: watcher
watcher: go nfd
	docker build -t seankhliao/ndn-watcher deploy/watcher


api/primary.pb.go: api/primary.proto
	go generate --tags=tools ./api/

.PHONY: go nfd
go:
	docker build -t seankhliao/ndn-mesh .
nfd:
	docker build -t seankhliao/ndn-base deploy/nfd


.PHONY: secondary
secondary: go nfd
	docker build -t seankhliao/ndn-secondary deploy/secondary

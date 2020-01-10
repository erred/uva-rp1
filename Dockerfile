FROM golang:alpine AS build
WORKDIR /workspace
COPY go.* ./
COPY sidecar sidecar
RUN CGO_ENABLED=0 go build -trimpath -o /bin/sidecar github.com/seankhliao/uva-rp1/sidecar

FROM ndn-server
COPY --from=build /bin/sidecar /bin/
COPY sidecar/start.sh /bin/
ENTRYPOINT start.sh

FROM golang:alpine AS build
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -mod=vendor -o /bin/sidecar github.com/seankhliao/uva-rp1/ndn-sidecar

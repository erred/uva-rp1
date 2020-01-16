FROM golang:alpine

ENV CGO_ENABLED=0
WORKDIR /workspace
COPY . .
# RUN go build -trimpath -mod=vendor -o /bin/primary ./cmd/primary
RUN go build -trimpath -mod=vendor -o /bin/secondary ./cmd/secondary
# RUN go build -trimpath -mod=vendor -o /bin/watcher ./cmd/watcher

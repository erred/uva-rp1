FROM golang:alpine AS build

WORKDIR /workspace
COPY watcher .
RUN CGO_ENABLED=0 go build -trimpath -o /bin/watcher


FROM prom/prometheus:v2.15.2
COPY --from=build /bin/watcher /bin/watcher
COPY prometheus.yml /etc/prometheus/
COPY start.sh /bin/start.sh

ENTRYPOINT /bin/start.sh

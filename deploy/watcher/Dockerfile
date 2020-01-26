FROM seankhliao/ndn-mesh AS build

# FROM prom/prometheus:v2.15.2
# USER root
FROM alpine:3.11.2
RUN apk add --update --no-cache bash runit
COPY conf /
COPY --from=build /bin/watcher /bin/watcher

ENTRYPOINT /entrypoint

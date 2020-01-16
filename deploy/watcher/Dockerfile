FROM seankhliao/ndn-mesh AS build


FROM prom/prometheus:v2.15.2
COPY conf /
COPY --from=build /bin/watcher /bin/watcher

ENTRYPOINT /entrypoint

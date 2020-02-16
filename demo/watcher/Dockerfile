FROM seankhliao/ndn-mesh AS build

FROM alpine:3.11.2
RUN apk add --update --no-cache bash runit
COPY conf /
COPY --from=build /bin/watcher /bin/watcher
ENV PROMETHEUS_DISABLED=1

ENTRYPOINT /entrypoint

FROM seankhliao/ndn-mesh AS build

FROM seankhliao/demo-traffic
COPY conf /
COPY --from=build /bin/primary /bin
ENV NDN_TRAFFIC_SERVER_DISABLED=
ENTRYPOINT /entrypoint

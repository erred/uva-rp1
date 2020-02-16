FROM seankhliao/ndn-mesh AS build

FROM seankhliao/demo-nfd
COPY conf /
COPY --from=build /bin/primary /bin
ENTRYPOINT /entrypoint

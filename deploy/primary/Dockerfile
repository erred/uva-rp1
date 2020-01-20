FROM seankhliao/ndn-mesh AS build

FROM seankhliao/ndn-base
COPY conf /
COPY --from=build /bin/primary /bin
ENTRYPOINT /entrypoint

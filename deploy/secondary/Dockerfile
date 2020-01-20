FROM seankhliao/ndn-mesh AS build

FROM seankhliao/ndn-nfd
COPY conf /
COPY --from=build /bin/secondary /bin
ENTRYPOINT /entrypoint

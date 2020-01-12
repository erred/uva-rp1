FROM seankhliao/ndn-sidecar AS sidecar

FROM seankhliao/ndn-base
COPY --from=sidecar /bin/sidecar /bin
COPY conf /

ENTRYPOINT /entrypoint

FROM sidecar AS sidecar

FROM ndn-base
COPY --from=sidecar /bin/sidecar /bin
COPY conf /

ENV NDN_TRAFFIC_SERVER=
ENTRYPOINT /entrypoint

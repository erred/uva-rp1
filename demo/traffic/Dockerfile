FROM seankhliao/demo-nfd

WORKDIR /workspace
RUN git clone --depth 1 --branch ndn-tools-0.7 https://github.com/named-data/ndn-tools && \
    cd ndn-tools && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf ndn-tools
RUN git clone --recursive --depth 1 https://github.com/named-data/ndn-traffic-generator && \
    cd ndn-traffic-generator && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf ndn-traffic-generator

COPY conf /
ENV NDN_TRAFFIC_SERVER_DISABLED=1 \
    NDN_TRAFFIC_CLIENT_DISABLED=1
ENTRYPOINT /entrypoint

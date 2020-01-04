FROM seankhliao/ndn-cxx

RUN git clone --recursive --depth 1 https://github.com/named-data/ndn-traffic-generator && \
    cd ndn-traffic-generator && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf ndn-traffic-generator

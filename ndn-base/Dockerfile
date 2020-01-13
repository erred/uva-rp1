FROM alpine

WORKDIR /workspace
RUN apk add --no-cache boost openssl sqlite libpcap bash runit
RUN apk add --no-cache --virtual build-deps \
    g++ python3 pkgconfig git bash \
    boost-dev openssl-dev sqlite-dev libpcap-dev

RUN git clone --depth 1 https://github.com/named-data/ndn-cxx && \
    cd ndn-cxx && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    ldconfig /usr/local/lib && \
    cd .. && \
    rm -rf ndn-cxx

RUN git clone --depth 1 https://github.com/named-data/ndn-tools && \
    cd ndn-tools && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf ndn-tools

RUN git clone --recursive --depth 1 https://github.com/named-data/NFD && \
    cd NFD && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf NFD

RUN git clone --recursive --depth 1 https://github.com/named-data/ndn-traffic-generator && \
    cd ndn-traffic-generator && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf ndn-traffic-generator

RUN git clone --depth 1 https://github.com/named-data/ChronoSync && \
    cd ChronoSync && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    ldconfig /usr/local/lib && \
    cd .. && \
    rm -rf ChronoSync
RUN git clone --depth 1 https://github.com/named-data/PSync && \
    cd PSync && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    ldconfig /usr/local/lib && \
    cd .. && \
    rm -rf PSync
RUN git clone --depth 1 https://github.com/named-data/NLSR && \
    cd NLSR && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf NLSR

COPY conf /
ENV NDN_TRAFFIC_SERVER_DISABLED=1 \
    NDN_TRAFFIC_CLIENT_DISABLED=1
ENTRYPOINT /entrypoint

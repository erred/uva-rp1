FROM alpine

WORKDIR /workspace
RUN apk add --no-cache boost openssl sqlite libpcap && \
    apk add --no-cache --virtual build-deps \
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

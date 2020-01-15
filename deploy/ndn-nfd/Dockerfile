FROM alpine:3.11.2

WORKDIR /workspace
RUN apk add --no-cache boost openssl sqlite libpcap bash runit
RUN apk add --no-cache --virtual build-deps \
    g++ python3 pkgconfig git bash \
    boost-dev openssl-dev sqlite-dev libpcap-dev

RUN git clone --depth 1 --branch ndn-cxx-0.7.0 https://github.com/named-data/ndn-cxx && \
    cd ndn-cxx && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    ldconfig /usr/local/lib && \
    cd .. && \
    rm -rf ndn-cxx

RUN git clone --depth 1 --branch NFD-0.7.0 --recursive https://github.com/named-data/NFD && \
    cd NFD && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf NFD

COPY conf /
ENV NDN_AUTOCONFIG_DISABLED=1 \
    NDN_AUTOCONFIG_SERVER_DISABLED=1
ENTRYPOINT /entrypoint

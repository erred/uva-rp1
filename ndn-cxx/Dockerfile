FROM alpine

WORKDIR /workspace
RUN apk add --no-cache g++ python3 boost-dev pkgconfig openssl-dev sqlite-dev git
RUN git clone https://github.com/named-data/ndn-cxx
RUN cd ndn-cxx && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    ldconfig /usr/local/lib && \
    cd .. && \
    rm -rf ndn-cxx

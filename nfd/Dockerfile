FROM seankhliao/ndn-cxx

WORKDIR /workspace
RUN git clone --recursive https://github.com/named-data/NFD && \
    cd NFD && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf NFD
RUN apk del build-deps

COPY start.sh /bin/start.sh
COPY nfd.conf /usr/local/etc/ndn/nfd.conf

ENTRYPOINT /bin/start.sh

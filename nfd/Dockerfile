FROM seankhliao/nfd

WORKDIR /workspace
RUN git clone --recursive https://github.com/named-data/NFD nfd
RUN cd nfd && \
    ./waf configure && \
    ./waf && \
    ./waf install && \
    cd .. && \
    rm -rf nfd

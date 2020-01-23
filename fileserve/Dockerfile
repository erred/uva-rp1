FROM seankhliao/ndn-traffic

WORKDIR /opt/fileserve
# RUN apk add --update --no-cache py3-requests && \
RUN pip3 install -U git+https://github.com/zjkmxy/python-ndn.git && \
    pip3 install requests
COPY conf /
COPY public public
COPY app.py .
COPY client.py .

ENTRYPOINT /entrypoint

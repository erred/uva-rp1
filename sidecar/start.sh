#!/bin/sh

if ! ndnsec-get-default &>/dev/null
then
  ndnsec-keygen /localhost/operator | ndnsec-install-cert -
fi

/usr/local/bin/nfd &
sleep 2
/usr/local/bin/nfd-status-http-server -p 80 -a 0.0.0.0 &
# /usr/local/bin/nfd-autoreg --prefix /example &
/usr/local/bin/ndn-traffic-server ndn-traffic-server.conf &
/bin/sidecar -addrs udp://172.17.0.3:56363 -name hello-world -servers 172.17.0.2:8125

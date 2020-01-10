#!/bin/sh

if ! ndnsec-get-default &>/dev/null
then
  ndnsec-keygen /localhost/operator | ndnsec-install-cert -
fi

/usr/local/bin/nfd &
sleep 2
nfdc face create remote udp://172.17.0.2
nfdc route add /example udp://172.17.0.2
/usr/local/bin/ndn-traffic-client ndn-traffic-client.conf

#!/bin/sh

if ! ndnsec-get-default &>/dev/null
then
  ndnsec-keygen /localhost/operator | ndnsec-install-cert -
fi

poststart() {
    # hopefully initialization will be done by then
    sleep 5

    if [ -f /usr/local/etc/ndn/nfd-init.sh ]; then
        /usr/local/etc/ndn/nfd-init.sh
    fi

    if [ -f /usr/local/etc/ndn/autoconfig.conf ]; then
        /usr/local/bin/ndn-autoconfig -d -c "/usr/local/etc/ndn/autoconfig.conf" &
    fi
}

poststart &
/usr/local/bin/nfd

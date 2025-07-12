#!/bin/bash

getent group manhttpd > /dev/null
if [[ $? -ne 0 ]]; then
  addgroup --system manhttpd
fi

getent passwd manhttpd > /dev/null
if [[ $? -ne 0 ]]; then
  adduser --system --disabled-password --home /var/lib/manhttpd --no-create-home --group manhttpd manhttpd
fi



if [[ ! -d /var/lib/manhttpd ]]; then
  mkdir -p /var/lib/manhttpd
  chown manhttpd:manhttpd /var/lib/manhttpd
  chmod 755 /var/lib/manhttpd
fi

# Create the configuration directory for manhttpd
if [[ ! -d /etc/manhttpd ]]; then
  mkdir -p /etc/manhttpd
fi

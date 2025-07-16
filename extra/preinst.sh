#!/bin/bash

getent group manhttpd > /dev/null
if [ $? -ne 0 ]; then
  addgroup --system manhttpd
fi

getent passwd manhttpd > /dev/null
if [ $? -ne 0 ]; then
  adduser --system --disabled-password --home /var/lib/manhttpd --no-create-home --group manhttpd manhttpd
fi

if [ ! -d /var/lib/manhttpd ]; then
  mkdir -p /var/lib/manhttpd
fi

if [ -d /var/lib/manhttpd ]; then
  chown manhttpd:manhttpd /var/lib/manhttpd
  chmod 0775 /var/lib/manhttpd
fi

if [ ! -d /etc/manhttpd ]; then
  mkdir -p /etc/manhttpd
fi

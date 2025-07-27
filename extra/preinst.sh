#!/bin/bash

getent group manweb > /dev/null
if [ $? -ne 0 ]; then
  addgroup --system manweb
fi

getent passwd manweb > /dev/null
if [ $? -ne 0 ]; then
  adduser --system manweb
  usermod -aG manweb manweb
  usermod --shell /bin/false manweb
  usermod --home /var/lib/manweb manweb
fi

if [ ! -d /var/lib/manweb ]; then
  mkdir -p /var/lib/manweb
fi

if [ -d /var/lib/manweb ]; then
  chown manweb:manweb /var/lib/manweb
  chmod 0775 /var/lib/manweb
fi

if [ ! -d /etc/manweb ]; then
  mkdir -p /etc/manweb
fi

if [ -d /etc/systemd/system/manhttpd.service ]; then
  systemctl disable manhttpd
fi


#!/bin/bash

if [ -f /etc/manhttpd/manhttpd.service ]; then
  ln -sf /etc/manhttpd/manhttpd.service /etc/systemd/system/manhttpd.service
fi

if [ ! -d /var/lib/manhttpd ]; then
  mkdir -p /var/lib/manhttpd
fi

if [ -d /var/lib/manhttpd ]; then
  chown manhttpd:manhttpd /var/lib/manhttpd
  chmod 755 /var/lib/manhttpd
fi

systemctl daemon-reload
systemctl enable manhttpd
systemctl start manhttpd

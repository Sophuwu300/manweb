#!/bin/bash

if [[ -f /etc/manhttpd.service ]]; then
  mv /etc/manhttpd.service /etc/manhttpd/manhttpd.service
fi
if [[ -f /etc/manhttpd/manhttpd.service ]]; then
  ln -s /etc/manhttpd/manhttpd.service /etc/systemd/system/manhttpd.service
fi
if [[ ! -d /var/lib/manhttpd ]]; then
  mkdir -p /var/lib/manhttpd
fi
chown manhttpd:manhttpd /var/lib/manhttpd

if [[ -f /etc/manhttpd/manhttpd.conf ]]; then
  mv /etc/manhttpd/manhttpd.conf /etc/manhttpd/manhttpd.conf.bak
  printf "%s %s\n" "hostname" "$(hostname)" >> /etc/manhttpd/manhttpd.conf
fi

systemctl daemon-reload
systemctl enable manhttpd
systemctl start manhttpd

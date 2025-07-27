#!/bin/bash

if [ -f /etc/manweb/manweb.service ]; then
  ln -sf /etc/manweb/manweb.service /etc/systemd/system/manweb.service
fi

if [ ! -d /var/lib/manweb ]; then
  mkdir -p /var/lib/manweb
fi

if [ -d /var/lib/manweb ]; then
  chown manweb:manweb /var/lib/manweb
  chmod 755 /var/lib/manweb
fi

systemctl daemon-reload
systemctl enable manweb
systemctl start manweb

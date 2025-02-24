#!/bin/bash

if [[ -f /etc/manhttpd.service ]]; then
	ln -s /etc/manhttpd.service /etc/systemd/system/manhttpd.service
fi
systemctl daemon-reload
systemctl enable manhttpd
systemctl start manhttpd

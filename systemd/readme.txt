

nano /lib/systemd/system/perfectsrv.service
>
[Unit]
Description=Perfect Server
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/perfect
Restart=always
LimitMEMLOCK=infinity
TimeoutStopSec=5s

[Install]
WantedBy=multi-user.target
Alias=perfect-server.service

systemctl enable perfectsrv.service
systemctl start perfectsrv.service

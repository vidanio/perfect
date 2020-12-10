

nano /lib/systemd/system/perfectsrv.service
>
[Unit]
Description=Perfect Server
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/autoserver onpremisesrt.com srtserver.com srtliveserver.com srtgateway.com
Restart=always

[Install]
WantedBy=multi-user.target
Alias=perfectsrv.service

systemctl enable perfectsrv.service
systemctl start perfectsrv.service

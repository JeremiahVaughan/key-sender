#!/bin/bash
set -e

target_host=$1

rsync -avzh --rsync-path='sudo rsync' --delete -e ssh ./usb-keyboard-setup.sh "$target_host:/usr/local/bin/usb-keyboard-setup.sh"
ssh $target_host "sudo chmod +x /usr/local/bin/usb-keyboard-setup.sh"
rsync -avzh --rsync-path='sudo rsync' --delete -e ssh ./usb-gadget.service "$target_host:/etc/systemd/system/usb-gadget.service"
ssh $target_host "sudo systemctl daemon-reload"
ssh $target_host "sudo systemctl enable usb-gadget.service"
ssh $target_host "sudo systemctl start usb-gadget.service"
ssh $target_host "sudo systemctl status usb-gadget.service"

if [[ "$2" = "arm32" ]]; then
    GOOS=linux GOARCH=arm GOARM=5 go build -o "/tmp/key-sender"
else 
    GOOS=linux GOARCH=arm64 go build -o "/tmp/key-sender"
fi

rsync -avzh --rsync-path='sudo rsync' --delete -e ssh /tmp/key-sender "$target_host:/usr/local/bin/key-sender"
rsync -avzh --rsync-path='sudo rsync' --delete -e ssh ./key-sender.service "$target_host:/etc/systemd/system/key-sender.service"
ssh $target_host "sudo chown root:root /root/secret.txt"
ssh $target_host "sudo chmod 400 /root/secret.txt"

ssh $target_host "sudo systemctl daemon-reload"
ssh $target_host "sudo systemctl enable key-sender.service"
ssh $target_host "sudo systemctl stop key-sender.service" || true
ssh $target_host "sudo systemctl start key-sender.service"
ssh $target_host "sudo systemctl status key-sender.service"


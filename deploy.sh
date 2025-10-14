#!/bin/bash
set -e

target_host=$1

GOOS=linux GOARCH=arm64 go build -o "/tmp/key-sender"
rsync -avzh --rsync-path='sudo rsync' --delete -e ssh /tmp/key-sender "$target_host:${HOME}/go/bin/key-sender"
rsync -avzh --rsync-path='sudo rsync' --delete -e ssh ./key-sender.service "$target_host:/etc/systemd/system/key-sender.service"
ssh $target_host "sudo chown root:root /root/secret.txt"
ssh $target_host "sudo chmod 400 /root/secret.txt"

ssh $target_host "sudo systemctl daemon-reload"
ssh $target_host "sudo systemctl enable key-sender.service"
ssh $target_host "sudo systemctl stop key-sender.service" || true
ssh $target_host "sudo systemctl start key-sender.service"
ssh $target_host "sudo systemctl status key-sender.service"


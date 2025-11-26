#!/usr/bin/env bash

# cleanup any existing gadget
cd /sys/kernel/config/usb_gadget
if [ -d kb ]; then
  cd kb
  echo "" > UDC 2>/dev/null || true
  cd ..
  rm -rf kb
fi

# giving time for the kernal to finish cleaning up the old gadget
sleep 1

cd /sys/kernel/config/usb_gadget
mkdir -p kb
cd kb

echo 0x1d6b > idVendor  # Linux Foundation
echo 0x0104 > idProduct # Multifunction Composite Gadget
echo 0x0100 > bcdDevice # v1.0.0
echo 0x0200 > bcdUSB    # USB2

echo 0x00 > bDeviceClass
echo 0x00 > bDeviceSubClass
echo 0x00 > bDeviceProtocol

mkdir -p strings/0x409
echo "90898c2300000103" > strings/0x409/serialnumber
echo "CanaKit" > strings/0x409/manufacturer
echo "Raspberry Pi Zero W Rev 1.1" > strings/0x409/product

mkdir -p configs/c.1/strings/0x409
echo "Config 1: ECM network" > configs/c.1/strings/0x409/configuration
echo 250 > configs/c.1/MaxPower

# Add HID function
mkdir -p functions/hid.usb0
echo 0 > functions/hid.usb0/subclass
echo 1 > functions/hid.usb0/protocol
echo 8 > functions/hid.usb0/report_length

# Write your USB keyboard report descriptor
# Reference for this file descriptor: https://randomnerdtutorials.com/raspberry-pi-zero-usb-keyboard-hid/
echo -ne \
'\x05\x01\x09\x06\xa1\x01\x05\x07\x19\xe0\x29\xe7\x15\x00\x25\x01' \
'\x75\x01\x95\x08\x81\x02\x95\x01\x75\x08\x81\x03\x95\x05\x75\x01' \
'\x05\x08\x19\x01\x29\x05\x91\x02\x95\x01\x75\x03\x91\x03\x95\x06' \
'\x75\x08\x15\x00\x25\x65\x05\x07\x19\x00\x29\x65\x81\x00\xc0' \
> functions/hid.usb0/report_desc

ln -s functions/hid.usb0 configs/c.1/

# Activate
ls /sys/class/udc > UDC


Manual required steps:
 - Provision SD card with the Raspberry Pi Imager application.
 - Make changes to the /boot/firmware/config.txt file on the SD card.
    - Note: you may see settings in sections [cm4] and [cm5] that already contain these values, but since raspberry pi zero is not a compute model (cm) these sections are ignored on boot
    - If your not editing the files through the raspberry pi OS then Find your SD card:
        - lsblk
        - For me there were two listings for sda -> sda1 and sda2. sda2 had the disk size that matched my SD card size, but it was actual sda1 that contained the config.txt file I was looking for.
    - append `dtoverlay=dwc2` -- usb gadget module
    - ensure the legacy approach isn't present in the file: `otg_mode=1`
 - Make changes to the /etc/modules file on the SD card.
    - append: `dwc2`
    - append: `libcomposite` -- this is the gadget framework that lets you setup multiple functions at the same time. We are just using dwc2 right now, but most tutorials assume your using multiple so the libcomposite approach appears to be better documented as of right now.
 - confirm both modules are loaded with `lsmod`
 - you can tell the libcomposite module is loaded too because the directory `/sys/kernel/config/usb_gadget` will be created automagically
 - Setup device init script in this project:
    - `sudo cp usb-keyboard-setup.sh /usr/local/bin/`
    - `sudo chmod +x /usr/local/bin/usb-keyboard-setup.sh`
    - `sudo cp usb-gadget.service /etc/systemd/system/usb-gadget.service`
    - `sudo systemctl enable usb-gadget.service`
    - `sudo systemctl start usb-gadget.service`
    - `sudo systemctl status usb-gadget.service`
Install:
    - `go install github.com/JeremiahVaughan/key-sender@latest`

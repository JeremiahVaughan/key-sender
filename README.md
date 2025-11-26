Manual required steps:
 - Provision SD card with the Raspberry Pi Imager application.
 - disable password auth:
   `sudo vim /etc/ssh/sshd_config`
     PasswordAuthentication no                                             
 - Make changes to the /boot/firmware/config.txt file on the SD card.
    - Note: you may see settings in sections [cm4] and [cm5] that already contain these values, but since raspberry pi zero is not a compute model (cm) these sections are ignored on boot
    - If your not editing the files through the raspberry pi OS then Find your SD card:
        - lsblk
        - For me there were two listings for sda -> sda1 and sda2. sda2 had the disk size that matched my SD card size, but it was actual sda1 that contained the config.txt file I was looking for.
    - append `dtoverlay=dwc2` -- usb gadget module
    - ensure the legacy approach isn't present in the file: `otg_mode=1`

 - Make changes to the /etc/modules-load.d/. file on the SD card add:
```
dwc2
libcomposite
```
    - append: `dwc2`
    - append: `libcomposite` -- this is the gadget framework that lets you setup multiple functions at the same time. We are just using dwc2 right now, but most tutorials assume your using multiple so the libcomposite approach appears to be better documented as of right now.
 - confirm both modules are loaded with `lsmod`
 - you can tell the libcomposite module is loaded too because the directory `/sys/kernel/config/usb_gadget` will be created automagically

Install:
    - `edit /etc/systemd/system/key-sender.service to include the desired paths for PASSWORD_16_FILE_LOCATION and PASSWORD_25_FILE_LOCATION or leave blank if you only want one or the other (single password)
    - `./deploy.sh <target-host>`

View logs with:
    - `journalctl -u key-sender.service`

Known issue on Mac. You cannot use a seperate cable to power the rasbperry pi to avoid the slow boot time of the pi zero.
Apparently you can get around this by using a data-only (no power) cable for the data cable. But I don't have one to test with.

Best description I can find:
```
• Macs often refuse to enumerate when the Pi Zero is already sourcing 5V on the USB data port.
  When you power the Pi from the PWR port, the +5V rail is back-fed to the “USB” OTG port, so
  the Mac sees a device that’s already driving VBUS and it treats the link as broken/illegal.
  That’s why it works when the Mac powers the Pi (VBUS comes from the host) but fails when the
  Pi is pre‑powered.
```

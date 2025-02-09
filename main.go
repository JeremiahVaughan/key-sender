package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"github.com/warthog618/go-gpiocdev/device/rpi"
)

var (
	DEV_HID     = "/dev/hidg0"
	REPORT_SIZE = 8

	// HID Key Codes (from USB HID usage tables)
	KEY_A = byte(0x04) // 'a'
	KEY_B = byte(0x05) // 'b'
	KEY_C = byte(0x06) // 'c'

	MOD_NONE = byte(0x00) // No modifier keys
)

var d *debouncer
var f *os.File

func main() {
	log.Println("program started")
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()
	c, err := gpiocdev.NewChip("/dev/gpiochip0")
	if err != nil {
		log.Fatalf("error, when NewChip() for main(). Error: %v", err)
	}
	d = newDebouncer(time.Second * 2)
	l, err := c.RequestLine(rpi.GPIO16, gpiocdev.WithEventHandler(handler), gpiocdev.WithFallingEdge)
	if err != nil {
		log.Fatalf("error, when RequestLine() for main(). Error: %v", err)
	}
	defer l.Close()

	err = gadgetInit()
	if err != nil {
		log.Fatalf("error, when gadgetInit() for main(). Error: %v", err)
	}

	f, err = os.OpenFile(DEV_HID, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening HID device:", err)
		return
	}
	defer f.Close()

	<-ctx.Done()
}

type debouncer struct {
	delay   time.Duration
	timer   *time.Timer
	mu      sync.Mutex
	started bool
}

func newDebouncer(delay time.Duration) *debouncer {
	return &debouncer{
		delay: delay,
	}
}

func (d *debouncer) debounce(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.started {
		return
	}
	d.started = true
	f()

	// Start a new timer that will call the function after the delay
	d.timer = time.AfterFunc(d.delay, func() {
		d.mu.Lock()       // we are locking before the function call because there could be multple
		d.started = false // Allow the next function call to use a new timer
		d.mu.Unlock()
	})
}

func handler(evt gpiocdev.LineEvent) {
	d.debounce(func() {
		// Send multiple keys
        log.Println("about to send keys")
		keys := []byte{KEY_A, KEY_B, KEY_C} // Typing 'abc'

		for _, key := range keys {
			if err := sendKey(f, key); err != nil {
				fmt.Println("Error:", err)
				return
			}
		}

		fmt.Println("Keys sent successfully!")
	})
}

func writeFile(path, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Fatalf("error, when writing to file on path: %s", path)
	}
}

func gadgetInit() error {
	gadgetPath := "/sys/kernel/config/usb_gadget/kb"

	// Create the USB gadget directory
	if err := os.MkdirAll(gadgetPath, 0755); err != nil {
		return fmt.Errorf("error, failed to create gadget directory. Error:", err)
	}

	// Set USB identifiers
	writeFile(filepath.Join(gadgetPath, "idVendor"), "0x1d6b")  // Linux Foundation
	writeFile(filepath.Join(gadgetPath, "idProduct"), "0x0104") // Multifunction Composite Gadget
	writeFile(filepath.Join(gadgetPath, "bcdDevice"), "0x0100") // v1.0.0
	writeFile(filepath.Join(gadgetPath, "bcdUSB"), "0x0200")    // USB2

	// Create strings directory
	stringsPath := filepath.Join(gadgetPath, "strings/0x409")
	os.MkdirAll(stringsPath, 0755)
	writeFile(filepath.Join(stringsPath, "serialnumber"), "90898c2300000100")
	writeFile(filepath.Join(stringsPath, "manufacturer"), "me")
	writeFile(filepath.Join(stringsPath, "product"), "kb")

	// Create configuration
	configPath := filepath.Join(gadgetPath, "configs/c.1")
	os.MkdirAll(filepath.Join(configPath, "strings/0x409"), 0755)
	writeFile(filepath.Join(configPath, "strings/0x409/configuration"), "Config 1: ECM network")
	writeFile(filepath.Join(configPath, "MaxPower"), "250")

	// Create HID function
	hidPath := filepath.Join(gadgetPath, "functions/hid.usb0")
	os.MkdirAll(hidPath, 0755)
	writeFile(filepath.Join(hidPath, "protocol"), "1")
	writeFile(filepath.Join(hidPath, "subclass"), "1")
	writeFile(filepath.Join(hidPath, "report_length"), "8")
	writeFile(filepath.Join(hidPath, "report_desc"),
		"\x05\x01\x09\x06\xa1\x01\x05\x07\x19\xe0\x29\xe7\x15\x00\x25\x01"+
			"\x75\x01\x95\x08\x81\x02\x95\x01\x75\x08\x81\x03\x95\x05\x75\x01"+
			"\x05\x08\x19\x01\x29\x05\x91\x02\x95\x01\x75\x03\x91\x03\x95\x06"+
			"\x75\x08\x15\x00\x25\x65\x05\x07\x19\x00\x29\x65\x81\x00\xc0")

	// Link HID function to configuration
	os.Symlink(hidPath, filepath.Join(configPath, "hid.usb0"))

	// Enable the USB gadget by binding it to a UDC
	udcPath := "/sys/class/udc"
	files, err := os.ReadDir(udcPath)
	if err != nil || len(files) == 0 {
		return errors.New("No UDC device found")
	}

	udcName := files[0].Name()
	writeFile(filepath.Join(gadgetPath, "UDC"), udcName)

	fmt.Println("USB gadget configured successfully.")
    return nil
}

func sendKey(f *os.File, key byte) error {
	// HID report: [Modifier, Reserved, Key1, Key2, Key3, Key4, Key5, Key6]
	press := []byte{MOD_NONE, 0x00, key, 0x00, 0x00, 0x00, 0x00, 0x00}
	release := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// Send key press
	if _, err := f.Write(press); err != nil {
		return fmt.Errorf("failed to send key press: %v", err)
	}

	// Small delay to simulate a real key press
	time.Sleep(50 * time.Millisecond)

	// Send key release
	if _, err := f.Write(release); err != nil {
		return fmt.Errorf("failed to send key release: %v", err)
	}

	return nil
}

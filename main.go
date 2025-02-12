package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"github.com/warthog618/go-gpiocdev/device/rpi"
)

var (
	DEV_HID     = "/dev/hidg0"
	CHIP        = "/dev/gpiochip0"
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
	c, err := gpiocdev.NewChip(CHIP)
	if err != nil {
		log.Fatalf("error, when NewChip() for main(). Error: %v", err)
	}
	log.Println("chip added")
	d = newDebouncer(time.Second * 2)
	l, err := c.RequestLine(rpi.GPIO16, gpiocdev.WithEventHandler(handler), gpiocdev.WithFallingEdge)
	if err != nil {
		log.Fatalf("error, when RequestLine() for main(). Error: %v", err)
	}
	log.Println("event handler registered")
	defer l.Close()

	f, err = os.OpenFile(DEV_HID, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening HID device:", err)
		return
	}
	log.Println("device file opened")

	defer f.Close()

	for {
		select {
		case <-ctx.Done():
		default:
			// test
			time.Sleep(time.Second * 4)
			keys := []byte{KEY_A} // Typing 'a'
			// keys := []byte{KEY_A, KEY_B, KEY_C} // Typing 'abc'

			for _, key := range keys {
				if err := sendKey(f, key); err != nil {
					fmt.Println("Error:", err)
					return
				}
			}
			// test
		}
	}
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
	fmt.Println("handler triggered")
	d.debounce(func() {
		fmt.Println("attempting to send keys!")
		// Send multiple keys
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

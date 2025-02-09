package main

import (
	"log"
	"sync"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"github.com/warthog618/go-gpiocdev/device/rpi"
)

func main() {
	log.Println("program started")
	c, err := gpiocdev.NewChip("/dev/gpiochip0")
	if err != nil {
		log.Fatalf("error, when NewChip() for main(). Error: %v", err)
	}
	l, err := c.RequestLine(rpi.GPIO16, gpiocdev.WithEventHandler(handler), gpiocdev.WithRisingEdge)
	if err != nil {
		log.Fatalf("error, when RequestLine() for main(). Error: %v", err)
	}
	defer l.Close()

	// err := rpio.Open()
	// if err != nil {
	// 	log.Fatalf("error, unable to open rpio. Error: %v", err)
	// }
	// defer rpio.Close()

	// crashing
	// pin := rpio.Pin(25) // physical 22
	// pin := rpio.Pin(16) // physical 36
	// pin.Input()
	// pin.PullUp()
	// pin.Detect(rpio.FallEdge)
	// pollRate := time.Second
	// d := newDebouncer(time.Second * 3)
	// for {
	// 	log.Println("polling")
	// 	if pin.EdgeDetected() {
	// 		log.Println("edge detected")
	// 		d.debounce(handleButtonPress)
	// 	}
	// 	time.Sleep(pollRate)
	// }
}

func handleButtonPress() {
	// todo implement
	log.Println("button pushed")
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

	// If a timer is already running, stop it
	if d.timer != nil {
		d.timer.Stop()
	}

	// Start a new timer that will call the function after the delay
	d.timer = time.AfterFunc(d.delay, func() {
		d.mu.Lock() // we are locking before the function call because there could be multple
		// clients connected, perhaps debugging on two different device clients (i.e., different OS or different browser)
		f()
		d.timer = nil // Allow the next function call to use a new timer
		d.mu.Unlock()
	})
}

func handler(evt gpiocdev.LineEvent) {
	// handle edge event
	log.Println("edge detected")
}

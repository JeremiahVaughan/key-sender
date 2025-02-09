package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"github.com/warthog618/go-gpiocdev/device/rpi"
)

var d *debouncer

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
		log.Println("edge detected")
	})
}

package main

import (
	"log"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

func main() {
	err := rpio.Open()
	if err != nil {
		log.Fatalf("error, unable to open rpio. Error: %v", err)
	}
	defer rpio.Close()

	pin := rpio.Pin(25) // physical 22
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.FallEdge)
	pollRate := time.Second
	debounceTime := time.Second * 3
	bounced := true
	for {
		if pin.EdgeDetected() && bounced {
			handleButtonPress()
			bounced = false
			go func() {
				time.Sleep(debounceTime)
				bounced = true
			}()
		}
		time.Sleep(pollRate)
	}
}

func handleButtonPress() {
	// todo implement
	log.Println("button pushed")
}

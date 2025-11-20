package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os/signal"
	"syscall"

	"github.com/exepirit/meshtastic-go/pkg/meshtastic"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/http"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/serial"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// parse CLI flags
	deviceURLStr := flag.String("device", "serial:/dev/ttyS0", "Device URL (supported schema: serial, http)")
	flag.Parse()
	deviceURL, err := url.Parse(*deviceURLStr)
	if err != nil {
		log.Fatalln("Device URL is not valid")
	}

	// setup connection to device via adapter called HardwareTransport
	log.Println("Connecting to device...")
	var transport meshtastic.HardwareTransport
	switch deviceURL.Scheme {
	case "serial":
		serialTransport, err := serial.NewTransport(deviceURL.Path)
		if err != nil {
			log.Fatalln("Failed to open port:", err)
		}
		defer serialTransport.Close()
		transport = serialTransport
	case "http", "https":
		transport = &http.Transport{URL: deviceURL.String()}
	default:
		log.Fatalln("Unsupported URL scheme", deviceURL.Scheme)
	}

	// connect to device
	device, err := meshtastic.NewConfiguredDevice(ctx, transport)
	if err != nil {
		log.Fatalln("Failed to connect to device:", err)
	}
	log.Println("Connected!")

	// query device state
	state, err := device.Config().GetState(ctx)
	if err != nil {
		log.Fatalln("Failed to query device for its state:", err)
	}

	fmt.Println("Known nodes:")
	for _, nodeInfo := range state.Nodes {
		fmt.Printf("[%*s] %s\n",
			4, nodeInfo.User.ShortName, nodeInfo.User.LongName)
	}
}

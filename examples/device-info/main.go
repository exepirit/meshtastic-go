package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/exepirit/meshtastic-go/pkg/meshtastic"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/connect"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// parse CLI flags
	deviceURL := flag.String("device", "serial:/dev/ttyS0", "Device URL")
	flag.Parse()

	// setup connection to device via adapter called HardwareTransport
	log.Println("Connecting to device...")
	transport, err := connect.NewTransport(*deviceURL)
	if err != nil {
		log.Fatalln("Unable to create device transport:", err)
	}
	defer connect.CloseTransport(transport)

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

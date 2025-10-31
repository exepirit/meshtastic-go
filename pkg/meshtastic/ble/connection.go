package ble

import (
	"context"
	"fmt"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
	"tinygo.org/x/bluetooth"
)

// ConnectMAC connects to a BLE device using the specified MAC address.
func ConnectMAC(ctx context.Context, address string) (Transport, error) {
	return connect(ctx, bluetooth.DefaultAdapter, matchMacAddress(address))
}

// ConnectNamed connects to a BLE device using the specified device name.
func ConnectNamed(ctx context.Context, deviceName string) (Transport, error) {
	return connect(ctx, bluetooth.DefaultAdapter, matchName(deviceName))
}

// connect is an internal function that performs the core logic for connecting
// to a BLE device. It scans for devices using the provided match function,
// connects to the first matching device, and initializes the Transport.
//
// TODO: handle context
func connect(_ context.Context, adapter *bluetooth.Adapter, matchFunc deviceMatchFunc) (Transport, error) {
	candidate := make(chan bluetooth.ScanResult, 1)
	err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if matchFunc(result) {
			_ = adapter.StopScan()
			candidate <- result
		}
	})
	if err != nil {
		return Transport{}, fmt.Errorf("failed to seek device: %w", err)
	}

	result := <-candidate
	device, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
	if err != nil {
		return Transport{}, fmt.Errorf("failed to connect device %s: %w", result.Address, err)
	}

	services, err := device.DiscoverServices([]bluetooth.UUID{MeshBluetoothServiceID})
	switch {
	case err != nil:
		return Transport{}, fmt.Errorf("failed to search MeshBluetoothService: %w", err)
	case len(services) < 1:
		return Transport{}, fmt.Errorf("no MeshBluetoothService on device %s", device.Address)
	}
	service := services[0]

	properties, err := service.DiscoverCharacteristics([]bluetooth.UUID{
		FromRadioPropertyID, ToRadioPropertyID, FromNumPropertyID,
	})
	if err != nil {
		return Transport{}, fmt.Errorf("failed to discover BLE characteristics: %w", err)
	}

	t := Transport{
		device:    device,
		fromRadio: properties[0],
		toRadio:   properties[1],
		fromNum:   properties[2],
		packets:   make(chan *proto.FromRadio, packetsBufferSize),
	}
	return t, t.start()
}

// deviceMatchFunc is a function that determines whether a given ScanResult matches a specific device.
// It is used during the scanning process to filter and identify the desired BLE device.
type deviceMatchFunc func(result bluetooth.ScanResult) bool

func matchMacAddress(address string) deviceMatchFunc {
	mac, err := bluetooth.ParseMAC(address)
	if err != nil {
		return matchNoDevice
	}
	return func(result bluetooth.ScanResult) bool {
		return result.Address.MAC == mac
	}
}

func matchName(name string) deviceMatchFunc {
	return func(result bluetooth.ScanResult) bool {
		return result.LocalName() == name
	}
}

func matchNoDevice(_ bluetooth.ScanResult) bool {
	return false
}

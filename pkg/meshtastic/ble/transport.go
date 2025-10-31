package ble

import (
	"context"
	"errors"
	"fmt"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
	protobuf "google.golang.org/protobuf/proto"
	"log/slog"
	"tinygo.org/x/bluetooth"
)

const (
	packetsBufferSize = 120
)

var (
	// errEmptyQueue is returned when no data is available in the packet queue.
	errEmptyQueue = errors.New("no data in queue")
)

// Transport represents a BLE-based transport layer for communication with a Meshtastic device.
// It manages the Bluetooth connection, data transmission, and reception via BLE characteristics.
type Transport struct {
	device    bluetooth.Device
	fromRadio bluetooth.DeviceCharacteristic
	fromNum   bluetooth.DeviceCharacteristic
	toRadio   bluetooth.DeviceCharacteristic

	// these fields are set and used internally
	packets chan *proto.FromRadio
}

// ReceiveFromRadio receives packet from radio via the BLE connection.
func (t Transport) ReceiveFromRadio(ctx context.Context) (*proto.FromRadio, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-t.packets:
		return p, nil
	}
}

// SendToRadio sends a packet to the radio via the BLE connection.
func (t Transport) SendToRadio(ctx context.Context, packet *proto.ToRadio) error {
	buf, err := protobuf.Marshal(packet)
	if err != nil {
		return fmt.Errorf("marshalling error: %w", err)
	}

	_, err = t.toRadio.WriteWithoutResponse(buf)
	return err
}

// Close disconnects the BLE device.
// After calling Close, the Transport instance must not be used again. A new Transport must be created
// for further operations.
func (t Transport) Close() error {
	_ = t.fromNum.EnableNotifications(nil)
	if t.packets != nil {
		close(t.packets)
	}
	return t.device.Disconnect()
}

// start initializes the transport by requesting the device configuration and setting up
// notifications for incoming data. It is called internally during transport initialization.
func (t Transport) start() error {
	buf, err := protobuf.Marshal(&proto.ToRadio{
		PayloadVariant: &proto.ToRadio_WantConfigId{},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal WantConfigId packet: %w", err)
	}

	// send the WantConfigId packet
	if _, err := t.toRadio.WriteWithoutResponse(buf); err != nil {
		return fmt.Errorf("failed to send WantConfigId packet: %w", err)
	}

	// read packets until an error occurs, or we get a non-empty queue
	var readErr error
	for {
		_, readErr = t.readPacket()
		if readErr != nil {
			break
		}
	}

	// check if the error is not the expected one (errEmptyQueue)
	if !errors.Is(readErr, errEmptyQueue) {
		return fmt.Errorf("unexpected error while reading packets: %w", readErr)
	}

	// enable notifications for fromNum
	return t.fromNum.EnableNotifications(func(_ []byte) {
		t.pullPackets()
	})
}

// pullPackets continuously reads packets from the 'fromRadio' characteristic and sends them
// to the packets channel. It stops when an empty queue error is encountered or on other errors.
func (t Transport) pullPackets() {
	for {
		packet, err := t.readPacket()
		switch {
		case errors.Is(err, errEmptyQueue):
			return
		case err != nil:
			slog.Warn("Read packet from device error", "error", err)
		default:
			if len(t.packets) == packetsBufferSize {
				<-t.packets
			}
			t.packets <- packet
		}
	}
}

// readPacket reads a single packet from the 'fromRadio' characteristic.
func (t Transport) readPacket() (*proto.FromRadio, error) {
	var buf []byte
	n, err := t.fromRadio.Read(buf)
	switch {
	case err != nil:
		return nil, err
	case n < 1:
		return nil, errEmptyQueue
	}

	packet := new(proto.FromRadio)
	if err = protobuf.Unmarshal(buf, packet); err != nil {
		return nil, meshtastic.ErrInvalidPacketFormat
	}
	return packet, nil
}

package meshtastic

import (
	"fmt"
	"go.bug.st/serial"
)

// NewSerialTransport creates a new SerialTransport instance for the given serial port.
// It opens the specified serial port with default settings (115200 baud rate).
func NewSerialTransport(port string) (*StreamTransport, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	p, err := serial.Open(port, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	return &StreamTransport{Stream: p}, nil
}

// Package connect provides a unified way to create device transport instances
// using simple URL strings, automatically handling the details of each
// transport type. This provides a consistent interface regardless of the
// underlying communication protocol.
//
// The package supports device transport types through a simple URL scheme:
//
// Network Transports:
//   - HTTP:    "http://server:8080"
//
// Local Transports:
//   - Serial:  "serial:/dev/ttyUSB0"
//   - BLE:     "ble://00:1A:7D:DA:71:13"
//
// Example:
//
//	transport, err := connect.NewTransport("serial:/dev/ttyUSB0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer connect.CloseTransport(transport)
package connect

import (
	"context"
	"fmt"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic"
	"io"
	"net/url"
)

// NewTransport creates a new device transport based on the provided connection string.
// The connection string must be a URL with a scheme specifying the transport type.
//
// Supported schemes:
//
//   - http, https: HTTP/HTTPS transport. Example: "http://localhost:8080" or "https://meshtastic.example.com"
//   - serial: Serial port transport. Example: "serial:/dev/ttyUSB0" or "serial:COM3"
//   - ble: Bluetooth Low Energy transport. Example: "ble://00:1A:7D:DA:71:13"
//
// The function automatically selects the appropriate transport implementation
// based on the URL scheme and returns a [meshtastic.HardwareTransport] interface that can be used
// to communicate with device.
//
// Some transports (BLE, Serial) need to be closed after use. Use the [CloseTransport] function
// to properly clean up resources.
func NewTransport(connectionString string) (meshtastic.HardwareTransport, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	switch u.Scheme {
	case "ble":
		return connectBLE(u)
	case "http", "https":
		return connectHTTP(u), nil
	case "serial":
		return connectSerial(u)
	default:
		return nil, fmt.Errorf("unsupported scheme %q", err)
	}
}

// CloseTransport closes the transport if it implements [io.Closer].
//
// It is safe to call Close on any transport, including those that don't need closing.
// Transports that implement [io.Closer] will be properly closed, others will be ignored.
func CloseTransport(transport meshtastic.HardwareTransport) error {
	if closer, ok := transport.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// NewMeshTransport creates a new mesh transport based on the provided connection string.
//
// Supported schemes:
//   - http, https: HTTP/HTTPS transport. Example: "http://localhost:8080" or "https://meshtastic.example.com"
//   - serial: Serial port transport. Example: "serial:/dev/ttyUSB0" or "serial:COM3"
//   - ble: Bluetooth Low Energy transport. Example: "ble://00:1A:7D:DA:71:13"
//   - udp: UDP transport. Example: "udp://localhost:4403" or "udp://192.168.1.100:4403"
//
// The function automatically selects the appropriate transport implementation
// based on the URL scheme and returns a [meshtastic.MeshTransport] interface that can be used
// to communicate with the mesh network.
//
// Transports need to be closed after use. Use the [CloseMeshTransport] function
// to properly clean up resources.
func NewMeshTransport(ctx context.Context, connectionString string) (meshtastic.MeshTransport, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	switch u.Scheme {
	case "http", "https", "serial", "ble":
		hardwareTransport, err := NewTransport(connectionString)
		if err != nil {
			return nil, err
		}
		return meshtastic.NewConfiguredDevice(ctx, hardwareTransport)
	case "udp":
		return connectUDP(u)
	default:
		return nil, fmt.Errorf("unsupported scheme for mesh transport %q", u.Scheme)
	}
}

// CloseMeshTransport closes the mesh transport if it implements [io.Closer].
//
// It is safe to call [CloseMeshTransport] on any mesh transport, including those that don't need closing.
// Transports that implement [io.Closer] will be properly closed, others will be ignored.
func CloseMeshTransport(transport meshtastic.MeshTransport) error {
	if closer, ok := transport.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

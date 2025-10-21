package meshtastic

import (
	"context"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
)

// Device represents a device, encapsulating the transport used to communicate with the hardware.
type Device struct {
	Transport Transport
}

// SendToMesh sends a mesh packet over the device's transport.
// It converts the provided MeshPacket into a ToRadio message with the appropriate payload variant.
func (d *Device) SendToMesh(ctx context.Context, packet *proto.MeshPacket) error {
	return d.Transport.SendToRadio(ctx, &proto.ToRadio{
		PayloadVariant: &proto.ToRadio_Packet{
			Packet: packet,
		},
	})
}

// ReceiveFromMesh blocks until a mesh packet is received from the device's transport.
// It continuously listens for incoming frames and returns the first MeshPacket found.
// Other packets will be ignored.
func (d *Device) ReceiveFromMesh(ctx context.Context) (*proto.MeshPacket, error) {
	for {
		frame, err := d.Transport.ReceiveFromRadio(ctx)
		if err != nil {
			return nil, err
		}

		if packet := frame.GetPacket(); packet != nil {
			return packet, nil
		}
	}
}

// Config returns a configuration module for the device.
func (d *Device) Config() *DeviceModuleConfig {
	return &DeviceModuleConfig{transport: d.Transport}
}

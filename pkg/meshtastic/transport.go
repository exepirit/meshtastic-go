package meshtastic

import (
	"context"

	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
)

// HardwareTransport defines methods for sending and receiving packets directly to and from the radio hardware.
type HardwareTransport interface {
	// SendToRadio sends a packet to the radio hardware.
	SendToRadio(ctx context.Context, packet *proto.ToRadio) error
	// ReceiveFromRadio receives a packet from the radio hardware.
	ReceiveFromRadio(ctx context.Context) (*proto.FromRadio, error)
}

// MeshTransport defines methods for sending and receiving mesh packets over the network,
// abstracting the underlying radio communication.
type MeshTransport interface {
	PacketSender
	PacketReceiver
}

// PacketSender defines the interface for sending mesh packets.
type PacketSender interface {
	// SendToMesh sends a mesh packet to the network.
	SendToMesh(ctx context.Context, packet *proto.MeshPacket) error
}

// PacketReceiver defines the interface for receiving mesh packets.
type PacketReceiver interface {
	// ReceiveFromMesh receives a mesh packet from the network.
	ReceiveFromMesh(ctx context.Context) (*proto.MeshPacket, error)
}

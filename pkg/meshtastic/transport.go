package meshtastic

import (
	"context"
	"github.com/exepirit/meshtastic_exporter/pkg/meshtastic/proto"
)

// Transport defines methods for sending and receiving packets directly to and from the radio hardware.
type Transport interface {
	// SendToRadio sends a packet to the radio hardware.
	SendToRadio(ctx context.Context, packet *proto.ToRadio) error
	// ReceiveFromRadio receives a packet from the radio hardware.
	ReceiveFromRadio(ctx context.Context) (*proto.FromRadio, error)
}

// MeshTransport defines methods for sending and receiving mesh packets over the network,
// abstracting the underlying radio communication.
type MeshTransport interface {
	// SendToMesh sends a mesh packet to the network.
	SendToMesh(ctx context.Context, packet *proto.MeshPacket) error
	// ReceiveFromMesh receives a mesh packet from the network.
	ReceiveFromMesh(ctx context.Context) (*proto.MeshPacket, error)
}

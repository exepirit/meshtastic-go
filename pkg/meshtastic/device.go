package meshtastic

import (
	"context"
	"fmt"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
	"math/rand"
)

// NewConfiguredDevice creates a new Device instance with a given transport and initializes
// it by retrieving the device's configuration from the hardware.
func NewConfiguredDevice(ctx context.Context, transport Transport) (*Device, error) {
	d := new(Device)
	d.Transport = transport
	config, err := d.Config().GetState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device configuration: %w", err)
	}
	d.NodeID = config.MyInfo.MyNodeNum
	return d, nil
}

// Device represents a device, encapsulating the transport used to communicate with the hardware.
type Device struct {
	Transport Transport
	NodeID    uint32

	lastPacketID uint32
}

// SendToMesh sends a mesh packet over the device's transport.
// It converts the provided MeshPacket into a ToRadio message with the appropriate payload variant.
func (d *Device) SendToMesh(ctx context.Context, packet *proto.MeshPacket) error {
	packet.From = d.NodeID

	if packet.Id == 0 {
		packet.Id = d.generatePacketID()
	}

	if packet.HopLimit == 0 {
		packet.HopLimit = 3
	}

	return d.Transport.SendToRadio(ctx, &proto.ToRadio{
		PayloadVariant: &proto.ToRadio_Packet{
			Packet: packet,
		},
	})
}

// SendDataParams holds parameters for sending data in a packet over the mesh network.
type SendDataParams struct {
	// PortNum specifies the target application port number.
	PortNum proto.PortNum
	// Payload is the encoded application-level data to be transmitted.
	Payload []byte
	// DestNodeNum is the destination node's number.
	DestNodeNum uint32
	// WantAck indicates whether an acknowledgment is requested for this transmission.
	WantAck bool
	// ChannelIndex specifies the channel index to use for transmission.
	ChannelIndex uint32
	// ReplyID is the ID of the packet to which this is a reply, if any.
	ReplyID uint32
}

// SendData sends a data payload over the mesh network using the specified parameters.
// It encapsulates a data and a MeshPacket, then sends the packet.
func (d *Device) SendData(ctx context.Context, params SendDataParams) error {
	data := &proto.Data{
		Portnum: params.PortNum,
		Payload: params.Payload,
		ReplyId: params.ReplyID,
	}

	packet := &proto.MeshPacket{
		To:      params.DestNodeNum,
		Channel: params.ChannelIndex,
		PayloadVariant: &proto.MeshPacket_Decoded{
			Decoded: data,
		},
		WantAck: params.WantAck,
	}

	return d.SendToMesh(ctx, packet)
}

// generatePacketID generates a unique packet ID.
func (d *Device) generatePacketID() uint32 {
	nextPacketId := d.lastPacketID + 1
	nextPacketId = nextPacketId & 0x3FF
	nextPacketId = nextPacketId | (rand.Uint32() << 10)
	d.lastPacketID = nextPacketId
	return nextPacketId
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

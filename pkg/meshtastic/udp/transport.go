package udp

import (
	"context"
	"fmt"
	"net"

	"github.com/exepirit/meshtastic-go/pkg/meshtastic"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
	protobuf "google.golang.org/protobuf/proto"
)

var _ meshtastic.MeshTransport = &Transport{}

// Transport represents a transport mechanism over UDP for communicating with a Meshtastic device.
type Transport struct {
	conn *net.UDPConn
}

// NewTransport creates a new UDP transport for the Meshtastic device, using the default port (4403).
func NewTransport(host string) (*Transport, error) {
	return NewTransportPort(host, 4403)
}

// NewTransportPort creates a new UDP transport for the Meshtastic device using the specified host and port.
// It resolves the correct network interface to use, listens on a multicast UDP address, and initializes
// the transport connection.
func NewTransportPort(host string, port int) (*Transport, error) {
	// Detect which interface routes to the Meshtastic device
	_, iface, err := resolveLocalAddress(host, port)
	Logger.Debug("Found device interface", "interface", iface)

	broadcastAddr := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 69), Port: port}
	conn, err := net.ListenMulticastUDP("udp4", &iface, broadcastAddr)
	if err != nil {
		return nil, err
	}

	return &Transport{
		conn: conn,
	}, nil
}

// SendToMesh sends a mesh packet over the UDP transport.
// Currently, this method is unimplemented and will return an error.
func (t *Transport) SendToMesh(ctx context.Context, packet *proto.MeshPacket) error {
	return fmt.Errorf("unimplemented")
}

// ReceiveFromMesh receives a mesh packet from the UDP transport.
func (t *Transport) ReceiveFromMesh(ctx context.Context) (*proto.MeshPacket, error) {
	// TODO: handle ctx.Done()
	buf := make([]byte, 1500)
	n, addr, err := t.conn.ReadFrom(buf)
	if err != nil {
		return nil, err
	}

	packet := new(proto.MeshPacket)
	if err := protobuf.Unmarshal(buf[:n], packet); err != nil {
		return nil, meshtastic.ErrInvalidPacketFormat
	}

	Logger.Debug("Received UDP packet",
		"from", addr,
		"meshID", fmt.Sprintf("%08x", packet.Id),
		"meshFrom", fmt.Sprintf("%08x", packet.From),
		"meshTo", fmt.Sprintf("%08x", packet.To),
		"meshChannel", uint64(packet.Channel))
	return packet, nil
}

// Close closes the UDP connection associated with the transport.
func (t *Transport) Close() error {
	return t.conn.Close()
}

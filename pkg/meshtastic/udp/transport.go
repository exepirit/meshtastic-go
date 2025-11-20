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

func NewTransport(host string) (*Transport, error) {
	// Detect which interface routes to the Meshtastic device
	addr, err := net.ResolveUDPAddr("udp", host+":4403")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		return nil, err
	}

	laddr := conn.LocalAddr().(*net.UDPAddr).IP

	err = conn.Close()
	if err != nil {
		return nil, err
	}

	intfs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var foundIntf *net.Interface
	for _, intf := range intfs {
		if addrs, err := intf.Addrs(); err == nil {
			for _, addr := range addrs {
				if addrNet, ok := addr.(*net.IPNet); ok && laddr.Equal(addrNet.IP) {
					foundIntf = &intf
					break
				}
			}
		}
		if foundIntf != nil {
			break
		}
	}

	if foundIntf == nil {
		return nil, fmt.Errorf("could not find interface for local address %+v", laddr)
	}

	Logger.Info("Found device interface", "interface", foundIntf)

	gaddr := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 69), Port: 4403}

	conn, err = net.ListenMulticastUDP("udp4", foundIntf, gaddr)
	if err != nil {
		return nil, err
	}

	return &Transport{
		conn: conn,
	}, nil
}

func (t *Transport) SendToMesh(ctx context.Context, packet *proto.MeshPacket) error {
	return fmt.Errorf("unimplemented")
}

func (t *Transport) ReceiveFromMesh(ctx context.Context) (*proto.MeshPacket, error) {
	// TODO: handle ctx.Done()
	buf := make([]byte, 1500)
	n, addr, err := t.conn.ReadFrom(buf)
	if err != nil {
		return nil, err
	}

	logAttrs := []any{
		"from", addr,
	}

	defer func() {
		Logger.Debug("Received UDP packet", logAttrs...)
	}()

	packet := new(proto.MeshPacket)
	err = protobuf.Unmarshal(buf[:n], packet)
	if err != nil {
		return nil, meshtastic.ErrInvalidPacketFormat
	}
	logAttrs = append(logAttrs,
		"meshID", fmt.Sprintf("%08x", packet.Id),
		"meshFrom", fmt.Sprintf("%08x", packet.From),
		"meshTo", fmt.Sprintf("%08x", packet.To),
		"meshChannel", uint64(packet.Channel),
	)
	return packet, nil
}

func (t *Transport) Close() error {
	return t.conn.Close()
}

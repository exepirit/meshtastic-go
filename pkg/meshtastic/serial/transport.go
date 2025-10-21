package serial

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
	"go.bug.st/serial"
	protobuf "google.golang.org/protobuf/proto"
	"io"
	"log/slog"
	"sync"
)

// NewTransport creates a new SerialTransport instance for the given serial port.
// It opens the specified serial port with default settings (115200 baud rate).
func NewTransport(port string) (*StreamTransport, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	p, err := serial.Open(port, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	return &StreamTransport{Stream: p}, nil
}

var _ meshtastic.Transport = &StreamTransport{}

// StreamTransport represents a transport layer using a Stream (e.g., TCP connection or serial port).
type StreamTransport struct {
	Stream io.ReadWriteCloser
	Logger *slog.Logger
	lock   sync.Mutex
}

// ReceiveFromRadio reads a single packet from the stream and returns it.
func (st *StreamTransport) ReceiveFromRadio(ctx context.Context) (*proto.FromRadio, error) {
	st.lock.Lock()
	buf, err := st.readBytes()
	st.lock.Unlock()
	if err != nil {
		return nil, err
	}

	packet := new(proto.FromRadio)
	err = protobuf.Unmarshal(buf, packet)
	if err != nil {
		return nil, meshtastic.ErrInvalidPacketFormat
	}
	return packet, nil
}

func (st *StreamTransport) readBytes() ([]byte, error) {
	header := make([]byte, 4)

	for {
		_, err := io.ReadFull(st.Stream, header[:1])
		if err != nil {
			return nil, err
		}
		if header[0] != 0x94 {
			continue
		}

		_, err = io.ReadFull(st.Stream, header[1:2])
		if err != nil {
			return nil, err
		}
		if header[1] != 0xc3 {
			continue
		}

		_, err = io.ReadFull(st.Stream, header[2:])
		if err != nil {
			return nil, err
		}

		pduLen := int(binary.BigEndian.Uint16(header[2:4]))
		if pduLen > 512 {
			continue
		}

		data := make([]byte, pduLen)
		_, err = io.ReadFull(st.Stream, data)
		return data, err
	}
}

// SendToRadio sends a protobuf message to the radio.
func (st *StreamTransport) SendToRadio(ctx context.Context, packet *proto.ToRadio) error {
	buf, err := protobuf.Marshal(packet)
	if err != nil {
		return fmt.Errorf("marshalling error: %w", err)
	}

	st.lock.Lock()
	defer st.lock.Unlock()
	return st.sendBytes(buf)
}

func (st *StreamTransport) sendBytes(data []byte) error {
	// TODO: handle context
	if len(data) > 512 {
		return errors.New("packet too long")
	}

	header := []byte{0x94, 0xc3, 0, 0}
	binary.BigEndian.PutUint16(header[2:4], uint16(len(data)))

	_, err := st.Stream.Write(header)
	if err != nil {
		return err
	}

	_, err = st.Stream.Write(data)
	return err
}

func (st *StreamTransport) Close() error {
	return st.Stream.Close()
}

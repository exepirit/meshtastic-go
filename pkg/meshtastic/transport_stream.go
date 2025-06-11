package meshtastic

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/exepirit/meshtastic_exporter/pkg/meshtastic/proto"
	protobuf "google.golang.org/protobuf/proto"
	"io"
	"iter"
	"sync"
)

// StreamTransport represents a transport layer using a Stream (e.g., TCP connection or serial port).
type StreamTransport struct {
	Stream io.ReadWriteCloser
	lock   sync.Mutex
}

// ReceiveStream continuously reads packets from the stream and yields them in a sequence.
func (st *StreamTransport) ReceiveStream(ctx context.Context) iter.Seq2[*proto.FromRadio, error] {
	return func(yield func(*proto.FromRadio, error) bool) {
		st.lock.Lock()
		defer st.lock.Unlock()
		for {
			buf, err := st.readBytes()
			if err != nil {
				if yield(nil, err) {
					continue
				} else {
					return
				}
			}

			packet := new(proto.FromRadio)
			err = protobuf.Unmarshal(buf, packet)
			if !yield(packet, err) {
				return
			}
		}
	}
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
		return nil, fmt.Errorf("unmarshalling error: %w", err)
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
	buf, err := packet.MarshalVT()
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

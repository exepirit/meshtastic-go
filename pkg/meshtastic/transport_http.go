package meshtastic

import (
	"bytes"
	"context"
	"fmt"
	"github.com/exepirit/meshtastic_exporter/pkg/meshtastic/proto"
	protobuf "google.golang.org/protobuf/proto"
	"io"
	"net/http"
)

var _ Transport = &HttpTransport{}

// HttpTransport represents a transport mechanism over HTTP for communicating with a Meshtastic device.
type HttpTransport struct {
	// URL is the base URL of the meshtastic API endpoint.
	URL string
	// Client is an HTTP client used to send requests.
	Client http.Client
}

// SendToRadio sends a protobuf message to the radio through the Meshtastic API.
func (ht *HttpTransport) SendToRadio(ctx context.Context, packet *proto.ToRadio) error {
	body, err := packet.MarshalVT()
	if err != nil {
		return fmt.Errorf("marshalling error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", ht.URL+"/api/v1/toradio", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-protobuf")

	response, err := ht.Client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status code %d", response.StatusCode)
	}
	return nil
}

// ReceiveFromRadio retrieves a protobuf message from the radio through the Meshtastic API.
func (ht *HttpTransport) ReceiveFromRadio(ctx context.Context) (*proto.FromRadio, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", ht.URL+"/api/v1/fromradio?all=false", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	response, err := ht.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	packet := new(proto.FromRadio)
	err = protobuf.Unmarshal(body, packet)
	if err != nil {
		return nil, ErrInvalidPacketFormat
	}
	return packet, nil
}

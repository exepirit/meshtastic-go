package mqtt

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// Transport is an MQTT-based transport for Meshtastic communication.
type Transport struct {
	// BrokerURL is the URL of the MQTT broker to connect to.
	BrokerURL string
	// Username is the username for MQTT authentication.
	Username string
	// Password is the password for MQTT authentication.
	Password string
	// AppName is a unique identifier for the application, used in the MQTT client ID.
	AppName string
	// RootTopic is the base topic for all messages.
	RootTopic string
	// SendOpts is the configuration options for sending a mesh packet.
	SendOpts SendPacketOptions

	client     mqtt.Client
	messagesCh chan mqtt.Message
}

// SendPacketOptions holds configuration options for sending a mesh packet.
type SendPacketOptions struct {
	DeviceID string
}

// SendEnvelope sends an envelope to MQTT.
func (mt *Transport) SendEnvelope(envelope *proto.ServiceEnvelope) error {
	msgData, err := protobuf.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshalling error: %w", err)
	}

	topic := fmt.Sprintf("%s/2/e/%s/%s", mt.RootTopic, envelope.GetChannelId(), envelope.GetGatewayId())
	token := mt.client.Publish(topic, 0, false, msgData)
	<-token.Done()
	return token.Error()
}

// ReceiveEnvelope receives a envelope from MQTT.
func (mt *Transport) ReceiveEnvelope(ctx context.Context) (*proto.ServiceEnvelope, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-mt.messagesCh:
		if msg == nil {
			return nil, nil // TODO: return error?
		}

		envelope := new(proto.ServiceEnvelope)
		if err := protobuf.Unmarshal(msg.Payload(), envelope); err != nil {
			return nil, meshtastic.ErrInvalidPacketFormat
		}

		return envelope, nil
	}
}

// ReceiveFromMesh receives a mesh packet from the network via MQTT.
func (mt *Transport) ReceiveFromMesh(ctx context.Context) (*proto.MeshPacket, error) {
	envelope, err := mt.ReceiveEnvelope(ctx)
	if err != nil {
		return nil, err
	}
	return envelope.GetPacket(), nil
}

// Connect establishes an MQTT connection to the broker.
// It generates a random client ID, connects to the broker, and subscribes
// to all subtopics under the RootTopic.
func (mt *Transport) Connect() error {
	if mt.client != nil && mt.client.IsConnected() {
		return nil
	}

	randomId := make([]byte, 4)
	_, _ = rand.Read(randomId)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(mt.BrokerURL)
	opts.SetUsername(mt.Username)
	opts.SetPassword(mt.Password)
	opts.SetClientID(fmt.Sprintf("%s-%x", mt.AppName, randomId))
	opts.SetOrderMatters(false)

	mt.client = mqtt.NewClient(opts)

	token := mt.client.Connect()
	<-token.Done()
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to connect MQTT: %w", err)
	}

	return nil
}

// HandleMessages starts incoming messages handling routing.
func (mt *Transport) HandleMessages(buffer int) error {
	if mt.client == nil || !mt.client.IsConnected() {
		return errors.New("connection is not established")
	}

	mt.messagesCh = make(chan mqtt.Message, buffer)

	token := mt.client.Subscribe(mt.RootTopic+"/#", 0, mt.handleMessage)
	<-token.Done()
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	return nil
}

// Disconnect closes the MQTT connection and the message channel.
// It ensures that the client is disconnected and the channel is closed
// to prevent further message processing.
func (mt *Transport) Disconnect() {
	if mt.client != nil && mt.client.IsConnected() {
		mt.client.Disconnect(1000)
		close(mt.messagesCh)
	}
}

func (mt *Transport) handleMessage(_ mqtt.Client, message mqtt.Message) {
	mt.messagesCh <- message
}

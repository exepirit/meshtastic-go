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
	"time"
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
	// BufferSize is the internal messages queue size.
	BufferSize int

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
	if mt.client == nil || !mt.client.IsConnected() || mt.messagesCh == nil {
		return nil, ErrNotConnected
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-mt.messagesCh:
		if msg == nil { // channel is closed
			return nil, ErrNotConnected
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
	bufferSize := mt.BufferSize
	if bufferSize <= 0 {
		bufferSize = 100
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(mt.BrokerURL)
	opts.SetUsername(mt.Username)
	opts.SetPassword(mt.Password)
	opts.SetClientID(fmt.Sprintf("%s-%x", mt.AppName, randomId))
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(5 * time.Second)
	opts.SetCleanSession(true)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if err := mt.handleMessages(); err != nil {
			Logger.Error("Failed to create subscriptions", "error", err)
		} else {
			Logger.Debug("Connection and subscriptions re-established")
		}
	})

	mt.client = mqtt.NewClient(opts)
	mt.messagesCh = make(chan mqtt.Message, bufferSize)

	token := mt.client.Connect()
	<-token.Done()
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to connect MQTT: %w", err)
	}

	Logger.Debug("Connected to broker")
	return nil
}

// handleMessages starts incoming messages handling routing.
func (mt *Transport) handleMessages() error {
	if mt.client == nil || !mt.client.IsConnected() {
		return errors.New("connection is not established")
	}
	Logger.Debug("Connection established. Creating subscription to root topic")

	token := mt.client.Subscribe(mt.RootTopic+"/#", 0, mt.handleMessage)
	<-token.Done()
	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	Logger.Debug("Subscribed to the root topic")
	return nil
}

// Disconnect closes the MQTT connection and the message channel.
// It ensures that the client is disconnected and the channel is closed
// to prevent further message processing.
func (mt *Transport) Disconnect() {
	if mt.client != nil && mt.client.IsConnected() {
		mt.client.Disconnect(1000)
		Logger.Debug("Disconnected from broker")
	}
	if mt.messagesCh != nil {
		close(mt.messagesCh)
		Logger.Debug("Buffer channel is closed")
	}
}

func (mt *Transport) handleMessage(_ mqtt.Client, message mqtt.Message) {
	Logger.Debug("A new message received")
	mt.messagesCh <- message
}

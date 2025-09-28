package meshtastic

import (
	"context"
	"crypto/rand"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// MqttTransport is an MQTT-based transport for Meshtastic communication.
type MqttTransport struct {
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
	RadioPreset RadioPreset
	ChannelID   string
	DeviceID    string
}

// SendToMesh sends a mesh packet to the network via MQTT.
func (mt *MqttTransport) SendToMesh(_ context.Context, packet *proto.MeshPacket) error {
	topic := fmt.Sprintf("%s/2/e/%s/%s", mt.RootTopic, mt.SendOpts.RadioPreset.Name, mt.SendOpts.DeviceID)

	envelope := proto.ServiceEnvelope{
		Packet:    packet,
		ChannelId: mt.SendOpts.ChannelID,
		GatewayId: mt.SendOpts.DeviceID,
	}
	msgData, err := protobuf.Marshal(&envelope)
	if err != nil {
		return fmt.Errorf("marshalling error: %w", err)
	}

	token := mt.client.Publish(topic, 0, false, msgData)
	<-token.Done()
	return token.Error()
}

// ReceiveFromMesh receives a mesh packet from the network via MQTT.
func (mt *MqttTransport) ReceiveFromMesh(ctx context.Context) (*proto.MeshPacket, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-mt.messagesCh:
		if msg == nil {
			return nil, nil // TODO: return error?
		}

		var envelope proto.ServiceEnvelope
		if err := protobuf.Unmarshal(msg.Payload(), &envelope); err != nil {
			return nil, ErrInvalidPacketFormat
		}

		return envelope.GetPacket(), nil
	}
}

// Connect establishes an MQTT connection to the broker.
// It generates a random client ID, connects to the broker, and subscribes
// to all subtopics under the RootTopic.
func (mt *MqttTransport) Connect(buffer int) error {
	if mt.client != nil && mt.client.IsConnected() {
		return nil
	}

	randomId := make([]byte, 4)
	_, _ = rand.Read(randomId)
	mt.messagesCh = make(chan mqtt.Message, buffer)

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

	token = mt.client.Subscribe(mt.RootTopic+"/#", 0, mt.handleMessage)
	<-token.Done()
	if err := token.Error(); err != nil {
		mt.Disconnect()
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	return nil
}

// Disconnect closes the MQTT connection and the message channel.
// It ensures that the client is disconnected and the channel is closed
// to prevent further message processing.
func (mt *MqttTransport) Disconnect() {
	if mt.client != nil && mt.client.IsConnected() {
		mt.client.Disconnect(1000)
		close(mt.messagesCh)
	}
}

func (mt *MqttTransport) handleMessage(_ mqtt.Client, message mqtt.Message) {
	mt.messagesCh <- message
}

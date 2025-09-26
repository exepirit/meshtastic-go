package meshtastic

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/exepirit/meshtastic_exporter/pkg/meshtastic/proto"
	protobuf "google.golang.org/protobuf/proto"
)

type MqttTransport struct {
	BrokerURL string
	Username  string
	Password  string
	AppName   string
	RootTopic string

	client     mqtt.Client
	messagesCh chan mqtt.Message
}

func (mt *MqttTransport) SendToMesh(ctx context.Context, packet *proto.MeshPacket) error {
	return errors.New("send to mesh: not implemented")
}

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

func (mt *MqttTransport) Disconnect() {
	if mt.client != nil && mt.client.IsConnected() {
		mt.client.Disconnect(1000)
		close(mt.messagesCh)
	}
}

func (mt *MqttTransport) handleMessage(_ mqtt.Client, message mqtt.Message) {
	mt.messagesCh <- message
}

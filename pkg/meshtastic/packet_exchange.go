package meshtastic

import (
	"context"
	"github.com/exepirit/meshtastic_exporter/pkg/meshtastic/proto"
	"log/slog"
	"sync"
)

// PacketPublisher implements part of the pubsub pattern allowing other parts of the system to subscribe and receive
// packets.
type PacketPublisher interface {
	Publish(packet *proto.FromRadio)
}

// PacketSubscriber handles packets received from a publisher.
type PacketSubscriber interface {
	OnPacket(packet *proto.FromRadio)
}

type FanOutPacketPublisher struct {
	Subscribers []PacketSubscriber
}

func (pub *FanOutPacketPublisher) Subscribe(subscriber PacketSubscriber) {
	pub.Subscribers = append(pub.Subscribers, subscriber)
}

func (pub *FanOutPacketPublisher) Publish(packet *proto.FromRadio) {
	wg := sync.WaitGroup{}
	wg.Add(len(pub.Subscribers))
	for _, sub := range pub.Subscribers {
		go func() {
			defer wg.Done()
			sub.OnPacket(packet)
		}()
	}
	wg.Wait()
}

func (pub *FanOutPacketPublisher) PublishAll(ctx context.Context, transport Transport) {
	packets := transport.ReceiveStream(ctx)
	for packet, err := range packets {
		if err != nil {
			slog.Error("Cannot read next packet from stream", "error", err)
			continue
		}
		pub.Publish(packet)
	}
}

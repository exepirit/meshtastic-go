package meshtastic

import (
	"context"
	"github.com/exepirit/meshtastic_exporter/pkg/meshtastic/proto"
	"iter"
)

type Transport interface {
	SendToRadio(ctx context.Context, packet *proto.ToRadio) error
	ReceiveFromRadio(ctx context.Context) (*proto.FromRadio, error)
	ReceiveStream(ctx context.Context) iter.Seq2[*proto.FromRadio, error]
}

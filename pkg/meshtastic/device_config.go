package meshtastic

import (
	"context"
	"fmt"
	"github.com/exepirit/meshtastic_exporter/pkg/meshtastic/proto"
	"log/slog"
	"math/rand"
)

// DeviceModuleConfig provides actions for device configuration.
type DeviceModuleConfig struct {
	transport Transport
}

// GetState sends a request for the current configuration to the radio and retrieves the state of the device.
func (m *DeviceModuleConfig) GetState(ctx context.Context) (DeviceState, error) {
	configId := uint32(rand.Int())
	err := m.transport.SendToRadio(ctx, &proto.ToRadio{
		PayloadVariant: &proto.ToRadio_WantConfigId{
			WantConfigId: configId,
		},
	})
	if err != nil {
		return DeviceState{}, fmt.Errorf("failed to request configuration: %w", err)
	}

	slog.Debug("Configuration request is sent")

	var state DeviceState
	for {
		packet, err := m.transport.ReceiveFromRadio(ctx)
		if err != nil {
			return state, fmt.Errorf("failed to read response: %w", err)
		}
		slog.Debug("Received packet from radio")

		switch payload := packet.PayloadVariant.(type) {
		case *proto.FromRadio_MyInfo:
			state.MyInfo = payload.MyInfo
		case *proto.FromRadio_NodeInfo:
			state.Nodes = append(state.Nodes, payload.NodeInfo)
		case *proto.FromRadio_Channel:
			state.Channels = append(state.Channels, payload.Channel)
		case *proto.FromRadio_Metadata:
			state.Device = payload.Metadata
		case *proto.FromRadio_ConfigCompleteId:
			if payload.ConfigCompleteId == configId {
				return state, nil
			}
		default:
			continue // unexpected payload. ignore it
		}
	}
}

// DeviceState represents the current state of a device.
type DeviceState struct {
	MyInfo   *proto.MyNodeInfo
	Nodes    []*proto.NodeInfo
	Channels []*proto.Channel
	Device   *proto.DeviceMetadata
}

// CurrentNodeInfo returns the current node info if available.
func (s DeviceState) CurrentNodeInfo() (*proto.NodeInfo, bool) {
	if s.MyInfo == nil {
		return nil, false
	}
	for _, node := range s.Nodes {
		if node.Num == s.MyInfo.MyNodeNum {
			return node, true
		}
	}
	return nil, false
}

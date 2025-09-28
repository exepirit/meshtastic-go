package meshtastic

import (
	"errors"
)

// ErrInvalidPacketFormat indicates a problem in structure of received packet.
var ErrInvalidPacketFormat = errors.New("invalid packet data format")

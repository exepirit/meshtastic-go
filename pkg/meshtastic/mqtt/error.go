package mqtt

import "errors"

// ErrNotConnected is returned when attempting to perform an operation on a client that is not connected to the broker.
var ErrNotConnected = errors.New("client is not connected to broker")

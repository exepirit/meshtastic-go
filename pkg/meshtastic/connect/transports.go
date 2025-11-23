package connect

import (
	"context"
	"fmt"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/ble"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/http"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/mqtt"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/serial"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/udp"
	"net/url"
	"strconv"
)

func connectHTTP(u *url.URL) *http.Transport {
	return &http.Transport{URL: u.String()}
}

func connectSerial(u *url.URL) (*serial.StreamTransport, error) {
	return serial.NewTransport(u.Path)
}

func connectBLE(u *url.URL) (*ble.Transport, error) {
	return ble.ConnectMAC(context.Background(), u.Hostname())
}

func connectMQTT(u *url.URL) (*mqtt.Transport, error) {
	transport := &mqtt.Transport{
		BrokerURL: fmt.Sprintf("%s://%s", u.Scheme, u.Host),
	}

	if u.User != nil {
		transport.Username = u.User.Username()
		password, _ := u.User.Password()
		transport.Password = password
	}

	// Parse query parameters for additional options
	query := u.Query()
	if appName := query.Get("app_name"); appName != "" {
		transport.AppName = appName
	}
	if rootTopic := query.Get("root_topic"); rootTopic != "" {
		transport.RootTopic = rootTopic
	}
	if deviceID := query.Get("device_id"); deviceID != "" {
		transport.SendOpts.DeviceID = deviceID
	}

	if err := transport.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	return transport, nil
}

func connectUDP(u *url.URL) (*udp.Transport, error) {
	host := u.Hostname()

	if u.Port() == "" {
		return udp.NewTransport(host)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid UDP port: %w", err)
	}

	return udp.NewTransportPort(host, port)
}

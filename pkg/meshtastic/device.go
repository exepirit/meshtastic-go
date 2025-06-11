package meshtastic

type Device struct {
	Transport Transport
}

func (d *Device) Config() *DeviceModuleConfig {
	return &DeviceModuleConfig{transport: d.Transport}
}

package ble

import "tinygo.org/x/bluetooth"

var (
	MeshBluetoothServiceID = must(bluetooth.ParseUUID("6ba1b218-15a8-461f-9fa8-5dcae273eafd"))
	FromRadioPropertyID    = must(bluetooth.ParseUUID("2c55e69e-4993-11ed-b878-0242ac120002"))
	ToRadioPropertyID      = must(bluetooth.ParseUUID("f75c76d2-129e-4dad-a1dd-7866124401e7"))
	FromNumPropertyID      = must(bluetooth.ParseUUID("ed9da18c-a800-4f66-a670-aa7547e34453"))
)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

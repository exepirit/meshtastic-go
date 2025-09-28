package meshtastic

// RadioPreset describes basic information about LoRa radio preset.
type RadioPreset struct {
	Name string
}

var (
	PresetShortTurbo   = RadioPreset{Name: "ShortTurbo"}
	PresetShortFast    = RadioPreset{Name: "ShortFast"}
	PresetShortSlow    = RadioPreset{Name: "ShortSlow"}
	PresetMediumFast   = RadioPreset{Name: "MediumFast"}
	PresetMediumSlow   = RadioPreset{Name: "MediumSlow"}
	PresetLongFast     = RadioPreset{Name: "LongFast"}
	PresetLongModerate = RadioPreset{Name: "LongModerate"}
	PresetLongSlow     = RadioPreset{Name: "LongSlow"}
)

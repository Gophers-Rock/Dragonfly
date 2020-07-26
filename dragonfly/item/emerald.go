package item

// Emerald is a rare mineral obtained from emerald ore or from villagers.
type Emerald struct{}

// EncodeItem ...
func (Emerald) EncodeItem() (id int32, meta int16) {
	return 388, 0
}

// PayableForBeacon ...
func (Emerald) PayableForBeacon() bool {
	return true
}
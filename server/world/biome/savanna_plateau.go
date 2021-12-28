package biome

// SavannaPlateau ...
type SavannaPlateau struct{}

// Temperature ...
func (SavannaPlateau) Temperature() float64 {
	return 1
}

// Rainfall ...
func (SavannaPlateau) Rainfall() float64 {
	return 0
}

// Ash ...
func (SavannaPlateau) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SavannaPlateau) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SavannaPlateau) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SavannaPlateau) RedSpores() float64 {
	return 0
}

// String ...
func (SavannaPlateau) String() string {
	return "savanna_plateau"
}

// EncodeBiome ...
func (SavannaPlateau) EncodeBiome() int {
	return 36
}

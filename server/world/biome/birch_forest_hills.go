package biome

// BirchForestHills ...
type BirchForestHills struct{}

// Temperature ...
func (BirchForestHills) Temperature() float64 {
	return 0.6
}

// Rainfall ...
func (BirchForestHills) Rainfall() float64 {
	return 0.6
}

// String ...
func (BirchForestHills) String() string {
	return "Birch Forest Hills"
}

// EncodeBiome ...
func (BirchForestHills) EncodeBiome() int {
	return 28
}

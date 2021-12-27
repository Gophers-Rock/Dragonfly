package biome

// MountainEdge ...
type MountainEdge struct{}

// Temperature ...
func (MountainEdge) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (MountainEdge) Rainfall() float64 {
	return 0.3
}

// String ...
func (MountainEdge) String() string {
	return "Mountain Edge"
}

// EncodeBiome ...
func (MountainEdge) EncodeBiome() int {
	return 20
}

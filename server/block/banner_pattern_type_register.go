package block

// init initializes all default banner patterns to the registry.
func init() {
	RegisterBannerPattern("bo", BorderBannerPattern())
	RegisterBannerPattern("bri", BricksBannerPattern())
	RegisterBannerPattern("mc", CircleBannerPattern())
	RegisterBannerPattern("cre", CreeperBannerPattern())
	RegisterBannerPattern("cr", CrossBannerPattern())
	RegisterBannerPattern("cbo", CurlyBorderBannerPattern())
	RegisterBannerPattern("lud", DiagonalLeftBannerPattern())
	RegisterBannerPattern("rd", DiagonalRightBannerPattern())
	RegisterBannerPattern("ld", DiagonalUpLeftBannerPattern())
	RegisterBannerPattern("rud", DiagonalUpRightBannerPattern())
	RegisterBannerPattern("flo", FlowerBannerPattern())
	RegisterBannerPattern("gra", GradientBannerPattern())
	RegisterBannerPattern("gru", GradientUpBannerPattern())
	RegisterBannerPattern("hh", HalfHorizontalBannerPattern())
	RegisterBannerPattern("hhb", HalfHorizontalBottomBannerPattern())
	RegisterBannerPattern("vh", HalfVerticalBannerPattern())
	RegisterBannerPattern("vhr", HalfVerticalRightBannerPattern())
	RegisterBannerPattern("moj", MojangBannerPattern())
	RegisterBannerPattern("mr", RhombusBannerPattern())
	RegisterBannerPattern("sku", SkullBannerPattern())
	RegisterBannerPattern("ss", SmallStripesBannerPattern())
	RegisterBannerPattern("bl", SquareBottomLeftBannerPattern())
	RegisterBannerPattern("br", SquareBottomRightBannerPattern())
	RegisterBannerPattern("tl", SquareTopLeftBannerPattern())
	RegisterBannerPattern("tr", SquareTopRightBannerPattern())
	RegisterBannerPattern("sc", StraightCrossBannerPattern())
	RegisterBannerPattern("bs", StripeBottomBannerPattern())
	RegisterBannerPattern("cs", StripeCenterBannerPattern())
	RegisterBannerPattern("dls", StripeDownLeftBannerPattern())
	RegisterBannerPattern("drs", StripeDownRightBannerPattern())
	RegisterBannerPattern("ls", StripeLeftBannerPattern())
	RegisterBannerPattern("ms", StripeMiddleBannerPattern())
	RegisterBannerPattern("rs", StripeRightBannerPattern())
	RegisterBannerPattern("ts", StripeTopBannerPattern())
	RegisterBannerPattern("bt", TriangleBottomBannerPattern())
	RegisterBannerPattern("tt", TriangleTopBannerPattern())
	RegisterBannerPattern("bts", TrianglesBottomBannerPattern())
	RegisterBannerPattern("tts", TrianglesTopBannerPattern())
	RegisterBannerPattern("glb", GlobeBannerPattern())
	RegisterBannerPattern("pig", PiglinBannerPattern())
}

var (
	bannerPatternsMap = map[string]BannerPatternType{}
	bannerPatternIDs  = map[BannerPatternType]string{}
)

// RegisterBannerPattern registers a banner pattern with the ID passed.
func RegisterBannerPattern(id string, pattern BannerPatternType) {
	bannerPatternsMap[id] = pattern
	bannerPatternIDs[pattern] = id
}

// BannerPatternByID returns a banner pattern by the ID it was registered with.
func BannerPatternByID(id string) (BannerPatternType, bool) {
	b, ok := bannerPatternsMap[id]
	if !ok {
		return BannerPatternType{}, false
	}
	return b, true
}

// BannerPatternID returns the ID a banner pattern was registered with.
func BannerPatternID(pattern BannerPatternType) string {
	id, ok := bannerPatternIDs[pattern]
	if !ok {
		panic("should never happen")
	}
	return id
}

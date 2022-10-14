package messages

type Ping struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&Ping{})
}

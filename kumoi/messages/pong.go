package messages

type Pong struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&Pong{})
}

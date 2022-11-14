package messages

type Close struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&Close{})
}

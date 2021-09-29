package messages

import (
	"reflect"

	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteFrame interface {
	VoteID() string
}

type VoteTransitFrame struct {
	transitFrame
	voteId string
}

func (c *VoteTransitFrame) VoteID() string {
	return c.voteId
}

func (c *VoteTransitFrame) ParseTransitFrame(tf *omega.TransitFrame) {
	c.transitFrame.ParseTransitFrame(tf)
	tfd := reflect.ValueOf(tf.GetData())
	if !tfd.IsValid() {
		return
	}

	tfdE := tfd.Elem()
	if tfdE.NumField() == 0 {
		return
	}

	// has channelId
	if tfdEv := tfdE.Field(0).Elem().FieldByName("VoteId"); tfdEv.IsValid() {
		c.voteId = tfdEv.String()
	}
}

type VoteOption struct {
	Id          string
	Name        string
	Count       int32
	SubjectList []string
}

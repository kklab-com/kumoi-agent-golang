package messages

import (
	"reflect"

	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteFrame interface {
	TransitFrame
	VoteID() string
}

type voteTransitFrame struct {
	transitFrame
	voteId string
}

func (c *voteTransitFrame) VoteID() string {
	return c.voteId
}

func (c *voteTransitFrame) ParseTransitFrame(tf *omega.TransitFrame) {
	c.transitFrame.ParseTransitFrame(tf)
	tfd := reflect.ValueOf(tf.GetData())
	if !tfd.IsValid() {
		return
	}

	tfdE := tfd.Elem()
	if tfdE.NumField() == 0 {
		return
	}

	// has VoteId
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

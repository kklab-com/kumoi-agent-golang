package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type GetSessionMeta struct {
	transitFrame
	SessionId string
	Subject   string
	Name      string
	Data      *base.Metadata
}

func (c *GetSessionMeta) ParseTransitFrame(tf *omega.TransitFrame) {
	c.SessionId = tf.GetGetSessionMeta().GetSessionId()
	c.Subject = tf.GetGetSessionMeta().GetSubject()
	c.Name = tf.GetGetSessionMeta().GetName()
	c.Data = tf.GetGetSessionMeta().GetData()
	c.transitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}

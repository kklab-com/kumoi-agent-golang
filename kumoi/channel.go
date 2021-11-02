package kumoi

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kklab-com/goth-kkutil/concurrent"
	"github.com/kklab-com/goth-kkutil/hex"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type ChannelInfo struct {
	channelId string
	name      string
	metadata  *base.Metadata
	createdAt int64
	omega     *Omega
}

func (c *ChannelInfo) ChannelId() string {
	return c.channelId
}

func (c *ChannelInfo) Name() string {
	return c.name
}

func (c *ChannelInfo) Metadata() *base.Metadata {
	return c.metadata
}

func (c *ChannelInfo) CreatedAt() int64 {
	return c.createdAt
}

func (c *ChannelInfo) Join(key string) *Channel {
	if v := c.omega.Agent().JoinChannel(c.ChannelId(), key).Get(); v != nil {
		if cv := v.(*omega.TransitFrame).GetJoinChannel(); cv != nil {
			nc := *c
			ch := &Channel{
				key:      key,
				omega:    c.omega,
				info:     &nc,
				roleName: cv.GetRole(),
				role:     cv.GetRoleIndicator(),
				onLeave:  func() {},
				onClose:  func() {},
				watch:    func(msg messages.ChannelFrame) {},
			}

			ch.info.name = cv.GetName()
			ch.info.metadata = cv.GetChannelMetadata()
			ch.init()
			return ch
		}
	}

	return nil
}

func (c *ChannelInfo) Close(key string) concurrent.Future {
	return c.omega.agent.CloseChannel(c.ChannelId(), key)
}

type Channel struct {
	key              string
	omega            *Omega
	info             *ChannelInfo
	role             omega.Role
	roleName         string
	onLeave, onClose func()
	watch            func(msg messages.ChannelFrame)
}

func (c *Channel) Id() string {
	return c.Info().channelId
}

func (c *Channel) Info() *ChannelInfo {
	return c.info
}

func (c *Channel) Role() omega.Role {
	return c.role
}

func (c *Channel) RoleName() string {
	return c.roleName
}

func (c *Channel) Name() string {
	return c.Info().name
}

func (c *Channel) SetName(name string) bool {
	if c.omega.Agent().SetChannelMetadata(c.Info().channelId, name, nil).Await().IsSuccess() {
		c.info.name = name
		return true
	}

	return false
}

func (c *Channel) Metadata() *base.Metadata {
	return c.info.Metadata()
}

func (c *Channel) SetMetadata(meta *base.Metadata) bool {
	return c.omega.Agent().SetChannelMetadata(c.Info().channelId, "", meta).Await().IsSuccess()
}

func (c *Channel) Leave() bool {
	return c.omega.Agent().LeaveChannel(c.Info().channelId).Await().IsSuccess()
}

func (c *Channel) Close() bool {
	return c.omega.Agent().CloseChannel(c.Info().channelId, c.key).Await().IsSuccess()
}

func (c *Channel) SendMessage(msg string) bool {
	return c.omega.Agent().ChannelMessage(c.Info().channelId, msg).Await().IsSuccess()
}

func (c *Channel) SendOwnerMessage(msg string) bool {
	return c.omega.Agent().ChannelOwnerMessage(c.Info().channelId, msg).Await().IsSuccess()
}

func (c *Channel) GetCount() *messages.ChannelCount {
	if tf := castTransitFrame(c.omega.Agent().ChannelCount(c.Info().channelId).Get()); tf != nil {
		r := &messages.ChannelCount{}
		r.ParseTransitFrame(tf)
		return r
	}

	return nil
}

func (c *Channel) ReplayChannelMessage(targetTimestamp int64, inverse bool, volume omega.Volume) TFPlayer {
	cp := &ChannelPlayer{
		omega:     c.omega,
		channelId: c.info.channelId,
	}

	cp.load(c.omega.Agent().ReplayChannelMessage(c.Info().channelId, targetTimestamp, inverse, volume, ""))
	if len(cp.tfs) == 0 {
		cp = nil
	}

	return cp
}

func (c *Channel) OnLeave(f func()) *Channel {
	c.onLeave = f
	return c
}

func (c *Channel) OnClose(f func()) *Channel {
	c.onClose = f
	return c
}

func (c *Channel) Watch(f func(msg messages.ChannelFrame)) *Channel {
	c.watch = f
	return c
}

func (c *Channel) init() {
	c.omega.onMessageHandlers.Store(c.Id(), func(tf *omega.TransitFrame) {
		if tf.GetClass() == omega.TransitFrame_ClassError {
			return
		}

		tfd := reflect.ValueOf(tf.GetData())
		if !tfd.IsValid() {
			return
		}

		tfdE := tfd.Elem()
		if tfdE.NumField() == 0 {
			return
		}

		// has same channelId
		if tfdEChId := tfdE.Field(0).Elem().FieldByName("ChannelId"); tfdEChId.IsValid() && tfdEChId.String() == c.Id() {
			switch tf.GetClass() {
			case omega.TransitFrame_ClassNotification:
				if tfd := tf.GetGetChannelMeta(); tfd != nil {
					c.info.name = tfd.GetName()
					c.info.metadata = tfd.GetData()
					c.info.createdAt = tfd.GetCreatedAt()
				}

				var rcf messages.ChannelFrame
				switch tf.GetData().(type) {
				case *omega.TransitFrame_ChannelCount:
					rcf = &messages.ChannelCount{}
				case *omega.TransitFrame_ChannelMessage:
					rcf = &messages.ChannelMessage{}
				case *omega.TransitFrame_ChannelOwnerMessage:
					rcf = &messages.ChannelOwnerMessage{}
				case *omega.TransitFrame_GetChannelMeta:
					rcf = &messages.GetChannelMeta{}
				case *omega.TransitFrame_JoinChannel:
					rcf = &messages.JoinChannel{}
				case *omega.TransitFrame_LeaveChannel:
					rcf = &messages.LeaveChannel{}
				case *omega.TransitFrame_CloseChannel:
					rcf = &messages.CloseChannel{}
				}

				if rcf != nil {
					if ccf, ok := rcf.(messages.TransitFrameParsable); ok {
						ccf.ParseTransitFrame(tf)
					}
				}

				if tfd := tf.GetLeaveChannel(); tfd != nil {
					c.onLeave()
					c.deInit()
				}

				if tfd := tf.GetCloseChannel(); tfd != nil {
					c.onClose()
					c.deInit()
				}

				c.watch(rcf)
			case omega.TransitFrame_ClassResponse:
				if tfd := tf.GetGetChannelMeta(); tfd != nil {
					c.info.name = tfd.GetName()
					c.info.metadata = tfd.GetData()
					c.info.createdAt = tfd.GetCreatedAt()
				}

				if tfd := tf.GetLeaveChannel(); tfd != nil {
					c.onLeave()
					c.deInit()
				}

				if tfd := tf.GetCloseChannel(); tfd != nil {
					c.onClose()
					c.deInit()
				}
			}
		}
	})
}

func (c *Channel) deInit() {
	c.omega.onMessageHandlers.Delete(c.Id())
}

type TFPlayer interface {
	Next() TFPlayerEntity
}

type TFPlayerEntity interface {
	Name() string
	TransitFrame() *omega.TransitFrame
}

type ChannelPlayer struct {
	omega     *Omega
	channelId string
	nextId    string
	cursor    int
	tfs       []*omega.TransitFrame
}

func (c *ChannelPlayer) Next() TFPlayerEntity {
	if c.cursor < len(c.tfs) {
		cpe := &ChannelPlayerEntity{tf: c.tfs[c.cursor]}
		c.cursor++
		return cpe
	} else if c.cursor == len(c.tfs) && len(c.tfs) > 0 {
		c.cursor = 0
		c.tfs = nil
		if c.nextId == "" {
			return nil
		}

		c.load(c.omega.Agent().ReplayChannelMessage(c.channelId, 0, false, omega.Volume_VolumePeek, c.nextId))
		return c.Next()
	}

	return nil
}

func (c *ChannelPlayer) load(f base.SendFuture) {
	bwg := concurrent.BurstWaitGroup{}
	transitId := f.SentTransitFrame().GetTransitId()
	refId := ""
	totalCount := int32(0)
	loadCount := int32(0)
	rcf := concurrent.NewFuture(nil)
	c.omega.onMessageHandlers.Store(fmt.Sprintf("load-%d", transitId), func(tf *omega.TransitFrame) {
		if tf.GetTransitId() == transitId && tf.GetClass() == omega.TransitFrame_ClassResponse {
			refId = hex.EncodeToString(tf.GetMessageId())
			switch rcm := tf.GetData().(type) {
			case *omega.TransitFrame_ReplayChannelMessage:
				totalCount = rcm.ReplayChannelMessage.GetCount()
			case *omega.TransitFrame_PlaybackChannelMessage:
				totalCount = rcm.PlaybackChannelMessage.GetCount()
			}

			if totalCount > 0 {
				bwg.Add(int(totalCount))
			} else {
				c.omega.onMessageHandlers.Delete(fmt.Sprintf("load-%d", transitId))
			}

			rcf.Completable().Complete(nil)
		}

		if len(tf.GetRefererMessageId()) > 0 && hex.EncodeToString(tf.GetRefererMessageId()) == refId {
			c.tfs = append(c.tfs, tf)
			loadCount++
			bwg.Done()
			if loadCount == totalCount {
				c.omega.onMessageHandlers.Delete(fmt.Sprintf("load-%d", transitId))
			}
		}
	})

	if v := f.Get(); v != nil {
		go func() {
			<-time.After(base.DefaultTransitTimeout)
			rcf.Completable().Fail(base.ErrTransitTimeout)
			bwg.Burst()
		}()

		rcf.Await()
		bwg.Wait()
	} else {
		c.omega.onMessageHandlers.Delete(fmt.Sprintf("load-%d", transitId))
	}
}

type ChannelPlayerEntity struct {
	tf *omega.TransitFrame
}

func (e *ChannelPlayerEntity) TransitFrame() *omega.TransitFrame {
	return e.tf
}

func (e *ChannelPlayerEntity) Name() string {
	tfd := reflect.ValueOf(e.tf.GetData())
	if !tfd.IsValid() {
		return ""
	}

	tfdE := tfd.Elem()
	if tfdE.NumField() == 0 {
		return ""
	}

	return tfdE.Field(0).Elem().Type().Name()
}

func (e *ChannelPlayerEntity) GetChannelMessage() *messages.ChannelMessage {
	if cm := e.tf.GetChannelMessage(); cm != nil {
		r := &messages.ChannelMessage{}
		r.ParseTransitFrame(e.tf)
		return r
	}

	return nil
}

func (e *ChannelPlayerEntity) GetOwnerChannelMessage() *messages.ChannelOwnerMessage {
	if cm := e.tf.GetChannelMessage(); cm != nil {
		r := &messages.ChannelOwnerMessage{}
		r.ParseTransitFrame(e.tf)
		return r
	}

	return nil
}

func (e *ChannelPlayerEntity) GetSetChannelMeta() *messages.SetChannelMeta {
	if cm := e.tf.GetSetChannelMeta(); cm != nil {
		r := &messages.SetChannelMeta{}
		r.ParseTransitFrame(e.tf)
		return r
	}

	return nil
}

func (e *ChannelPlayerEntity) GetGetChannelMeta() *messages.GetChannelMeta {
	if cm := e.tf.GetGetChannelMeta(); cm != nil {
		r := &messages.GetChannelMeta{}
		r.ParseTransitFrame(e.tf)
		return r
	}

	return nil
}

func (e *ChannelPlayerEntity) GetJoinChannel() *messages.JoinChannel {
	if cm := e.tf.GetJoinChannel(); cm != nil {
		r := &messages.JoinChannel{}
		r.ParseTransitFrame(e.tf)
		return r
	}

	return nil
}

func (e *ChannelPlayerEntity) GetLeaveChannel() *messages.LeaveChannel {
	if cm := e.tf.GetLeaveChannel(); cm != nil {
		r := &messages.LeaveChannel{}
		r.ParseTransitFrame(e.tf)
		return r
	}

	return nil
}

func (e *ChannelPlayerEntity) GetCloseChannel() *messages.CloseChannel {
	if cm := e.tf.GetCloseChannel(); cm != nil {
		r := &messages.CloseChannel{}
		r.ParseTransitFrame(e.tf)
		return r
	}

	return nil
}

package kumoi

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	kklogger "github.com/kklab-com/goth-kklogger"
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

func (c *Channel) SetName(name string) SendFuture[*messages.SetChannelMeta] {
	return wrapSendFuture[*messages.SetChannelMeta](c.omega.Agent().SetChannelMetadata(c.Info().channelId, name, nil, nil))
}

func (c *Channel) Fetch() SendFuture[*messages.GetChannelMeta] {
	return wrapSendFuture[*messages.GetChannelMeta](c.omega.Agent().GetChannelMetadata(c.Id()))
}

func (c *Channel) Metadata() *base.Metadata {
	return c.info.Metadata()
}

func (c *Channel) SetMetadata(meta *base.Metadata) SendFuture[*messages.SetChannelMeta] {
	return wrapSendFuture[*messages.SetChannelMeta](c.omega.Agent().SetChannelMetadata(c.Info().channelId, "", meta, nil))
}

func (c *Channel) SetSkill(skill *omega.Skill) SendFuture[*messages.SetChannelMeta] {
	return wrapSendFuture[*messages.SetChannelMeta](c.omega.Agent().SetChannelMetadata(c.Info().channelId, "", nil, skill))
}

func (c *Channel) Leave() SendFuture[*messages.LeaveChannel] {
	return wrapSendFuture[*messages.LeaveChannel](c.omega.Agent().LeaveChannel(c.Info().channelId))
}

func (c *Channel) Close() SendFuture[*messages.CloseChannel] {
	return wrapSendFuture[*messages.CloseChannel](c.omega.Agent().CloseChannel(c.Info().channelId, c.key))
}

func (c *Channel) SendMessage(msg string, meta *base.Metadata) SendFuture[*messages.ChannelMessage] {
	return wrapSendFuture[*messages.ChannelMessage](c.omega.Agent().ChannelMessage(c.Info().channelId, msg, meta))
}

func (c *Channel) SendOwnerMessage(msg string, meta *base.Metadata) SendFuture[*messages.ChannelOwnerMessage] {
	return wrapSendFuture[*messages.ChannelOwnerMessage](c.omega.Agent().ChannelOwnerMessage(c.Info().channelId, msg, meta))
}

func (c *Channel) GetCount() SendFuture[*messages.ChannelCount] {
	return wrapSendFuture[*messages.ChannelCount](c.omega.Agent().ChannelCount(c.Info().channelId))
}

func (c *Channel) ReplayChannelMessage(targetTimestamp int64, inverse bool, volume omega.Volume) Player {
	omg := c.omega
	cp := &channelPlayer{
		omega:           omg,
		channelId:       c.Info().ChannelId(),
		targetTimestamp: targetTimestamp,
		inverse:         inverse,
		volume:          volume,
		loadFutureFunc: func(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) base.SendFuture {
			return omg.Agent().ReplayChannelMessage(channelId, targetTimestamp, inverse, volume, nextId)
		},
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

				if ctf := getParsedTransitFrameFromBaseTransitFrame(tf).Cast().ChannelFrame(); ctf != nil {
					c.watch(ctf)
				} else {
					kklogger.WarnJ("kumoi:Channel.init", fmt.Sprintf("%s should not be here", tf.String()))
				}

				if tfd := tf.GetLeaveChannel(); tfd != nil {
					c.onLeave()
					c.deInit()
				}

				if tfd := tf.GetCloseChannel(); tfd != nil {
					c.onClose()
					c.deInit()
				}
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

type Player interface {
	Next() messages.TransitFrame
}

type channelPlayer struct {
	omega           *Omega
	channelId       string
	targetTimestamp int64
	inverse         bool
	volume          omega.Volume
	nextId          string
	cursor          int
	loadFutureFunc  func(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) base.SendFuture
	tfs             []*omega.TransitFrame
	eof             bool
}

func (p *channelPlayer) Next() (t messages.TransitFrame) {
	if p.eof {
		return
	}

	if p.cursor < len(p.tfs) {
		t = getParsedTransitFrameFromBaseTransitFrame(p.tfs[p.cursor])
		p.cursor++
		return
	} else if p.cursor == len(p.tfs) && len(p.tfs) > 0 {
		p.cursor = 0
		p.tfs = nil
		if p.nextId == "" {
			p.eof = true
			return
		}
	}

	p.load(p.loadFutureFunc(p.channelId, p.targetTimestamp, p.inverse, p.volume, p.nextId))
	return p.Next()
}

func (p *channelPlayer) load(f base.SendFuture) {
	bwg := concurrent.WaitGroup{}
	transitId := f.SentTransitFrame().GetTransitId()
	var refId []byte
	totalCount := int32(0)
	loadCount := int32(0)
	rcf := concurrent.NewFuture()
	player := p
	player.omega.onMessageHandlers.Store(fmt.Sprintf("load-%d", transitId), func(tf *omega.TransitFrame) {
		if tf.GetTransitId() == transitId && tf.GetClass() == omega.TransitFrame_ClassResponse {
			refId = tf.GetMessageId()
			switch rcm := tf.GetData().(type) {
			case *omega.TransitFrame_ReplayChannelMessage:
				totalCount = rcm.ReplayChannelMessage.GetCount()
			case *omega.TransitFrame_PlaybackChannelMessage:
				totalCount = rcm.PlaybackChannelMessage.GetCount()
			}

			if totalCount > 0 {
				bwg.Add(int(totalCount))
			} else {
				player.omega.onMessageHandlers.Delete(fmt.Sprintf("load-%d", transitId))
			}

			rcf.Completable().Complete(nil)
		}

		if len(tf.GetRefererMessageId()) > 0 && bytes.Compare(tf.GetRefererMessageId(), refId) == 0 {
			player.tfs = append(player.tfs, tf)
			loadCount++
			bwg.Done()
			if loadCount == totalCount {
				player.omega.onMessageHandlers.Delete(fmt.Sprintf("load-%d", transitId))
			}
		}
	})

	if v := f.GetTimeout(base.DefaultTransitTimeout); v != nil {
		go func() {
			<-time.After(base.DefaultTransitTimeout)
			rcf.Completable().Fail(base.ErrTransitTimeout)
			bwg.Reset()
		}()

		rcf.AwaitTimeout(base.DefaultTransitTimeout)
		bwg.Wait()
	}

	p.omega.onMessageHandlers.Delete(fmt.Sprintf("load-%d", transitId))
}

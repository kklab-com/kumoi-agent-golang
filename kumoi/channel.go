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
	meta  *omega.GetChannelMeta
	omega *Omega
}

func (c *ChannelInfo) ChannelId() string {
	return c.meta.ChannelId
}

func (c *ChannelInfo) Name() string {
	return c.meta.Name
}

func (c *ChannelInfo) Metadata() map[string]any {
	return base.SafeGetStructMap(c.meta.Data)
}

func (c *ChannelInfo) CreatedAt() int64 {
	return c.meta.CreatedAt
}

func (c *ChannelInfo) Join(key string) *Channel {
	if v := c.omega.Agent().JoinChannel(c.ChannelId(), key).Get(); v != nil {
		if cv := v.GetJoinChannel(); cv != nil {
			nc := *c
			ch := &Channel{
				key:      key,
				omega:    c.omega,
				info:     &nc,
				roleName: cv.GetRole(),
				role:     cv.GetRoleIndicator(),
				onLeave:  func() {},
				onClose:  func() {},
				watch:    func(msg messages.TransitFrame) {},
			}

			ch.info.meta.Skill = cv.Skill
			ch.init()
			return ch
		}
	}

	return nil
}

func (c *ChannelInfo) Close(key string) SendFuture[*messages.CloseChannel] {
	return wrapSendFuture[*messages.CloseChannel](c.omega.agent.CloseChannel(c.ChannelId(), key))
}

type Channel struct {
	key              string
	omega            *Omega
	info             *ChannelInfo
	role             omega.Role
	roleName         string
	onLeave, onClose func()
	watch            func(msg messages.TransitFrame)
}

func (c *Channel) watchId() string {
	return fmt.Sprintf("ch-watch-%s", c.info.ChannelId())
}

func (c *Channel) Id() string {
	return c.Info().ChannelId()
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
	return c.Info().Name()
}

func (c *Channel) SetName(name string) SendFuture[*messages.SetChannelMeta] {
	return wrapSendFuture[*messages.SetChannelMeta](c.omega.Agent().SetChannelMetadata(c.Info().ChannelId(), name, nil, nil))
}

func (c *Channel) Fetch() SendFuture[*messages.GetChannelMeta] {
	return wrapSendFuture[*messages.GetChannelMeta](c.omega.Agent().GetChannelMetadata(c.Id()))
}

func (c *Channel) Metadata() map[string]any {
	return c.info.Metadata()
}

func (c *Channel) SetMetadata(metadata map[string]any) SendFuture[*messages.SetChannelMeta] {
	return wrapSendFuture[*messages.SetChannelMeta](c.omega.Agent().SetChannelMetadata(c.Info().ChannelId(), "", base.NewMetadata(metadata), nil))
}

func (c *Channel) SetSkill(skill *omega.Skill) SendFuture[*messages.SetChannelMeta] {
	return wrapSendFuture[*messages.SetChannelMeta](c.omega.Agent().SetChannelMetadata(c.Info().ChannelId(), "", nil, skill))
}

func (c *Channel) Leave() SendFuture[*messages.LeaveChannel] {
	wsf := wrapSendFuture[*messages.LeaveChannel](c.omega.Agent().LeaveChannel(c.Info().ChannelId()))
	fc := c
	wsf.Base().Chainable().Then(func(parent concurrent.Future) interface{} {
		if parent.IsSuccess() {
			fc.invokeOnLeaveChannelSuccess()
		}

		return nil
	})

	return wsf
}

func (c *Channel) Close() SendFuture[*messages.CloseChannel] {
	wsf := wrapSendFuture[*messages.CloseChannel](c.omega.Agent().CloseChannel(c.Info().ChannelId(), c.key))
	fc := c
	wsf.Base().Chainable().Then(func(parent concurrent.Future) interface{} {
		if parent.IsSuccess() {
			fc.invokeOnCloseChannelSuccess()
		}

		return nil
	})

	return wsf
}

func (c *Channel) SendMessage(msg string, metadata map[string]any) SendFuture[*messages.ChannelMessage] {
	return wrapSendFuture[*messages.ChannelMessage](c.omega.Agent().ChannelMessage(c.Info().ChannelId(), msg, base.NewMetadata(metadata)))
}

func (c *Channel) SendOwnerMessage(msg string, metadata map[string]any) SendFuture[*messages.ChannelOwnerMessage] {
	return wrapSendFuture[*messages.ChannelOwnerMessage](c.omega.Agent().ChannelOwnerMessage(c.Info().ChannelId(), msg, base.NewMetadata(metadata)))
}

func (c *Channel) Count() SendFuture[*messages.ChannelCount] {
	return wrapSendFuture[*messages.ChannelCount](c.omega.Agent().ChannelCount(c.Info().ChannelId()))
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

func (c *Channel) Watch(f func(msg messages.TransitFrame)) *Channel {
	c.watch = f
	return c
}

func (c *Channel) init() {
	fc := c
	c.omega.onMessageHandlers.Store(c.watchId(), func(tf *omega.TransitFrame) {
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
		if tfdEChId := tfdE.Field(0).Elem().FieldByName("ChannelId"); tfdEChId.IsValid() && tfdEChId.String() == fc.Id() {
			switch tf.GetClass() {
			case omega.TransitFrame_ClassNotification:
				if tfd := tf.GetGetChannelMeta(); tfd != nil {
					fc.info.meta = tfd
				}

				if ctf := messages.WrapTransitFrame(tf); ctf != nil {
					fc.watch(ctf)
				} else {
					kklogger.WarnJ("kumoi:Channel.init", fmt.Sprintf("%s should not be here", tf.String()))
				}

				if tfd := tf.GetLeaveChannel(); tfd != nil && tfd.GetSessionId() == c.omega.Session().GetId() {
					fc.invokeOnLeaveChannelSuccess()
				}

				if tfd := tf.GetCloseChannel(); tfd != nil {
					fc.invokeOnCloseChannelSuccess()
				}
			case omega.TransitFrame_ClassResponse:
				if tfd := tf.GetGetChannelMeta(); tfd != nil {
					fc.info.meta = tfd
					break
				}

				if tfd := tf.GetSetChannelMeta(); tfd != nil {
					fc.info.meta.Data = tfd.Data
					fc.info.meta.Name = tfd.Name
				}
			}
		}
	})
}

func (c *Channel) invokeOnLeaveChannelSuccess() {
	c.onLeave()
	c.deInit()
}

func (c *Channel) invokeOnCloseChannelSuccess() {
	c.onClose()
	c.deInit()
}

func (c *Channel) deInit() {
	c.omega.onMessageHandlers.Delete(c.watchId())
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

func (p *channelPlayer) HasNext() bool {
	if p.eof {
		return false
	}

	if p.cursor < len(p.tfs) {
		return true
	} else if p.cursor == len(p.tfs) && len(p.tfs) > 0 {
		p.cursor = 0
		p.tfs = nil
		if p.nextId == "" {
			p.eof = true
			return false
		}
	}

	p.load(p.loadFutureFunc(p.channelId, p.targetTimestamp, p.inverse, p.volume, p.nextId))
	return p.HasNext()
}

func (p *channelPlayer) Next() (t messages.TransitFrame) {
	if p.HasNext() {
		t = messages.WrapTransitFrame(p.tfs[p.cursor])
		p.cursor++
		return
	}

	return
}

func (p *channelPlayer) Range(f func(frame messages.TransitFrame)) {
	for tf := p.Next(); tf != nil; tf = p.Next() {
		f(tf)
	}
}

func (p *channelPlayer) load(f base.SendFuture) {
	bwg := concurrent.WaitGroup{}
	transitId := f.SentTransitFrame().GetTransitId()
	watchId := fmt.Sprintf("ch-load-%d", transitId)
	var refId []byte
	totalCount := int32(0)
	loadCount := int32(0)
	rcf := concurrent.NewFuture()
	player := p
	player.omega.onMessageHandlers.Store(watchId, func(tf *omega.TransitFrame) {
		if tf.GetTransitId() == transitId && tf.GetClass() == omega.TransitFrame_ClassResponse {
			refId = tf.GetMessageId()
			switch rcm := tf.GetData().(type) {
			case *omega.TransitFrame_ReplayChannelMessage:
				totalCount = rcm.ReplayChannelMessage.GetCount()
				player.nextId = rcm.ReplayChannelMessage.GetNextId()
			case *omega.TransitFrame_PlaybackChannelMessage:
				totalCount = rcm.PlaybackChannelMessage.GetCount()
				player.nextId = rcm.PlaybackChannelMessage.GetNextId()
			}

			if totalCount > 0 {
				bwg.Add(int(totalCount))
			} else {
				player.eof = true
				player.omega.onMessageHandlers.Delete(watchId)
			}

			rcf.Completable().Complete(nil)
		}

		if len(tf.GetRefererMessageId()) > 0 && bytes.Compare(tf.GetRefererMessageId(), refId) == 0 {
			player.tfs = append(player.tfs, tf)
			loadCount++
			bwg.Done()
			if loadCount == totalCount {
				player.omega.onMessageHandlers.Delete(watchId)
			}
		}
	})

	if v := f.GetTimeout(p.omega.Agent().Session().GetEngine().Config.TransitTimeout); v != nil {
		go func() {
			<-time.After(p.omega.Agent().Session().GetEngine().Config.TransitTimeout)
			rcf.Completable().Fail(base.ErrTransitTimeout)
			bwg.Reset()
		}()

		rcf.AwaitTimeout(p.omega.Agent().Session().GetEngine().Config.TransitTimeout)
		bwg.Wait()
	}

	p.omega.onMessageHandlers.Delete(watchId)
}

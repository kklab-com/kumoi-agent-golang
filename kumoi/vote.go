package kumoi

import (
	"fmt"
	"reflect"

	concurrent "github.com/kklab-com/goth-concurrent"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/value"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteInfo struct {
	meta  *omega.GetVoteMeta
	omega *Omega
}

func (v *VoteInfo) VoteId() string {
	return v.meta.VoteId
}

func (v *VoteInfo) Name() string {
	return v.meta.Name
}

func (v *VoteInfo) Metadata() map[string]any {
	return base.SafeGetStructMap(v.meta.Data)
}

func (v *VoteInfo) VoteOptions() []*omega.Vote_Option {
	return v.meta.VoteOptions
}

func (v *VoteInfo) CreatedAt() int64 {
	return v.meta.CreatedAt
}

func (v *VoteInfo) Join(key string) *Vote {
	if get := v.omega.Agent().JoinVote(v.VoteId(), key).Get(); get != nil {
		if jv := value.Cast[*omega.TransitFrame](get).GetJoinVote(); jv != nil {
			nv := *v
			vt := &Vote{
				key:     key,
				omega:   v.omega,
				info:    &nv,
				onLeave: func() {},
				onClose: func() {},
				watch:   func(msg messages.TransitFrame) {},
			}

			vt.init()
			return vt
		}
	}

	return nil
}

func (v *VoteInfo) Close(key string) SendFuture[*messages.CloseVote] {
	return wrapSendFuture[*messages.CloseVote](v.omega.agent.CloseVote(v.VoteId(), key))
}

type Vote struct {
	key              string
	omega            *Omega
	info             *VoteInfo
	voteOptions      []VoteOption
	onLeave, onClose func()
	watch            func(msg messages.TransitFrame)
}

func (v *Vote) watchId() string {
	return fmt.Sprintf("vt-watch-%s", v.info.VoteId())
}

func (v *Vote) Id() string {
	return v.Info().VoteId()
}

func (v *Vote) Info() *VoteInfo {
	return v.info
}

func (v *Vote) VoteOptions() []VoteOption {
	return v.voteOptions
}

func (v *Vote) Name() string {
	return v.Info().Name()
}

func (v *Vote) SetName(name string) SendFuture[*messages.SetVoteMeta] {
	return wrapSendFuture[*messages.SetVoteMeta](v.omega.Agent().SetVoteMetadata(v.Info().VoteId(), name, nil))

}

func (v *Vote) Fetch() SendFuture[*messages.GetVoteMeta] {
	return wrapSendFuture[*messages.GetVoteMeta](v.omega.Agent().GetVoteMetadata(v.Id()))
}

func (v *Vote) Metadata() map[string]any {
	return v.info.Metadata()
}

func (v *Vote) SetMetadata(metadata map[string]any) SendFuture[*messages.SetVoteMeta] {
	return wrapSendFuture[*messages.SetVoteMeta](v.omega.Agent().SetVoteMetadata(v.Info().VoteId(), "", base.NewMetadata(metadata)))
}

func (v *Vote) Leave() SendFuture[*messages.LeaveVote] {
	return wrapSendFuture[*messages.LeaveVote](v.omega.Agent().LeaveVote(v.Info().VoteId())).
		AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
			if f.IsSuccess() {
				v.invokeOnLeaveVoteSuccess()
			}
		}))
}

func (v *Vote) Close() SendFuture[*messages.CloseVote] {
	return wrapSendFuture[*messages.CloseVote](v.omega.Agent().CloseVote(v.Info().VoteId(), v.key)).
		AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
			if f.IsSuccess() {
				v.invokeOnCloseVoteSuccess()
			}
		}))
}

func (v *Vote) SendMessage(msg string) SendFuture[*messages.VoteMessage] {
	return wrapSendFuture[*messages.VoteMessage](v.omega.Agent().VoteMessage(v.Info().VoteId(), msg))
}

func (v *Vote) SendOwnerMessage(msg string) SendFuture[*messages.VoteOwnerMessage] {
	return wrapSendFuture[*messages.VoteOwnerMessage](v.omega.Agent().VoteOwnerMessage(v.Info().VoteId(), msg))
}

func (v *Vote) Count() SendFuture[*messages.VoteCount] {
	return wrapSendFuture[*messages.VoteCount](v.omega.Agent().VoteCount(v.Info().VoteId()))
}

func (v *Vote) Select(voteOptionId string) bool {
	f := v.omega.Agent().VoteSelect(v.Info().VoteId(), voteOptionId).Await()
	if !f.IsSuccess() {
		return false
	}

	return !f.Get().GetVoteSelect().Deny
}

func (v *Vote) Status(statusType omega.Vote_Status) SendFuture[*messages.VoteStatus] {
	return wrapSendFuture[*messages.VoteStatus](v.omega.Agent().VoteStatus(v.Info().VoteId(), statusType))
}

func (v *Vote) OnLeave(f func()) *Vote {
	v.onLeave = f
	return v
}

func (v *Vote) OnClose(f func()) *Vote {
	v.onClose = f
	return v
}

func (v *Vote) Watch(f func(msg messages.TransitFrame)) *Vote {
	v.watch = f
	return v
}

func (v *Vote) invokeOnLeaveVoteSuccess() {
	v.onLeave()
	v.deInit()
}

func (v *Vote) invokeOnCloseVoteSuccess() {
	v.onClose()
	v.deInit()
}

func (v *Vote) init() {
	for _, option := range v.Info().VoteOptions() {
		v.voteOptions = append(v.voteOptions, VoteOption{
			vote: v,
			Id:   option.Id,
			Name: option.Name,
		})
	}

	v.omega.onMessageHandlers.Store(v.watchId(), func(tf *omega.TransitFrame) {
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

		// has same voteId
		if tfdEVtId := tfdE.Field(0).Elem().FieldByName("VoteId"); tfdEVtId.IsValid() && tfdEVtId.String() == v.Id() {
			switch tf.GetClass() {
			case omega.TransitFrame_ClassNotification:
				if tfd := tf.GetGetVoteMeta(); tfd != nil {
					v.info.meta.Name = tfd.GetName()
					v.info.meta.Data = tfd.GetData()
					v.info.meta.CreatedAt = tfd.GetCreatedAt()
				}

				if vtf := messages.WrapTransitFrame(tf); vtf != nil {
					v.watch(vtf)
				} else {
					kklogger.WarnJ("kumoi:Vote.init", fmt.Sprintf("%s should not be here", tf.String()))
				}

				if tfd := tf.GetLeaveVote(); tfd != nil && tfd.GetSessionId() == v.omega.Session().GetId() {
					v.invokeOnLeaveVoteSuccess()
				}

				if tfd := tf.GetCloseVote(); tfd != nil {
					v.invokeOnCloseVoteSuccess()
				}
			case omega.TransitFrame_ClassResponse:
				if tfd := tf.GetGetVoteMeta(); tfd != nil {
					v.info.meta = tfd
					break
				}

				if tfd := tf.GetSetVoteMeta(); tfd != nil {
					v.info.meta.Data = tfd.Data
					v.info.meta.Name = tfd.Name
				}
			}
		}
	})
}

func (v *Vote) deInit() {
	v.omega.onMessageHandlers.Delete(v.watchId())
}

type VoteOption struct {
	vote *Vote
	Id   string
	Name string
}

func (vo *VoteOption) Select() bool {
	if vo.vote == nil {
		return false
	}

	return vo.vote.Select(vo.Id)
}

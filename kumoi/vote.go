package kumoi

import (
	"fmt"
	"reflect"

	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/value"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteInfo struct {
	voteId      string
	name        string
	metadata    *base.Metadata
	voteOptions []VoteOption
	createdAt   int64
	omega       *Omega
}

func (v *VoteInfo) VoteId() string {
	return v.voteId
}

func (v *VoteInfo) Name() string {
	return v.name
}

func (v *VoteInfo) Metadata() *base.Metadata {
	return v.metadata
}

func (v *VoteInfo) VoteOptions() []VoteOption {
	return v.voteOptions
}

func (v *VoteInfo) CreatedAt() int64 {
	return v.createdAt
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

			vt.info.name = jv.GetName()
			vt.info.metadata = jv.GetVoteMetadata()
			vt.info.voteOptions = nil
			for _, vto := range jv.GetVoteOptions() {
				vt.info.voteOptions = append(vt.info.voteOptions, VoteOption{
					vote: vt,
					Id:   vto.GetId(),
					Name: vto.GetName(),
				})
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
	onLeave, onClose func()
	watch            func(msg messages.TransitFrame)
}

func (v *Vote) watchId() string {
	return fmt.Sprintf("vt-watch-%s", v.info.VoteId())
}

func (v *Vote) Id() string {
	return v.Info().voteId
}

func (v *Vote) Info() *VoteInfo {
	return v.info
}

func (v *Vote) Name() string {
	return v.Info().name
}

func (v *Vote) SetName(name string) SendFuture[*messages.SetVoteMeta] {
	return wrapSendFuture[*messages.SetVoteMeta](v.omega.Agent().SetVoteMetadata(v.Info().voteId, name, nil))

}

func (v *Vote) Fetch() SendFuture[*messages.GetVoteMeta] {
	return wrapSendFuture[*messages.GetVoteMeta](v.omega.Agent().GetVoteMetadata(v.Id()))
}

func (v *Vote) Metadata() *base.Metadata {
	return v.info.Metadata()
}

func (v *Vote) SetMetadata(meta *base.Metadata) SendFuture[*messages.SetVoteMeta] {
	return wrapSendFuture[*messages.SetVoteMeta](v.omega.Agent().SetVoteMetadata(v.Info().voteId, "", meta))
}

func (v *Vote) Leave() SendFuture[*messages.LeaveVote] {
	return wrapSendFuture[*messages.LeaveVote](v.omega.Agent().LeaveVote(v.Info().voteId))
}

func (v *Vote) Close() SendFuture[*messages.CloseVote] {
	return wrapSendFuture[*messages.CloseVote](v.omega.Agent().CloseVote(v.Info().voteId, v.key))
}

func (v *Vote) SendMessage(msg string) SendFuture[*messages.VoteMessage] {
	return wrapSendFuture[*messages.VoteMessage](v.omega.Agent().VoteMessage(v.Info().voteId, msg))
}

func (v *Vote) SendOwnerMessage(msg string) SendFuture[*messages.VoteOwnerMessage] {
	return wrapSendFuture[*messages.VoteOwnerMessage](v.omega.Agent().VoteOwnerMessage(v.Info().voteId, msg))
}

func (v *Vote) Count() SendFuture[*messages.VoteCount] {
	return wrapSendFuture[*messages.VoteCount](v.omega.Agent().VoteCount(v.Info().voteId))
}

func (v *Vote) Select(voteOptionId string) bool {
	f := v.omega.Agent().VoteSelect(v.Info().voteId, voteOptionId).Await()
	if !f.IsSuccess() {
		return false
	}

	return !f.Get().GetVoteSelect().Deny
}

func (v *Vote) Status(statusType omega.Vote_Status) SendFuture[*messages.VoteStatus] {
	return wrapSendFuture[*messages.VoteStatus](v.omega.Agent().VoteStatus(v.Info().voteId, statusType))
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

func (v *Vote) init() {
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
		if tfdEChId := tfdE.Field(0).Elem().FieldByName("VoteId"); tfdEChId.IsValid() && tfdEChId.String() == v.Id() {
			switch tf.GetClass() {
			case omega.TransitFrame_ClassNotification:
				if tfd := tf.GetGetVoteMeta(); tfd != nil {
					var vtos []VoteOption
					for _, vto := range tfd.GetVoteOptions() {
						vtos = append(vtos, VoteOption{
							vote: v,
							Id:   vto.GetId(),
							Name: vto.GetName(),
						})
					}

					v.info.name = tfd.GetName()
					v.info.metadata = tfd.GetData()
					v.info.voteOptions = vtos
					v.info.createdAt = tfd.GetCreatedAt()
				}

				if tfd := tf.GetVoteCount(); tfd != nil {
					var vtos []VoteOption
					for _, vto := range tf.GetVoteCount().GetVoteOptions() {
						vtos = append(vtos, VoteOption{
							vote: v,
							Id:   vto.GetId(),
							Name: vto.GetName(),
						})
					}

					v.info.voteOptions = vtos
				}

				if vtf := getParsedTransitFrameFromBaseTransitFrame(tf); vtf != nil {
					v.watch(vtf)
				} else {
					kklogger.WarnJ("kumoi:Vote.init", fmt.Sprintf("%s should not be here", tf.String()))
				}

				if tfd := tf.GetLeaveVote(); tfd != nil {
					v.onLeave()
					v.deInit()
				}

				if tfd := tf.GetCloseVote(); tfd != nil {
					v.onClose()
					v.deInit()
				}
			case omega.TransitFrame_ClassResponse:
				if tfd := tf.GetGetVoteMeta(); tfd != nil {
					v.info.name = tfd.GetName()
					v.info.metadata = tfd.GetData()
					v.info.createdAt = tfd.GetCreatedAt()
				}

				if tfd := tf.GetLeaveVote(); tfd != nil {
					v.onLeave()
					v.deInit()
				}

				if tfd := tf.GetCloseVote(); tfd != nil {
					v.onClose()
					v.deInit()
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

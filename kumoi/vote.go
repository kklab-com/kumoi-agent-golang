package kumoi

import (
	"reflect"

	"github.com/kklab-com/goth-kkutil/concurrent"
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
		if jv := get.(*omega.TransitFrame).GetJoinVote(); jv != nil {
			nv := *v
			vt := &Vote{
				key:     key,
				omega:   v.omega,
				info:    &nv,
				onLeave: func() {},
				onClose: func() {},
				watch:   func(msg messages.VoteFrame) {},
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

func (v *VoteInfo) Close(key string) concurrent.Future {
	return v.omega.agent.CloseVote(v.VoteId(), key)
}

type Vote struct {
	key              string
	omega            *Omega
	info             *VoteInfo
	onLeave, onClose func()
	watch            func(msg messages.VoteFrame)
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

func (v *Vote) SetName(name string) bool {
	if v.omega.Agent().SetVoteMetadata(v.Info().voteId, name, nil).Await().IsSuccess() {
		v.info.name = name
		return true
	}

	return false
}

func (v *Vote) Metadata() *base.Metadata {
	return v.info.Metadata()
}

func (v *Vote) SetMetadata(meta *base.Metadata) bool {
	return v.omega.Agent().SetVoteMetadata(v.Info().voteId, "", meta).Await().IsSuccess()
}

func (v *Vote) Leave() bool {
	f := v.omega.Agent().LeaveVote(v.Info().voteId)
	f.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			v.onLeave()
			v.deInit()
		}
	}))

	return f.Await().IsSuccess()
}

func (v *Vote) Close() bool {
	f := v.omega.Agent().CloseVote(v.Info().voteId, v.key)
	f.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			v.deInit()
		}
	}))

	return f.Await().IsSuccess()
}

func (v *Vote) SendMessage(msg string) bool {
	return v.omega.Agent().VoteMessage(v.Info().voteId, msg).Await().IsSuccess()
}

func (v *Vote) SendOwnerMessage(msg string) bool {
	return v.omega.Agent().VoteOwnerMessage(v.Info().voteId, msg).Await().IsSuccess()
}

func (v *Vote) GetCount() *messages.VoteCount {
	if tf := convertTransitFrame(v.omega.Agent().VoteCount(v.Info().voteId).Get()); tf != nil {
		r := &messages.VoteCount{}
		r.ParseTransitFrame(tf)
		return r
	}

	return nil
}

func (v *Vote) Select(voteOptionId string) bool {
	f := v.omega.Agent().VoteSelect(v.Info().voteId, voteOptionId).Await()
	if !f.IsSuccess() {
		return false
	}

	return !f.Get().(*omega.TransitFrame).GetVoteSelect().Deny
}

func (v *Vote) Status(statusType omega.Vote_Status) bool {
	f := v.omega.Agent().VoteStatus(v.Info().voteId, statusType).Await()
	return f.IsSuccess()
}

func (v *Vote) OnLeave(f func()) *Vote {
	v.onLeave = f
	return v
}

func (v *Vote) OnClose(f func()) *Vote {
	v.onClose = f
	return v
}

func (v *Vote) Watch(f func(msg messages.VoteFrame)) *Vote {
	v.watch = f
	return v
}

func (v *Vote) init() {
	v.omega.onMessageHandlers.Store(v.Id(), func(tf *omega.TransitFrame) {
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
				if tfd := tf.GetCloseVote(); tfd != nil {
					v.onClose()
					v.deInit()
				}

				if tfd := tf.GetLeaveVote(); tfd != nil {
					v.onLeave()
					v.deInit()
				}

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

				var vf messages.VoteFrame
				switch tf.GetData().(type) {
				case *omega.TransitFrame_VoteCount:
					vf = &messages.VoteCount{}
				case *omega.TransitFrame_VoteMessage:
					vf = &messages.VoteMessage{}
				case *omega.TransitFrame_VoteOwnerMessage:
					vf = &messages.VoteOwnerMessage{}
				case *omega.TransitFrame_GetVoteMeta:
					vf = &messages.GetVoteMeta{}
				case *omega.TransitFrame_JoinVote:
					vf = &messages.JoinVote{}
				case *omega.TransitFrame_LeaveVote:
					vf = &messages.LeaveVote{}
				case *omega.TransitFrame_CloseVote:
					vf = &messages.CloseVote{}
				case *omega.TransitFrame_VoteStatus:
					vf = &messages.VoteStatus{}
				case *omega.TransitFrame_VoteSelect:
					vf = &messages.VoteSelect{}
				}

				if vf != nil {
					if ccf, ok := vf.(messages.TransitFrameParsable); ok {
						ccf.ParseTransitFrame(tf)
					}
				}

				v.watch(vf)
			case omega.TransitFrame_ClassResponse:
				if tfd := tf.GetGetVoteMeta(); tfd != nil {
					v.info.name = tfd.GetName()
					v.info.metadata = tfd.GetData()
					v.info.createdAt = tfd.GetCreatedAt()
				}
			}
		}
	})
}

func (v *Vote) deInit() {
	v.omega.onMessageHandlers.Delete(v.Id())
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

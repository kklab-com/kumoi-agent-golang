package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type VoteMessage struct {
	TransitFrame
}

func (x *VoteMessage) GetVoteId() string {
	return x.BaseTransitFrame().GetVoteMessage().GetVoteId()
}

func (x *VoteMessage) GetFromSession() string {
	return x.BaseTransitFrame().GetVoteMessage().GetFromSession()
}

func (x *VoteMessage) GetSubject() string {
	return x.BaseTransitFrame().GetVoteMessage().GetSubject()
}

func (x *VoteMessage) GetMessage() string {
	return x.BaseTransitFrame().GetVoteMessage().GetMessage()
}

func (x *VoteMessage) GetSequential() bool {
	return x.BaseTransitFrame().GetVoteMessage().GetSequential()
}

func (x *VoteMessage) GetRoleIndicator() omega.Role {
	return x.BaseTransitFrame().GetVoteMessage().GetRoleIndicator()
}

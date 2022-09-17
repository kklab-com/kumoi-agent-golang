package messages

type VoteOwnerMessage struct {
	TransitFrame
}

func (x *VoteOwnerMessage) GetVoteId() string {
	return x.BaseTransitFrame().GetVoteOwnerMessage().GetVoteId()
}

func (x *VoteOwnerMessage) GetMessage() string {
	return x.BaseTransitFrame().GetVoteOwnerMessage().GetMessage()
}

func (x *VoteOwnerMessage) GetSequential() bool {
	return x.BaseTransitFrame().GetVoteOwnerMessage().GetSequential()
}

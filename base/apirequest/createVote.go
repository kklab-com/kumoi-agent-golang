package apirequest

type CreateVote struct {
	Name              string             `json:"name"`
	Options           ResourceOptions    `json:"options,omitempty"`
	VoteOptions       []CreateVoteOption `json:"vote_options,omitempty"`
	IdleTimeoutSecond int32              `json:"idle_timeout,omitempty"`
}

type CreateVoteOption struct {
	Name string `json:"name,omitempty"`
}

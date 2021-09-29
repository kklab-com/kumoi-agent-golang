package apiresponse

type CreateVote struct {
	AppId             string              `json:"app_id"`
	Name              string              `json:"name"`
	VoteId            string              `json:"vote_id"`
	VoteOptions       []CreateVoteOptions `json:"vote_options,omitempty"`
	Key               string              `json:"key,omitempty"`
	IdleTimeoutSecond int               `json:"idle_timeout,omitempty"`
}

type CreateVoteOptions struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

package apiresponse

type CreateChannel struct {
	AppId             string `json:"app_id,omitempty"`
	ChannelId         string `json:"channel_id,omitempty"`
	Name              string `json:"name,omitempty"`
	OwnerKey          string `json:"owner_key,omitempty"`
	ParticipatorKey   string `json:"participator_key,omitempty"`
	IdleTimeoutSecond int    `json:"idle_timeout,omitempty"`
}

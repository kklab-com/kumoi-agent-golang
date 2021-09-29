package apirequest

type CreateChannel struct {
	Name                 string          `json:"name"`
	Options              ResourceOptions `json:"options,omitempty"`
	ParticipatorRestrict bool            `json:"participator_restrict,omitempty"`
	IdleTimeoutSecond    int             `json:"idle_timeout,omitempty"`
}

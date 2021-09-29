package apirequest

type ResourceOptions struct {
	Notification ResourceOptionsNotification `json:"notification"`
}

type ResourceOptionsNotification struct {
	Join  bool `json:"join"`
	Leave bool `json:"leave"`
}

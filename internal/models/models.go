package models

type Webhook struct {
	ID     int64  `json:"id"`
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

type RawEvent struct {
	ID      int64  `json:"id"`
	HookID  int64  `json:"hook_id"`
	Payload string `json:"payload"`
}

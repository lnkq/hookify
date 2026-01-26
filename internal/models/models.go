package models

import "encoding/json"

type Webhook struct {
	ID     int64  `json:"id"`
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

type RawEvent struct {
	ID        int64  `json:"id"`
	WebhookID int64  `json:"webhook_id"`
	Payload   string `json:"payload"`
}

func (e *RawEvent) UnmarshalJSON(data []byte) error {
	type rawEventAlias RawEvent

	var v struct {
		rawEventAlias
		HookID int64 `json:"hook_id"`
	}

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*e = RawEvent(v.rawEventAlias)
	if e.WebhookID == 0 && v.HookID != 0 {
		e.WebhookID = v.HookID
	}

	return nil
}

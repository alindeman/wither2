package slack

import (
	"net/http"
)

type WebhookPayload struct {
	Token    string
	UserName string
	Text     string
}

func ParseWebhookPayload(r *http.Request) *WebhookPayload {
	return &WebhookPayload{
		Token:    r.FormValue("token"),
		UserName: r.FormValue("user_name"),
		Text:     r.FormValue("text"),
	}
}

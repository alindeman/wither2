package slack

import (
	"bufio"
	"io"
	"strings"
)

type WebhookPayload struct {
	Token    string
	UserName string
	Text     string
}

func ParseWebhookPayload(r io.Reader) (*WebhookPayload, error) {
	payload := new(WebhookPayload)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "token":
			payload.Token = parts[1]
		case "user_name":
			payload.UserName = parts[1]
		case "text":
			payload.Text = parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return payload, nil
}

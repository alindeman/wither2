package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	WebhookURL string
}

func (c *Client) Post(ctx context.Context, text string) error {
	buf, err := json.Marshal(map[string]interface{}{
		"text": text,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.WebhookURL, bytes.NewReader(buf))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}

	return nil
}

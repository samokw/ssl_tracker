package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Sender interface {
	Send(webhookURL, message string) error
}

type DiscordSender struct {
}

func (d *DiscordSender) Send(webhookURL, message string) error {
	content := map[string]interface{}{
		"content": message,
	}
	discordJson, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(discordJson))
	if err != nil {
		return fmt.Errorf("failed to create a request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get a response: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned %d %s", resp.StatusCode, resp.Status)
	}
	return nil
}

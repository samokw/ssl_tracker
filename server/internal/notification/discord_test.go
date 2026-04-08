package notification

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscordSender_Success(t *testing.T) {
	expectedMessage := "test message"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload struct {
			Content string `json:"content"`
		}

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, expectedMessage, payload.Content)
		w.WriteHeader(http.StatusNoContent)
	}))

	defer srv.Close()

	discordSender := &DiscordSender{}
	err := discordSender.Send(srv.URL, expectedMessage)
	assert.NoError(t, err)
}

func TestDiscordSender_Non2xx_ReturnsError(t *testing.T) {
	expectedMessage := "test message"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var payload struct {
			Content string `json:"content"`
		}
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, expectedMessage, payload.Content)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	discordSender := &DiscordSender{}
	err := discordSender.Send(srv.URL, expectedMessage)
	assert.Error(t, err)
}

func TestDiscordSenderNetworkError_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	discordSender := &DiscordSender{}
	err := discordSender.Send(url, "msg")
	assert.Error(t, err)
}

package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"tblocker/config"
	"testing"
)

func TestSendTelegramNotification(t *testing.T) {
	var gotPath string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldBase := telegramAPIBase
	telegramAPIBase = server.URL
	defer func() { telegramAPIBase = oldBase }()

	config.SendTelegram = true
	config.TelegramBotToken = "123:ABC"
	config.TelegramChatID = "-1001234"
	config.TelegramTemplate = "user %s ip %s server %s action %s dur %d ts %s"
	config.Hostname = "test-host"
	config.BlockDuration = 10
	defer func() {
		config.SendTelegram = false
		config.TelegramBotToken = ""
		config.TelegramChatID = ""
	}()

	SendTelegramNotification("testuser", "1.2.3.4", "block")

	if gotPath != "/bot123:ABC/sendMessage" {
		t.Errorf("Expected path '/bot123:ABC/sendMessage', got '%s'", gotPath)
	}

	var payload struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("Failed to parse request body: %v", err)
	}

	if payload.ChatID != "-1001234" {
		t.Errorf("Expected chat_id '-1001234', got '%s'", payload.ChatID)
	}

	expectedPrefix := "user testuser ip 1.2.3.4 server test-host action block dur 10 ts "
	if len(payload.Text) < len(expectedPrefix) || payload.Text[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Unexpected message text: '%s'", payload.Text)
	}
}

func TestSendTelegramNotificationDisabled(t *testing.T) {
	requested := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = true
	}))
	defer server.Close()

	oldBase := telegramAPIBase
	telegramAPIBase = server.URL
	defer func() { telegramAPIBase = oldBase }()

	config.SendTelegram = false

	SendTelegramNotification("testuser", "1.2.3.4", "block")

	if requested {
		t.Error("Expected no request when SendTelegram is false")
	}
}

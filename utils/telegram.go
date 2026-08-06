package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"tblocker/config"
	"time"
)

var telegramAPIBase = "https://api.telegram.org"

var telegramClient = &http.Client{Timeout: 10 * time.Second}

func SendTelegramNotification(username string, ip string, action string) {
	if !config.SendTelegram || config.TelegramBotToken == "" || config.TelegramChatID == "" {
		return
	}

	cleanUsername := processUsernameForWebhook(username)

	text := fmt.Sprintf(
		config.TelegramTemplate,
		cleanUsername,
		ip,
		config.Hostname,
		action,
		config.BlockDuration,
		time.Now().Format(time.RFC3339),
	)

	body, err := json.Marshal(struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: config.TelegramChatID,
		Text:   text,
	})
	if err != nil {
		log.Printf("Error building telegram message: %v", err)
		return
	}

	url := telegramAPIBase + "/bot" + config.TelegramBotToken + "/sendMessage"
	resp, err := telegramClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		// The error may embed the request URL, which contains the bot token.
		log.Printf("Error sending telegram notification (request failed)")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Telegram API returned unexpected status code: %d", resp.StatusCode)
	}
}

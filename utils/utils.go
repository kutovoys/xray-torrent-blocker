package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"tblocker/config"
	"tblocker/storage"
	"time"

	"github.com/hpcloud/tail"
)

var ipStorage *storage.IPStorage

func SetIPStorage(storage *storage.IPStorage) {
	ipStorage = storage
}

func StartLogMonitor() {
	t, err := tail.TailFile(config.LogFile, tail.Config{
		Follow:    true,
		ReOpen:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
	})
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	for line := range t.Lines {
		if strings.Contains(line.Text, config.TorrentTag) {
			handleLogEntry(line.Text)
		}
	}
}

func handleLogEntry(line string) {
	ip := config.IpRegex.FindString(line)
	var tid []string
	var username []string

	if config.TidRegex != nil {
		tid = config.TidRegex.FindStringSubmatch(line)
	}

	if config.UsernameRegex != nil {
		username = config.UsernameRegex.FindStringSubmatch(line)
	}

	if ip == "" || len(username) < 2 {
		log.Println("Invalid log entry format: IP or username missing")
		return
	}

	if IsBypassedIP(ip) {
		// log.Printf("IP %s is in the bypass list. Skipping...\n", ip)
		// printing removed due to large amount of log strings
		return
	}

	if ipStorage.IsBlocked(ip) {
		log.Printf("User %s with IP: %s is already blocked. Skipping...\n", username[1], ip)
		return
	}

	if err := ipStorage.AddBlockedIP(ip, username[1], time.Duration(config.BlockDuration)*time.Minute); err != nil {
		log.Printf("Error saving blocked IP to storage: %v", err)
	}

	if config.SendUserMessage && len(tid) >= 2 {
		go SendTelegramMessage(tid[1], config.Message, config.BotToken, "HTML", true)
	}

	if config.SendAdminMessage {
		adminMsg := fmt.Sprintf(config.AdminBlockTemplate, username[1], ip, config.Hostname, username[1])
		go SendTelegramMessage(config.AdminChatID, adminMsg, config.AdminBotToken, "HTML", true)
	}

	go BlockIP(ip)
	log.Printf("User %s with IP: %s blocked for %d minutes\n", username[1], ip, config.BlockDuration)

	if config.SendWebhook {
		go SendWebhook(username[1], ip, "block")
	}

	go UnblockIPAfterDelay(ip, time.Duration(config.BlockDuration)*time.Minute, username[1])
}

func ScheduleBlockedIPsUpdate() {
	UpdateBlockedIPs()
	go func() {
		for range time.Tick(time.Duration(config.BlockDuration) * time.Minute) {
			UpdateBlockedIPs()
		}
	}()
}

func UpdateBlockedIPs() {
	cmd := exec.Command("ufw", "status")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error checking ufw status: %v", err)
		return
	}

	currentBlockedIPs := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		ip := config.IpRegex.FindString(line)
		if ip != "" {
			currentBlockedIPs[ip] = true
		}
	}

	blockedInStorage := ipStorage.GetBlockedIPs()

	for ip, info := range blockedInStorage {
		if time.Now().Before(info.BlockedUntil) && !currentBlockedIPs[ip] {
			go BlockIP(ip)
			log.Printf("Restoring block for IP: %s (user: %s)\n", ip, info.Username)
		}
	}
}

func SendTelegramMessage(chatID string, message string, botToken string, parseMode string, disablePreview bool) {
	urlStr := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("text", message)
	data.Set("parse_mode", parseMode)
	if disablePreview {
		data.Set("disable_web_page_preview", "true")
	}

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending HTTP request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
	}
}

func IsBypassedIP(ip string) bool {
	_, exists := config.BypassIPSet[ip]
	return exists
}

func BlockIP(ip string) {
	var cmd *exec.Cmd

	if config.BlockMode == "iptables" {
		cmd = exec.Command("iptables", "-I", "INPUT", "-s", ip, "-j", "DROP")
	} else {
		cmd = exec.Command("ufw", "insert", "1", "deny", "from", ip, "to", "any")
	}

	err := cmd.Run()
	if err != nil {
		log.Printf("Error blocking IP: %v", err)
		return
	}
}

func UnblockIPAfterDelay(ip string, delay time.Duration, username string) {
	time.Sleep(delay)

	if ipStorage.IsBlocked(ip) {
		log.Printf("Skipping unblock for IP %s as it has an active block", ip)
		return
	}

	var cmd *exec.Cmd

	if config.BlockMode == "iptables" {
		cmd = exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP")
	} else {
		cmd = exec.Command("ufw", "delete", "deny", "from", ip, "to", "any")
	}

	err := cmd.Run()
	if err != nil {
		log.Printf("Error unblocking IP: %v", err)
		return
	}

	if err := ipStorage.RemoveBlockedIP(ip); err != nil {
		log.Printf("Error removing IP from storage: %v", err)
	}

	log.Printf("User %s with IP: %s has been unblocked\n", username, ip)

	if config.SendWebhook {
		go SendWebhook(username, ip, "unblock")
	}

	if config.SendAdminMessage {
		adminMsg := fmt.Sprintf(config.AdminUnblockTemplate, username, ip, config.Hostname, username)
		go SendTelegramMessage(config.AdminChatID, adminMsg, config.AdminBotToken, "HTML", true)
	}
}

func SendWebhook(username string, ip string, action string) {
	if !config.SendWebhook || config.WebhookURL == "" {
		return
	}

	payload := fmt.Sprintf(
		config.WebhookTemplate,
		username,
		ip,
		config.Hostname,
		action,
		config.BlockDuration,
		time.Now().Format(time.RFC3339),
	)

	req, err := http.NewRequest("POST", config.WebhookURL, strings.NewReader(payload))
	if err != nil {
		log.Printf("Error creating webhook request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.WebhookHeaders {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Webhook returned unexpected status code: %d", resp.StatusCode)
	}
}

package utils

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"tblocker/config"
	"tblocker/firewall"
	"tblocker/storage"
	"time"

	"github.com/hpcloud/tail"
)

var ipStorage *storage.IPStorage
var firewallManager *firewall.Manager

func SetIPStorage(storage *storage.IPStorage) {
	ipStorage = storage
}

func SetFirewallManager(manager *firewall.Manager) {
	firewallManager = manager
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
	var username []string

	if config.UsernameRegex != nil {
		username = config.UsernameRegex.FindStringSubmatch(line)
	}

	if ip == "" || len(username) < 2 {
		log.Println("Invalid log entry format: IP or username missing")
		return
	}

	if IsBypassedIP(ip) {
		return
	}

	if ipStorage.IsBlocked(ip) {
		log.Printf("User %s with IP: %s is already blocked. Skipping...\n", username[1], ip)
		return
	}

	if err := ipStorage.AddBlockedIP(ip, username[1], time.Duration(config.BlockDuration)*time.Minute); err != nil {
		log.Printf("Error saving blocked IP to storage: %v", err)
	}

	go BlockIP(ip)
	log.Printf("User %s with IP: %s blocked for %d minutes\n", username[1], ip, config.BlockDuration)

	if config.SendWebhook {
		go SendWebhook(username[1], ip, "block")
	}

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
	if firewallManager == nil {
		log.Printf("Firewall manager not initialized")
		return
	}

	currentBlockedIPs, err := firewallManager.GetBlockedIPs()
	if err != nil {
		log.Printf("Error checking firewall status: %v", err)
		return
	}

	blockedInStorage := ipStorage.GetBlockedIPs()

	for ip, info := range blockedInStorage {
		if time.Now().Before(info.BlockedUntil) && !currentBlockedIPs[ip] {
			log.Printf("Restoring block for IP: %s (user: %s) using %s", ip, info.Username, firewallManager.GetFirewallName())
			go BlockIP(ip)
		}
	}
}

func IsBypassedIP(ip string) bool {
	_, exists := config.BypassIPSet[ip]
	return exists
}

func BlockIP(ip string) {
	if firewallManager == nil {
		log.Printf("Firewall manager not initialized")
		return
	}

	err := firewallManager.BlockIP(ip)
	if err != nil {
		log.Printf("Error blocking IP %s: %v", ip, err)
		return
	}

	if conntrackManager != nil && conntrackManager.IsAvailable() {
		if err := conntrackManager.DropConnections(ip); err != nil {
			log.Printf("Warning: failed to drop connections for IP %s: %v", ip, err)
		}
	}
}

func UnblockIPAfterDelay(ip string, delay time.Duration, username string) {
	time.Sleep(delay)

	if ipStorage.IsBlocked(ip) {
		log.Printf("Skipping unblock for IP %s as it has an active block", ip)
		return
	}

	if firewallManager == nil {
		log.Printf("Firewall manager not initialized")
		return
	}

	blockedIPs := ipStorage.GetBlockedIPs()
	if _, exists := blockedIPs[ip]; !exists {
		log.Printf("IP %s not found in storage, skipping unblock", ip)
		return
	}

	err := firewallManager.UnblockIP(ip)
	if err != nil {
		if strings.Contains(err.Error(), "no rule found") || strings.Contains(err.Error(), "exit status 1") {
			log.Printf("IP %s already unblocked or rule not found, continuing...", ip)
		} else {
			log.Printf("Error unblocking IP %s: %v", ip, err)
			return
		}
	}

	if err := ipStorage.RemoveBlockedIP(ip); err != nil {
		log.Printf("Error removing IP from storage: %v", err)
	}

	log.Printf("User %s with IP: %s has been unblocked\n", username, ip)

	if config.SendWebhook {
		go SendWebhook(username, ip, "unblock")
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

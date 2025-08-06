package main

import (
	"os"
	"path/filepath"
	"tblocker/config"
	"tblocker/firewall"
	"tblocker/storage"
	"tblocker/utils"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 1
TorrentTag: "TORRENT"
UsernameRegex: "user: (\\S+)"
BlockMode: "iptables"
BypassIPS:
  - "127.0.0.1"
  - "192.168.1.1"
SendWebhook: false
StorageDir: "/tmp/test_storage"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	err = config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	firewallManager, err := firewall.NewManager(config.BlockMode)
	if err != nil {
		t.Fatalf("Failed to create firewall manager: %v", err)
	}
	utils.SetFirewallManager(firewallManager)

	storageDir := filepath.Join(tempDir, "storage")
	store, err := storage.NewIPStorage(storageDir, utils.UnblockIPAfterDelay)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	utils.SetIPStorage(store)

	if store == nil {
		t.Fatal("Storage is nil")
	}

	if firewallManager == nil {
		t.Fatal("Firewall manager is nil")
	}

	err = store.AddBlockedIP("192.168.1.100", "testuser", 1*time.Minute)
	if err != nil {
		t.Fatalf("Failed to add blocked IP: %v", err)
	}

	if !store.IsBlocked("192.168.1.100") {
		t.Error("IP should be blocked")
	}

	if store.IsBlocked("127.0.0.1") {
		t.Error("Bypass IP should not be blocked")
	}

	if !utils.IsBypassedIP("127.0.0.1") {
		t.Error("Expected 127.0.0.1 to be bypassed")
	}

	if utils.IsBypassedIP("192.168.1.200") {
		t.Error("Expected 192.168.1.200 to not be bypassed")
	}

	blockedIPs := store.GetBlockedIPs()
	if len(blockedIPs) != 1 {
		t.Errorf("Expected 1 blocked IP, got %d", len(blockedIPs))
	}

	blockedIP, exists := blockedIPs["192.168.1.100"]
	if !exists {
		t.Error("Blocked IP not found in storage")
	}

	if blockedIP.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", blockedIP.Username)
	}

	err = store.RemoveBlockedIP("192.168.1.100")
	if err != nil {
		t.Fatalf("Failed to remove blocked IP: %v", err)
	}

	if store.IsBlocked("192.168.1.100") {
		t.Error("IP should not be blocked after removal")
	}

	blockedIPs = store.GetBlockedIPs()
	if len(blockedIPs) != 0 {
		t.Errorf("Expected 0 blocked IPs, got %d", len(blockedIPs))
	}

	t.Logf("Integration test completed successfully using firewall: %s", firewallManager.GetFirewallName())
}

func TestIntegrationWithExpiration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 1
TorrentTag: "TORRENT"
UsernameRegex: "user: (\\S+)"
BlockMode: "iptables"
SendWebhook: false
StorageDir: "/tmp/test_storage"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	err = config.LoadConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	firewallManager, err := firewall.NewManager(config.BlockMode)
	if err != nil {
		t.Fatalf("Failed to create firewall manager: %v", err)
	}
	utils.SetFirewallManager(firewallManager)

	storageDir := filepath.Join(tempDir, "storage")
	store, err := storage.NewIPStorage(storageDir, utils.UnblockIPAfterDelay)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	utils.SetIPStorage(store)

	err = store.AddBlockedIP("192.168.1.200", "testuser", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to add blocked IP: %v", err)
	}

	if !store.IsBlocked("192.168.1.200") {
		t.Error("IP should be blocked initially")
	}

	time.Sleep(50 * time.Millisecond)

	if store.IsBlocked("192.168.1.200") {
		t.Error("IP should not be blocked after expiration")
	}

	t.Log("Integration test with expiration completed successfully")
}

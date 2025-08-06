package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 15
TorrentTag: "TEST_TORRENT"
UsernameRegex: "user: (\\S+)"
BlockMode: "iptables"
BypassIPS:
  - "127.0.0.1"
  - "192.168.1.1"
SendWebhook: true
WebhookURL: "https://test.com/webhook"
WebhookTemplate: '{"test":"%s"}'
StorageDir: "/tmp/test"
WebhookHeaders:
  Authorization: "Bearer test-token"
`

	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tmpFile.Close()

	err = LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if LogFile != "/var/log/test.log" {
		t.Errorf("Expected LogFile '/var/log/test.log', got '%s'", LogFile)
	}

	if BlockDuration != 15 {
		t.Errorf("Expected BlockDuration 15, got %d", BlockDuration)
	}

	if TorrentTag != "TEST_TORRENT" {
		t.Errorf("Expected TorrentTag 'TEST_TORRENT', got '%s'", TorrentTag)
	}

	if BlockMode != "iptables" {
		t.Errorf("Expected BlockMode 'iptables', got '%s'", BlockMode)
	}

	if !SendWebhook {
		t.Error("Expected SendWebhook to be true")
	}

	if WebhookURL != "https://test.com/webhook" {
		t.Errorf("Expected WebhookURL 'https://test.com/webhook', got '%s'", WebhookURL)
	}

	if StorageDir != "/tmp/test" {
		t.Errorf("Expected StorageDir '/tmp/test', got '%s'", StorageDir)
	}

	if _, exists := BypassIPSet["127.0.0.1"]; !exists {
		t.Error("Expected 127.0.0.1 to be in BypassIPSet")
	}

	if _, exists := BypassIPSet["192.168.1.1"]; !exists {
		t.Error("Expected 192.168.1.1 to be in BypassIPSet")
	}

	if WebhookHeaders["Authorization"] != "Bearer test-token" {
		t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", WebhookHeaders["Authorization"])
	}
}

func TestLoadConfigWithDefaults(t *testing.T) {
	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 10
TorrentTag: "TORRENT"
`

	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tmpFile.Close()

	err = LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if BlockMode != "iptables" {
		t.Errorf("Expected default BlockMode 'iptables', got '%s'", BlockMode)
	}

	if SendWebhook {
		t.Error("Expected default SendWebhook to be false")
	}

	if StorageDir != "/opt/tblocker" {
		t.Errorf("Expected default StorageDir '/opt/tblocker', got '%s'", StorageDir)
	}

	if UsernameRegex == nil {
		t.Error("Expected UsernameRegex to be compiled")
	}
}

func TestLoadConfigInvalidFile(t *testing.T) {
	err := LoadConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error when loading nonexistent file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: "invalid"
`

	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tmpFile.Close()

	err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}

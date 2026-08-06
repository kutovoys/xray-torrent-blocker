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
IgnoreEmail: true
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

	if !IgnoreEmail {
		t.Error("Expected IgnoreEmail to be true")
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

func TestLoadConfigAPIAndTelegram(t *testing.T) {
	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 10
TorrentTag: "TORRENT"
EnableAPI: true
APIAddress: "0.0.0.0:9090"
APIToken: "secret-token"
SendTelegram: true
TelegramBotToken: "123:ABC"
TelegramChatID: "-1001234"
TelegramTemplate: "user %s ip %s server %s action %s dur %d ts %s"
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

	if !EnableAPI {
		t.Error("Expected EnableAPI to be true")
	}
	if APIAddress != "0.0.0.0:9090" {
		t.Errorf("Expected APIAddress '0.0.0.0:9090', got '%s'", APIAddress)
	}
	if APIToken != "secret-token" {
		t.Errorf("Expected APIToken 'secret-token', got '%s'", APIToken)
	}
	if !SendTelegram {
		t.Error("Expected SendTelegram to be true")
	}
	if TelegramBotToken != "123:ABC" {
		t.Errorf("Expected TelegramBotToken '123:ABC', got '%s'", TelegramBotToken)
	}
	if TelegramChatID != "-1001234" {
		t.Errorf("Expected TelegramChatID '-1001234', got '%s'", TelegramChatID)
	}
	if TelegramTemplate != "user %s ip %s server %s action %s dur %d ts %s" {
		t.Errorf("Unexpected TelegramTemplate: '%s'", TelegramTemplate)
	}
}

func TestLoadConfigAPITelegramDefaults(t *testing.T) {
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

	if EnableAPI {
		t.Error("Expected default EnableAPI to be false")
	}
	if APIAddress != "127.0.0.1:8085" {
		t.Errorf("Expected default APIAddress '127.0.0.1:8085', got '%s'", APIAddress)
	}
	if SendTelegram {
		t.Error("Expected default SendTelegram to be false")
	}
	if TelegramTemplate == "" {
		t.Error("Expected default TelegramTemplate to be set")
	}
}

func TestLoadConfigAPIWithoutToken(t *testing.T) {
	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 10
TorrentTag: "TORRENT"
EnableAPI: true
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
		t.Error("Expected error when EnableAPI is true without APIToken")
	}
}

func TestLoadConfigTelegramWithoutCredentials(t *testing.T) {
	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 10
TorrentTag: "TORRENT"
SendTelegram: true
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
		t.Error("Expected error when SendTelegram is true without credentials")
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

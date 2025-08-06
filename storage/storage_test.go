package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewIPStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	unblockFunc := func(ip string, delay time.Duration, username string) {
	}

	storage, err := NewIPStorage(tempDir, unblockFunc)
	if err != nil {
		t.Fatalf("Failed to create IP storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Storage is nil")
	}

	if storage.filepath != filepath.Join(tempDir, "blocked_ips.json") {
		t.Errorf("Expected filepath %s, got %s", filepath.Join(tempDir, "blocked_ips.json"), storage.filepath)
	}

	err = storage.AddBlockedIP("192.168.1.100", "testuser", 1*time.Minute)
	if err != nil {
		t.Fatalf("Failed to add test IP: %v", err)
	}

	if _, err := os.Stat(storage.filepath); os.IsNotExist(err) {
		t.Error("Storage file was not created after adding data")
	}
}

func TestAddBlockedIP(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	unblockFunc := func(ip string, delay time.Duration, username string) {
	}

	storage, err := NewIPStorage(tempDir, unblockFunc)
	if err != nil {
		t.Fatalf("Failed to create IP storage: %v", err)
	}

	err = storage.AddBlockedIP("192.168.1.100", "testuser", 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to add blocked IP: %v", err)
	}

	if !storage.IsBlocked("192.168.1.100") {
		t.Error("IP should be blocked")
	}

	if storage.IsBlocked("192.168.1.200") {
		t.Error("Non-existent IP should not be blocked")
	}

	blockedIPs := storage.GetBlockedIPs()
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

	if blockedIP.IP != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", blockedIP.IP)
	}
}

func TestRemoveBlockedIP(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	unblockFunc := func(ip string, delay time.Duration, username string) {
	}

	storage, err := NewIPStorage(tempDir, unblockFunc)
	if err != nil {
		t.Fatalf("Failed to create IP storage: %v", err)
	}

	err = storage.AddBlockedIP("192.168.1.100", "testuser", 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to add blocked IP: %v", err)
	}

	if !storage.IsBlocked("192.168.1.100") {
		t.Error("IP should be blocked")
	}

	err = storage.RemoveBlockedIP("192.168.1.100")
	if err != nil {
		t.Fatalf("Failed to remove blocked IP: %v", err)
	}

	if storage.IsBlocked("192.168.1.100") {
		t.Error("IP should not be blocked after removal")
	}

	blockedIPs := storage.GetBlockedIPs()
	if len(blockedIPs) != 0 {
		t.Errorf("Expected 0 blocked IPs, got %d", len(blockedIPs))
	}
}

func TestIsBlockedExpired(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	unblockFunc := func(ip string, delay time.Duration, username string) {
	}

	storage, err := NewIPStorage(tempDir, unblockFunc)
	if err != nil {
		t.Fatalf("Failed to create IP storage: %v", err)
	}

	err = storage.AddBlockedIP("192.168.1.100", "testuser", 1*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to add blocked IP: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if storage.IsBlocked("192.168.1.100") {
		t.Error("IP should not be blocked after expiration")
	}
}

func TestLoadExistingData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	existingData := `{
		"192.168.1.100": {
			"ip": "192.168.1.100",
			"username": "testuser",
			"blocked_until": "2024-12-31T23:59:59Z"
		}
	}`

	storageFile := filepath.Join(tempDir, "blocked_ips.json")
	err = os.WriteFile(storageFile, []byte(existingData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	unblockFunc := func(ip string, delay time.Duration, username string) {
	}

	storage, err := NewIPStorage(tempDir, unblockFunc)
	if err != nil {
		t.Fatalf("Failed to create IP storage: %v", err)
	}

	blockedIPs := storage.GetBlockedIPs()
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
}

func TestConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	unblockFunc := func(ip string, delay time.Duration, username string) {
	}

	storage, err := NewIPStorage(tempDir, unblockFunc)
	if err != nil {
		t.Fatalf("Failed to create IP storage: %v", err)
	}

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			ip := fmt.Sprintf("192.168.1.%d", id)
			err := storage.AddBlockedIP(ip, fmt.Sprintf("user%d", id), 1*time.Minute)
			if err != nil {
				t.Errorf("Failed to add IP %s: %v", ip, err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	blockedIPs := storage.GetBlockedIPs()
	if len(blockedIPs) != 10 {
		t.Errorf("Expected 10 blocked IPs, got %d", len(blockedIPs))
	}
}

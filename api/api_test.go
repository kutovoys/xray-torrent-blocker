package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tblocker/config"
	"tblocker/firewall"
	"tblocker/storage"
	"tblocker/utils"
)

const testToken = "test-api-token"

func setupTestServer(t *testing.T) (*httptest.Server, *storage.IPStorage) {
	t.Helper()

	config.APIToken = testToken
	config.Hostname = "test-host"
	config.BlockDuration = 10
	config.TorrentTag = "TORRENT"

	if fm, err := firewall.NewManager("iptables"); err == nil {
		utils.SetFirewallManager(fm)
	}

	store, err := storage.NewIPStorage(t.TempDir(), utils.UnblockIPAfterDelay)
	if err != nil {
		t.Fatalf("Failed to create IP storage: %v", err)
	}
	utils.SetIPStorage(store)
	SetIPStorage(store)

	server := httptest.NewServer(NewMux())
	t.Cleanup(server.Close)

	return server, store
}

func doRequest(t *testing.T, method, url, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	return resp
}

func TestDashboardServedWithoutAuth(t *testing.T) {
	server, _ := setupTestServer(t)

	resp := doRequest(t, "GET", server.URL+"/", "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for dashboard, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Expected HTML content type, got '%s'", ct)
	}
}

func TestAPIRequiresAuth(t *testing.T) {
	server, _ := setupTestServer(t)

	paths := []struct {
		method string
		path   string
	}{
		{"GET", "/api/blocked"},
		{"GET", "/api/stats"},
		{"DELETE", "/api/blocked/1.2.3.4"},
	}

	for _, p := range paths {
		resp := doRequest(t, p.method, server.URL+p.path, "")
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s without token: expected 401, got %d", p.method, p.path, resp.StatusCode)
		}

		resp = doRequest(t, p.method, server.URL+p.path, "wrong-token")
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s with wrong token: expected 401, got %d", p.method, p.path, resp.StatusCode)
		}
	}
}

func TestListBlocked(t *testing.T) {
	server, store := setupTestServer(t)

	if err := store.AddBlockedIP("10.0.0.2", "user2", 10*time.Minute); err != nil {
		t.Fatalf("Failed to seed blocked IP: %v", err)
	}
	if err := store.AddBlockedIP("10.0.0.1", "user1", 10*time.Minute); err != nil {
		t.Fatalf("Failed to seed blocked IP: %v", err)
	}

	resp := doRequest(t, "GET", server.URL+"/api/blocked", testToken)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var list []storage.BlockedIP
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("Expected 2 blocked IPs, got %d", len(list))
	}
	if list[0].IP != "10.0.0.1" || list[1].IP != "10.0.0.2" {
		t.Errorf("Expected list sorted by IP, got %s, %s", list[0].IP, list[1].IP)
	}
	if list[0].Username != "user1" {
		t.Errorf("Expected username 'user1', got '%s'", list[0].Username)
	}
}

func TestUnblock(t *testing.T) {
	server, store := setupTestServer(t)

	if err := store.AddBlockedIP("10.0.0.5", "user5", 10*time.Minute); err != nil {
		t.Fatalf("Failed to seed blocked IP: %v", err)
	}

	resp := doRequest(t, "DELETE", server.URL+"/api/blocked/10.0.0.5", testToken)
	resp.Body.Close()

	// Without a usable firewall (e.g. on non-Linux dev machines) ForceUnblock
	// fails with 500 before touching storage; with one, the entry is removed.
	switch resp.StatusCode {
	case http.StatusNoContent:
		if store.IsBlocked("10.0.0.5") {
			t.Error("Expected IP to be removed from storage after unblock")
		}
	case http.StatusInternalServerError:
		t.Log("Firewall unavailable on this platform; got expected 500")
	default:
		t.Errorf("Expected 204 or 500, got %d", resp.StatusCode)
	}
}

func TestUnblockNotFound(t *testing.T) {
	server, _ := setupTestServer(t)

	resp := doRequest(t, "DELETE", server.URL+"/api/blocked/9.9.9.9", testToken)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for unknown IP, got %d", resp.StatusCode)
	}
}

func TestUnblockInvalidIP(t *testing.T) {
	server, _ := setupTestServer(t)

	resp := doRequest(t, "DELETE", server.URL+"/api/blocked/not-an-ip", testToken)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid IP, got %d", resp.StatusCode)
	}
}

func TestStats(t *testing.T) {
	server, store := setupTestServer(t)

	if err := store.AddBlockedIP("10.0.0.9", "statsapiuser", 10*time.Minute); err != nil {
		t.Fatalf("Failed to seed blocked IP: %v", err)
	}
	utils.RecordBlock("statsapiuser")

	resp := doRequest(t, "GET", server.URL+"/api/stats", testToken)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var stats struct {
		Server           string           `json:"server"`
		CurrentlyBlocked int              `json:"currently_blocked"`
		TotalBlocks      int64            `json:"total_blocks"`
		PerUser          map[string]int64 `json:"per_user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.Server != "test-host" {
		t.Errorf("Expected server 'test-host', got '%s'", stats.Server)
	}
	if stats.CurrentlyBlocked != 1 {
		t.Errorf("Expected 1 currently blocked, got %d", stats.CurrentlyBlocked)
	}
	if stats.TotalBlocks < 1 {
		t.Errorf("Expected at least 1 total block, got %d", stats.TotalBlocks)
	}
	if stats.PerUser["statsapiuser"] < 1 {
		t.Errorf("Expected statsapiuser to have at least 1 block, got %d", stats.PerUser["statsapiuser"])
	}
}

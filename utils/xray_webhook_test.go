package utils

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"tblocker/config"
	"testing"
)

func TestHandleXrayWebhook(t *testing.T) {
	config.TorrentTag = "TORRENT"
	config.BypassIPSet = map[string]struct{}{"203.0.113.7": {}, "198.51.100.1": {}}

	// bypassed IPs return before storage/firewall, so only parsing is covered
	cases := []struct {
		name, body string
		status     int
	}{
		{"torrent", `{"email":"a@b","source":"203.0.113.7:1234","outboundTag":"TORRENT"}`, http.StatusNoContent},
		{"other outbound", `{"email":"a@b","source":"198.51.100.1:1234","outboundTag":"direct"}`, http.StatusNoContent},
		{"ipv6 source", `{"email":"a@b","source":"[2001:db8::1]:1234","outboundTag":"TORRENT"}`, http.StatusNoContent},
		{"no email", `{"source":"198.51.100.1:1234","outboundTag":"TORRENT"}`, http.StatusNoContent},
		{"bad json", `{`, http.StatusBadRequest},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		handleXrayWebhook(rec, httptest.NewRequest("POST", "/webhook/xray", strings.NewReader(c.body)))
		if rec.Code != c.status {
			t.Errorf("%s: status %d, want %d", c.name, rec.Code, c.status)
		}
	}
}

func TestListenXrayWebhook(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "w.sock")
	for _, addr := range []string{"127.0.0.1:0", "unix:" + sock, sock, "@tblocker-test", "@@tblocker-test-padded"} {
		l, err := listenXrayWebhook(addr)
		if err != nil {
			t.Errorf("%q: %v", addr, err)
			continue
		}
		l.Close()
	}
}

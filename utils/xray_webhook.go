package utils

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"syscall"
	"tblocker/config"
	"time"
)

// xrayEvent is the subset of Xray-core's routing webhook payload we need.
type xrayEvent struct {
	Email       *string `json:"email"`
	Source      *string `json:"source"`
	OutboundTag *string `json:"outboundTag"`
}

// StartXrayWebhookServer receives Xray routing webhooks instead of tailing logs.
func StartXrayWebhookServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/xray", handleXrayWebhook)

	listener, err := listenXrayWebhook(addr)
	if err != nil {
		log.Fatalf("Error listening for Xray webhooks on %s: %v", addr, err)
	}

	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Xray webhook receiver listening on %s", addr)
	log.Fatalf("Xray webhook receiver stopped: %v", srv.Serve(listener))
}

// listenXrayWebhook accepts host:port, unix:/path, /path, @name or @@name.
func listenXrayWebhook(addr string) (net.Listener, error) {
	addr = strings.TrimPrefix(addr, "unix:")
	if addr == "" || (addr[0] != '/' && addr[0] != '@') {
		return net.Listen("tcp", addr)
	}

	if addr[0] == '/' {
		// remove a stale socket, never a regular file
		if fi, err := os.Lstat(addr); err == nil && fi.Mode()&os.ModeSocket != 0 {
			_ = os.Remove(addr)
		}
	} else if strings.HasPrefix(addr, "@@") {
		// pad like Xray's ResolveSocketPath
		full := make([]byte, len(syscall.RawSockaddrUnix{}.Path))
		copy(full, addr[1:])
		addr = string(full)
	}
	return net.Listen("unix", addr)
}

func handleXrayWebhook(w http.ResponseWriter, r *http.Request) {
	var ev xrayEvent
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		http.Error(w, "invalid webhook payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)

	if ev.OutboundTag == nil || *ev.OutboundTag != config.TorrentTag || ev.Source == nil {
		return
	}

	ap, err := netip.ParseAddrPort(*ev.Source)
	if err != nil || !ap.Addr().Unmap().Is4() {
		log.Printf("Invalid webhook source: %s", *ev.Source)
		return
	}

	username := config.DefaultNoEmailUsername
	if ev.Email != nil && *ev.Email != "" {
		username = *ev.Email
	} else if !config.IgnoreEmail {
		log.Println("Invalid webhook: email missing")
		return
	}

	blockUser(ap.Addr().Unmap().String(), username)
}

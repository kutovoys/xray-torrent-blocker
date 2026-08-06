package api

import (
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"tblocker/config"
	"tblocker/storage"
	"tblocker/utils"
)

//go:embed dashboard.html
var dashboardHTML []byte

var ipStorage *storage.IPStorage

func SetIPStorage(s *storage.IPStorage) {
	ipStorage = s
}

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handleDashboard)
	mux.HandleFunc("GET /api/blocked", requireAuth(handleListBlocked))
	mux.HandleFunc("DELETE /api/blocked/{ip}", requireAuth(handleUnblock))
	mux.HandleFunc("GET /api/stats", requireAuth(handleStats))

	return mux
}

func StartServer() {
	srv := &http.Server{
		Addr:              config.APIAddress,
		Handler:           NewMux(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Printf("API server listening on %s", config.APIAddress)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("API server error: %v", err)
	}
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || config.APIToken == "" ||
			subtle.ConstantTimeCompare([]byte(token), []byte(config.APIToken)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dashboardHTML)
}

func handleListBlocked(w http.ResponseWriter, r *http.Request) {
	blocked := ipStorage.GetBlockedIPs()

	list := make([]storage.BlockedIP, 0, len(blocked))
	for _, info := range blocked {
		list = append(list, info)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].IP < list[j].IP
	})

	writeJSON(w, http.StatusOK, list)
}

func handleUnblock(w http.ResponseWriter, r *http.Request) {
	ip := r.PathValue("ip")

	if !utils.IsValidIPFormat(ip) {
		writeError(w, http.StatusBadRequest, "invalid IP address")
		return
	}

	if err := utils.ForceUnblock(ip); err != nil {
		if strings.Contains(err.Error(), "is not blocked") {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	stats := utils.GetStats()

	writeJSON(w, http.StatusOK, struct {
		Server           string           `json:"server"`
		CurrentlyBlocked int              `json:"currently_blocked"`
		TotalBlocks      int64            `json:"total_blocks"`
		PerUser          map[string]int64 `json:"per_user"`
		StartedAt        time.Time        `json:"started_at"`
	}{
		Server:           config.Hostname,
		CurrentlyBlocked: len(ipStorage.GetBlockedIPs()),
		TotalBlocks:      stats.TotalBlocks,
		PerUser:          stats.PerUser,
		StartedAt:        stats.StartedAt,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding API response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

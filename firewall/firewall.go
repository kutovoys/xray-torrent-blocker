package firewall

import (
	"log"
	"os/exec"
	"strings"
)

type Firewall interface {
	Initialize() error
	BlockIP(ip string) error
	UnblockIP(ip string) error

	GetBlockedIPs() (map[string]bool, error)

	IsAvailable() bool

	GetName() string
}

type Manager struct {
	firewall Firewall
}

func NewManager(blockMode string) (*Manager, error) {
	var firewall Firewall

	switch strings.ToLower(blockMode) {
	case "iptables":
		firewall = NewIPTablesFirewall()
	case "nft":
		firewall = NewNFTFirewall()
	default:
		log.Printf("Unknown firewall mode: %s, falling back to iptables", blockMode)
		firewall = NewIPTablesFirewall()
	}

	if !firewall.IsAvailable() {
		log.Printf("Firewall %s is not available, trying alternatives", firewall.GetName())

		alternatives := []Firewall{
			NewIPTablesFirewall(),
			NewNFTFirewall(),
		}

		for _, alt := range alternatives {
			if alt.IsAvailable() {
				firewall = alt
				log.Printf("Using %s as fallback", alt.GetName())
				break
			}
		}
	}

	if err := firewall.Initialize(); err != nil {
		log.Printf("Error initializing firewall: %v", err)
		return nil, err
	}

	return &Manager{firewall: firewall}, nil
}

func (m *Manager) BlockIP(ip string) error {
	return m.firewall.BlockIP(ip)
}

func (m *Manager) UnblockIP(ip string) error {
	return m.firewall.UnblockIP(ip)
}

func (m *Manager) GetBlockedIPs() (map[string]bool, error) {
	return m.firewall.GetBlockedIPs()
}

func (m *Manager) GetFirewallName() string {
	return m.firewall.GetName()
}

func execCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	return string(output), err
}

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

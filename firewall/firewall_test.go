package firewall

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	testCases := []string{"iptables", "nft", "unknown"}

	for _, blockMode := range testCases {
		manager, err := NewManager(blockMode)
		if err != nil {
			t.Errorf("Failed to create manager for %s: %v", blockMode, err)
			continue
		}

		if manager == nil {
			t.Errorf("Manager is nil for %s", blockMode)
			continue
		}

		firewallName := manager.GetFirewallName()
		if firewallName == "" {
			t.Errorf("Firewall name is empty for %s", blockMode)
		}

		t.Logf("Created manager for %s, using firewall: %s", blockMode, firewallName)
	}
}

func TestFirewallAvailability(t *testing.T) {
	firewalls := []Firewall{
		NewIPTablesFirewall(),
		NewNFTFirewall(),
	}

	for _, fw := range firewalls {
		name := fw.GetName()
		available := fw.IsAvailable()
		t.Logf("Firewall %s: available = %v", name, available)
	}
}

func TestFirewallNames(t *testing.T) {
	expectedNames := map[string]string{
		"iptables": "iptables",
		"nft":      "nftables",
	}

	firewalls := map[string]Firewall{
		"iptables": NewIPTablesFirewall(),
		"nft":      NewNFTFirewall(),
	}

	for key, fw := range firewalls {
		expected := expectedNames[key]
		actual := fw.GetName()
		if actual != expected {
			t.Errorf("Expected name %s for %s, got %s", expected, key, actual)
		}
	}
}

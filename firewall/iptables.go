package firewall

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/coreos/go-iptables/iptables"
)

type IPTablesFirewall struct {
	ipRegex     *regexp.Regexp
	ipt         *iptables.IPTables
	chainName   string
	initialized bool
}

func NewIPTablesFirewall() *IPTablesFirewall {
	ipt, err := iptables.New()
	if err != nil {
		log.Printf("Error creating iptables instance: %v", err)
		return &IPTablesFirewall{
			ipRegex:     regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`),
			ipt:         nil,
			chainName:   "TBLOCKER_BLOCKED",
			initialized: false,
		}
	}

	return &IPTablesFirewall{
		ipRegex:     regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`),
		ipt:         ipt,
		chainName:   "TBLOCKER_BLOCKED",
		initialized: false,
	}
}

func (f *IPTablesFirewall) Initialize() error {
	if f.initialized {
		return nil
	}

	if f.ipt == nil {
		return fmt.Errorf("iptables not available")
	}

	_, err := f.ipt.List("raw", "PREROUTING")
	if err != nil {
		log.Printf("IPTables is not available on this system: %v", err)
		return fmt.Errorf("iptables not available: %v", err)
	}

	log.Printf("Initializing iptables firewall...")

	exists, err := f.ipt.ChainExists("raw", f.chainName)
	if err != nil {
		log.Printf("Error checking chain existence: %v", err)
		return err
	}

	if !exists {
		err = f.ipt.NewChain("raw", f.chainName)
		if err != nil {
			log.Printf("Error creating chain %s: %v", f.chainName, err)
			return err
		}
		log.Printf("Created chain %s in raw table", f.chainName)
	}

	rules, err := f.ipt.List("raw", "PREROUTING")
	if err != nil {
		log.Printf("Error listing PREROUTING rules: %v", err)
		return err
	}

	jumpRuleExists := false
	for _, rule := range rules {
		if strings.Contains(rule, f.chainName) {
			jumpRuleExists = true
			break
		}
	}

	if !jumpRuleExists {
		err = f.ipt.Insert("raw", "PREROUTING", 1, "-j", f.chainName)
		if err != nil {
			log.Printf("Error adding jump rule to %s: %v", f.chainName, err)
			return err
		}
		log.Printf("Added jump rule to %s in PREROUTING chain", f.chainName)
	}

	f.initialized = true
	log.Printf("IPTables firewall initialized successfully with custom chain %s", f.chainName)
	return nil
}

func (f *IPTablesFirewall) BlockIP(ip string) error {
	if f.ipt == nil {
		return fmt.Errorf("iptables not available")
	}

	if !f.initialized {
		if err := f.Initialize(); err != nil {
			return err
		}
	}

	rules, err := f.ipt.List("raw", f.chainName)
	if err != nil {
		log.Printf("Error getting rules from chain %s: %v", f.chainName, err)
		return err
	}

	for _, rule := range rules {
		if strings.Contains(rule, ip) && strings.Contains(rule, "DROP") {
			log.Printf("IP %s is already blocked in chain %s", ip, f.chainName)
			return nil
		}
	}

	err = f.ipt.Append("raw", f.chainName, "-s", ip, "-j", "DROP")
	if err != nil {
		log.Printf("Error blocking IP %s in chain %s: %v", ip, f.chainName, err)
		return err
	}

	log.Printf("IP %s blocked in chain %s", ip, f.chainName)
	return nil
}

func (f *IPTablesFirewall) UnblockIP(ip string) error {
	if f.ipt == nil {
		return fmt.Errorf("iptables not available")
	}

	if !f.initialized {
		if err := f.Initialize(); err != nil {
			return err
		}
	}

	err := f.ipt.Delete("raw", f.chainName, "-s", ip, "-j", "DROP")
	if err != nil {
		log.Printf("Error unblocking IP %s from chain %s: %v", ip, f.chainName, err)
		return err
	}

	log.Printf("IP %s unblocked from chain %s", ip, f.chainName)
	return nil
}

func (f *IPTablesFirewall) GetBlockedIPs() (map[string]bool, error) {
	if f.ipt == nil {
		return nil, fmt.Errorf("iptables not available")
	}

	if !f.initialized {
		if err := f.Initialize(); err != nil {
			return nil, err
		}
	}

	rules, err := f.ipt.List("raw", f.chainName)
	if err != nil {
		log.Printf("Error getting rules from chain %s: %v", f.chainName, err)
		return nil, err
	}

	blockedIPs := make(map[string]bool)

	for _, rule := range rules {
		if strings.Contains(rule, "DROP") {
			ip := f.ipRegex.FindString(rule)
			if ip != "" && ip != "0.0.0.0" {
				blockedIPs[ip] = true
			}
		}
	}

	return blockedIPs, nil
}

func (f *IPTablesFirewall) IsAvailable() bool {
	if f.ipt == nil {
		return false
	}

	_, err := f.ipt.List("raw", "PREROUTING")
	return err == nil
}

func (f *IPTablesFirewall) GetName() string {
	return "iptables"
}

func (f *IPTablesFirewall) FlushChain() error {
	if f.ipt == nil {
		return fmt.Errorf("iptables not available")
	}

	if !f.initialized {
		return nil
	}

	err := f.ipt.ClearChain("raw", f.chainName)
	if err != nil {
		log.Printf("Error flushing chain %s: %v", f.chainName, err)
		return err
	}

	log.Printf("Chain %s flushed successfully", f.chainName)
	return nil
}

func (f *IPTablesFirewall) RemoveChain() error {
	if f.ipt == nil {
		return fmt.Errorf("iptables not available")
	}

	if !f.initialized {
		return nil
	}

	err := f.ipt.Delete("raw", "PREROUTING", "-j", f.chainName)
	if err != nil {
		log.Printf("Warning: Could not remove jump rule to %s: %v", f.chainName, err)
	}

	err = f.ipt.ClearChain("raw", f.chainName)
	if err != nil {
		log.Printf("Error clearing chain %s: %v", f.chainName, err)
		return err
	}

	err = f.ipt.DeleteChain("raw", f.chainName)
	if err != nil {
		log.Printf("Error deleting chain %s: %v", f.chainName, err)
		return err
	}

	f.initialized = false
	log.Printf("Chain %s removed successfully", f.chainName)
	return nil
}

func (f *IPTablesFirewall) GetChainName() string {
	return f.chainName
}

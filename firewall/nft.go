package firewall

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

type NFTFirewall struct {
	ipRegex *regexp.Regexp
}

func NewNFTFirewall() *NFTFirewall {
	return &NFTFirewall{
		ipRegex: regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+|[0-9a-fA-F:]+:[0-9a-fA-F:]*[0-9a-fA-F]+)`),
	}
}

func (f *NFTFirewall) BlockIP(ip string) error {
	_, err := execCommand("nft", "add", "table", "inet", "filter")
	if err != nil {
		log.Printf("Table filter might already exist: %v", err)
	}

	_, err = execCommand("nft", "add", "chain", "inet", "filter", "input", "{", "type", "filter", "hook", "input", "priority", "0", ";", "}")
	if err != nil {
		log.Printf("Chain input might already exist: %v", err)
	}

	_, err = execCommand("nft", "add", "rule", "inet", "filter", "input", "ip", "saddr", ip, "drop")
	if err != nil {
		log.Printf("Error blocking IP %s with nftables: %v", ip, err)
		return err
	}
	log.Printf("IP %s blocked with nftables", ip)
	return nil
}

func (f *NFTFirewall) UnblockIP(ip string) error {
	output, err := execCommand("nft", "-a", "list", "ruleset")
	if err != nil {
		log.Printf("Error listing nftables rules: %v", err)
		return err
	}

	var handle string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "ip saddr "+ip) && strings.Contains(line, "drop") {
			handleMatch := regexp.MustCompile(`# handle (\d+)`).FindStringSubmatch(line)
			if len(handleMatch) > 1 {
				handle = handleMatch[1]
				break
			}
		}
	}

	if handle == "" {
		log.Printf("No rule found for IP %s in nftables", ip)
		return fmt.Errorf("no rule found for IP %s", ip)
	}

	_, err = execCommand("nft", "delete", "rule", "inet", "filter", "input", "handle", handle)
	if err != nil {
		log.Printf("Error unblocking IP %s with nftables: %v", ip, err)
		return err
	}
	log.Printf("IP %s unblocked with nftables (handle: %s)", ip, handle)
	return nil
}

func (f *NFTFirewall) GetBlockedIPs() (map[string]bool, error) {
	output, err := execCommand("nft", "list", "ruleset")
	if err != nil {
		log.Printf("Error checking nftables status: %v", err)
		return nil, err
	}

	blockedIPs := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		if (strings.Contains(line, "ip saddr") || strings.Contains(line, "ip6 saddr")) &&
			strings.Contains(line, "drop") {
			ip := f.ipRegex.FindString(line)
			if ip != "" {
				blockedIPs[ip] = true
			}
		}
	}

	return blockedIPs, nil
}

func (f *NFTFirewall) IsAvailable() bool {
	return isCommandAvailable("nft")
}

func (f *NFTFirewall) GetName() string {
	return "nftables"
}

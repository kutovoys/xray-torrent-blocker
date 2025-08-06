package firewall

import (
	"log"
	"regexp"
	"strings"
)

type IPTablesFirewall struct {
	ipRegex *regexp.Regexp
}

func NewIPTablesFirewall() *IPTablesFirewall {
	return &IPTablesFirewall{
		ipRegex: regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`),
	}
}

func (f *IPTablesFirewall) BlockIP(ip string) error {
	_, err := execCommand("iptables", "-I", "INPUT", "-s", ip, "-j", "DROP")
	if err != nil {
		log.Printf("Error blocking IP %s with iptables: %v", ip, err)
		return err
	}
	log.Printf("IP %s blocked with iptables", ip)
	return nil
}

func (f *IPTablesFirewall) UnblockIP(ip string) error {
	output, err := execCommand("iptables", "-L", "INPUT", "-n", "--line-numbers")
	if err != nil {
		log.Printf("Error getting iptables rules: %v", err)
		return err
	}

	lines := strings.Split(output, "\n")
	ruleFound := false

	for _, line := range lines {
		if strings.Contains(line, "DROP") && strings.Contains(line, ip) {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				lineNum := fields[0]
				_, err := execCommand("iptables", "-D", "INPUT", lineNum)
				if err != nil {
					log.Printf("Error unblocking IP %s with iptables (line %s): %v", ip, lineNum, err)
					return err
				}
				log.Printf("IP %s unblocked with iptables (line: %s)", ip, lineNum)
				ruleFound = true
				break
			}
		}
	}

	if !ruleFound {
		_, err = execCommand("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP")
		if err != nil {
			log.Printf("Error unblocking IP %s with iptables (exact match): %v", ip, err)
			return err
		}
		log.Printf("IP %s unblocked with iptables (exact match)", ip)
	}

	return nil
}

func (f *IPTablesFirewall) GetBlockedIPs() (map[string]bool, error) {
	output, err := execCommand("iptables", "-L", "INPUT", "-n", "--line-numbers")
	if err != nil {
		log.Printf("Error checking iptables status: %v", err)
		return nil, err
	}

	blockedIPs := make(map[string]bool)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "DROP") {
			ip := f.ipRegex.FindString(line)
			if ip != "" && ip != "0.0.0.0" {
				blockedIPs[ip] = true
			}
		}
	}

	return blockedIPs, nil
}

func (f *IPTablesFirewall) IsAvailable() bool {
	return isCommandAvailable("iptables")
}

func (f *IPTablesFirewall) GetName() string {
	return "iptables"
}

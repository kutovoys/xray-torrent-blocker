package firewall

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
)

type NFTFirewall struct {
	ipRegex     *regexp.Regexp
	conn        *nftables.Conn
	initialized bool
}

func NewNFTFirewall() *NFTFirewall {
	return &NFTFirewall{
		ipRegex:     regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+|[0-9a-fA-F:]+:[0-9a-fA-F:]*[0-9a-fA-F]+)`),
		conn:        &nftables.Conn{},
		initialized: false,
	}
}

func (f *NFTFirewall) Initialize() error {
	if f.initialized {
		return nil
	}

	log.Printf("Initializing nftables firewall...")

	table := &nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   "tblocker",
	}
	f.conn.AddTable(table)

	policy := nftables.ChainPolicyAccept
	chain := &nftables.Chain{
		Name:     "TBLOCKER_BLOCKED",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &policy,
	}
	f.conn.AddChain(chain)

	set := &nftables.Set{
		Table:   table,
		Name:    "TBLOCKER_BLOCKED_IPS",
		KeyType: nftables.TypeIPAddr,
	}
	f.conn.AddSet(set, []nftables.SetElement{})

	rule := &nftables.Rule{
		Table: table,
		Chain: chain,
		Exprs: []expr.Any{
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       12, // IP source address offset
				Len:          4,
			},
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        set.Name,
				SetID:          set.ID,
			},
			&expr.Verdict{
				Kind: expr.VerdictDrop,
			},
		},
	}
	f.conn.AddRule(rule)

	err := f.conn.Flush()
	if err != nil {
		log.Printf("Error initializing nftables: %v", err)
		return f.initializeViaCommand()
	}

	log.Printf("Nftables firewall initialized successfully")
	f.initialized = true
	return nil
}

func (f *NFTFirewall) initializeViaCommand() error {
	log.Printf("Initializing nftables via commands...")

	_, err := execCommand("nft", "add", "table", "inet", "tblocker")
	if err != nil {
		log.Printf("Warning: table might already exist: %v", err)
	}

	_, err = execCommand("nft", "add", "chain", "inet", "tblocker", "TBLOCKER_BLOCKED", "{", "type", "filter", "hook", "prerouting", "priority", "-100", ";", "policy", "accept", ";", "}")
	if err != nil {
		log.Printf("Warning: chain might already exist: %v", err)
	}

	_, err = execCommand("nft", "add", "set", "inet", "tblocker", "TBLOCKER_BLOCKED_IPS", "{", "type", "ipv4_addr", ";", "}")
	if err != nil {
		log.Printf("Warning: set might already exist: %v", err)
	}

	_, err = execCommand("nft", "add", "rule", "inet", "tblocker", "TBLOCKER_BLOCKED", "ip", "saddr", "@TBLOCKER_BLOCKED_IPS", "drop")
	if err != nil {
		log.Printf("Warning: rule might already exist: %v", err)
	}

	log.Printf("Nftables firewall initialized via commands")
	f.initialized = true
	return nil
}

func (f *NFTFirewall) BlockIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	if !f.initialized {
		if err := f.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize firewall: %v", err)
		}
	}

	table := &nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   "tblocker",
	}
	set := &nftables.Set{
		Table: table,
		Name:  "TBLOCKER_BLOCKED_IPS",
	}

	element := nftables.SetElement{
		Key: parsedIP.To4(),
	}
	f.conn.SetAddElements(set, []nftables.SetElement{element})

	if err := f.conn.Flush(); err != nil {
		log.Printf("Error adding IP %s to nftables set: %v", ip, err)
		return f.blockIPViaCommand(ip)
	}

	log.Printf("IP %s blocked with nftables", ip)
	return nil
}

func (f *NFTFirewall) blockIPViaCommand(ip string) error {
	_, err := execCommand("nft", "add", "element", "inet", "tblocker", "TBLOCKER_BLOCKED_IPS", "{", ip, "}")
	if err != nil {
		log.Printf("Error adding IP %s to nftables set via command: %v", ip, err)
		return err
	}
	log.Printf("IP %s blocked with nftables via command", ip)
	return nil
}

func (f *NFTFirewall) UnblockIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	table := &nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   "tblocker",
	}
	set := &nftables.Set{
		Table: table,
		Name:  "TBLOCKER_BLOCKED_IPS",
	}

	element := nftables.SetElement{
		Key: parsedIP.To4(),
	}
	f.conn.SetDeleteElements(set, []nftables.SetElement{element})

	if err := f.conn.Flush(); err != nil {
		log.Printf("Error unblocking IP %s with nftables: %v", ip, err)
		return f.unblockIPViaCommand(ip)
	}

	log.Printf("IP %s unblocked with nftables", ip)
	return nil
}

func (f *NFTFirewall) unblockIPViaCommand(ip string) error {
	_, err := execCommand("nft", "delete", "element", "inet", "tblocker", "TBLOCKER_BLOCKED_IPS", "{", ip, "}")
	if err != nil {
		log.Printf("Error unblocking IP %s with nftables via command: %v", ip, err)
		return err
	}
	log.Printf("IP %s unblocked with nftables via command", ip)
	return nil
}

func (f *NFTFirewall) GetBlockedIPs() (map[string]bool, error) {
	table := &nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   "tblocker",
	}

	sets, err := f.conn.GetSets(table)
	if err != nil {
		log.Printf("Error getting sets via API, falling back to nft command: %v", err)
		return f.getBlockedIPsViaCommand()
	}

	blockedIPs := make(map[string]bool)
	for _, s := range sets {
		if s.Name == "TBLOCKER_BLOCKED_IPS" {
			elements, err := f.conn.GetSetElements(s)
			if err != nil {
				log.Printf("Error listing nftables set via API, falling back to nft command: %v", err)
				return f.getBlockedIPsViaCommand()
			}

			for _, element := range elements {
				if len(element.Key) == 4 {
					ip := net.IP(element.Key).String()
					blockedIPs[ip] = true
				}
			}
			break
		}
	}

	return blockedIPs, nil
}

func (f *NFTFirewall) getBlockedIPsViaCommand() (map[string]bool, error) {
	output, err := execCommand("nft", "list", "set", "inet", "tblocker", "TBLOCKER_BLOCKED_IPS")
	if err != nil {
		log.Printf("Error listing nftables set via command: %v", err)
		return nil, err
	}

	blockedIPs := make(map[string]bool)
	for _, line := range strings.Split(output, "\n") {
		ip := f.ipRegex.FindString(line)
		if ip != "" {
			blockedIPs[ip] = true
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

package firewall

import (
	"fmt"
	"log"
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
)

type NFTFirewall struct {
	conn        *nftables.Conn
	initialized bool
}

func NewNFTFirewall() *NFTFirewall {
	return &NFTFirewall{
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

	if !f.ruleExists(table, chain) {
		rule := &nftables.Rule{
			Table: table,
			Chain: chain,
			Exprs: []expr.Any{
				&expr.Payload{
					DestRegister: 1,
					Base:         expr.PayloadBaseNetworkHeader,
					Offset:       12,
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
		log.Printf("Rule added to nftables")
	} else {
		log.Printf("Rule already exists in nftables")
	}

	err := f.conn.Flush()
	if err != nil {
		log.Printf("Error initializing nftables: %v", err)
		return fmt.Errorf("failed to initialize nftables: %v", err)
	}

	log.Printf("Nftables firewall initialized successfully")
	f.initialized = true
	return nil
}

func (f *NFTFirewall) ruleExists(table *nftables.Table, chain *nftables.Chain) bool {
	rules, err := f.conn.GetRules(table, chain)
	if err != nil {
		log.Printf("Error checking existing rules: %v", err)
		return false
	}

	for _, rule := range rules {
		if len(rule.Exprs) >= 3 {
			if payload, ok := rule.Exprs[0].(*expr.Payload); ok {
				if payload.DestRegister == 1 && payload.Base == expr.PayloadBaseNetworkHeader &&
					payload.Offset == 12 && payload.Len == 4 {
					if lookup, ok := rule.Exprs[1].(*expr.Lookup); ok {
						if lookup.SourceRegister == 1 && lookup.SetName == "TBLOCKER_BLOCKED_IPS" {
							if verdict, ok := rule.Exprs[2].(*expr.Verdict); ok {
								if verdict.Kind == expr.VerdictDrop {
									return true
								}
							}
						}
					}
				}
			}
		}
	}
	return false
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
		return fmt.Errorf("failed to add IP %s to nftables set: %v", ip, err)
	}

	log.Printf("IP %s blocked with nftables", ip)
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
		return fmt.Errorf("failed to unblock IP %s with nftables: %v", ip, err)
	}

	log.Printf("IP %s unblocked with nftables", ip)
	return nil
}

func (f *NFTFirewall) GetBlockedIPs() (map[string]bool, error) {
	table := &nftables.Table{
		Family: nftables.TableFamilyINet,
		Name:   "tblocker",
	}

	sets, err := f.conn.GetSets(table)
	if err != nil {
		log.Printf("Error getting sets via API: %v", err)
		return nil, fmt.Errorf("failed to get sets via API: %v", err)
	}

	blockedIPs := make(map[string]bool)
	for _, s := range sets {
		if s.Name == "TBLOCKER_BLOCKED_IPS" {
			elements, err := f.conn.GetSetElements(s)
			if err != nil {
				log.Printf("Error listing nftables set via API: %v", err)
				return nil, fmt.Errorf("failed to list nftables set via API: %v", err)
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

func (f *NFTFirewall) IsAvailable() bool {
	return isCommandAvailable("nft")
}

func (f *NFTFirewall) GetName() string {
	return "nftables"
}

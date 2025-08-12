package utils

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"strings"

	"github.com/ti-mo/conntrack"
)

type ConntrackManager struct {
	available bool
	conn      *conntrack.Conn
}

var conntrackManager *ConntrackManager

func InitConntrackManager() *ConntrackManager {
	manager := &ConntrackManager{}

	if err := manager.ensureKernelModule(); err != nil {
		log.Fatalf("Error working with nf_conntrack kernel module: %v", err)
	}

	conn, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatalf("Failed to connect to conntrack via netlink: %v. Please ensure proper kernel support and permissions.", err)
	}

	manager.available = true
	manager.conn = conn

	conntrackManager = manager
	log.Println("Conntrack manager initialized successfully via netlink")
	return manager
}

func GetConntrackManager() *ConntrackManager {
	return conntrackManager
}

func (cm *ConntrackManager) ensureKernelModule() error {
	if cm.isModuleLoaded() {
		log.Println("Kernel module nf_conntrack is already loaded")
		return nil
	}

	log.Println("Kernel module nf_conntrack not found, attempting to load...")

	if err := cm.loadModule(); err != nil {
		return fmt.Errorf("failed to load nf_conntrack module: %v", err)
	}

	if !cm.isModuleLoaded() {
		return fmt.Errorf("nf_conntrack module failed to load properly")
	}

	if err := cm.setupAutoload(); err != nil {
		log.Printf("Warning: failed to setup module autoload: %v", err)
	}

	log.Println("Kernel module nf_conntrack loaded and configured successfully")
	return nil
}

func (cm *ConntrackManager) isModuleLoaded() bool {
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: failed to check loaded modules: %v", err)
		return false
	}

	return strings.Contains(string(output), "nf_conntrack")
}

func (cm *ConntrackManager) loadModule() error {
	cmd := exec.Command("modprobe", "nf_conntrack")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("modprobe command failed: %v", err)
	}
	return nil
}

func (cm *ConntrackManager) setupAutoload() error {
	cmd := exec.Command("mkdir", "-p", "/etc/modules-load.d")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create modules-load.d directory: %v", err)
	}

	cmd = exec.Command("sh", "-c", "echo 'nf_conntrack' > /etc/modules-load.d/conntrack.conf")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create autoload configuration: %v", err)
	}

	log.Println("Module autoload configured in /etc/modules-load.d/conntrack.conf")
	return nil
}

func (cm *ConntrackManager) DropConnections(ip string) error {
	if !cm.available || cm.conn == nil {
		return fmt.Errorf("conntrack is not available")
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	return cm.dropConnectionsViaLibrary(parsedIP)
}

func (cm *ConntrackManager) dropConnectionsViaLibrary(ip net.IP) error {
	netipAddr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return fmt.Errorf("failed to convert IP address: %s", ip.String())
	}

	flows, err := cm.conn.Dump(nil)
	if err != nil {
		return fmt.Errorf("failed to dump conntrack table: %v", err)
	}

	for _, flow := range flows {
		shouldDelete := false

		if flow.TupleOrig.IP.SourceAddress.IsValid() && flow.TupleOrig.IP.SourceAddress == netipAddr {
			shouldDelete = true
		}

		if flow.TupleOrig.IP.DestinationAddress.IsValid() && flow.TupleOrig.IP.DestinationAddress == netipAddr {
			shouldDelete = true
		}

		if flow.TupleReply.IP.SourceAddress.IsValid() && flow.TupleReply.IP.SourceAddress == netipAddr {
			shouldDelete = true
		}

		if flow.TupleReply.IP.DestinationAddress.IsValid() && flow.TupleReply.IP.DestinationAddress == netipAddr {
			shouldDelete = true
		}

		if shouldDelete {
			err := cm.conn.Delete(flow)
			if err != nil {
				log.Printf("Warning: failed to delete connection for IP %s: %v", ip.String(), err)
			}
		}
	}

	log.Printf("Connections for IP %s have been dropped via conntrack library", ip.String())
	return nil
}

func (cm *ConntrackManager) IsAvailable() bool {
	return cm.available
}

func (cm *ConntrackManager) Close() error {
	if cm.conn != nil {
		return cm.conn.Close()
	}
	return nil
}

func (cm *ConntrackManager) GetConnectionCount(ip string) (int, error) {
	if !cm.available || cm.conn == nil {
		return 0, fmt.Errorf("conntrack is not available")
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ip)
	}

	netipAddr, ok := netip.AddrFromSlice(parsedIP)
	if !ok {
		return 0, fmt.Errorf("failed to convert IP address: %s", ip)
	}

	flows, err := cm.conn.Dump(nil)
	if err != nil {
		return 0, fmt.Errorf("failed to dump conntrack table: %v", err)
	}

	count := 0
	for _, flow := range flows {
		if (flow.TupleOrig.IP.SourceAddress.IsValid() && flow.TupleOrig.IP.SourceAddress == netipAddr) ||
			(flow.TupleOrig.IP.DestinationAddress.IsValid() && flow.TupleOrig.IP.DestinationAddress == netipAddr) ||
			(flow.TupleReply.IP.SourceAddress.IsValid() && flow.TupleReply.IP.SourceAddress == netipAddr) ||
			(flow.TupleReply.IP.DestinationAddress.IsValid() && flow.TupleReply.IP.DestinationAddress == netipAddr) {
			count++
		}
	}

	return count, nil
}

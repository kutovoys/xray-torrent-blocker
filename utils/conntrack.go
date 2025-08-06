package utils

import (
	"log"
	"os/exec"
	"strings"
)

type ConntrackManager struct {
	available bool
}

var conntrackManager *ConntrackManager

func InitConntrackManager() *ConntrackManager {
	manager := &ConntrackManager{}
	manager.available = manager.checkAvailability()

	if !manager.available {
		log.Fatal("conntrack-tools is not installed. Please install conntrack-tools package to continue.")
	}

	if err := manager.ensureKernelModule(); err != nil {
		log.Fatalf("Error working with nf_conntrack kernel module: %v", err)
	}

	conntrackManager = manager
	log.Println("Conntrack manager initialized successfully")
	return manager
}

func GetConntrackManager() *ConntrackManager {
	return conntrackManager
}

func (cm *ConntrackManager) checkAvailability() bool {
	_, err := exec.LookPath("conntrack")
	return err == nil
}

func (cm *ConntrackManager) ensureKernelModule() error {
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	if strings.Contains(string(output), "nf_conntrack") {
		log.Println("Kernel module nf_conntrack is already loaded")
		return nil
	}

	log.Println("Loading kernel module nf_conntrack...")
	cmd = exec.Command("modprobe", "nf_conntrack")
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to load nf_conntrack module: %v", err)
	}

	cmd = exec.Command("sh", "-c", "echo 'nf_conntrack' > /etc/modules-load.d/conntrack.conf")
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to create module autoload configuration file: %v", err)
	}

	log.Println("Kernel module nf_conntrack configured")
	return nil
}

func (cm *ConntrackManager) DropConnections(ip string) error {
	if !cm.available {
		log.Printf("Conntrack is not available, skipping connection drop for IP: %s", ip)
		return nil
	}

	cmd := exec.Command("conntrack", "-D", "-s", ip)
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to drop incoming connections for IP %s: %v", ip, err)
	}

	cmd = exec.Command("conntrack", "-D", "-d", ip)
	if err := cmd.Run(); err != nil {
		log.Printf("Warning: failed to drop outgoing connections for IP %s: %v", ip, err)
	}

	log.Printf("Connections for IP %s dropped via conntrack", ip)
	return nil
}

func (cm *ConntrackManager) IsAvailable() bool {
	return cm.available
}

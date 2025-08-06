package utils

import (
	"os/exec"
	"testing"
)

func TestConntrackManager_CheckAvailability(t *testing.T) {
	manager := &ConntrackManager{}
	available := manager.checkAvailability()

	if available {
		t.Log("conntrack is available in the system")
	} else {
		t.Log("conntrack is not available in the system")
	}
}

func TestConntrackManager_EnsureKernelModule(t *testing.T) {
	manager := &ConntrackManager{}

	err := manager.ensureKernelModule()
	if err != nil {
		t.Logf("Error checking kernel module: %v", err)
	} else {
		t.Log("Kernel module checked successfully")
	}
}

func TestConntrackManager_DropConnections(t *testing.T) {
	manager := &ConntrackManager{}
	manager.available = true

	err := manager.DropConnections("192.168.1.999")
	if err != nil {
		t.Logf("Error dropping connections (expected for non-existent IP): %v", err)
	} else {
		t.Log("Connections dropped successfully")
	}
}

func TestConntrackManager_IsAvailable(t *testing.T) {
	manager := &ConntrackManager{}
	manager.available = true

	if !manager.IsAvailable() {
		t.Error("IsAvailable should return true when available = true")
	}

	manager.available = false
	if manager.IsAvailable() {
		t.Error("IsAvailable should return false when available = false")
	}
}

func TestConntrackCommandExists(t *testing.T) {
	_, err := exec.LookPath("conntrack")
	if err != nil {
		t.Log("conntrack command not found in system")
	} else {
		t.Log("conntrack command found in system")
	}
}

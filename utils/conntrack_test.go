package utils

import (
	"testing"
)

func TestNewConntrackManager(t *testing.T) {
	manager := &ConntrackManager{}

	if manager.available {
		t.Error("Expected available to be false initially")
	}

	if manager.conn != nil {
		t.Error("Expected conn to be nil initially")
	}

	t.Log("ConntrackManager structure created successfully")
}

func TestConntrackManager_DropConnections(t *testing.T) {
	manager := &ConntrackManager{available: false}

	err := manager.DropConnections("192.168.1.1")
	if err == nil {
		t.Error("Expected error when conntrack is not available")
	}

	manager.available = true
	err = manager.DropConnections("invalid-ip")
	if err == nil {
		t.Error("Expected error for invalid IP address")
	}
}

func TestConntrackManager_GetConnectionCount(t *testing.T) {
	manager := &ConntrackManager{available: false}

	_, err := manager.GetConnectionCount("192.168.1.1")
	if err == nil {
		t.Error("Expected error when conntrack is not available")
	}

	manager.available = true
	_, err = manager.GetConnectionCount("invalid-ip")
	if err == nil {
		t.Error("Expected error for invalid IP address")
	}
}

func TestConntrackManager_IsAvailable(t *testing.T) {
	manager := &ConntrackManager{available: true}
	if !manager.IsAvailable() {
		t.Error("Expected IsAvailable to return true")
	}

	manager.available = false
	if manager.IsAvailable() {
		t.Error("Expected IsAvailable to return false")
	}
}

func TestConntrackManager_Close(t *testing.T) {
	manager := &ConntrackManager{}

	err := manager.Close()
	if err != nil {
		t.Errorf("Close should not return error when no connection: %v", err)
	}
}

func TestConntrackManager_LibraryOnly(t *testing.T) {
	manager := &ConntrackManager{}

	err := manager.DropConnections("192.168.1.1")
	if err == nil {
		t.Error("Expected error when no connection established")
	}

	_, err = manager.GetConnectionCount("192.168.1.1")
	if err == nil {
		t.Error("Expected error when no connection established")
	}

	t.Log("Library-only mode validation passed")
}

func TestConntrackManager_ensureKernelModule(t *testing.T) {
	manager := &ConntrackManager{}

	err := manager.ensureKernelModule()
	if err != nil {
		t.Logf("ensureKernelModule failed (expected without root): %v", err)
	} else {
		t.Log("ensureKernelModule succeeded")
	}
}

func TestConntrackManager_isModuleLoaded(t *testing.T) {
	manager := &ConntrackManager{}

	loaded := manager.isModuleLoaded()
	t.Logf("Module nf_conntrack loaded: %v", loaded)
}

func TestConntrackManager_loadModule(t *testing.T) {
	manager := &ConntrackManager{}

	err := manager.loadModule()
	if err != nil {
		t.Logf("loadModule failed (expected without root): %v", err)
	} else {
		t.Log("loadModule succeeded")
	}
}

func TestConntrackManager_setupAutoload(t *testing.T) {
	manager := &ConntrackManager{}

	err := manager.setupAutoload()
	if err != nil {
		t.Logf("setupAutoload failed (expected without root): %v", err)
	} else {
		t.Log("setupAutoload succeeded")
	}
}

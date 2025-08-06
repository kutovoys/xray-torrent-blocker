package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestMainVersionFlag(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	os.Args = []string{"tblocker", "-v"}
	Version = "test-version"

	tempDir, err := os.MkdirTemp("", "main_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 10
TorrentTag: "TORRENT"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

}

func TestMainConfigFlag(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	tempDir, err := os.MkdirTemp("", "main_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 15
TorrentTag: "TEST_TORRENT"
BlockMode: "iptables"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	os.Args = []string{"tblocker", "-c", configFile}
}

func TestMainDefaultConfig(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	tempDir, err := os.MkdirTemp("", "main_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `
LogFile: "/var/log/test.log"
BlockDuration: 10
TorrentTag: "TORRENT"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	os.Args = []string{"tblocker"}
}

func TestMainInvalidConfig(t *testing.T) {
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	os.Args = []string{"tblocker", "-c", "/nonexistent/config.yaml"}
}

package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"
)

const (
	DefaultNoEmailUsername = "__NO_USER_NAME__"
)

var (
	LogFile       string
	BlockDuration int
	TorrentTag    string
	BlockMode     string
	BypassIPSet   = make(map[string]struct{})
	IgnoreEmail   bool
	StorageDir    string

	SendWebhook     bool
	WebhookURL      string
	WebhookTemplate string
	WebhookHeaders  map[string]string

	UsernameRegex        *regexp.Regexp
	DefaultUsernameRegex = `^(.+)$`

	Hostname string

	EnablePerformanceMetrics bool

	EnableAPI  bool
	APIAddress string
	APIToken   string

	SendTelegram     bool
	TelegramBotToken string
	TelegramChatID   string
	TelegramTemplate string
)

type Config struct {
	LogFile         string            `yaml:"LogFile"`
	BlockDuration   int               `yaml:"BlockDuration"`
	TorrentTag      string            `yaml:"TorrentTag"`
	UsernameRegex   string            `yaml:"UsernameRegex"`
	BlockMode       string            `yaml:"BlockMode"`
	BypassIPS       []string          `yaml:"BypassIPS"`
	IgnoreEmail     bool              `yaml:"IgnoreEmail"`
	SendWebhook     bool              `yaml:"SendWebhook"`
	WebhookURL      string            `yaml:"WebhookURL"`
	WebhookTemplate string            `yaml:"WebhookTemplate"`
	StorageDir      string            `yaml:"StorageDir"`
	WebhookHeaders  map[string]string `yaml:"WebhookHeaders"`
	Hostname        string            `yaml:"Hostname"`

	EnableAPI  bool   `yaml:"EnableAPI"`
	APIAddress string `yaml:"APIAddress"`
	APIToken   string `yaml:"APIToken"`

	SendTelegram     bool   `yaml:"SendTelegram"`
	TelegramBotToken string `yaml:"TelegramBotToken"`
	TelegramChatID   string `yaml:"TelegramChatID"`
	TelegramTemplate string `yaml:"TelegramTemplate"`
}

func LoadConfig(configPath string) error {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg Config
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return err
	}

	LogFile = cfg.LogFile
	BlockDuration = cfg.BlockDuration
	TorrentTag = cfg.TorrentTag
	IgnoreEmail = cfg.IgnoreEmail
	SendWebhook = cfg.SendWebhook
	WebhookURL = cfg.WebhookURL
	WebhookHeaders = cfg.WebhookHeaders

	if cfg.UsernameRegex != "" {
		UsernameRegex, err = regexp.Compile(cfg.UsernameRegex)
	} else {
		UsernameRegex, err = regexp.Compile(DefaultUsernameRegex)
	}
	if err != nil {
		return fmt.Errorf("invalid UsernameRegex pattern: %v", err)
	}

	if cfg.Hostname != "" {
		Hostname = cfg.Hostname
	} else {
		Hostname, err = os.Hostname()
	}

	if cfg.BlockMode != "" {
		BlockMode = cfg.BlockMode
	} else {
		BlockMode = "iptables"
	}
	if cfg.BypassIPS != nil {
		fmt.Println("Bypass IPS list:")
		BypassIPSet = make(map[string]struct{})
		for _, ip := range cfg.BypassIPS {
			BypassIPSet[ip] = struct{}{}
			fmt.Printf("- %s\n", ip)
		}
	} else {
		BypassIPSet = make(map[string]struct{})
	}
	if WebhookHeaders == nil {
		WebhookHeaders = make(map[string]string)
	}
	if cfg.WebhookTemplate != "" {
		WebhookTemplate = cfg.WebhookTemplate
	} else {
		WebhookTemplate = `{"username":"%s","ip":"%s","server":"%s","action":"%s","duration":%d,"timestamp":"%s"}`
	}

	StorageDir = cfg.StorageDir
	if StorageDir == "" {
		StorageDir = "/opt/tblocker"
	}

	EnableAPI = cfg.EnableAPI
	APIToken = cfg.APIToken
	if cfg.APIAddress != "" {
		APIAddress = cfg.APIAddress
	} else {
		APIAddress = "127.0.0.1:8085"
	}
	if EnableAPI && APIToken == "" {
		return fmt.Errorf("APIToken is required when EnableAPI is true")
	}

	SendTelegram = cfg.SendTelegram
	TelegramBotToken = cfg.TelegramBotToken
	TelegramChatID = cfg.TelegramChatID
	if cfg.TelegramTemplate != "" {
		TelegramTemplate = cfg.TelegramTemplate
	} else {
		TelegramTemplate = "🚫 Torrent detected!\nUser: %s\nIP: %s\nServer: %s\nAction: %s\nDuration: %d min\nTime: %s"
	}
	if SendTelegram && (TelegramBotToken == "" || TelegramChatID == "") {
		return fmt.Errorf("TelegramBotToken and TelegramChatID are required when SendTelegram is true")
	}

	return err
}

package config

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	DefaultMessageTemplate = `⚠️ <b>Замечено использование торрентов</b> ⚠️

Возможно вы забыли отключить торрент-клиент и сейчас качаете или раздаете торренты через VPN. 
У нас запрещено скачивать или раздавать торренты.

Мы заблокируем ваше подключение на сервере на %d минут, выключите пожалуйста торрент-клиент полностью. 
Через %d минут вас разблокирует, но если вы продолжите качать торренты, вас снова заблокирует и вы получите данное сообщение снова.`

	AdminBlockTemplate = `⛔️ <b>#Blocked</b>
➖➖➖➖➖➖➖➖➖
<b>Username :</b> %s
<b>IP :</b> %s
<b>Server :</b> %s
➖➖➖➖➖➖➖➖➖
<b>User tag :</b> #%s`

	AdminUnblockTemplate = `☑️ <b>#Unblocked</b>
➖➖➖➖➖➖➖➖➖
<b>Username :</b> %s
<b>IP :</b> %s
<b>Server :</b> %s
➖➖➖➖➖➖➖➖➖
<b>User tag :</b> #%s`

	SendWebhook     bool
	WebhookURL      string
	WebhookTemplate string
	WebhookHeaders  map[string]string
	StorageDir      string
)

var (
	LogFile              string
	BotToken             string
	AdminBotToken        string
	AdminChatID          string
	BlockDuration        int
	TorrentTag           string
	Hostname             string
	Message              string
	IpRegex              = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`)
	TidRegex             *regexp.Regexp
	DefaultUsernameRegex = `email: \d+\.(\S+)`
	UsernameRegex        *regexp.Regexp
	BlockedIPs           = make(map[string]bool)
	Mu                   sync.Mutex
	SendUserMessage      bool
	SendAdminMessage     bool
	BlockMode            string
)

type Config struct {
	LogFile             string            `yaml:"LogFile"`
	BotToken            string            `yaml:"BotToken"`
	AdminBotToken       string            `yaml:"AdminBotToken"`
	AdminChatID         string            `yaml:"AdminChatID"`
	BlockDuration       int               `yaml:"BlockDuration"`
	TorrentTag          string            `yaml:"TorrentTag"`
	TidRegex            string            `yaml:"TidRegex"`
	UsernameRegex       string            `yaml:"UsernameRegex"`
	SendUserMessage     bool              `yaml:"SendUserMessage"`
	SendAdminMessage    bool              `yaml:"SendAdminMessage"`
	UserMessageTemplate string            `yaml:"UserMessageTemplate"`
	BlockMode           string            `yaml:"BlockMode"`
	SendWebhook         bool              `yaml:"SendWebhook"`
	WebhookURL          string            `yaml:"WebhookURL"`
	WebhookTemplate     string            `yaml:"WebhookTemplate"`
	StorageDir          string            `yaml:"StorageDir"`
	WebhookHeaders      map[string]string `yaml:"WebhookHeaders"`
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
	BotToken = cfg.BotToken
	AdminBotToken = cfg.AdminBotToken
	AdminChatID = cfg.AdminChatID
	BlockDuration = cfg.BlockDuration
	TorrentTag = cfg.TorrentTag
	SendUserMessage = cfg.SendUserMessage
	SendAdminMessage = cfg.SendAdminMessage
	SendWebhook = cfg.SendWebhook
	WebhookURL = cfg.WebhookURL
	WebhookHeaders = cfg.WebhookHeaders

	if cfg.UserMessageTemplate != "" {
		Message = cfg.UserMessageTemplate
	} else {
		Message = fmt.Sprintf(DefaultMessageTemplate, BlockDuration, BlockDuration)
	}

	if cfg.UsernameRegex != "" {
		UsernameRegex, err = regexp.Compile(cfg.UsernameRegex)
	} else {
		UsernameRegex, err = regexp.Compile(DefaultUsernameRegex)
	}
	if err != nil {
		return err
	}
	if cfg.TidRegex != "" {
		TidRegex, err = regexp.Compile(cfg.TidRegex)
	}
	if err != nil {
		return err
	}

	Hostname, err = os.Hostname()
	if cfg.BlockMode != "" {
		BlockMode = cfg.BlockMode
	} else {
		BlockMode = "ufw"
	}
	if WebhookHeaders == nil {
		WebhookHeaders = make(map[string]string)
	}
	if cfg.WebhookTemplate != "" {
		WebhookTemplate = cfg.WebhookTemplate
	} else {
		WebhookTemplate = `{"username":"%s","ip":"%s","server":"%s","action":"%s","timestamp":"%s"}`
	}

	StorageDir = cfg.StorageDir
	if StorageDir == "" {
		StorageDir = "/opt/tblocker"
	}

	return err
}

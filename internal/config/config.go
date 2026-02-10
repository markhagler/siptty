package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the top-level siptty configuration.
type Config struct {
	General  GeneralConfig   `toml:"general"`
	Accounts []AccountConfig `toml:"accounts"`
	Audio    AudioConfig     `toml:"audio"`
}

// GeneralConfig holds global application settings.
type GeneralConfig struct {
	LogLevel  int    `toml:"log_level"`
	LogFile   string `toml:"log_file"`
	UserAgent string `toml:"user_agent"`
}

// AccountConfig holds a single SIP account's settings.
type AccountConfig struct {
	Name         string            `toml:"name"`
	Enabled      bool              `toml:"enabled"`
	SipURI       string            `toml:"sip_uri"`
	AuthUser     string            `toml:"auth_user"`
	AuthPassword string            `toml:"auth_password"`
	Registrar    string            `toml:"registrar"`
	Transport    string            `toml:"transport"`
	Register     bool              `toml:"register"`
	RegExpiry    int               `toml:"reg_expiry"`
	Headers      map[string]string `toml:"headers"`
}

// AudioConfig holds audio/media settings.
type AudioConfig struct {
	Mode      string `toml:"mode"`
	PlayFile  string `toml:"play_file"`
	RecordDir string `toml:"record_dir"`
}

// rawAccountConfig mirrors AccountConfig but uses *bool for fields that
// default to true, so we can distinguish "not set" from "explicitly false".
type rawAccountConfig struct {
	Name         string            `toml:"name"`
	Enabled      *bool             `toml:"enabled"`
	SipURI       string            `toml:"sip_uri"`
	AuthUser     string            `toml:"auth_user"`
	AuthPassword string            `toml:"auth_password"`
	Registrar    string            `toml:"registrar"`
	Transport    string            `toml:"transport"`
	Register     *bool             `toml:"register"`
	RegExpiry    int               `toml:"reg_expiry"`
	Headers      map[string]string `toml:"headers"`
}

type rawConfig struct {
	General  GeneralConfig    `toml:"general"`
	Accounts []rawAccountConfig `toml:"accounts"`
	Audio    AudioConfig      `toml:"audio"`
}

// Load reads and parses a TOML config file, applies defaults, and validates.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var raw rawConfig
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	cfg := fromRaw(&raw)
	applyDefaults(cfg)

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// FindConfigFile searches for a config file in standard locations.
// Search order: ./siptty.toml, then os.UserConfigDir()/siptty/config.toml.
func FindConfigFile() (string, error) {
	local := "siptty.toml"
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("finding user config dir: %w", err)
	}

	userPath := filepath.Join(configDir, "siptty", "config.toml")
	if _, err := os.Stat(userPath); err == nil {
		return userPath, nil
	}

	return "", fmt.Errorf("no config file found (searched ./siptty.toml and %s)", userPath)
}

func fromRaw(raw *rawConfig) *Config {
	cfg := &Config{
		General: raw.General,
		Audio:   raw.Audio,
	}
	for _, ra := range raw.Accounts {
		a := AccountConfig{
			Name:         ra.Name,
			Enabled:      boolDefault(ra.Enabled, true),
			SipURI:       ra.SipURI,
			AuthUser:     ra.AuthUser,
			AuthPassword: ra.AuthPassword,
			Registrar:    ra.Registrar,
			Transport:    ra.Transport,
			Register:     boolDefault(ra.Register, true),
			RegExpiry:    ra.RegExpiry,
			Headers:      ra.Headers,
		}
		cfg.Accounts = append(cfg.Accounts, a)
	}
	return cfg
}

func boolDefault(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func applyDefaults(cfg *Config) {
	if cfg.General.LogLevel == 0 {
		cfg.General.LogLevel = 3
	}
	if cfg.General.UserAgent == "" {
		cfg.General.UserAgent = "siptty/0.1"
	}
	if cfg.Audio.Mode == "" {
		cfg.Audio.Mode = "null"
	}

	for i := range cfg.Accounts {
		if cfg.Accounts[i].Transport == "" {
			cfg.Accounts[i].Transport = "udp"
		}
		if cfg.Accounts[i].RegExpiry == 0 {
			cfg.Accounts[i].RegExpiry = 300
		}
		if cfg.Accounts[i].AuthUser == "" {
			cfg.Accounts[i].AuthUser = deriveAuthUser(cfg.Accounts[i].SipURI)
		}
	}
}

func deriveAuthUser(sipURI string) string {
	uri := sipURI
	uri = strings.TrimPrefix(uri, "sips:")
	uri = strings.TrimPrefix(uri, "sip:")

	if idx := strings.Index(uri, "@"); idx >= 0 {
		return uri[:idx]
	}
	return ""
}

func validate(cfg *Config) error {
	if len(cfg.Accounts) == 0 {
		return fmt.Errorf("at least one account is required")
	}

	for i, a := range cfg.Accounts {
		if a.SipURI == "" {
			return fmt.Errorf("account %d: sip_uri is required", i)
		}
		if a.Registrar == "" {
			return fmt.Errorf("account %d: registrar is required", i)
		}
		if !isValidTransport(a.Transport) {
			return fmt.Errorf("account %d: invalid transport %q (must be udp, tcp, or tls)", i, a.Transport)
		}
	}

	if !isValidAudioMode(cfg.Audio.Mode) {
		return fmt.Errorf("invalid audio mode %q (must be null or file)", cfg.Audio.Mode)
	}

	return nil
}

func isValidTransport(t string) bool {
	switch t {
	case "udp", "tcp", "tls":
		return true
	}
	return false
}

func isValidAudioMode(m string) bool {
	switch m {
	case "null", "file":
		return true
	}
	return false
}

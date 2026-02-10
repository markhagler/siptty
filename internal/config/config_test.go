package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test config: %v", err)
	}
	return path
}

const fullConfig = `
[general]
log_level = 4
log_file = "/var/log/siptty.log"
user_agent = "test-agent/1.0"

[[accounts]]
name = "work"
enabled = true
sip_uri = "sip:alice@example.com"
auth_user = "alice"
auth_password = "secret"
registrar = "sip:registrar.example.com"
transport = "tcp"
register = true
reg_expiry = 600

[accounts.headers]
X-Custom = "value1"
X-Other = "value2"

[[accounts]]
name = "home"
sip_uri = "sip:bob@home.example.com"
auth_password = "pass"
registrar = "sip:reg.home.example.com"

[audio]
mode = "file"
play_file = "/tmp/test.wav"
record_dir = "/tmp/recordings"
`

func TestLoadValidConfig(t *testing.T) {
	path := writeTestConfig(t, fullConfig)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// General
	if cfg.General.LogLevel != 4 {
		t.Errorf("LogLevel = %d, want 4", cfg.General.LogLevel)
	}
	if cfg.General.LogFile != "/var/log/siptty.log" {
		t.Errorf("LogFile = %q, want /var/log/siptty.log", cfg.General.LogFile)
	}
	if cfg.General.UserAgent != "test-agent/1.0" {
		t.Errorf("UserAgent = %q, want test-agent/1.0", cfg.General.UserAgent)
	}

	// Accounts
	if len(cfg.Accounts) != 2 {
		t.Fatalf("len(Accounts) = %d, want 2", len(cfg.Accounts))
	}

	a0 := cfg.Accounts[0]
	if a0.Name != "work" {
		t.Errorf("account 0 Name = %q, want work", a0.Name)
	}
	if !a0.Enabled {
		t.Error("account 0 Enabled = false, want true")
	}
	if a0.Transport != "tcp" {
		t.Errorf("account 0 Transport = %q, want tcp", a0.Transport)
	}
	if a0.RegExpiry != 600 {
		t.Errorf("account 0 RegExpiry = %d, want 600", a0.RegExpiry)
	}

	// Audio
	if cfg.Audio.Mode != "file" {
		t.Errorf("Audio.Mode = %q, want file", cfg.Audio.Mode)
	}
	if cfg.Audio.PlayFile != "/tmp/test.wav" {
		t.Errorf("Audio.PlayFile = %q, want /tmp/test.wav", cfg.Audio.PlayFile)
	}
}

func TestLoadMissingSipURI(t *testing.T) {
	tomlData := `
[[accounts]]
name = "bad"
registrar = "sip:reg.example.com"
`
	path := writeTestConfig(t, tomlData)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing sip_uri")
	}
	if !strings.Contains(err.Error(), "sip_uri") {
		t.Errorf("error %q should mention sip_uri", err)
	}
}

func TestLoadMissingRegistrar(t *testing.T) {
	tomlData := `
[[accounts]]
name = "bad"
sip_uri = "sip:alice@example.com"
`
	path := writeTestConfig(t, tomlData)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing registrar")
	}
	if !strings.Contains(err.Error(), "registrar") {
		t.Errorf("error %q should mention registrar", err)
	}
}

func TestLoadNoAccounts(t *testing.T) {
	tomlData := `
[general]
log_level = 3
`
	path := writeTestConfig(t, tomlData)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for no accounts")
	}
	if !strings.Contains(err.Error(), "at least one account") {
		t.Errorf("error %q should mention accounts requirement", err)
	}
}

func TestDefaultsApplied(t *testing.T) {
	tomlData := `
[[accounts]]
name = "minimal"
sip_uri = "sip:alice@example.com"
auth_password = "secret"
registrar = "sip:reg.example.com"
`
	path := writeTestConfig(t, tomlData)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// General defaults
	if cfg.General.LogLevel != 3 {
		t.Errorf("default LogLevel = %d, want 3", cfg.General.LogLevel)
	}
	if cfg.General.UserAgent != "siptty/0.1" {
		t.Errorf("default UserAgent = %q, want siptty/0.1", cfg.General.UserAgent)
	}

	// Account defaults
	a := cfg.Accounts[0]
	if !a.Enabled {
		t.Error("default Enabled = false, want true")
	}
	if a.Transport != "udp" {
		t.Errorf("default Transport = %q, want udp", a.Transport)
	}
	if !a.Register {
		t.Error("default Register = false, want true")
	}
	if a.RegExpiry != 300 {
		t.Errorf("default RegExpiry = %d, want 300", a.RegExpiry)
	}

	// Audio defaults
	if cfg.Audio.Mode != "null" {
		t.Errorf("default Audio.Mode = %q, want null", cfg.Audio.Mode)
	}
}

func TestInvalidTransport(t *testing.T) {
	tomlData := `
[[accounts]]
name = "bad-transport"
sip_uri = "sip:alice@example.com"
registrar = "sip:reg.example.com"
transport = "ws"
`
	path := writeTestConfig(t, tomlData)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid transport")
	}
	if !strings.Contains(err.Error(), "invalid transport") {
		t.Errorf("error %q should mention invalid transport", err)
	}
}

func TestInvalidAudioMode(t *testing.T) {
	tomlData := `
[[accounts]]
name = "ok"
sip_uri = "sip:alice@example.com"
registrar = "sip:reg.example.com"

[audio]
mode = "pulse"
`
	path := writeTestConfig(t, tomlData)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid audio mode")
	}
	if !strings.Contains(err.Error(), "invalid audio mode") {
		t.Errorf("error %q should mention invalid audio mode", err)
	}
}

func TestHeaderOverrides(t *testing.T) {
	tomlData := `
[[accounts]]
name = "headers-test"
sip_uri = "sip:alice@example.com"
registrar = "sip:reg.example.com"

[accounts.headers]
X-Tenant = "acme"
X-Region = "us-east"
`
	path := writeTestConfig(t, tomlData)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	headers := cfg.Accounts[0].Headers
	if len(headers) != 2 {
		t.Fatalf("len(Headers) = %d, want 2", len(headers))
	}
	if headers["X-Tenant"] != "acme" {
		t.Errorf("Headers[X-Tenant] = %q, want acme", headers["X-Tenant"])
	}
	if headers["X-Region"] != "us-east" {
		t.Errorf("Headers[X-Region] = %q, want us-east", headers["X-Region"])
	}
}

func TestAuthUserDerivedFromSipURI(t *testing.T) {
	tomlData := `
[[accounts]]
name = "derive-test"
sip_uri = "sip:charlie@voip.example.com"
auth_password = "pass"
registrar = "sip:reg.example.com"
`
	path := writeTestConfig(t, tomlData)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Accounts[0].AuthUser != "charlie" {
		t.Errorf("AuthUser = %q, want charlie", cfg.Accounts[0].AuthUser)
	}
}

func TestAuthUserDerivedFromSipsURI(t *testing.T) {
	tomlData := `
[[accounts]]
name = "sips-test"
sip_uri = "sips:secure@tls.example.com"
auth_password = "pass"
registrar = "sip:reg.example.com"
transport = "tls"
`
	path := writeTestConfig(t, tomlData)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Accounts[0].AuthUser != "secure" {
		t.Errorf("AuthUser = %q, want secure", cfg.Accounts[0].AuthUser)
	}
}

func TestAuthUserExplicitNotOverridden(t *testing.T) {
	tomlData := `
[[accounts]]
name = "explicit-auth"
sip_uri = "sip:alice@example.com"
auth_user = "custom-user"
auth_password = "pass"
registrar = "sip:reg.example.com"
`
	path := writeTestConfig(t, tomlData)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Accounts[0].AuthUser != "custom-user" {
		t.Errorf("AuthUser = %q, want custom-user", cfg.Accounts[0].AuthUser)
	}
}

func TestMinimalConfig(t *testing.T) {
	tomlData := `
[[accounts]]
name = "main"
sip_uri = "sip:user@pbx.local"
auth_password = "pw"
registrar = "sip:pbx.local"
`
	path := writeTestConfig(t, tomlData)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(cfg.Accounts) != 1 {
		t.Fatalf("len(Accounts) = %d, want 1", len(cfg.Accounts))
	}
	a := cfg.Accounts[0]
	if a.Name != "main" {
		t.Errorf("Name = %q, want main", a.Name)
	}
	if a.SipURI != "sip:user@pbx.local" {
		t.Errorf("SipURI = %q, want sip:user@pbx.local", a.SipURI)
	}
	if a.Registrar != "sip:pbx.local" {
		t.Errorf("Registrar = %q, want sip:pbx.local", a.Registrar)
	}
	if a.AuthUser != "user" {
		t.Errorf("AuthUser = %q, want user (derived)", a.AuthUser)
	}
}

func TestEnabledExplicitlyFalse(t *testing.T) {
	tomlData := `
[[accounts]]
name = "disabled"
enabled = false
sip_uri = "sip:alice@example.com"
registrar = "sip:reg.example.com"
`
	path := writeTestConfig(t, tomlData)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Accounts[0].Enabled {
		t.Error("Enabled = true, want false (explicitly set)")
	}
}

func TestRegisterExplicitlyFalse(t *testing.T) {
	tomlData := `
[[accounts]]
name = "no-register"
sip_uri = "sip:alice@example.com"
registrar = "sip:reg.example.com"
register = false
`
	path := writeTestConfig(t, tomlData)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Accounts[0].Register {
		t.Error("Register = true, want false (explicitly set)")
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadInvalidTOML(t *testing.T) {
	path := writeTestConfig(t, "this is not valid [[[ toml")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

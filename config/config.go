package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the top-level configuration.
type Config struct {
	Servers map[string]ServerConfig `toml:"servers"`
}

// ServerConfig holds connection details for one TrueNAS server.
type ServerConfig struct {
	Host               string     `toml:"host"`
	Port               int        `toml:"port"`
	Username           string     `toml:"username"`
	APIKey             string     `toml:"api_key"`
	InsecureSkipVerify bool       `toml:"insecure_skip_verify"`
	SSH                *SSHConfig `toml:"ssh"`
}

// SSHConfig holds optional SSH connection details for filesystem operations.
type SSHConfig struct {
	Host               string `toml:"host"`
	Port               int    `toml:"port"`
	Username           string `toml:"username"`
	PrivateKeyPath     string `toml:"private_key_path"`
	HostKeyFingerprint string `toml:"host_key_fingerprint"`
}

// DefaultPath returns the default config file path using XDG conventions.
func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "truenas-tui", "config.toml")
}

// LoadFrom reads and parses the config file at the given path.
// It applies defaults for SSH config fields after parsing.
func LoadFrom(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("loading config from %s: %w", path, err)
	}
	if len(cfg.Servers) == 0 {
		return nil, fmt.Errorf("config has no servers defined")
	}
	for name, server := range cfg.Servers {
		if server.SSH != nil {
			if server.SSH.Port == 0 {
				server.SSH.Port = 22
			}
			if server.SSH.Username == "" {
				server.SSH.Username = server.Username
			}
			server.SSH.PrivateKeyPath = expandPath(server.SSH.PrivateKeyPath)
		}
		cfg.Servers[name] = server
	}
	return &cfg, nil
}

// expandPath expands ~ to $HOME and then expands all environment variables.
func expandPath(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		path = "$HOME" + path[1:]
	}
	return os.ExpandEnv(path)
}

// ServerNames returns the sorted list of server profile names.
func (c *Config) ServerNames() []string {
	names := make([]string, 0, len(c.Servers))
	for name := range c.Servers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

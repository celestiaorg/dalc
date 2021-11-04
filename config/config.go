package config

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

const (
	DefaultDirName = ".dalc"
	ConfigFileName = "dalc.toml"
)

var (
	DefaultConfigPath string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	DefaultConfigPath = fmt.Sprintf("%s/%s/", homeDir, DefaultDirName)
}

type ServerConfig struct {
	BlockSubmitterConfig `toml:"block-submitter"`
	LightClientConfig    `toml:"light-client"`
	KeyringConfig        `toml:"keyring"`
}

func (cfg ServerConfig) Save(path string) error {
	cfgFile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	err = toml.NewEncoder(cfgFile).Encode(cfg)
	if err != nil {
		return err
	}
	return cfgFile.Close()
}

func Load(path string) (ServerConfig, error) {
	var cfg ServerConfig
	rawCfg, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	_, err = toml.Decode(string(rawCfg), &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		// todo(evan): add default configs later
		// BlockSubmitterConfig: ,
		// LightClientConfig: ,
		KeyringConfig: DefaultKeyringConfig(),
	}
}

type BlockSubmitterConfig struct {
}

type KeyringConfig struct {
	KeyringBackend string `toml:"backend"`
	KeyringPath    string `toml:"path"`
}

func DefaultKeyringConfig() KeyringConfig {
	return KeyringConfig{
		KeyringBackend: "os",
		KeyringPath:    DefaultConfigPath,
	}
}

type LightClientConfig struct {
}

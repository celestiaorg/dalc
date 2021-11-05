package config

import (
	"fmt"
	"log"
	"os"
	"time"

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
		BlockSubmitterConfig: DefaultBlockSubmitterConfig(),
		KeyringConfig:        DefaultKeyringConfig(),
	}
}

// BlockSubmitterConfig holds the settings relevant for submitting a block to Celestia
// Config holds all configuration required by Celestia DA layer client.
type BlockSubmitterConfig struct {
	// temporary fee fields
	GasLimit  uint64
	FeeAmount uint64
	Denom     string

	// RPC related params
	RPCAddress string
	ChainID    string
	Timeout    time.Duration

	// BroadcastMode determines what the light client does after submitting a
	// WirePayForMessage. 0 Unspecified, 1 Block until included in a block, 2
	// Syncronous, 3 Asyncronous
	BroadcastMode int // see https://github.com/celestiaorg/cosmos-sdk/blob/51997c8de9c54e279f303a556ab59ea5dd28f1e2/types/tx/service.pb.go#L71-L83 // nolint: lll

	// KeyringAccName is the name of the account registered in the keyring
	// for the `From` address field
	KeyringAccName string
}

func DefaultBlockSubmitterConfig() BlockSubmitterConfig {
	return BlockSubmitterConfig{
		GasLimit:       2000000,
		FeeAmount:      1,
		Denom:          "tia",
		RPCAddress:     "127.0.0.1:9090",
		KeyringAccName: "test",
	}
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

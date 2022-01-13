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
	HomeDir string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	HomeDir = homeDir
}

func ConfigPath(home string) string {
	return fmt.Sprintf("%s/%s/%s", home, DefaultDirName, ConfigFileName)
}

func DirectoryPath(home string) string {
	return fmt.Sprintf("%s/%s", home, DefaultDirName)
}

type ServerConfig struct {
	BaseConfig           `toml:"base"`
	BlockSubmitterConfig `toml:"block-submitter"`
	LightClientConfig    `toml:"light-client"`
	KeyringConfig        `toml:"keyring"`
}

// Save saves the server config to a specific path
func (cfg ServerConfig) Save(home string) error {
	cfgFile, err := os.OpenFile(ConfigPath(home), os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	err = toml.NewEncoder(cfgFile).Encode(cfg)
	if err != nil {
		return err
	}
	return cfgFile.Close()
}

// Load attempts to load the dalc.toml file from the provided path
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

// DefaultServerConfig returns the default ServerConfig
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		BaseConfig:           DefaultBaseConfig(),
		BlockSubmitterConfig: DefaultBlockSubmitterConfig(),
		KeyringConfig:        DefaultKeyringConfig(),
	}
}

// BaseConfig contains the basic configurations required for the grpc server
type BaseConfig struct {
	ListenAddr string `toml:"laddr"`
	Namespace  string `toml:"namespace"`
}

func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		ListenAddr: "0.0.0.0:4200",
		Namespace:  "0102030405060708",
	}
}

// BlockSubmitterConfig holds the settings relevant for submitting a block to Celestia
// Config holds all configuration required by Celestia DA layer client.
type BlockSubmitterConfig struct {
	// GasLimit is the gas limit used to submit celestia txs. Defaults to 200000
	GasLimit uint64 `toml:"gas-limit"`
	// FeeAmount specifies the fee to be used per amount of gas used. Defaults
	// to 1
	FeeAmount uint64 `toml:"fee-amount"`
	// Denomination is the token denomination of the celestia chain being used
	// for a data availability layer. Defaults to "tia"
	Denom string `toml:"denomination"`
	// RPCAddress is the rpc address to submit celestia transactions to and
	// become a light client of. Defaults to "127.0.0.1:9090"
	GRPCAddress string `toml:"celestia-grpc-addr"`
	// RestRPCAddress is the ip and port of the celestia rest API that is used
	// to create a remote node. Will be removed with future updates to celestia-node
	RestRPCAddress string `toml:"celestia-rest-addr"`
	// ChainID is the chainID of the celstia chain being used as a data availability layer
	ChainID string `toml:"chain-id"`
	// Timeout is the amount of time in seconds waited for a tx to be included in a block. Defaults to 180 seconds
	Timeout time.Duration `toml:"timeout"` // todo: actually implement a timeout
	// BroadcastMode determines what the light client does after submitting a
	// WirePayForMessage. 0 Unspecified, 1 Block until included in a block, 2
	// Synchronous, 3 Asynchronous. Defaults to 1 Note: due to the difference
	// between WirePayForMessage and PayForMessage, celestia-core currently can
	// not properly notify the dalc that the WirePayForMessage was included in
	// the block, so we are defaulting to 2 at the moment.
	BroadcastMode int `toml:"broadcast-mode"` // see https://github.com/celestiaorg/cosmos-sdk/blob/51997c8de9c54e279f303a556ab59ea5dd28f1e2/types/tx/service.pb.go#L71-L83 // nolint: lll
	// KeyringAccName is the name of the account registered in the keyring
	// for the `From` address field. Defaults to "test"
	KeyringAccName string `toml:"keyring-account-name"`
}

// DefaultBlockSubmitterConfig returns the default configurations for the
// BlockSubmitter portion of the server config1
func DefaultBlockSubmitterConfig() BlockSubmitterConfig {
	return BlockSubmitterConfig{
		GasLimit:       2000000,
		FeeAmount:      1,
		Denom:          "tia",
		GRPCAddress:    "127.0.0.1:9090",
		RestRPCAddress: "127.0.0.1:26657",
		KeyringAccName: "dalc",
		BroadcastMode:  1,
		Timeout:        time.Minute * 3,
		ChainID:        "test",
	}
}

// KeyringConfig contains the info relevant to using a keyring
type KeyringConfig struct {
	// KeyringBackend indicates which type of backend is to be used by the
	// keyring
	KeyringBackend string `toml:"backend"`
	// KeyringPath specifies the path do which any keyring data is stored.
	KeyringPath string `toml:"path"`
}

// DefaultKeyringConfig returns the default configuration of the Keyring portion
// of the ServerConfig
func DefaultKeyringConfig() KeyringConfig {
	return KeyringConfig{
		KeyringBackend: "test",
		KeyringPath:    "~/.dalc",
	}
}

type LightClientConfig struct {
}

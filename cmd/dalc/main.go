package main

import (
	"os"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/dalc/config"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/spf13/cobra"
	"github.com/tendermint/spm/cosmoscmd"
)

func main() {
	root := rootCmd()

	cosmoscmd.SetPrefixes(app.AccountAddressPrefix)

	root.AddCommand(
		keys.Commands(config.DefaultConfigPath),
		initCmd(),
	)
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dalc",
		Short: "data availability light client for celestia",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return rootCmd
}

func initCmd() *cobra.Command {
	const pathFlag = "path"
	command := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := cmd.Flags().GetString(pathFlag)
			if err != nil {
				return err
			}

			err = os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}

			cfg := config.DefaultServerConfig()
			err = cfg.Save(path + config.ConfigFileName)
			if err != nil {
				return err
			}

			return nil
		},
	}
	command.Flags().String(pathFlag, config.DefaultConfigPath, "specific the home path")
	return command
}

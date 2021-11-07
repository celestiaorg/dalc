package main

import (
	"log"
	"net"
	"os"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/server"
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
		startCmd(),
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

			err = os.MkdirAll(path, 0664)
			if err != nil {
				return err
			}

			cfg := config.DefaultServerConfig()
			err = cfg.Save(path)
			if err != nil {
				return err
			}

			return nil
		},
	}
	command.Flags().String(pathFlag, config.DefaultConfigPath, "specific the home path")
	return command
}

func startCmd() *cobra.Command {
	const pathFlag = "path"
	command := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			// load the config
			path, err := cmd.Flags().GetString(pathFlag)
			if err != nil {
				return err
			}

			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			// create the grpc server
			srv, err := server.New(cfg)
			if err != nil {
				return err
			}

			// boot the server
			lis, err := net.Listen("tcp", cfg.ListenAddr)
			if err != nil {
				log.Panic(err)
			}
			log.Println("Listening on:", lis.Addr())

			return srv.Serve(lis)
		},
	}
	command.Flags().String(pathFlag, config.DefaultConfigPath, "specific the home path")
	return command
}

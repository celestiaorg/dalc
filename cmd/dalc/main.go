package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/server"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/spf13/cobra"
	"github.com/tendermint/spm/cosmoscmd"
)

func init() {
	cosmoscmd.SetPrefixes(app.AccountAddressPrefix)
}

func main() {
	root := rootCmd()

	root.AddCommand(
		keys.Commands(config.ConfigPath(config.HomeDir)),
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
	const homeFlag = "home"
	command := &cobra.Command{
		Use: "init",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := cmd.Flags().GetString(homeFlag)
			if err != nil {
				return err
			}

			// todo: update permissions to something more reasonable
			err = os.MkdirAll(path+"/"+config.DefaultDirName, 0777)
			if err != nil {
				return err
			}

			cfg := config.DefaultServerConfig()
			err = cfg.Save(path)
			if err != nil {
				return err
			}

			hm := server.HeightMapper{}
			err = hm.SaveToFile(filepath.Join(path, server.HeightMapFileName))
			if err != nil {
				return err
			}

			fmt.Println("Please add a key to the keyring via `dalc keys add`")
			fmt.Println("Currently referencing this key using \"dalc\", but this can be changed in the config under the BlockSubmitter section")

			return nil
		},
	}
	command.Flags().String(homeFlag, config.HomeDir, "specific the home path")
	return command
}

func startCmd() *cobra.Command {
	const homeFlag = "home"
	command := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			// load the config
			home, err := cmd.Flags().GetString(homeFlag)
			if err != nil {
				return err
			}

			dalcHome := config.ConfigPath(home)

			cfg, err := config.Load(dalcHome)
			if err != nil {
				return err
			}

			// create the grpc server
			srv, err := server.New(cfg, home, filepath.Join(home, config.CelestiaNodeHome))
			if err != nil {
				return err
			}

			// boot the server
			lis, err := net.Listen("tcp", cfg.ListenAddr)
			if err != nil {
				log.Panic(err)
			}
			log.Println("DALC listening on:", lis.Addr())

			return srv.Serve(lis)
		},
	}
	command.Flags().String(homeFlag, config.HomeDir, "specify the home path")
	return command
}

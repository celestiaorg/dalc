package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := rootCmd()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Short: "data availability light client for celestia",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return nil
		},
	}
	return rootCmd
}

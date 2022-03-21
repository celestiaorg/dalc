package main

import (
	"context"
	"log"

	"github.com/celestiaorg/celestia-app/app"
	nodecmd "github.com/celestiaorg/celestia-node/cmd"
	"github.com/celestiaorg/dalc/server"
	"github.com/tendermint/spm/cosmoscmd"
)

func init() {
	cosmoscmd.SetPrefixes(app.AccountAddressPrefix)
}

func main() {
	plugin := server.NodePlugin{}

	root := nodecmd.NewRootCmd(&plugin)

	if err := root.ExecuteContext(nodecmd.WithEnv(context.Background())); err != nil {
		log.Fatal(err)
	}
}

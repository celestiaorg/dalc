package server

import (
	"github.com/celestiaorg/celestia-node/libs/utils"
	"github.com/celestiaorg/celestia-node/node"
	"github.com/celestiaorg/celestia-node/node/fxutil"
	"github.com/celestiaorg/dalc/config"
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/fx"
)

var log = logging.Logger("dalc/server")

type NodePlugin struct{}

func (onp *NodePlugin) Name() string {
	return "Optimint GRPC Adapter"
}

func (onp *NodePlugin) Initialize(path string) error {
	cfgPath := config.ConfigPath(path)
	if !utils.Exists(cfgPath) {
		cfg := config.DefaultServerConfig(path)
		err := cfg.Save(path)
		if err != nil {
			return err
		}
		log.Infow("Saving default dalc server config", "path", cfgPath)
	} else {
		log.Infow("Config already exists", "path", cfgPath)
	}
	return nil
}

func (onp *NodePlugin) Components(cfg *node.Config, store node.Store) fxutil.Option {
	configLoader, err := LoadConfig(store)
	if err != nil {
		log.Fatal(err)
	}
	annotated := fxutil.Raw(
		fx.Provide(
			fx.Annotate(
				GRPCServer,
				fx.ResultTags(`group:"plugins"`),
			),
		),
	)
	return fxutil.Options(
		fxutil.Provide(configLoader),
		fxutil.Provide(DALC),
		annotated,
	)
}

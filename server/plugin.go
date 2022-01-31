package server

import (
	"github.com/celestiaorg/celestia-node/libs/utils"
	"github.com/celestiaorg/celestia-node/node"
	"github.com/celestiaorg/celestia-node/node/fxutil"
	"github.com/celestiaorg/dalc/config"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("dalc/server")

type NodePlugin struct{}

func (onp *NodePlugin) Name() string {
	return "Optimint GRPC Adapter"
}

func (onp *NodePlugin) Initialize(path string) error {
	cfgPath := config.ConfigPath(path)
	hmPath := HeightMapPath(path)
	if !utils.Exists(cfgPath) {
		cfg := config.DefaultServerConfig(path)
		err := cfg.Save(cfgPath)
		if err != nil {
			return err
		}
		log.Infow("Saving default dalc server config", "path", cfgPath)
	} else {
		log.Infow("Config already exists", "path", cfgPath)
	}
	if !utils.Exists(hmPath) {
		hm := NewHeightMapper()
		err := hm.SaveToFile(hmPath)
		if err != nil {
			return err
		}
		log.Infow("Saving HeightMapper", "path", cfgPath)
	} else {
		log.Infow("HeightMapper already exists", "path", cfgPath)
	}
	return nil
}

func (onp *NodePlugin) Components(cfg *node.Config, store node.Store) fxutil.Option {
	return fxutil.Options(
		fxutil.Provide(LoadConfig),
		fxutil.Provide(LoadHeightMapper),
		fxutil.Provide(DALC),
		fxutil.Provide(GRPCServer),
	)
}

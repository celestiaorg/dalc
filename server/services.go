package server

import (
	"context"
	"net"

	"github.com/celestiaorg/celestia-node/node"
	"github.com/celestiaorg/celestia-node/service/header"
	"github.com/celestiaorg/celestia-node/service/share"
	"github.com/celestiaorg/dalc/config"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func DALC(
	cfg config.ServerConfig,
	ss share.Service,
	hstore header.Store,
) (*grpc.Server, error) {
	return New(cfg, ss, hstore)
}

func LoadConfig(store node.Store) (func() config.ServerConfig, error) {
	cfg, err := config.Load(store.Path())
	if err != nil {
		return func() config.ServerConfig { return config.ServerConfig{} }, err
	}
	return func() config.ServerConfig { return cfg }, nil
}

func GRPCServer(lc fx.Lifecycle, srv *grpc.Server, cfg config.ServerConfig) node.PluginResult {
	lc.Append(
		fx.Hook{
			OnStart: func(c context.Context) error {
				// listen to the client
				lis, err := net.Listen("tcp", cfg.ListenAddr)
				if err != nil {
					return err
				}
				err = srv.Serve(lis)
				if err != nil {
					return err
				}
				return nil
			},
			OnStop: func(c context.Context) error {
				srv.Stop()
				return nil
			},
		},
	)
	return node.PluginResult(struct{}{})
}

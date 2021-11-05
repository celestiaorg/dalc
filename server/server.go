package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/proto/dalc"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
)

func NewServer(cfg config.ServerConfig) (*grpc.Server, error) {
	logger := tmlog.NewTMLogger(os.Stdout)
	// todo(evan): handle password or at least something geeze
	kr, err := keyring.New("celestia", cfg.KeyringPath, cfg.KeyringBackend, strings.NewReader(""))

	bs, err := newBlockSubmitter(cfg.BlockSubmitterConfig, kr)
	if err != nil {
		return nil, err
	}

	lc := &DataAvailabilityLightClient{
		logger:         logger,
		blockSubmitter: bs,
	}

	srv := grpc.NewServer()
	dalc.RegisterDALCServiceServer(srv, lc)

	return srv, nil
}

type DataAvailabilityLightClient struct {
	logger tmlog.Logger

	blockSubmitter blockSubmitter
}

// SubmitBlock posts an optimint blcok
func (d *DataAvailabilityLightClient) SubmitBlock(ctx context.Context, blockReq *dalc.SubmitBlockRequest) (*dalc.SubmitBlockResponse, error) {
	broadcastResp, err := d.blockSubmitter.SubmitBlock(ctx, blockReq.Block)
	if err != nil {
		return &dalc.SubmitBlockResponse{
			Result: &dalc.DAResponse{Code: dalc.StatusCode_STATUS_CODE_ERROR, Message: err.Error()},
		}, err
	}
	resp := broadcastResp.TxResponse
	if broadcastResp.TxResponse.Code != 0 {
		return &dalc.SubmitBlockResponse{
			Result: &dalc.DAResponse{
				Code:    dalc.StatusCode_STATUS_CODE_ERROR,
				Message: fmt.Sprintf("failed to broadcast tx: code %d", resp.Code),
			},
		}, err
	}
	d.logger.Info("Submitted block to celstia", "height", resp.Height, "gas used", resp.GasUsed, "hash", resp.TxHash)
	return &dalc.SubmitBlockResponse{Result: &dalc.DAResponse{Code: dalc.StatusCode_STATUS_CODE_SUCCESS}}, nil
}

func (d *DataAvailabilityLightClient) CheckBlockAvailability(context.Context, *dalc.CheckBlockAvailabilityRequest) (*dalc.CheckBlockAvailabilityResponse, error) {
	return nil, nil
}

func (d *DataAvailabilityLightClient) RetrieveBlock(context.Context, *dalc.RetrieveBlockRequest) (*dalc.RetrieveBlockResponse, error) {
	return nil, nil
}

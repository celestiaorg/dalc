package server

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	coretypes "github.com/tendermint/tendermint/types"

	"github.com/celestiaorg/celestia-node/service/header"
	"github.com/celestiaorg/celestia-node/service/share"
	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/proto/dalc"
	"github.com/celestiaorg/dalc/proto/optimint"
)

// New creates a grpc server ready to listen for incoming messages from optimint
func New(cfg config.ServerConfig, ss share.Service, hstore header.Store) (*grpc.Server, error) {
	// connect to a celestia full node to submit txs/query todo: change when
	// celestia-node does this for us
	client, err := grpc.Dial(cfg.GRPCAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// open a keyring using the configured settings
	ring, err := keyring.New(cfg.KeyringAccName, cfg.KeyringBackend, cfg.KeyringPath, strings.NewReader("."))
	if err != nil {
		return nil, err
	}

	bs, err := newBlockSubmitter(cfg.BlockSubmitterConfig, client, ring)
	if err != nil {
		return nil, err
	}

	namespace, err := hex.DecodeString(cfg.Namespace)
	if err != nil {
		return nil, err
	}

	lc := &DataAvailabilityLightClient{
		namespace:      namespace,
		blockSubmitter: bs,
		ss:             ss,
		hstore:         hstore,
	}

	srv := grpc.NewServer()
	dalc.RegisterDALCServiceServer(srv, lc)

	return srv, nil
}

type DataAvailabilityLightClient struct {
	namespace      []byte
	blockSubmitter blockSubmitter
	hstore         header.Store
	ss             share.Service
}

// SubmitBlock posts an optimint block to celestia
func (d *DataAvailabilityLightClient) SubmitBlock(ctx context.Context, blockReq *dalc.SubmitBlockRequest) (*dalc.SubmitBlockResponse, error) {
	// submit the block
	broadcastResp, err := d.blockSubmitter.SubmitBlock(ctx, blockReq.Block)
	if err != nil {
		return &dalc.SubmitBlockResponse{
			Result: &dalc.DAResponse{Code: dalc.StatusCode_STATUS_CODE_ERROR, Message: err.Error()},
		}, err
	}

	// handle response
	resp := broadcastResp.TxResponse
	if resp.Code != 0 {
		return &dalc.SubmitBlockResponse{
			Result: &dalc.DAResponse{
				Code:    dalc.StatusCode_STATUS_CODE_ERROR,
				Message: fmt.Sprintf("failed to submit tx: code %d: %s", resp.Code, resp.RawLog),
			},
		}, err
	}

	return &dalc.SubmitBlockResponse{Result: &dalc.DAResponse{Code: dalc.StatusCode_STATUS_CODE_SUCCESS}}, nil
}

// CheckBlockAvailability samples shares from the underlying data availability layer
func (d *DataAvailabilityLightClient) CheckBlockAvailability(ctx context.Context, req *dalc.CheckBlockAvailabilityRequest) (*dalc.CheckBlockAvailabilityResponse, error) {
	extHeader, err := d.hstore.GetByHeight(ctx, req.DAHeight)
	if err != nil {
		return nil, err
	}

	err = d.ss.SharesAvailable(ctx, extHeader.DAH)
	switch err {
	case nil:
		return &dalc.CheckBlockAvailabilityResponse{
			Result: &dalc.DAResponse{
				Code: dalc.StatusCode_STATUS_CODE_SUCCESS,
			},
			DataAvailable: true,
		}, nil
	default:
		return &dalc.CheckBlockAvailabilityResponse{
			Result: &dalc.DAResponse{
				Code:    dalc.StatusCode_STATUS_CODE_UNSPECIFIED,
				Message: err.Error(),
			},
			DataAvailable: false,
		}, err
	}
}

func (d *DataAvailabilityLightClient) RetrieveBlocks(ctx context.Context, req *dalc.RetrieveBlocksRequest) (*dalc.RetrieveBlocksResponse, error) {
	extHeader, err := d.hstore.GetByHeight(ctx, req.DAHeight)
	if err != nil {
		return nil, err
	}

	// todo include namespace inside the request, not preconfigured
	shares, err := d.ss.GetSharesByNamespace(ctx, extHeader.DAH, d.namespace)
	if err != nil {
		return nil, err
	}

	rawShares := make([][]byte, len(shares))
	for i, share := range shares {
		rawShares[i] = share.Data()
	}

	msgs, err := coretypes.ParseMsgs(rawShares)
	if err != nil {
		return nil, err
	}

	var blocks []*optimint.Block
	for _, msg := range msgs.MessagesList {
		var block optimint.Block
		err = proto.Unmarshal(msg.Data, &block)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, &block)
	}

	return &dalc.RetrieveBlocksResponse{
		Result: &dalc.DAResponse{
			Code: dalc.StatusCode_STATUS_CODE_SUCCESS,
		},
		Blocks: blocks,
	}, nil
}

func (d *DataAvailabilityLightClient) Start(ctx context.Context) error {
	return nil
}

func (d *DataAvailabilityLightClient) Stop(ctx context.Context) error {
	return nil
}

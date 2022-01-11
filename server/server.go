package server

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/celestiaorg/celestia-node/core"
	nodecore "github.com/celestiaorg/celestia-node/core"
	cnode "github.com/celestiaorg/celestia-node/node"
	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/proto/dalc"
	"github.com/celestiaorg/dalc/proto/optimint"
	"github.com/gogo/protobuf/proto"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/pkg/da"
	coretypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
)

// New creates a grpc server ready to listen for incoming messages from optimint
func New(cfg config.ServerConfig, nodePath string) (*grpc.Server, error) {
	logger := tmlog.NewTMLogger(os.Stdout)

	bs, err := newBlockSubmitter(cfg)
	if err != nil {
		return nil, err
	}

	repo, err := cnode.Open(nodePath, cnode.Light)
	if err != nil {
		return nil, err
	}

	node, err := cnode.New(cnode.Light, repo)
	if err != nil {
		return nil, err
	}

	coreClient, err := nodecore.NewRemote("tcp", strings.Replace(cfg.RPCAddress, "9090", "26657", -1))
	if err != nil {
		return nil, err
	}

	node.CoreClient = coreClient

	lc := &DataAvailabilityLightClient{
		logger: logger,

		blockSubmitter: bs,
		node:           node,
	}

	srv := grpc.NewServer()
	dalc.RegisterDALCServiceServer(srv, lc)

	return srv, nil
}

type DataAvailabilityLightClient struct {
	logger tmlog.Logger

	blockSubmitter blockSubmitter
	node           *cnode.Node
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

	d.logger.Info("Submitted block to celstia", "height", resp.Height, "gas used", resp.GasUsed, "hash", resp.TxHash)
	return &dalc.SubmitBlockResponse{Result: &dalc.DAResponse{Code: dalc.StatusCode_STATUS_CODE_SUCCESS}}, nil
}

// CheckBlockAvailability samples shares from the underlying data availability layer
func (d *DataAvailabilityLightClient) CheckBlockAvailability(ctx context.Context, req *dalc.CheckBlockAvailabilityRequest) (*dalc.CheckBlockAvailabilityResponse, error) {
	// get the dah for the block
	// todo(evan): change the optimint header to include some height for celestia instead of using incorrect optimint height
	dah, err := getDAH(ctx, d.node.CoreClient, int64(req.Header.Height))
	if err != nil {
		return nil, err
	}

	err = d.node.ShareServ.SharesAvailable(ctx, dah)
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

func (d *DataAvailabilityLightClient) RetrieveBlock(ctx context.Context, req *dalc.RetrieveBlockRequest) (*dalc.RetrieveBlockResponse, error) {
	// todo(evan) don't use optimint heigt use correct celestia height
	dah, err := getDAH(ctx, d.node.CoreClient, int64(req.Height))
	if err != nil {
		return nil, err
	}

	// todo include namespace inside the request, not preconfigured
	shares, err := d.node.ShareServ.GetSharesByNamespace(ctx, dah, []byte{1, 1, 1, 1, 1, 1, 1, 1})
	if err != nil {
		fmt.Println("error here2", err)
		return nil, err
	}

	rawShares := make([][]byte, len(shares))
	for i, share := range shares {
		rawShares[i] = share.Data()
	}

	msgs, err := coretypes.ParseMsgs(rawShares)
	if err != nil {
		fmt.Println("error here3", err)
		return nil, err
	}
	if len(msgs.MessagesList) != 1 {
		return nil, fmt.Errorf("only expected a single message: got %d", len(msgs.MessagesList))
	}

	var block optimint.Block
	err = proto.Unmarshal(msgs.MessagesList[0].Data, &block)
	if err != nil {
		fmt.Println("error here4", err)
		return nil, err
	}

	return &dalc.RetrieveBlockResponse{
		Result: &dalc.DAResponse{
			Code: dalc.StatusCode_STATUS_CODE_SUCCESS,
		},
		Block: &block,
	}, nil
}

// getDAH is a stop gap measure until we have header service implemented in celestia-node. This should be deleted ASAP
func getDAH(ctx context.Context, client core.Client, hate int64) (*da.DataAvailabilityHeader, error) {
	blockResp, err := client.Block(ctx, &hate)
	if err != nil {
		return nil, err
	}

	shares, _ := blockResp.Block.Data.ComputeShares()
	rawShares := shares.RawShares()

	squareSize := uint64(math.Sqrt(float64(len(shares))))

	eds, err := da.ExtendShares(squareSize, rawShares)
	if err != nil {
		panic(err)
	}

	dah := da.NewDataAvailabilityHeader(eds)

	return &dah, nil
}

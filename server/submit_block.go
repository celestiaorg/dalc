package server

import (
	"context"

	apptypes "github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/proto/optimint"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/pkg/consts"
	"google.golang.org/grpc"
)

// "github.com/celestiaorg/optimint/types"

func newBlockSubmitter(cfg config.BlockSubmitterConfig, ring keyring.Keyring) (blockSubmitter, error) {
	client, err := grpc.Dial(cfg.RPCAddress, grpc.WithInsecure())
	if err != nil {
		return blockSubmitter{}, err
	}

	signer := apptypes.NewKeyringSigner(ring, cfg.KeyringAccName, cfg.ChainID)
	return blockSubmitter{
		config:      cfg,
		signer:      signer,
		celestiaRPC: client,
	}, nil
}

// blockSubmitter submits optimint blocks to celestia
type blockSubmitter struct {
	config config.BlockSubmitterConfig
	signer *apptypes.KeyringSigner

	celestiaRPC *grpc.ClientConn
}

func (bs *blockSubmitter) buildPayForMessage(ctx context.Context, block *optimint.Block) (*apptypes.MsgWirePayForMessage, error) {
	// TODO(evan): change this when implementing ADR007
	message, err := proto.Marshal(block)
	if err != nil {
		return nil, err
	}

	pfmMsg, err := apptypes.NewWirePayForMessage(block.Header.NamespaceId, message, bs.squareSizes()...)
	if err != nil {
		return nil, err
	}

	err = bs.signer.QueryAccountNumber(ctx, bs.celestiaRPC)
	if err != nil {
		return nil, err
	}

	err = pfmMsg.SignShareCommitments(bs.signer, bs.newTxBuilder())
	if err != nil {
		return nil, err
	}

	return pfmMsg, nil
}

// SubmitBlock prepares a WirePayForMessage that contains the provided block data
func (bs *blockSubmitter) SubmitBlock(ctx context.Context, block *optimint.Block) (*tx.BroadcastTxResponse, error) {
	pfmMsg, err := bs.buildPayForMessage(ctx, block)
	if err != nil {
		return nil, err
	}

	err = bs.signer.QueryAccountNumber(ctx, bs.celestiaRPC)
	if err != nil {
		return nil, err
	}

	wirePFMtx, err := bs.signer.BuildSignedTx(bs.newTxBuilder(), pfmMsg)
	if err != nil {
		return nil, err
	}

	rawTx, err := bs.signer.EncodeTx(wirePFMtx)
	if err != nil {
		return nil, err
	}

	txClient := tx.NewServiceClient(bs.celestiaRPC)

	return txClient.BroadcastTx(
		ctx,
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode(bs.config.BroadcastMode),
			TxBytes: rawTx,
		},
	)
}

func (bs *blockSubmitter) squareSizes() []uint64 {
	// todo: don't hardcode the square sizes
	return []uint64{
		// only use a single square size until the app is fixed in #144
		// consts.MaxSquareSize / 8,
		// consts.MaxSquareSize / 4,
		// consts.MaxSquareSize / 2,
		consts.MaxSquareSize,
	}
}

func (bs *blockSubmitter) newTxBuilder() client.TxBuilder {
	// todo: don't hardcode the gas limit and fees
	builder := bs.signer.NewTxBuilder()
	fee := sdk.Coins{sdk.NewCoin(bs.config.Denom, sdk.NewInt(int64(bs.config.FeeAmount)))}
	builder.SetFeeAmount(fee)
	builder.SetGasLimit(bs.config.GasLimit)
	return builder
}
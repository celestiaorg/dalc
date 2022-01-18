package server

import (
	"context"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/x/payment/types"
	apptypes "github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/proto/optimint"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/spm/cosmoscmd"
	"github.com/tendermint/tendermint/pkg/consts"
	"google.golang.org/grpc"
)

func newBlockSubmitter(cfg config.BlockSubmitterConfig, conn *grpc.ClientConn, ring keyring.Keyring) (blockSubmitter, error) {
	encCfg := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	signer := apptypes.NewKeyringSigner(ring, cfg.KeyringAccName, cfg.ChainID)

	return blockSubmitter{
		config:      cfg,
		signer:      signer,
		celestiaRPC: conn,
		encCfg:      encCfg,
	}, nil
}

// blockSubmitter submits optimint blocks to celestia
type blockSubmitter struct {
	config config.BlockSubmitterConfig
	signer *apptypes.KeyringSigner

	encCfg cosmoscmd.EncodingConfig

	celestiaRPC *grpc.ClientConn
}

func (bs *blockSubmitter) buildPayForMessage(block *optimint.Block) (*apptypes.MsgWirePayForMessage, error) {
	// TODO(evan): change this when implementing ADR007
	message, err := proto.Marshal(block)
	if err != nil {
		return nil, err
	}

	pfmMsg, err := apptypes.NewWirePayForMessage(block.Header.NamespaceId, message, bs.squareSizes()...)
	if err != nil {
		return nil, err
	}

	err = pfmMsg.SignShareCommitments(
		bs.signer,
		types.SetFeeAmount(
			sdk.NewCoins(
				sdk.NewCoin(
					bs.config.Denom,
					sdk.NewInt(int64(bs.config.FeeAmount)),
				),
			),
		),
		types.SetGasLimit(bs.config.GasLimit),
	)
	if err != nil {
		return nil, err
	}

	return pfmMsg, nil
}

// SubmitBlock prepares a WirePayForMessage that contains the provided block data
func (bs *blockSubmitter) SubmitBlock(ctx context.Context, block *optimint.Block) (*tx.BroadcastTxResponse, error) {
	err := bs.signer.QueryAccountNumber(ctx, bs.celestiaRPC)
	if err != nil {
		return nil, err
	}

	pfmMsg, err := bs.buildPayForMessage(block)
	if err != nil {
		return nil, err
	}
	wirePFMtx, err := bs.signer.BuildSignedTx(bs.newTxBuilder(), pfmMsg)
	if err != nil {
		return nil, err
	}

	rawTx, err := bs.encCfg.TxConfig.TxEncoder()(wirePFMtx)
	if err != nil {
		return nil, err
	}

	txClient := tx.NewServiceClient(bs.celestiaRPC)

	return txClient.BroadcastTx(
		ctx,
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode(1),
			TxBytes: rawTx,
		},
	)
}

func (bs *blockSubmitter) squareSizes() []uint64 {
	// todo: don't hardcode the square sizes
	return []uint64{
		consts.MaxSquareSize / 8,
		consts.MaxSquareSize / 4,
		consts.MaxSquareSize / 2,
		consts.MaxSquareSize,
	}
}

// todo: refactor this out
func (bs *blockSubmitter) newTxBuilder() client.TxBuilder {
	// todo: don't hardcode the gas limit and fees
	builder := bs.signer.NewTxBuilder()
	fee := sdk.Coins{sdk.NewCoin(bs.config.Denom, sdk.NewInt(int64(bs.config.FeeAmount)))}
	builder.SetFeeAmount(fee)
	builder.SetGasLimit(bs.config.GasLimit)

	return builder
}

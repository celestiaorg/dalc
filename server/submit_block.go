package server

import (
	"context"
	"strings"

	"github.com/celestiaorg/celestia-app/app"
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

func newBlockSubmitter(cfg config.ServerConfig) (blockSubmitter, error) {
	client, err := grpc.Dial(cfg.RPCAddress, grpc.WithInsecure())
	if err != nil {
		return blockSubmitter{}, err
	}

	ring, err := keyring.New("", cfg.KeyringBackend, cfg.KeyringPath, strings.NewReader(""))
	if err != nil {
		return blockSubmitter{}, err
	}

	encCfg := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)

	signer := apptypes.NewKeyringSigner(ring, cfg.KeyringAccName, cfg.ChainID)

	return blockSubmitter{
		config:      cfg.BlockSubmitterConfig,
		signer:      signer,
		celestiaRPC: client,
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

	err = pfmMsg.SignShareCommitments(bs.signer, bs.newTxBuilder())
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
			Mode:    tx.BroadcastMode(2),
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

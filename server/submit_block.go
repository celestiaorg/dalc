package server

import (
	"context"
	"time"

	apptypes "github.com/celestiaorg/celestia-app/x/payment/types"
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

// SubmitBlockConfig holds the settings relevant for submitting a block to Celestia
// Config holds all configuration required by Celestia DA layer client.
type SubmitBlockConfig struct {
	// temporary fee fields
	GasLimit  uint64
	FeeAmount uint64
	Denom     string

	// RPC related params
	RPCAddress string
	ChainID    string
	Timeout    time.Duration

	// BroadcastMode determines what the light client does after submitting a
	// WirePayForMessage. 0 Unspecified, 1 Block until included in a block, 2
	// Syncronous, 3 Asyncronous
	BroadcastMode int // see https://github.com/celestiaorg/cosmos-sdk/blob/51997c8de9c54e279f303a556ab59ea5dd28f1e2/types/tx/service.pb.go#L71-L83 // nolint: lll

	// keyring related params

	// KeyringAccName is the name of the account registered in the keyring
	// for the `From` address field
	KeyringAccName string
	// // Backend is the backend of keyring that contains the KeyringAccName
	// Backend    string
	// KeyringDir string
}

func DefaultSubmitBlockConfig() SubmitBlockConfig {
	return SubmitBlockConfig{
		GasLimit:   2000000,
		FeeAmount:  1,
		Denom:      "tia",
		RPCAddress: "127.0.0.1:9090",
	}
}

func (sbc SubmitBlockConfig) newBlockSubmitter(ring keyring.Keyring) (blockSubmitter, error) {
	client, err := grpc.Dial(sbc.RPCAddress, grpc.WithInsecure())
	if err != nil {
		return blockSubmitter{}, err
	}

	signer := apptypes.NewKeyringSigner(ring, sbc.KeyringAccName, sbc.ChainID)
	return blockSubmitter{
		config:      sbc,
		signer:      signer,
		celestiaRPC: client,
	}, nil
}

func (sbc SubmitBlockConfig) ValidateBasic() error {
	// todo(evan)
	return nil
}

// blockSubmitter submits optimint blocks to celestia
type blockSubmitter struct {
	config SubmitBlockConfig
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
		consts.MaxSquareSize / 8,
		consts.MaxSquareSize / 4,
		consts.MaxSquareSize / 2,
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

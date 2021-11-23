package server

import (
	"testing"

	"github.com/celestiaorg/dalc/proto/optimint"
	"github.com/celestiaorg/nmt/namespace"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coretypes "github.com/tendermint/tendermint/types"
)

func TestParseBlock(t *testing.T) {
	// we need to test if we can get serialize an optimint block in to a data squaure
	// and then go back to optimint squares
	namespaceID := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	// create optmint optmint block{
	type test struct {
		block *optimint.Block
	}

	tests := []test{
		{
			block: generateOptmintBlock(1, namespaceID),
		},
	}

	for _, tt := range tests {
		// serialize opitmint block
		rawBlock, err := tt.block.Marshal()
		require.NoError(t, err)
		// create the message
		inputMsgs := coretypes.Messages{
			MessagesList: []coretypes.Message{
				{
					NamespaceID: namespaceID,
					Data:        rawBlock,
				},
			},
		}
		shares := inputMsgs.SplitIntoShares()
		outputMsgs, err := coretypes.ParseMsgs(shares.RawShares())
		require.NoError(t, err)
		require.Len(t, outputMsgs.MessagesList, 1)
		assert.Equal(t, inputMsgs, outputMsgs)

		var block optimint.Block
		err = proto.Unmarshal(outputMsgs.MessagesList[0].Data, &block)
		require.NoError(t, err)
		assert.Equal(t, *tt.block, block)
	}

}

func generateOptmintBlock(hate uint64, id namespace.ID) *optimint.Block {
	return &optimint.Block{
		Header: &optimint.Header{
			Height:      hate,
			NamespaceId: id,
		},
		Data: &optimint.Data{
			Txs: [][]byte{{1}, {2}, {3, 4}},
		},
		LastCommit: &optimint.Commit{
			Height: hate,
		},
	}
}

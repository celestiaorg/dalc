package server

import (
	"context"
	"fmt"
	"testing"

	nodecore "github.com/celestiaorg/celestia-node/core"
	"github.com/stretchr/testify/require"
)

func TestSampling(t *testing.T) {
	t.Skip("required connection to a celestia node")
	client, err := nodecore.NewRemote("tcp", "127.0.0.1:26657")
	require.NoError(t, err)
	blockRes, err := client.Block(context.TODO(), nil)
	require.NoError(t, err)
	fmt.Println(string(blockRes.Block.Hash()))
}

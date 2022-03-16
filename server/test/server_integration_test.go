package test

import (
	"bytes"
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/proto/dalc"
	"github.com/celestiaorg/dalc/proto/optimint"
	"github.com/celestiaorg/dalc/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/spm/cosmoscmd"
	"github.com/umee-network/umee/app"
	"google.golang.org/grpc"
)

var (
	dalcClient dalc.DALCServiceClient
)

// TestIntegration is only meant to run when connected to a celestia full node
// and a celestia-node
func TestIntegration(t *testing.T) {
	t.Skip("test requires connection to a full node and celestia-node")
	// start a new dalc server
	cfg := config.DefaultServerConfig()
	cosmoscmd.SetPrefixes(app.AccountAddressPrefix)

	// this config uses the keyring that is in celestia-app
	// this account is already funded
	// funds are needed to submit blocks
	// this can be replicated by using the "single-node.sh" script
	cfg.BlockSubmitterConfig.KeyringAccName = "user1"
	cfg.KeyringConfig.KeyringPath = "~/.celestia-app"
	cfg.Denom = "stake"

	// start the DALC grpc server
	srv, err := server.New(cfg, "~/.dalc", "~/.celestia-light")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		// listen to the client
		lis, err := net.Listen("tcp", "127.0.0.1:4200")
		if err != nil {
			log.Panic(err)
		}
		err = srv.Serve(lis)
		if err != nil {
			log.Println("failure to serve grpc: ", err)
		}
	}()

	// create a client connection to the server
	time.Sleep(1 * time.Second)
	conn, err := grpc.Dial("127.0.0.1:4200", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	// set the global client
	dalcClient = dalc.NewDALCServiceClient(conn)

	optimintBlock, daHeight := testSubmitBlock(t)

	testBlockAvailability(t, daHeight)

	req := dalc.RetrieveBlockRequest{
		Height: daHeight,
	}
	resp, err := dalcClient.RetrieveBlock(context.TODO(), &req)
	require.NoError(t, err)
	assert.Equal(t, dalc.StatusCode_STATUS_CODE_SUCCESS, resp.Result.Code)
	assert.Equal(t, optimintBlock, resp.Blocks[0])
	srv.Stop()
}

//nolint:unused
func testBlockAvailability(t *testing.T, height uint64) {
	resp, err := dalcClient.CheckBlockAvailability(
		context.TODO(),
		&dalc.CheckBlockAvailabilityRequest{
			Height: height,
		},
	)
	require.NoError(t, err)
	assert.True(t, resp.DataAvailable)
	assert.Equal(t, dalc.StatusCode_STATUS_CODE_SUCCESS, resp.Result.Code)
}

//nolint:unused
func testSubmitBlock(t *testing.T) (block *optimint.Block, daHeight uint64) {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	hate := uint64(8)
	block = &optimint.Block{
		Header: &optimint.Header{
			Height:      hate,
			NamespaceId: id,
		},
		Data: &optimint.Data{
			Txs: [][]byte{bytes.Repeat([]byte{1, 2, 3, 4}, 2000), {2}, {3, 4}},
		},
		LastCommit: &optimint.Commit{
			Height: hate,
		},
	}
	req := dalc.SubmitBlockRequest{
		Block: block,
	}
	resp, err := dalcClient.SubmitBlock(context.TODO(), &req)
	require.NoError(t, err)
	require.Equal(t, dalc.StatusCode_STATUS_CODE_SUCCESS, resp.Result.Code)
	return block, resp.Result.DataLayerHeight
}

//nolint
func testRetrieveBlock(t *testing.T, block *optimint.Block) {
	req := &dalc.RetrieveBlockRequest{
		Height: block.Header.Height,
	}

	resp, err := dalcClient.RetrieveBlock(context.TODO(), req)
	require.NoError(t, err)
	require.Equal(t, dalc.StatusCode_STATUS_CODE_SUCCESS, resp.Result.Code)

}

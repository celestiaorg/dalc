package test

import (
	"bytes"
	"context"
	"testing"

	"github.com/celestiaorg/dalc/proto/dalc"
	"github.com/celestiaorg/dalc/proto/optimint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dalcClient dalc.DALCServiceClient
)

// // TestIntegration is only meant to run when connected to a celestia full node
// // and a celestia-node
// func TestIntegration(t *testing.T) {
// 	t.Skip("test requires connection to a full node and celestia-node")
// 	// start a new dalc server
// 	cfg := config.DefaultServerConfig(".")
// 	cosmoscmd.SetPrefixes(app.AccountAddressPrefix)

// 	// this config uses the keyring that is in celestia-app
// 	// this account is already funded
// 	// funds are needed to submit blocks
// 	// this can be replicated by using the "single-node.sh" script
// 	cfg.BlockSubmitterConfig.KeyringAccName = "user1"
// 	cfg.KeyringConfig.KeyringPath = "~/.celestia-app"
// 	cfg.Denom = "stake"

// 	// start the DALC grpc server
// 	srv, err := server.New(cfg, "~/.dalc")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	go func() {
// 		// listen to the client
// 		lis, err := net.Listen("tcp", "127.0.0.1:4200")
// 		if err != nil {
// 			log.Panic(err)
// 		}
// 		err = srv.Serve(lis)
// 		if err != nil {
// 			log.Println("failure to serve grpc: ", err)
// 		}
// 	}()

// 	// create a client connection to the server
// 	time.Sleep(1 * time.Second)
// 	conn, err := grpc.Dial("127.0.0.1:4200", grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// set the global client
// 	dalcClient = dalc.NewDALCServiceClient(conn)

// 	optimintBlock := testSubmitBlock(t)

// 	testBlockAvailability(t, optimintBlock.Header)

// 	req := dalc.RetrieveBlockRequest{
// 		Height: optimintBlock.Header.Height,
// 	}
// 	resp, err := dalcClient.RetrieveBlock(context.TODO(), &req)
// 	require.NoError(t, err)
// 	assert.Equal(t, dalc.StatusCode_STATUS_CODE_SUCCESS, resp.Result.Code)
// 	assert.Equal(t, optimintBlock, resp.Block)
// 	srv.Stop()
// }

//nolint:unused
func testBlockAvailability(t *testing.T, header *optimint.Header) {
	resp, err := dalcClient.CheckBlockAvailability(
		context.TODO(),
		&dalc.CheckBlockAvailabilityRequest{
			Header: header,
		},
	)
	require.NoError(t, err)
	assert.True(t, resp.DataAvailable)
	assert.Equal(t, dalc.StatusCode_STATUS_CODE_SUCCESS, resp.Result.Code)
}

//nolint:unused
func testSubmitBlock(t *testing.T) *optimint.Block {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	hate := uint64(8)
	block := &optimint.Block{
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
	return block
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

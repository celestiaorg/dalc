// //go:build integrations

package test

import (
	"context"
	"errors"
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
	"google.golang.org/grpc"
)

var (
	dalcClient dalc.DALCServiceClient
)

// TestIntegration is only meant to run when connected to celestia network
func TestIntegration(t *testing.T) {
	// start a new dalc server
	cfg := config.DefaultServerConfig()
	srv, err := server.New(cfg, "~/.celestia-light")
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
	time.Sleep(3 * time.Second)
	conn, err := grpc.Dial("127.0.0.1:4200", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	dalcClient = dalc.NewDALCServiceClient(conn)

	block := testSubmitBlock(t)
	testBlockAvailability(t, 1)
	testRetrieveBlock(t, block)

	srv.Stop()
}

func testBlockAvailability(t *testing.T, height uint64) {
	resp, err := dalcClient.RetrieveBlock(context.TODO(), &dalc.RetrieveBlockRequest{
		Height: height,
	})
	if err != nil {
		log.Fatal(err)
	}
	hash := resp.Block.Header.DataHash
	if len(hash) > 0 {
		t.Error(errors.New(""))
	}
}

func testSubmitBlock(t *testing.T) *optimint.Block {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	hate := uint64(8)
	block := &optimint.Block{
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
	req := dalc.SubmitBlockRequest{
		Block: block,
	}
	_, err := dalcClient.SubmitBlock(context.TODO(), &req)
	require.NoError(t, err)
	// require.Zero(t, resp.Result.Code, resp.Result.Message)
	return block
}

func testRetrieveBlock(t *testing.T, block *optimint.Block) {
	req := dalc.RetrieveBlockRequest{
		Height: 8,
	}
	resp, err := dalcClient.RetrieveBlock(context.TODO(), &req)
	require.NoError(t, err)

	assert.Equal(t, dalc.StatusCode_STATUS_CODE_SUCCESS, resp.Result.Code)

	assert.Equal(t, block, resp.Block)
}

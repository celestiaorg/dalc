package test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	nodecmd "github.com/celestiaorg/celestia-node/cmd"
	"github.com/celestiaorg/dalc/server"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type IntegrationSuite struct {
	suite.Suite
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, newIntegrationTestSuite())
}

func newIntegrationTestSuite() *IntegrationSuite {
	return &IntegrationSuite{}
}

func (s *IntegrationSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	s.cancel = cancel
	s.ctx = ctx
	s.wg = &sync.WaitGroup{}

	tmpDir := s.T().TempDir()

	plugin := server.NodePlugin{}

	root := nodecmd.NewRootCmd(&plugin)

	root.SetArgs([]string{
		"light",
		"init",
		"--node.store",
		fmt.Sprintf("%s/.celestia-light", tmpDir),
	})

	if err := root.ExecuteContext(nodecmd.WithEnv(ctx)); err != nil {
		log.Fatal(err)
	}

	root.SetArgs([]string{
		"light",
		"start",
		"--node.store",
		fmt.Sprintf("%s/.celestia-light", tmpDir),
	})

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := root.ExecuteContext(nodecmd.WithEnv(ctx)); err != nil {
			if strings.Compare(err.Error(), "context canceled") == 0 {
				return
			}
			log.Fatal(err)
		}
	}()
	time.Sleep(time.Second * 15)
}

func (s *IntegrationSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.cancel()
	s.wg.Wait()
}

func (s *IntegrationSuite) TestPing() {
	_, err := grpc.Dial("tcp://localhost:4200", grpc.WithInsecure())
	s.NoError(err)
}

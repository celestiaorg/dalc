package server

import (
	"testing"

	"github.com/celestiaorg/dalc/config"
	"github.com/celestiaorg/dalc/proto/optimint"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPFM(t *testing.T) {
	bs, kr := testBlockSubmitter(t)
	block := &optimint.Block{
		Header: &optimint.Header{
			Height: 1,
		},
		Data: &optimint.Data{
			Txs: [][]byte{{0, 1, 2, 3, 4}, {4, 3, 2, 1, 0}},
		},
		LastCommit: &optimint.Commit{
			Height: 1,
		},
	}
	pfm, err := bs.buildPayForMessage(block)
	require.NoError(t, err)

	signerInfo, err := kr.Key(testAccName)
	require.NoError(t, err)

	rawBlock, err := block.Marshal()
	require.NoError(t, err)

	assert.Equal(t, signerInfo.GetAddress().String(), pfm.Signer)
	assert.Contains(t, string(pfm.Message), string(rawBlock))
}

func testBlockSubmitter(t *testing.T) (blockSubmitter, keyring.Keyring) {
	t.Helper()
	kr := generateKeyring(t)
	testBS, err := newBlockSubmitter(config.DefaultServerConfig())
	require.NoError(t, err)
	return testBS, kr
}

func generateKeyring(t *testing.T, accts ...string) keyring.Keyring {
	t.Helper()
	kb := keyring.NewInMemory()

	for _, acc := range accts {
		_, _, err := kb.NewMnemonic(acc, keyring.English, "", "", hd.Secp256k1)
		if err != nil {
			t.Error(err)
		}
	}

	_, err := kb.NewAccount(testAccName, testMnemo, "1234", "", hd.Secp256k1)
	if err != nil {
		panic(err)
	}

	return kb
}

const (
	// nolint:lll
	testMnemo   = `ramp soldier connect gadget domain mutual staff unusual first midnight iron good deputy wage vehicle mutual spike unlock rocket delay hundred script tumble choose`
	testAccName = "test"
)

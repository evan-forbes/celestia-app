package app_test

import (
	"fmt"
	"testing"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/test/util/testnode"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
)

func TestSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SDK integration test in short mode.")
	}
	suite.Run(t, new(Summary))
}

type Summary struct {
	suite.Suite

	accounts []string
	cctx     testnode.Context
	ecfg     encoding.Config
}

func (s *Summary) SetupSuite() {
	t := s.T()
	t.Log("setting up integration test suite")

	accounts := make([]string, 300)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = tmrand.Str(9)
	}

	cfg := testnode.DefaultConfig().WithAccounts(accounts)
	cctx, _, _ := testnode.NewNetwork(t, cfg)
	s.accounts = cfg.Accounts
	s.ecfg = encoding.MakeConfig(app.ModuleEncodingRegisters...)
	s.cctx = cctx
}

func (s *Summary) TestQuery() {
	t := s.T()

	_, err := s.cctx.WaitForHeight(10)

	_, err = s.cctx.FillBlock(64, s.accounts, flags.BroadcastSync)

	require.NoError(t, err)

	_, err = s.cctx.WaitForHeight(20)

	for i := 10; i < 20; i++ {
		respd, err := s.cctx.Client.NamespaceSummary(s.cctx.GoContext(), 10)
		require.NoError(t, err)
		for k, t := range respd.DataSquare {
			fmt.Println(k, t)
		}
	}

}

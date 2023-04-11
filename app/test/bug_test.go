package app_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/testutil/testnode"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	uptypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestBugSuite(t *testing.T) {
	suite.Run(t, new(BugSuite))
}

type BugSuite struct {
	suite.Suite

	accounts []string
	cctx     testnode.Context
	ecfg     encoding.Config

	mut            sync.Mutex
	accountCounter int
}

func (s *BugSuite) SetupSuite() {
	t := s.T()
	t.Log("setting up integration test suite")
	accounts, cctx := testnode.DefaultNetwork(t, time.Millisecond*400)
	s.accounts = accounts
	s.ecfg = encoding.MakeConfig(app.ModuleEncodingRegisters...)
	s.cctx = cctx
}

func (s *BugSuite) unusedAccount() string {
	s.mut.Lock()
	acc := s.accounts[s.accountCounter]
	s.accountCounter++
	s.mut.Unlock()
	return acc
}

func (s *BugSuite) TestBug() {
	t := s.T()

	// retrieve the gov module account via grpc
	aqc := authtypes.NewQueryClient(s.cctx.GRPCClient)
	resp, err := aqc.ModuleAccountByName(
		s.cctx.GoContext(), &authtypes.QueryModuleAccountByNameRequest{Name: "gov"},
	)
	s.Require().NoError(err)
	var acc authtypes.AccountI
	err = s.ecfg.InterfaceRegistry.UnpackAny(resp.Account, &acc)
	s.Require().NoError(err)

	govModuleAddress := acc.GetAddress().String()

	su := &uptypes.MsgSoftwareUpgrade{
		Authority: govModuleAddress,
		Plan: uptypes.Plan{
			Name:   "all-good",
			Info:   "some text here",
			Height: 123450000,
		},
	}

	sat := s.unusedAccount()
	satAcc := getAddress(sat, s.cctx.Keyring)

	// check the balance of the sender using grpc and the bank module query client
	bqc := banktypes.NewQueryClient(s.cctx.GRPCClient)
	satResp, err := bqc.AllBalances(s.cctx.GoContext(), &banktypes.QueryAllBalancesRequest{Address: satAcc.String()})
	require.NoError(t, err)
	fmt.Println("sat resp bals", satResp.Balances)

	res, err := testnode.SignAndBroadcastTx(s.ecfg, s.cctx.Context, sat, su)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	fmt.Println(res.Code, res.Logs, res.Data, res.Events, res.Info, res.RawLog)
	require.Equal(t, abci.CodeTypeOK, res.Code, res.Code)

	require.NoError(t, s.cctx.WaitForNextBlock())
	require.NoError(t, s.cctx.WaitForNextBlock())

	ress, err := queryTx(s.cctx.Context, res.TxHash, false)
	assert.NoError(t, err)
	assert.Equal(t, abci.CodeTypeOK, ress.TxResult.Code)

}

package summary

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/celestiaorg/celestia-app/pkg/square"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types"

	"github.com/celestiaorg/celestia-app/pkg/shares"
)

func QueryNamespaceSummary(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	if len(path) != 1 {
		return nil, fmt.Errorf("expected query path length: 1 actual: %d ", len(path))
	}

	// unmarshal the block data that is passed from the ABCI client
	pbb := new(tmproto.Block)
	err := pbb.Unmarshal(req.Data)
	if err != nil {
		return nil, fmt.Errorf("error reading block: %w", err)
	}

	data, err := types.DataFromProto(&pbb.Data)
	if err != nil {
		panic(fmt.Errorf("error from proto block: %w", err))
	}

	// build the square from the set of valid and prioritised transactions.
	// The txs returned are the ones used in the square and block
	dataSquare, _, err := square.Build(data.Txs.ToSliceOfBytes(), 1, 64)
	if err != nil {
		return nil, err
	}

	namespaces := make(map[string]int)

	for _, sh := range []shares.Share(dataSquare) {
		ns, err := sh.Namespace()
		if err != nil {
			return nil, err
		}
		namespaces[hex.EncodeToString(ns.ID)]++
	}

	return json.Marshal(namespaces)
}

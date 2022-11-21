package blob

import (
	"context"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdk_tx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/celestiaorg/celestia-app/x/blob/types"
	"github.com/celestiaorg/nmt/namespace"
)

// SubmitPayForBlob builds, signs, and synchronously submits a PayForBlob
// transaction. It returns a sdk.TxResponse after submission.
func SubmitPayForBlob(
	ctx context.Context,
	signer *types.KeyringSigner,
	conn *grpc.ClientConn,
	nID namespace.ID,
	data []byte,
	gasLim uint64,
	opts ...types.TxBuilderOption,
) (*sdk.TxResponse, error) {
	opts = append(opts, types.SetGasLimit(gasLim))

	pfb, err := BuildPayForBlob(ctx, signer, conn, nID, data, opts...)
	if err != nil {
		return nil, err
	}

	signed, err := SignPayForBlob(signer, pfb, opts...)
	if err != nil {
		return nil, err
	}

	rawTx, err := signer.EncodeTx(signed)
	if err != nil {
		return nil, err
	}

	txResp, err := types.BroadcastTx(ctx, conn, sdk_tx.BroadcastMode_BROADCAST_MODE_BLOCK, rawTx)
	if err != nil {
		return nil, err
	}
	return txResp.TxResponse, nil
}

// BuildPayForBlob builds a PayForBlob message.
func BuildPayForBlob(
	ctx context.Context,
	signer *types.KeyringSigner,
	conn *grpc.ClientConn,
	nID namespace.ID,
	message []byte,
	opts ...types.TxBuilderOption,
) (*types.MsgPayForBlob, error) {
	rec := signer.GetSignerInfo()
	addr, err := rec.GetAddress()
	if err != nil {
		return nil, err
	}

	// create the raw WirePayForBlob transaction
	wpfb, err := types.NewPayForBlob(addr.String(), nID, message)
	if err != nil {
		return nil, err
	}

	// query for account information necessary to sign a valid tx
	err = signer.QueryAccountNumber(ctx, conn)
	if err != nil {
		return nil, err
	}

	return wpfb, nil
}

// SignPayForBlob signs a PayForBlob transaction.
func SignPayForBlob(
	signer *types.KeyringSigner,
	pfb *types.MsgPayForBlob,
	opts ...types.TxBuilderOption,
) (signing.Tx, error) {
	// Build and sign the final `PayForBlob` tx
	builder := signer.NewTxBuilder(opts...)
	return signer.BuildSignedTx(
		builder,
		pfb,
	)
}

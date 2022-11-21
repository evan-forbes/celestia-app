package types

import (
	"bytes"
	"errors"
	"math"

	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	shares "github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/nmt/namespace"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"golang.org/x/exp/constraints"
)

const (
	URLBLobTx = "/blob.BlobTx"
)

// ProcessedBlobTx caches the unmarshalled result of the BlobTx
type ProcessedBlobTx struct {
	Tx    sdk.Tx
	Blobs [][]byte // todo, probably switch this to coretypes.Blob after we rename that
	PFBs  []*MsgPayForBlob
}

// ProcessWirePayForBlob performs the malleation process that occurs before
// creating a block. It unmarshals and parses the BlobTx.
func ProcessBlobTx(encfg encoding.Config, bTx *BlobTx) (ProcessedBlobTx, error) {
	sdkTx, err := encfg.TxConfig.TxDecoder()(bTx.Tx)
	if err != nil {
		return ProcessedBlobTx{}, err
	}

	coreMsg := tmproto.Message{
		NamespaceId: msg.GetNamespaceId(),
		Data:        bTx.Blob,
	}

	// wrap the signed transaction data
	pfb, err := msg.unsignedPayForBlob()
	if err != nil {
		return nil, nil, nil, err
	}

	return &coreMsg, pfb, nil
}

// ValidateBasic checks for valid namespace length, declared blob size, share
// commitments, signatures for those share commitment, and fulfills the sdk.Msg
// interface.
func (msg *MsgWirePayForBlob) ValidateBasic() error {
	if err := ValidateMessageNamespaceID(msg.GetNamespaceId()); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid 'from' address: %s", err)
	}

	// make sure that the blob size matches the actual size of the blob
	if msg.BlobSize != uint64(len(msg.Blob)) {
		return ErrDeclaredActualDataSizeMismatch.Wrapf(
			"declared: %d vs actual: %d",
			msg.BlobSize,
			len(msg.Blob),
		)
	}

	return msg.ValidateMessageShareCommitment()
}

// ValidateMessageShareCommitment returns an error if the share
// commitment is invalid.
func (msg *MsgWirePayForBlob) ValidateMessageShareCommitment() error {
	// check that the commit is valid
	commit := msg.ShareCommitment
	calculatedCommit, err := CreateCommitment(msg.GetNamespaceId(), msg.Blob)
	if err != nil {
		return ErrCalculateCommit.Wrap(err.Error())
	}

	if !bytes.Equal(calculatedCommit, commit.ShareCommitment) {
		return ErrInvalidShareCommit
	}

	return nil
}

// ValidateMessageNamespaceID returns an error if the provided namespace.ID is an invalid or reserved namespace id.
func ValidateMessageNamespaceID(ns namespace.ID) error {
	// ensure that the namespace id is of length == NamespaceIDSize
	if nsLen := len(ns); nsLen != NamespaceIDSize {
		return ErrInvalidNamespaceLen.Wrapf("got: %d want: %d",
			nsLen,
			NamespaceIDSize,
		)
	}
	// ensure that a reserved namespace is not used
	if bytes.Compare(ns, appconsts.MaxReservedNamespace) < 1 {
		return ErrReservedNamespace.Wrapf("got namespace: %x, want: > %x", ns, appconsts.MaxReservedNamespace)
	}

	// ensure that ParitySharesNamespaceID is not used
	if bytes.Equal(ns, appconsts.ParitySharesNamespaceID) {
		return ErrParitySharesNamespace
	}

	// ensure that TailPaddingNamespaceID is not used
	if bytes.Equal(ns, appconsts.TailPaddingNamespaceID) {
		return ErrTailPaddingNamespace
	}

	return nil
}

// HasWirePayForBlob performs a quick but not definitive check to see if a tx
// contains a MsgWirePayForBlob. The check is quick but not definitive because
// it only uses a proto.Message generated method instead of performing a full
// type check.
func HasWirePayForBlob(tx sdk.Tx) bool {
	for _, msg := range tx.GetMsgs() {
		msgName := sdk.MsgTypeURL(msg)
		if msgName == URLMsgWirePayForBlob {
			return true
		}
	}
	return false
}

// ExtractMsgWirePayForBlob attempts to extract a MsgWirePayForBlob from a
// provided sdk.Tx. It returns an error if no MsgWirePayForBlob is found.
func ExtractMsgWirePayForBlob(tx sdk.Tx) (*MsgWirePayForBlob, error) {
	noWirePFBError := errors.New("sdk.Tx does not contain MsgWirePayForBlob sdk.Msg")
	// perform a quick check before attempting a type check
	if !HasWirePayForBlob(tx) {
		return nil, noWirePFBError
	}

	// only support malleated transactions that contain a single sdk.Msg
	if len(tx.GetMsgs()) != 1 {
		return nil, errors.New("sdk.Txs with a single MsgWirePayForBlob are currently supported")
	}

	msg := tx.GetMsgs()[0]
	wireMsg, ok := msg.(*MsgWirePayForBlob)
	if !ok {
		return nil, noWirePFBError
	}

	return wireMsg, nil
}

// MsgMinSquareSize returns the minimum square size that msgSize can be included
// in. The returned square size does not account for the associated transaction
// shares or non-interactive defaults so it is a minimum.
func MsgMinSquareSize[T constraints.Integer](msgSize T) T {
	shareCount := shares.MsgSharesUsed(int(msgSize))
	return T(MinSquareSize(shareCount))
}

// MinSquareSize returns the minimum square size that can contain shareCount
// number of shares.
func MinSquareSize[T constraints.Integer](shareCount T) T {
	return T(shares.RoundUpPowerOfTwo(uint64(math.Ceil(math.Sqrt(float64(shareCount))))))
}

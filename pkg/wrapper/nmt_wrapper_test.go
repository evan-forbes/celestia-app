package wrapper

import (
	"bytes"
	"crypto/sha256"
	"sort"
	"testing"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/pkg/namespace"
	appns "github.com/celestiaorg/celestia-app/pkg/namespace"
	"github.com/celestiaorg/nmt"
	nmtnamespace "github.com/celestiaorg/nmt/namespace"
	"github.com/celestiaorg/rsmt2d"
	"github.com/stretchr/testify/assert"
	tmrand "github.com/tendermint/tendermint/libs/rand"
)

func TestPushErasuredNamespacedMerkleTree(t *testing.T) {
	testCases := []struct {
		name       string
		squareSize int
	}{
		{
			name:       "squareSize = 8, extendedSquareSize = 16",
			squareSize: 8,
		},
		{
			name:       "squareSize = 128, extendedSquareSize = 256",
			squareSize: 128,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewErasuredNamespacedMerkleTree(uint64(tc.squareSize), 0)

			for _, d := range generateErasuredData(t, tc.squareSize, appconsts.DefaultCodec()) {
				err := tree.Push(d)
				assert.NoError(t, err)
			}
		})
	}
}

// TestRootErasuredNamespacedMerkleTree checks that the root of an erasured NMT
// is different from the root of a standard NMT. The roots should be different
// because the erasured NMT should use the parity namespace ID for leaves pushed
// to the second half of the tree.
func TestRootErasuredNamespacedMerkleTree(t *testing.T) {
	size := 8
	data := generateRandNamespacedRawData(size)
	nmtErasured := NewErasuredNamespacedMerkleTree(uint64(size), 0)
	nmtStandard := nmt.New(sha256.New(), nmt.NamespaceIDSize(namespace.NamespaceSize), nmt.IgnoreMaxNamespace(true))

	for _, d := range data {
		err := nmtErasured.Push(d)
		if err != nil {
			t.Error(err)
		}
		err = nmtStandard.Push(d)
		if err != nil {
			t.Error(err)
		}
	}

	rootErasured, err := nmtErasured.Root()
	assert.NoError(t, err)

	rootStandard, err := nmtStandard.Root()
	assert.NoError(t, err)

	assert.NotEqual(t, rootStandard, rootErasured)
}

// TestErasuredNamespacedMerkleTreeEmptyRoot checks that the root of an empty erasured NMT is always the same
func TestErasuredNamespacedMerkleTreeEmptyRoot(t *testing.T) {
	// set up a first tree with some parameters
	tree1 := NewErasuredNamespacedMerkleTree(1, 0)
	r1, err := tree1.Root()
	assert.NoError(t, err)

	// set up a second tree with different parameters
	tree2 := NewErasuredNamespacedMerkleTree(2, 1)
	r2, err := tree2.Root()
	assert.NoError(t, err)

	// as they are empty, the roots should be the same
	assert.True(t, bytes.Equal(r1, r2))
}

func TestErasureNamespacedMerkleTreePushErrors(t *testing.T) {
	squareSize := 16

	dataOverSquareSize := generateErasuredData(t, squareSize+1, appconsts.DefaultCodec())
	dataReversed := generateErasuredData(t, squareSize, appconsts.DefaultCodec())
	sort.Slice(dataReversed, func(i, j int) bool {
		return bytes.Compare(dataReversed[i], dataReversed[j]) > 0
	})
	dataWithoutNamespace := [][]byte{{0x1}}

	testCases := []struct {
		name string
		data [][]byte
	}{
		{
			name: "push over square size",
			data: dataOverSquareSize,
		},
		{
			name: "push in incorrect lexicographic order",
			data: dataReversed,
		},
		{
			name: "push data that is too short to contain a namespace",
			data: dataWithoutNamespace,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := NewErasuredNamespacedMerkleTree(uint64(squareSize), 0)
			var err error
			for _, d := range tc.data {
				err = tree.Push(d)
			}
			assert.Error(t, err)
		})
	}
}

func TestComputeExtendedDataSquare(t *testing.T) {
	squareSize := 4
	// data for a 4X4 square
	data := generateRandNamespacedRawData(squareSize * squareSize)

	_, err := rsmt2d.ComputeExtendedDataSquare(data, appconsts.DefaultCodec(), NewConstructor(uint64(squareSize)))
	assert.NoError(t, err)
}

// generateErasuredData generates random data and then erasure codes it. It
// returns a slice that is twice as long as numLeaves because it returns the
// original data + erasured data.
func generateErasuredData(t *testing.T, numLeaves int, codec rsmt2d.Codec) [][]byte {
	raw := generateRandNamespacedRawData(numLeaves)
	erasuredData, err := codec.Encode(raw)
	if err != nil {
		t.Error(err)
	}
	return append(raw, erasuredData...)
}

// generateRandNamespacedRawData returns random data of length count. Each chunk
// of random data is of size shareSize and is prefixed with a random blob
// namespace.
func generateRandNamespacedRawData(count int) (result [][]byte) {
	for i := 0; i < count; i++ {
		rawData := tmrand.Bytes(appconsts.ShareSize)
		namespace := appns.RandomBlobNamespace().Bytes()
		copy(rawData, namespace)
		result = append(result, rawData)
	}

	sortByteArrays(result)
	return result
}

func sortByteArrays(src [][]byte) {
	sort.Slice(src, func(i, j int) bool { return bytes.Compare(src[i], src[j]) < 0 })
}

// TestErasuredNamespacedMerkleTree_ProveRange checks that the proof returned by the ProveRange for all the shares within the erasured data is non-empty.
func TestErasuredNamespacedMerkleTree_ProveRange(t *testing.T) {
	for sqaureSize := 1; sqaureSize <= 16; sqaureSize++ {
		tree := NewErasuredNamespacedMerkleTree(uint64(sqaureSize), 0, nmt.IgnoreMaxNamespace(true))
		data := generateErasuredData(t, sqaureSize, appconsts.DefaultCodec())
		for _, d := range data {
			err := tree.Push(d)
			assert.NoError(t, err)
		}

		root, err := tree.Root()
		assert.NoError(t, err)
		// iterate over all the shares and check that the proof is non-empty and can be verified
		for i := 0; i < len(data); i++ {
			proof, err := tree.ProveRange(i, i+1)
			assert.NoError(t, err)
			assert.NotEmpty(t, proof.Nodes())
			assert.False(t, proof.IsEmptyProof())

			var namespaceID nmtnamespace.ID
			if i < sqaureSize {
				namespaceID = data[i][:appconsts.NamespaceSize]
			} else {
				namespaceID = appns.ParitySharesNamespace.Bytes()
			}
			verfied := proof.VerifyInclusion(appconsts.NewBaseHashFunc(), namespaceID, [][]byte{data[i]}, root)
			assert.True(t, verfied)
		}
	}
}

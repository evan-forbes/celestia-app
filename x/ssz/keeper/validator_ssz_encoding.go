// Code generated by fastssz. DO NOT EDIT.
// Hash: d85983dadcadc38fefec1b058a66bdd73d092b284876502590b81923d1d0263d
// Version: 0.1.3
package keeper

import (
	ssz "github.com/ferranbt/fastssz"
)

// MarshalSSZ ssz marshals the ValidatorSetSSZ object
func (v *ValidatorSetSSZ) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the ValidatorSetSSZ object to a target array
func (v *ValidatorSetSSZ) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf
	offset := int(4)

	// Offset (0) 'Validators'
	dst = ssz.WriteOffset(dst, offset)
	offset += len(v.Validators) * 40

	// Field (0) 'Validators'
	if size := len(v.Validators); size > 1024 {
		err = ssz.ErrListTooBigFn("ValidatorSetSSZ.Validators", size, 1024)
		return
	}
	for ii := 0; ii < len(v.Validators); ii++ {
		if dst, err = v.Validators[ii].MarshalSSZTo(dst); err != nil {
			return
		}
	}

	return
}

// UnmarshalSSZ ssz unmarshals the ValidatorSetSSZ object
func (v *ValidatorSetSSZ) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size < 4 {
		return ssz.ErrSize
	}

	tail := buf
	var o0 uint64

	// Offset (0) 'Validators'
	if o0 = ssz.ReadOffset(buf[0:4]); o0 > size {
		return ssz.ErrOffset
	}

	if o0 < 4 {
		return ssz.ErrInvalidVariableOffset
	}

	// Field (0) 'Validators'
	{
		buf = tail[o0:]
		num, err := ssz.DivideInt2(len(buf), 40, 1024)
		if err != nil {
			return err
		}
		v.Validators = make([]*ValidatorSSZ, num)
		for ii := 0; ii < num; ii++ {
			if v.Validators[ii] == nil {
				v.Validators[ii] = new(ValidatorSSZ)
			}
			if err = v.Validators[ii].UnmarshalSSZ(buf[ii*40 : (ii+1)*40]); err != nil {
				return err
			}
		}
	}
	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the ValidatorSetSSZ object
func (v *ValidatorSetSSZ) SizeSSZ() (size int) {
	size = 4

	// Field (0) 'Validators'
	size += len(v.Validators) * 40

	return
}

// HashTreeRoot ssz hashes the ValidatorSetSSZ object
func (v *ValidatorSetSSZ) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the ValidatorSetSSZ object with a hasher
func (v *ValidatorSetSSZ) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'Validators'
	{
		subIndx := hh.Index()
		num := uint64(len(v.Validators))
		if num > 1024 {
			err = ssz.ErrIncorrectListSize
			return
		}
		for _, elem := range v.Validators {
			if err = elem.HashTreeRootWith(hh); err != nil {
				return
			}
		}
		hh.MerkleizeWithMixin(subIndx, num, 1024)
	}

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the ValidatorSetSSZ object
func (v *ValidatorSetSSZ) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(v)
}

// MarshalSSZ ssz marshals the ValidatorSSZ object
func (v *ValidatorSSZ) MarshalSSZ() ([]byte, error) {
	return ssz.MarshalSSZ(v)
}

// MarshalSSZTo ssz marshals the ValidatorSSZ object to a target array
func (v *ValidatorSSZ) MarshalSSZTo(buf []byte) (dst []byte, err error) {
	dst = buf

	// Field (0) 'PubKey'
	if size := len(v.PubKey); size != 32 {
		err = ssz.ErrBytesLengthFn("ValidatorSSZ.PubKey", size, 32)
		return
	}
	dst = append(dst, v.PubKey...)

	// Field (1) 'VotingPower'
	dst = ssz.MarshalUint64(dst, v.VotingPower)

	return
}

// UnmarshalSSZ ssz unmarshals the ValidatorSSZ object
func (v *ValidatorSSZ) UnmarshalSSZ(buf []byte) error {
	var err error
	size := uint64(len(buf))
	if size != 40 {
		return ssz.ErrSize
	}

	// Field (0) 'PubKey'
	if cap(v.PubKey) == 0 {
		v.PubKey = make([]byte, 0, len(buf[0:32]))
	}
	v.PubKey = append(v.PubKey, buf[0:32]...)

	// Field (1) 'VotingPower'
	v.VotingPower = ssz.UnmarshallUint64(buf[32:40])

	return err
}

// SizeSSZ returns the ssz encoded size in bytes for the ValidatorSSZ object
func (v *ValidatorSSZ) SizeSSZ() (size int) {
	size = 40
	return
}

// HashTreeRoot ssz hashes the ValidatorSSZ object
func (v *ValidatorSSZ) HashTreeRoot() ([32]byte, error) {
	return ssz.HashWithDefaultHasher(v)
}

// HashTreeRootWith ssz hashes the ValidatorSSZ object with a hasher
func (v *ValidatorSSZ) HashTreeRootWith(hh ssz.HashWalker) (err error) {
	indx := hh.Index()

	// Field (0) 'PubKey'
	if size := len(v.PubKey); size != 32 {
		err = ssz.ErrBytesLengthFn("ValidatorSSZ.PubKey", size, 32)
		return
	}
	hh.PutBytes(v.PubKey)

	// Field (1) 'VotingPower'
	hh.PutUint64(v.VotingPower)

	hh.Merkleize(indx)
	return
}

// GetTree ssz hashes the ValidatorSSZ object
func (v *ValidatorSSZ) GetTree() (*ssz.Node, error) {
	return ssz.ProofTree(v)
}
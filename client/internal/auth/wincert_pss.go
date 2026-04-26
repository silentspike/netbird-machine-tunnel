package auth

import (
	"crypto"
	"crypto/rsa"
	"fmt"
)

// ncryptPSSSaltLength converts Go's sentinel salt lengths to the concrete
// ULONG value that Windows CNG expects in BCRYPT_PSS_PADDING_INFO.
func ncryptPSSSaltLength(opts *rsa.PSSOptions) (uint32, error) {
	if opts == nil {
		return 0, fmt.Errorf("RSA-PSS options are nil")
	}
	switch opts.SaltLength {
	case rsa.PSSSaltLengthAuto, rsa.PSSSaltLengthEqualsHash:
		size, err := supportedHashSize(opts.HashFunc())
		if err != nil {
			return 0, err
		}
		return uint32(size), nil
	default:
		if opts.SaltLength < 0 {
			return 0, fmt.Errorf("unsupported RSA-PSS salt length: %d", opts.SaltLength)
		}
		if uint64(opts.SaltLength) > uint64(^uint32(0)) {
			return 0, fmt.Errorf("RSA-PSS salt length exceeds NCrypt limit: %d", opts.SaltLength)
		}
		return uint32(opts.SaltLength), nil
	}
}

func supportedHashSize(hash crypto.Hash) (int, error) {
	switch hash {
	case crypto.SHA1:
		return crypto.SHA1.Size(), nil
	case crypto.SHA256:
		return crypto.SHA256.Size(), nil
	case crypto.SHA384:
		return crypto.SHA384.Size(), nil
	case crypto.SHA512:
		return crypto.SHA512.Size(), nil
	default:
		return 0, fmt.Errorf("unsupported hash function: %v", hash)
	}
}

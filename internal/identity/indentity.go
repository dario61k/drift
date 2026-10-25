package identity

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"

	"drift/internal/errs"
)

type Keys struct {
	Public  string `toml:"public"`
	Private string `toml:"private"`
}

func (k *Keys) Generate() error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return errs.ErrCreatingKeys
	}

	k.Public = hex.EncodeToString(publicKey)
	k.Private = hex.EncodeToString(privateKey)

	return nil
}

func (k *Keys) Validate() error {
	if k.Public == "" && k.Private == "" {
		return errs.ErrMissingKeys
	}

	publicKey, err := hex.DecodeString(k.Public)
	if err != nil || len(publicKey) != ed25519.PublicKeySize {
		return errs.ErrInvalidPublicKey
	}

	privateKey, err := hex.DecodeString(k.Private)
	if err != nil || len(privateKey) != ed25519.PrivateKeySize {
		return errs.ErrInvalidPrivateKey
	}

	derivedPrivateKey := ed25519.NewKeyFromSeed(privateKey[:ed25519.SeedSize])

	if !bytes.Equal(privateKey, derivedPrivateKey) {
		return errs.ErrInvalidPrivateKey
	}

	derivedPublicKey := derivedPrivateKey.Public().(ed25519.PublicKey)

	if !bytes.Equal(publicKey, derivedPublicKey) {
		return errs.ErrMismatchedKeyPair
	}

	return nil
}

package model

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
)

type EvmAccount struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

type EclipseAccount struct {
	PublicKey  solana.PublicKey
	PrivateKey solana.PrivateKey
}

func NewEvmAccount(address common.Address, privateKey *ecdsa.PrivateKey) *EvmAccount {
	return &EvmAccount{
		Address:    address,
		PrivateKey: privateKey,
	}
}

func NewEclipseAccount(publicKey solana.PublicKey, privateKey solana.PrivateKey) *EclipseAccount {
	return &EclipseAccount{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}

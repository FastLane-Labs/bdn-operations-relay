package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

func generateConfig() {
	dappPrivateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	dAppAddress := crypto.PubkeyToAddress(dappPrivateKey.PublicKey).String()
	dappPrivateKeyHex := hex.EncodeToString(crypto.FromECDSA(dappPrivateKey))

	solverProvateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	solverPrivateKeyHex := hex.EncodeToString(crypto.FromECDSA(solverProvateKey))

	fmt.Println("dapp-private-key", dappPrivateKeyHex)
	fmt.Println("dapp-address", dAppAddress)
	fmt.Println("solver-private-key", solverPrivateKeyHex)
}

func main() {
	generateConfig()
}

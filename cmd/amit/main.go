package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

func main() {
	t := "020000"

	i := big.NewInt(1)
	for {
		s := hex.EncodeToString(i.Bytes())
		if s == t {
			fmt.Println(i)
			break
		}
		i.Mul(i, big.NewInt(2))
	}
}

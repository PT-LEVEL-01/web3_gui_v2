package coin_address

import (
	"encoding/hex"
	"fmt"
	"testing"
)

const testAddr = "bc1q6s2s556t57jqnyy7qdtr2yu54hpkahkrp87wxd"

func TestBase58(t *testing.T) {
	bs, err := Decode(testAddr, BitcoinAlphabet)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(bs))
}

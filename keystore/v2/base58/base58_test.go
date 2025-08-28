package base58

import (
	"encoding/hex"
	"fmt"
	"testing"
)

const addr = "bc1q6s2s556t57jqnyy7qdtr2yu54hpkahkrp87wxd"

func TestBase58(t *testing.T) {
	bs := Decode(addr)
	fmt.Println(hex.EncodeToString(bs))
}

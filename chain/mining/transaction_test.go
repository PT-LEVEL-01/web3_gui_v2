package mining

import (
	"fmt"
	"testing"
	"web3_gui/keystore/adapter/crypto"
)

func Test_MergeVouts(t *testing.T) {
	vout1 := []*Vout{
		&Vout{
			Value:        10,
			Address:      crypto.AddressFromB58String("iComDK4yrCppT7im4T9UjbCrhz8KFVaPunyk35"),
			FrozenHeight: 0,
		},
		&Vout{
			Value:        20,
			Address:      crypto.AddressFromB58String("iComDK4yrCppT7im4T9UjbCrhz8KFVaPunyk35"),
			FrozenHeight: 0,
		},
	}

	vout2 := MergeVouts(&vout1)
	for _, one := range vout2 {
		fmt.Println(one.Address.B58String(), one.Value)
	}
}

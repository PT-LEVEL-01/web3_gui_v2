package full_node

import (
	"fmt"
	"github.com/craftto/go-tron/pkg/keystore"
	"math/big"
	"testing"
	usdtconfig "web3_gui/cross_chain/usdt/config"
)

func TestTrc20(t *testing.T) {
	priKey := ""
	toAddress := ""
	amounts := big.NewInt(1)
	err := ConnFullNode()
	if err != nil {
		panic(err)
	}
	ks, errs := keystore.ImportFromPrivateKey(priKey) //私钥
	if errs != nil {
		panic(err)
	}

	trxC, err := NewTrc20(rpcClient, usdtconfig.TrxUsdtAddr)
	if err != nil {
		panic(err)
	}
	trxT, errs := trxC.Transfer(ks, toAddress, amounts)
	if errs != nil {
		panic(err)
	}
	fmt.Printf("交易详情:%+v", trxT)
}

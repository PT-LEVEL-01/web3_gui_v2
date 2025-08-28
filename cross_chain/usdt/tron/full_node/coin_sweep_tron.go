package full_node

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	keystore2 "github.com/craftto/go-tron/pkg/keystore"
	"github.com/craftto/go-tron/pkg/proto/core"
	"github.com/ethereum/go-ethereum/crypto"
	"google.golang.org/protobuf/proto"
	"math/big"
	"time"
	usdtconfig "web3_gui/cross_chain/usdt/config"
	"web3_gui/cross_chain/usdt/wallet"
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/utils"
)

//var newCoinNeedSweepChan = make(chan bool, 1)

var rootAddressInfoTrx *keystore.CoinAddressInfo

/*
构建大钱包地址
*/
func BuildRootAddress() utils.ERROR {
	//大钱包地址
	addr, ERR := usdtconfig.Wallet_keystore.GetAddressCoinType(coin_address.COINTYPE_TRX, 0, 0, usdtconfig.WALLET_password)
	if ERR.CheckFail() {
		return ERR
	}
	rootAddressInfoTrx = usdtconfig.Wallet_keystore.FindCoinAddr(addr)
	return utils.NewErrorSuccess()
}

/*
币归集到大钱包
*/
func loopTriggerSweep() {
	oldHeight := int64(0)
	for range time.NewTicker(time.Minute).C {
		height, ERR := GetNodeBlockHeight()
		if ERR.CheckFail() {
			continue
		}
		if oldHeight+usdtconfig.Balance_sweep_interval_height_trx > height {
			continue
		}
		ERR = CheckAddressBalance()
		if ERR.CheckFail() {
			utils.Log.Error().Str("ERR", ERR.String()).Send()
		}
	}
}

/*
检查一遍所有地址余额
*/
func CheckAddressBalance() utils.ERROR {
	_, prkBs, ERR := rootAddressInfoTrx.DecryptPrkPuk(usdtconfig.WALLET_password)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return ERR
	}
	prk, err := crypto.ToECDSA(prkBs)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return ERR
	}

	prkStr := ""
	prkStr = hex.EncodeToString(prk.Serialize())

	addrs := wallet.GetAddressAll()
	for _, one := range addrs {
		//获取地址上的usdt余额
		amount, err := GetUSDTBalance(one)
		if err != nil {
			utils.Log.Error().Str("GetUSDTBalance", err.Error()).Send()
			continue
		}
		if amount.Cmp(big.NewInt(0)) == 0 {
			continue
		}
		//余额不为0，需要归集

		//获取trx余额
		trxBalance, err := GetTrxBalance(one)
		if err != nil {
			ERR = utils.NewErrorSysSelf(err)
			utils.Log.Error().Str("ERR", ERR.String()).Send()
			continue
		}
		if trxBalance < usdtconfig.GAS_trx {
			//先给这个地址转入足够的trx作为手续费
			_, err = TransferTRX(rootAddressInfoTrx.GetAddrStr(), one, usdtconfig.GAS_trx, "", prk)
			if err != nil {
				utils.Log.Error().Str("TransferTRX", err.Error()).Send()
			}
			continue
		}
		//开始归集
		_, err = TransferUSDT(one, rootAddressInfoTrx.GetAddrStr(), amount.Int64(), prkStr, prk)
		if err != nil {
			utils.Log.Error().Str("TransferTRX", err.Error()).Send()
		}
	}
	return utils.NewErrorSuccess()
}

/*
转账 TRX
*/
func TransferTRX(formAddr string, toAddr string, amount int64, privkey string, prk *ecdsa.PrivateKey) (string, error) {
	tx, err := rpcClient.Transfer(formAddr, toAddr, amount)
	if err != nil {
		utils.Log.Error().Str("sign SignTransaction fail", err.Error()).Send()
		return "", err
	}
	signTx, err := TrxSignTransaction(tx.Transaction, privkey, prk) //*core.Transaction
	if err != nil {
		utils.Log.Error().Str("TrxSignTransaction fail", err.Error()).Send()
		return "", err
	}
	_, err = rpcClient.Broadcast(signTx) //*core.Transaction
	//fmt.Println(rt)
	if err != nil {
		utils.Log.Error().Str("BroadcastTransaction fail", err.Error()).Send()
		return "", err
	}
	return BytesToHexString(tx.GetTxid()), nil
}

// BytesToHexString encodes bytes as a hex string.
func BytesToHexString(bytes []byte) string {
	encode := make([]byte, len(bytes)*2)
	hex.Encode(encode, bytes)
	return string(encode)
}

// 离线签名
func TrxSignTransaction(transaction *core.Transaction, privateKey string, prk *ecdsa.PrivateKey) (*core.Transaction, error) {
	//privateBytes, err := hex.DecodeString(privateKey)
	//if err != nil {
	//	return nil, fmt.Errorf("hex decode private key error: %v", err)
	//}
	//priv := crypto.ToECDSAUnsafe(privateBytes)
	//defer zeroKey(priv)
	rawData, err := proto.Marshal(transaction.GetRawData())
	if err != nil {
		return nil, fmt.Errorf("proto marshal tx raw data error: %v", err)
	}
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	signature, err := crypto.Sign(hash, prk)
	if err != nil {
		return nil, fmt.Errorf("sign error: %v", err)
	}
	transaction.Signature = append(transaction.Signature, signature)
	return transaction, nil
}

// zeroKey zeroes a private key in memory.
func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}

/*
转账 USDT
*/
func TransferUSDT(formAddr string, toAddr string, amount int64, privkey string, prk *ecdsa.PrivateKey) (string, error) {
	ks, err := keystore2.ImportFromPrivateKey(privkey) //私钥
	if err != nil {
		return "", err
	}
	trxC, err := NewTrc20(rpcClient, usdtconfig.TrxUsdtAddr)
	if err != nil {
		return "", err
	}
	trxT, errs := trxC.Transfer(ks, toAddress, amounts)
	if errs != nil {
		return "", err
	}
}

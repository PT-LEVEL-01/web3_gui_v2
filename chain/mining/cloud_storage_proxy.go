package mining

import (
	"web3_gui/chain/config"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
)

/*
构建云存储代理发奖励交易
*/
func CreateTxCloudStorageProxy(address crypto.AddressCoin, blockHeight uint64, payload []byte) *Tx_Contract {
	start := config.TimeNow()
	//构建输入
	baseCoinAddr := Area.Keystore.GetCoinbase()
	vins := make([]*Vin, 0)
	vin := Vin{
		Puk:  baseCoinAddr.Puk, //公钥
		Sign: nil,              //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	}
	vins = append(vins, &vin)

	//构建云存储奖励输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Address: address, //钱包地址
	}
	vouts = append(vouts, &vout)

	var txReward *Tx_Contract
	for i := uint64(0); i < 10000; i++ {
		base := TxBase{
			Type:       config.Wallet_tx_type_contract, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  1,
			Vin:        vins,
			Vout_total: uint64(len(vouts)), //输出交易数量
			Vout:       vouts,              //交易输出
			LockHeight: blockHeight + i,    //锁定高度
			Payload:    payload,
			Comment:    []byte{},
		}
		txReward = &Tx_Contract{
			TxBase:        base,
			ContractClass: uint64(config.CLOUD_STORAGE_PROXY_CONTRACT), //云存储类型合约
		}

		//给输出签名，防篡改
		for i, one := range txReward.Vin {
			_, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, config.Wallet_keystore_default_pwd)
			if err != nil {
				engine.Log.Error("build reward error:%s", err.Error())
				return nil
			}
			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
			sign := txReward.GetSign(&prk, uint64(i))
			txReward.Vin[i].Sign = *sign
		}

		txReward.BuildHash()
		if txReward.CheckHashExist() {
			txReward = nil
			// engine.Log.Info("hash is exist:%s", hex.EncodeToString(txReward.Hash))
			continue
		} else {
			break
		}
	}

	engine.Log.Info("构建云存储代理合约奖励消耗时间 %s", config.TimeNow().Sub(start))
	return txReward
}

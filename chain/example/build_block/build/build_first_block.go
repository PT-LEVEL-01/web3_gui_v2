package build

import (
	"bytes"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/db/leveldb"
	"web3_gui/chain/mining"
	"web3_gui/libp2parea/adapter"
	"web3_gui/utils"
)

/*
构建创世块
创世块生成两个组，一个组一个块
第一个块给3个见证者账户分配初始额度，下一个组见证者投票结果
第二个块
*/
func BuildFirstBlock(area *libp2parea.Area, lDB *leveldb.LevelDB, firstBlockTime int64) (*mining.BlockHeadVO, error) {

	//如果模块未启用，则不执行

	//if !config.InitNode || !config.DB_is_null {
	//	// if !config.InitNode {
	//	// utils.LogZero.Info().Msgf("不是创始节点")
	//	return nil, nil
	//}
	// config.InitNode = true

	// fmt.Println("开始创建创世区块")

	//db.InitDB(config.DB_path, config.DB_path_temp)

	//----------------生成第一个区块-----------------

	balanceTotal := uint64(config.Mining_coin_premining)
	//构建交易
	txHashes := make([][]byte, 0)
	txs := make([]mining.TxItr, 0)
	//创建首块预挖奖励
	reward := BuildReward(balanceTotal, area)
	txs = append(txs, reward)
	txHashes = append(txHashes, *reward.GetHash())
	//创建奖励合约交易
	//rewardContractTx := BuildRewardContractTx()
	//txs = append(txs, rewardContractTx)
	//txHashes = append(txHashes, *rewardContractTx.GetHash())
	//创建见证人押金交易
	depositIn := BuildDepositIn(area)
	txs = append(txs, depositIn)
	txHashes = append(txHashes, *depositIn.GetHash())
	//创建云存储合约交易
	//cloudStorageProxyContractTx := BuildCloudStorageProxyContractTx()
	//txs = append(txs, cloudStorageProxyContractTx)
	//txHashes = append(txHashes, *cloudStorageProxyContractTx.GetHash())

	//区块头
	blockHead1 := mining.BlockHead{
		Height:      config.Mining_block_start_height, //区块高度(每秒产生一个块高度，也足够使用上千亿年)
		GroupHeight: config.Mining_group_start_height, //
		NTx:         uint64(len(txHashes)),            //交易数量
		Tx:          txHashes,                         //本区块包含的交易id
		Time:        firstBlockTime,                   //  time.Now().Unix(),                //unix时间戳
		//Time:    config.TimeNow().Unix(),                //pl time
		Witness: area.Keystore.GetCoinbase().Addr, //
	}
	blockHead1.BuildMerkleRoot()
	blockHead1.BuildSign(area.Keystore.GetCoinbase().Addr, area.Keystore)
	blockHead1.BuildBlockHash()

	// blockHead1.FindNonce(1, make(chan bool, 1))
	bhbs, _ := blockHead1.Proto()
	// bhbs, _ := blockHead1.Json()
	bhashkey := config.BuildBlockHead(blockHead1.Hash)
	err := lDB.Save(bhashkey, bhbs)
	if err != nil {
		// utils.LogZero.Info().Msgf("1111111111111")
		return nil, err
	}
	// fmt.Println("key", "blockHead", hex.EncodeToString(blockHead1.Hash))
	// fmt.Println("value", "blockHead", string(*bhbs), "\n")

	lDB.Save(config.Key_block_start, &blockHead1.Hash)

	hashExist := false

	//保存到数据库
	for _, one := range txs {
		utils.LogZero.Info().Hex("hash1", *one.GetHash()).Hex("hash2", blockHead1.Hash).Send()
		// one.SetBlockHash(blockHead1.Hash)
		db.SaveTxToBlockHash(one.GetHash(), &blockHead1.Hash)
		// one.TxBase.BlockHash = blockHead1.Hash
		bs, err := one.Proto()
		// bs, err := one.Json()
		if err != nil {
			// fmt.Println("2 json格式化错误", err)
			return nil, err
		}
		if one.CheckHashExist() {
			//utils.LogZero.Info().Str("已经存在", "").Send()
			hashExist = true
		}

		txhashkey := config.BuildBlockTx(*one.GetHash())
		err = lDB.Save(txhashkey, bs)
		if err != nil {
			return nil, err
		}

		// fmt.Println("key", "tx", hex.EncodeToString(*one.GetHash()))
		// fmt.Println("value", "tx", string(*bs))
	}

	//保存见证人押金交易
	db.SaveTxToBlockHash(depositIn.GetHash(), &blockHead1.Hash)
	// depositIn.SetBlockHash(blockHead1.Hash)
	bs, err := depositIn.Proto()
	// bs, err := depositIn.Json()
	if err != nil {
		return nil, err
	}

	txhashkey := config.BuildBlockTx(*depositIn.GetHash())
	err = lDB.Save(txhashkey, bs)
	if err != nil {
		return nil, err
	}
	// fmt.Println("key", "tx", hex.EncodeToString(*depositIn.GetHash()))
	// fmt.Println("value", "tx", string(*bs))

	// db.SaveBlockHeight(blockHead1.Height, &blockHead1.Hash)
	// fmt.Println("创建初始块完成")

	bhvo := mining.BlockHeadVO{
		BH:  &blockHead1, //区块
		Txs: txs,         //交易明细
	}

	if hashExist {
		//return BuildFirstBlock(area, lDB)
	}

	err = lDB.StoreSnap(nil)
	if err != nil {
		panic(err)
	}
	lDB.ResetSnap()

	return &bhvo, nil
}

/*
构建创始块奖励
*/
func BuildReward(balanceTotal uint64, area *libp2parea.Area) mining.TxItr {
	//创世块矿工奖励
	baseCoinAddr := area.Keystore.GetCoinbase()

	//构建输入

	vins := make([]*mining.Vin, 0)
	vin := mining.Vin{
		Puk:  baseCoinAddr.Puk, //公钥
		Sign: nil,              //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	}
	vins = append(vins, &vin)

	vouts := make([]*mining.Vout, 0)
	vouts = append(vouts, &mining.Vout{
		Value:   balanceTotal,                     //输出金额 = 实际金额 * 100000000
		Address: area.Keystore.GetCoinbase().Addr, //钱包地址
	})
	base := mining.TxBase{
		Type:       config.Wallet_tx_type_mining, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  1,
		Vin:        vins,
		Vout_total: 1,                                //输出交易数量
		Vout:       vouts,                            //交易输出
		LockHeight: config.Mining_block_start_height, //
		//		CreateTime: config.TimeNow().Unix(),            //创建时间
	}
	reward := &mining.Tx_reward{
		TxBase: base,
	}

	//给输出签名，防篡改
	for i, one := range reward.Vin {
		for _, key := range area.Keystore.GetAddrAll() {

			puk, ok := area.Keystore.GetPukByAddr(key.Addr)
			if !ok {
				return nil //config.ERROR_public_key_not_exist
			}

			if bytes.Equal(puk, one.Puk) {
				_, prk, _, err := area.Keystore.GetKeyByAddr(key.Addr, config.Wallet_keystore_default_pwd)
				// prk, err := key.GetPriKey(pwd)
				if err != nil {
					return nil
				}
				sign := reward.GetSign(&prk, uint64(i))
				//				sign := pay.GetVoutsSign(prk, uint64(i))
				reward.Vin[i].Sign = *sign
			}
		}
	}
	reward.BuildHash()
	return reward
}

func BuildDepositIn(area *libp2parea.Area) mining.TxItr {
	coinbase := area.Keystore.GetCoinbase()

	// puk, _ := keystore.GetPukByAddr(coinbase)

	//首个见证人押金
	// mining.CreateTxDepositIn()
	txin := mining.Tx_deposit_in{
		Puk: coinbase.Puk,
	}
	//创世块矿工奖励
	// vins := make([]mining.Vin, 0)
	// vins = append(vins, mining.Vin{
	// 	Txid: txid,        //UTXO 前一个交易的id
	// 	Vout: 0,           //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	// 	Puk:  coinbasepuk, //公钥
	// 	// Sign :, //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	// })
	vouts := make([]*mining.Vout, 0)
	vouts = append(vouts, &mining.Vout{
		Value:   config.Mining_deposit, //  config.Wallet_MDL_mining, //输出金额 = 实际金额 * 100000000
		Address: coinbase.Addr,         //钱包地址
	})
	depositInBase := mining.TxBase{
		Type: config.Wallet_tx_type_deposit_in, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		// Vin_total:  uint64(len(vins)),                //输入交易数量
		// Vin:        vins,                             //交易输入
		Vout_total: uint64(len(vouts)), //
		Vout:       vouts,              //
		LockHeight: 1,                  //锁定高度
		//		CreateTime: config.TimeNow().Unix(),                //创建时间
		Payload: []byte("Hive Network Foundation"),
	}
	txin.TxBase = depositInBase
	txin.Rate = 90
	txin.BuildHash()
	return &txin
}

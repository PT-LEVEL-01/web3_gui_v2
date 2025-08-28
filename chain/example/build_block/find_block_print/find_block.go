package main

import (
	"encoding/hex"
	jsoniter "github.com/json-iterator/go"
	"path/filepath"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	bconfig "web3_gui/chain/example/build_block/config"
	"web3_gui/chain/mining"
	_ "web3_gui/chain/mining/tx_name_in"
	_ "web3_gui/chain/mining/tx_name_out"
	"web3_gui/chain/rpc"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	// find("D:/workspaces/go/src/icom_chain/example/peer_super/wallet/data")
	//path := filepath.Join(bconfig.WalletDirPath, bconfig.WalletDBName)
	// find(path)
	FindNextBlock()
	utils.LogZero.Info().Msgf("finish!")
}

func FindNextBlock() {
	path := filepath.Join(bconfig.WalletDirPath, bconfig.WalletDBName)
	//启动和链接leveldb数据库
	err := db.InitDB(path, "")
	if err != nil {
		panic(err)
	}
	//db.LevelDB.PrintAll()
	// db.InitDB(dir)
	beforBlockHash, err := db.LevelDB.Find(config.Key_block_start)
	if err != nil {
		utils.LogZero.Info().Msgf("111 查询起始块id错误 " + err.Error())
		return
	}

	utils.LogZero.Info().Hex("hex", *beforBlockHash).Send()

	beforGroupHeight := uint64(0)
	beforBlockHeight := uint64(0)

	for beforBlockHash != nil {
		bs, err := db.LevelDB.Find(config.BuildBlockHead(*beforBlockHash))
		if err != nil {
			utils.LogZero.Info().Msgf("查询第 个块错误" + err.Error())
			return
		}
		bh, err := mining.ParseBlockHeadProto(bs)
		if err != nil {

			utils.LogZero.Info().Msgf("查询第 个块错误" + err.Error())
			// fmt.Println(string(*bs))
			utils.LogZero.Info().Msgf(string(*bs))
			return
		}
		if bh.Nextblockhash == nil {
			utils.LogZero.Info().Msgf("第%d个块 -----------------------------------\n%s\n", bh.Height)
		} else {
			utils.LogZero.Info().Msgf("第%d个块 -----------------------------------\n%s\n", bh.Height,
				hex.EncodeToString(bh.Hash))
			// fmt.Println("第", bh.Height, "个块 -----------------------------------\n",
			// 	hex.EncodeToString(bh.Hash), "\n", string(*bs), "\nnext区块个数", len(bh.Nextblockhash))
		}
		utils.LogZero.Info().Msgf("交易数量 %d", len(bh.Tx))

		txs := make([]string, 0)
		for _, one := range bh.Tx {
			txs = append(txs, hex.EncodeToString(one))
		}
		bhvo := rpc.BlockHeadVO{
			Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
			Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
			GroupHeight:       bh.GroupHeight,                           //矿工组高度
			GroupHeightGrowth: bh.GroupHeightGrowth,                     //
			Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
			Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
			NTx:               bh.NTx,                                   //交易数量
			MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
			Tx:                txs,                                      //本区块包含的交易id
			Time:              bh.Time,                                  //出块时间，unix时间戳
			Witness:           bh.Witness.B58String(),                   //此块见证人地址
			Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		}
		*bs, _ = json.Marshal(bhvo)
		utils.LogZero.Info().Msgf(string(*bs))

		//计算是否跳过了组
		intervalGroup := bhvo.GroupHeight - beforGroupHeight
		if intervalGroup > 1 {
			utils.LogZero.Warn().Msgf("跳过了组高度 %d", intervalGroup)
		}
		beforGroupHeight = bhvo.GroupHeight

		if beforBlockHeight+1 != bhvo.Height {
			utils.LogZero.Info().Msgf("高度不连续，前高度:%d :%d", beforBlockHeight, bhvo.Height)
			panic("高度不连续")
		}
		beforBlockHeight = bhvo.Height

		// for _, one := range bh.Nextblockhash {
		// 	fmt.Println("下一个块hash", hex.EncodeToString(one))
		// }

		for _, one := range bh.Tx {
			txhashkey := config.BuildBlockTx(one)
			tx, err := db.LevelDB.Find(txhashkey)
			if err != nil {
				utils.LogZero.Info().Msgf("查询第 %d 个块的交易错误"+err.Error(), bh.Height)
				panic("error:查询交易错误")
				return
			}
			txBase, err := mining.ParseTxBaseProto(mining.ParseTxClass(one), tx)
			if err != nil {
				utils.LogZero.Info().Msgf("解析第 %d 个块的交易错误"+err.Error(), bh.Height)
				// fmt.Println("解析第", bh.Height, "个块的交易错误", err)
				panic("error:解析交易错误")
				return
			}

			txid := txBase.GetHash()
			//				if txBase.Class() == config.Wallet_tx_type_deposit_in {
			//					deposit := txBase.(*mining.Tx_deposit_in)
			//					txid = deposit.Hash
			//				}
			utils.LogZero.Info().Msgf("交易id " + string(hex.EncodeToString(*txid)))
			if len(*txBase.GetVin()) > 0 {
				utils.LogZero.Info().Msgf("交易 nonce:%+v ", (*txBase.GetVin())[0])
			}

			itr := txBase.GetVOJSON()
			bs, _ := json.Marshal(itr)
			utils.LogZero.Info().Msgf(string(bs))

			//如果是区块奖励，则计算奖励总和
			if txBase.Class() == config.Wallet_tx_type_mining {
				rewardTotal := uint64(0)
				for _, one := range *txBase.GetVout() {
					rewardTotal += one.Value
				}
				utils.LogZero.Info().Msgf("区块奖励 %d", rewardTotal)
			}

			// switch txBase.Class() {
			// case config.Wallet_tx_type_vote_in:
			// 	tx := txBase.(*mining.Tx_vote_in)
			// 	utils.LogZero.Info().Msgf("%d %s %s", tx.VoteType, hex.EncodeToString((*tx.GetVout())[0].Address), tx.Vote.B58String())
			// case config.Wallet_tx_type_vote_out:
			// 	tx := txBase.(*mining.Tx_vote_out)
			// 	utils.LogZero.Info().Msgf("%s", hex.EncodeToString((*tx.GetVin())[0].Txid))
			// }
		}

		utils.LogZero.Info().Msgf("下一个块hash %s \n", hex.EncodeToString(bh.Nextblockhash))

		if bh.Nextblockhash != nil {
			beforBlockHash = &bh.Nextblockhash
		} else {
			beforBlockHash = nil
		}
	}
}

package full_node

import (
	"bytes"
	"encoding/hex"
	"github.com/craftto/go-tron/pkg/proto/api"
	"github.com/craftto/go-tron/pkg/proto/core"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"math/big"
	"time"
	usdtconfig "web3_gui/cross_chain/usdt/config"
	"web3_gui/cross_chain/usdt/db"
	"web3_gui/cross_chain/usdt/models"
	tronconfig "web3_gui/cross_chain/usdt/tron/config"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/utils"
)

var scanBlockHeight = uint64(0)

func ScanTron() utils.ERROR {
	var ERR utils.ERROR
	//查询数据库，重启之前已经统计的高度
	scanBlockHeight, ERR = db.GetScanBlockHeight(coin_address.COINTYPE_TRX)
	if ERR.CheckFail() {
		return ERR
	}
	if scanBlockHeight == 0 {
		scanBlockHeight = usdtconfig.SCAN_BLOCK_HEIGHT_trx
	}
	ERR = BuildRootAddress()
	if ERR.CheckFail() {
		return ERR
	}
	go loopScanTron()
	go loopTriggerSweep()
	return utils.NewErrorSuccess()
}

/*
反复连接节点，直到节点连接成功
*/
func loopScanTron() {
	utils.Log.Info().Str("连接全节点", "").Send()
	for {
		err := ConnFullNode()
		if err != nil {
			ERR := utils.NewErrorBus(usdtconfig.ERROR_CODE_conn_node_fail_trx, err.Error())
			utils.Log.Error().Str("ERROR", ERR.String()).Send()
			time.Sleep(time.Minute)
			continue
		}
		ERR := loopScanChain()
		if ERR.CheckFail() {
			if ERR.Code == usdtconfig.ERROR_CODE_conn_node_fail_trx {
				utils.Log.Error().Str("ERROR", ERR.String()).Send()
				continue
			}
		}
		time.Sleep(time.Minute)
	}
}

/*
循环查询区块高度，有新高度后开始统计
*/
func loopScanChain() utils.ERROR {
	utils.Log.Info().Str("扫描全节点", "").Send()
	for {
		//获取节点区块高度
		height, ERR := GetNodeBlockHeight()
		if ERR.CheckFail() {
			return ERR
		}
		//检查区块确认数量
		ok := usdtconfig.CheckBlockConfirmationNumber(usdtconfig.Block_confirmation_number_trx, scanBlockHeight+1, uint64(height))
		if !ok {
			time.Sleep(time.Second)
			continue
		}
		ERR = GetBlockHeight(int64(scanBlockHeight + 1))
		if ERR.CheckFail() {
			return ERR
		}
	}
}

/*
查询一个区块信息，并统计其中的交易
*/
func GetBlockHeight(countHeight int64) utils.ERROR {
	utils.Log.Info().Int64("根据区块高度查询区块信息", countHeight).Send()
	v, err := rpcClient.GetBlockByLimitNext(int64(countHeight), int64(countHeight+1))
	if err != nil {
		ERR := utils.NewErrorBus(usdtconfig.ERROR_CODE_conn_node_fail_trx, err.Error())
		utils.Log.Error().Str("ERROR", ERR.String()).Send()
		return ERR
	}
	utils.Log.Info().Int64("根据区块高度查询区块信息", countHeight).Send()
	for _, b := range v.Block {
		txs := make([]*models.TxInfo, 0)
		height := b.GetBlockHeader().GetRawData().GetNumber()
		//fmt.Println("获取的此区块高度", height)
		for _, v2 := range b.Transactions {
			txInfos, ERR := CountTransfer(v2)
			if ERR.CheckFail() {
				return ERR
			}
			now := time.Now().Unix()
			for i, _ := range txInfos {
				txInfoOne := txInfos[i]
				txInfoOne.BlockHeight = uint64(height)
				txInfoOne.BlockHash = hex.EncodeToString(b.Blockid)
				txInfoOne.TimeBlock = b.GetBlockHeader().GetRawData().Timestamp * 1000
				txInfoOne.TimeStamp = now
				txInfoOne.Confirm = usdtconfig.Block_confirmation_number_trx
			}
			txs = append(txs, txInfos...)
		}
		ERR := SaveTxInfos(txs)
		if ERR.CheckFail() {
			return ERR
		}
		scanBlockHeight = uint64(height)
	}
	return utils.NewErrorSuccess()
}

/*
解析一个交易
*/
func CountTransfer(tx *api.TransactionExtention) ([]*models.TxInfo, utils.ERROR) {
	txInfos := make([]*models.TxInfo, 0)
	for _, v1 := range tx.Transaction.RawData.Contract {
		if v1.Type == core.Transaction_Contract_TransferContract { //转账合约
			// trx 转账
		} else if v1.Type == core.Transaction_Contract_TriggerSmartContract { //调用智能合约
			// trc20 转账
			unObj := &core.TriggerSmartContract{}
			err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
			if err != nil {
				continue
			}
			contract := tronconfig.EncodeCheck(unObj.GetContractAddress())
			from := tronconfig.EncodeCheck(unObj.GetOwnerAddress())
			objData := unObj.GetData()
			// unObj.Data  https://goethereumbook.org/en/transfer-tokens/ 参考eth 操作
			to, amount, flag := processTransferData(objData)
			txidStr := hex.EncodeToString(tx.Txid)
			txInfoTrx, err := rpcClient.GetTransactionInfoByID(txidStr)
			if err != nil {
				return nil, utils.NewErrorSysSelf(err)
			}
			status := "failed"
			if txInfoTrx.Result == core.TransactionInfo_SUCESS {
				status = "success"
			} else {
				status = "failed"
			}
			//fmt.Printf("转账USDT1:区块高度:%d 区块hash:%s 链上时间:%d 交易hash:%s 合约地址:%s from:%s to:%s 手续费:%s 转账金额:%s 状态:%s\n",
			//	height, hex.EncodeToString(b.Blockid), b.GetBlockHeader().GetRawData().Timestamp*1000,
			//	hex.EncodeToString(tx.Txid), contract, from, to, decimal.New(txInfo.Fee, 0).Div(decimal.New(1, 6)).String(),
			//	decimal.New(amount, 0).Div(decimal.New(1, 6)).String(), status)
			if flag { // 只有调用了 transfer(address,uint256) 才是转账
				txInfo := models.TxInfo{
					ChainName:       "trx",
					TxId:            txidStr,
					Status:          status,
					From:            from,
					To:              to,
					ContractAddress: contract,
					Fee:             txInfoTrx.Fee,
					Amount:          amount,
					TokenName:       "TRX",
					//Send:            false,
					Type: usdtconfig.ContractTransfer,
				}
				txInfos = append(txInfos, &txInfo)
			}
		} else {
			//合约调用
		}
	}
	return txInfos, utils.NewErrorSuccess()
}

func processTransferData(trc20 []byte) (to string, amount int64, flag bool) {
	if len(trc20) < 68 {
		return
	}
	if !bytes.Equal(trc20[:4], usdtconfig.Trc20PreBytes) {
		return
	}
	add, _ := hex.DecodeString("41")
	add = append(add, common.TrimLeftZeroes(trc20[4:36])...)
	to = tronconfig.EncodeCheck(add)
	if string(to[0]) != "T" {
		to = tronconfig.EncodeCheck(common.TrimLeftZeroes(trc20[4:36]))
	}
	amount = new(big.Int).SetBytes(common.TrimLeftZeroes(trc20[36:68])).Int64()
	flag = true
	return
}

/*
保存交易信息
*/
func SaveTxInfos(txInfos []*models.TxInfo) utils.ERROR {
	utils.Log.Info().Str("保存交易信息", "").Send()
	for _, one := range txInfos {
		utils.Log.Info().Msgf("保存交易信息:%+v", one)
	}
	return db.SaveTxInfo(txInfos)
}

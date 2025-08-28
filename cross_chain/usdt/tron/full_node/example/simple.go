package main

import (
	"encoding/hex"
	"fmt"
	"github.com/craftto/go-tron/pkg/client"
	"github.com/craftto/go-tron/pkg/proto/core"
	"github.com/craftto/go-tron/pkg/trc20"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"math/big"
	"time"
	usdtconfig "web3_gui/cross_chain/usdt/config"
	tronconfig "web3_gui/cross_chain/usdt/tron/config"
	"web3_gui/utils"
)

func main() {
	height := GetBlockHeight()
	GetBalance("TFjvhpeTGfDcFBpmgBv7aAwBteZ1uAiYjP")
	GetContractBalance(usdtconfig.TrxUsdtAddr, "TKYjUfYzBaNgfZcLZFNmCTgQPFtWARKFdi")

	GetBlockHeightRange(height, height+1)
}

// 全节点列表
// var TronGrpcMainNetWorkAddrs = []string{"127.0.0.1:50051", "127.0.0.1:50051"}
var TronGrpcMainNetWorkAddrs = usdtconfig.NODE_rpc_addr_trx

func GetBlockHeight() int64 {
	//currentUrl := ""
	height := int64(0)
	for _, tronAddr := range TronGrpcMainNetWorkAddrs {
		c1, err := client.NewGrpcClient(tronAddr,
			grpc.WithInsecure(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                time.Minute,
				PermitWithoutStream: true,
			}),
		)
		if err != nil {
			continue
		}
		nowBlock, err := c1.GetNowBlock()
		if err != nil {
			c1.Close()
			continue
		}
		if nowBlock.BlockHeader.RawData.Number > height {
			height = nowBlock.BlockHeader.RawData.Number
			//currentUrl = tronAddr
		}
		c1.Close()
	}
	fmt.Println("区块最新高度", height)
	return height
}

func GetBalance(address string) error {
	c1, err := client.NewGrpcClient(TronGrpcMainNetWorkAddrs[0],
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Minute,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return err
	}
	res, err := c1.GetAccount(address)
	if err != nil {
		c1.Close()
		return err
	}
	fmt.Println("地址余额", decimal.New(res.Balance, 0).Div(decimal.New(1, 6)).String())
	return err
}

func GetContractBalance(contractAddress, address string) error {
	c1, err := client.NewGrpcClient(TronGrpcMainNetWorkAddrs[0],
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Minute,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return err
	}

	t, err := trc20.NewTrc20(c1, contractAddress)
	if err != nil {
		utils.Log.Error().Msgf("GetTrc20Amount NewTrc20 addr:%s err:%v", address, err)
		return err
	}
	amount, err := t.GetBalance(address)
	if err != nil {
		return err
	}
	fmt.Println("地址余额", decimal.NewFromBigInt(amount, 0).Div(decimal.New(1, 6)).String())
	return err
}

/*
获取一个高度范围的区块信息
*/
func GetBlockHeightRange(startHeight, endHeight int64) error {
	c1, err := client.NewGrpcClient(TronGrpcMainNetWorkAddrs[0],
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Minute,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return err
	}
	v, err := c1.GetBlockByLimitNext(int64(startHeight), int64(endHeight))
	if err != nil {
		//engine.Log.Error("GetBlockByLimitNext", err, scanHeight, toHeight+1)
		return err
	}

	for _, b := range v.Block {
		height := b.GetBlockHeader().GetRawData().GetNumber()
		fmt.Println("获取的此区块高度", height)

		for _, v2 := range b.Transactions {
			for _, v1 := range v2.Transaction.RawData.Contract {
				if v1.Type == core.Transaction_Contract_TransferContract { //转账合约
					// trx 转账
					unObj := &core.TransferContract{}
					err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
					if err != nil {
						continue
					}
					from := tronconfig.EncodeCheck(unObj.GetOwnerAddress())
					to := tronconfig.EncodeCheck(unObj.GetToAddress())

					txInfo, err := c1.GetTransactionInfoByID(hex.EncodeToString(v2.Txid))
					if err != nil {
						return err
					}

					status := "failed"
					if txInfo.Result == core.TransactionInfo_SUCESS {
						status = "success"
					} else {
						status = "failed"
					}

					fmt.Printf("转账TRX:区块高度:%d 区块hash:%s 链上时间:%d 交易hash:%s from:%s to:%s 手续费:%s 转账金额:%s 状态:%s\n",
						height, hex.EncodeToString(b.Blockid), b.GetBlockHeader().GetRawData().Timestamp*1000,
						hex.EncodeToString(v2.Txid), from, to, decimal.New(txInfo.Fee, 0).Div(decimal.New(1, 6)).String(),
						decimal.New(unObj.Amount, 0).Div(decimal.New(1, 6)).String(), status)

					// break
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

					txInfo, err := c1.GetTransactionInfoByID(hex.EncodeToString(v2.Txid))
					if err != nil {
						return err
					}
					status := "failed"
					if txInfo.Result == core.TransactionInfo_SUCESS {
						status = "success"
					} else {
						status = "failed"
					}

					fmt.Printf("转账USDT1:区块高度:%d 区块hash:%s 链上时间:%d 交易hash:%s 合约地址:%s from:%s to:%s 手续费:%s 转账金额:%s 状态:%s\n",
						height, hex.EncodeToString(b.Blockid), b.GetBlockHeader().GetRawData().Timestamp*1000,
						hex.EncodeToString(v2.Txid), contract, from, to, decimal.New(txInfo.Fee, 0).Div(decimal.New(1, 6)).String(),
						decimal.New(amount, 0).Div(decimal.New(1, 6)).String(), status)

					if flag { // 只有调用了 transfer(address,uint256) 才是转账
						fmt.Printf("转账USDT2:区块高度:%d 区块hash:%s 链上时间:%d 交易hash:%s 合约地址:%s from:%s to:%s 手续费:%s 转账金额:%s 状态:%s\n",
							height, hex.EncodeToString(b.Blockid), b.GetBlockHeader().GetRawData().Timestamp*1000,
							hex.EncodeToString(v2.Txid), contract, from, to, decimal.New(txInfo.Fee, 0).Div(decimal.New(1, 6)).String(),
							decimal.New(amount, 0).Div(decimal.New(1, 6)).String(), status)

					}

					// break
				} else {
					unObj := &core.TransferContract{}
					err := proto.Unmarshal(v1.Parameter.GetValue(), unObj)
					if err != nil {
						continue
					}
					from := tronconfig.EncodeCheck(unObj.GetOwnerAddress())
					to := tronconfig.EncodeCheck(unObj.GetToAddress())

					txInfo, err := c1.GetTransactionInfoByID(hex.EncodeToString(v2.Txid))
					if err != nil {
						return err
					}
					status := "failed"
					if txInfo.Result == core.TransactionInfo_SUCESS {
						status = "success"
					} else {
						status = "failed"
					}

					fmt.Printf("合约:区块高度:%d 区块hash:%s 链上时间:%d 交易hash:%s from:%s to:%s 手续费:%s 转账金额:%s 状态:%s\n",
						height, hex.EncodeToString(b.Blockid), b.GetBlockHeader().GetRawData().Timestamp*1000,
						hex.EncodeToString(v2.Txid), from, to, decimal.New(txInfo.Fee, 0).Div(decimal.New(1, 6)).String(),
						decimal.New(unObj.Amount, 0).Div(decimal.New(1, 6)).String(), status)

				}

			}

		}
	}
	return nil
}

func processTransferData(trc20 []byte) (to string, amount int64, flag bool) {
	if len(trc20) >= 68 {
		if hex.EncodeToString(trc20[:4]) != "a9059cbb" {
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
	}
	return
}

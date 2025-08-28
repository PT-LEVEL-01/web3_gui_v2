package mining

import (
	"web3_gui/chain/config"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

var checkSwapTxQueue = make(chan *TxSwap, 100000)

func init() {
	go startCheckSwapQueue()
}

func startCheckSwapQueue() {
	utils.Go(func() {
		NumCPUTokenChan := make(chan bool, config.CPUNUM)
		for txItr := range checkSwapTxQueue {
			NumCPUTokenChan <- false
			go checkMulticastSwapTransaction(txItr, NumCPUTokenChan)
		}
	}, nil)
}

/*
验证广播消息
*/
func checkMulticastSwapTransaction(txbase *TxSwap, tokenCPU chan bool) {
	defer func() {
		<-tokenCPU
	}()
	//验证交易
	if err := txbase.CheckLockHeight(GetLongChain().GetCurrentBlock()); err != nil {
		engine.Log.Error("Failed to verify transaction lock height")
		return
	}

	forks.GetLongChain().transactionSwapManager.AddTx(txbase, Swap_Step_Promoter)
}

package mining

import (
	"sync"

	"web3_gui/chain/config"
	"web3_gui/chain/rpc/limiter"
)

var limiter_handletx = "handleTxLimit"

var txLimitLock = sync.Mutex{}

// 添加Gas限流
func AddTxLimit(tx TxItr) bool {
	txLimitLock.Lock()
	defer txLimitLock.Unlock()

	//排除不限流交易类型
	if checkReward(tx) {
		return true
	}

	// 合约交易消耗更多的令牌=手续费+额外token
	gasUsed := tx.GetGasUsed()
	if tx.Class() == config.Wallet_tx_type_deposit_in ||
		tx.Class() == config.Wallet_tx_type_deposit_out ||
		tx.Class() == config.Wallet_tx_type_contract {
		gasUsed += limiter.ContractExtraUsed
	}

	return limiter.RpcReqLimiter.AllowN(limiter_handletx, gasUsed) //添加令牌
}

// 减少交易Gas限流
func DelTxLimit(tx TxItr) {
	txLimitLock.Lock()
	defer txLimitLock.Unlock()

	if checkReward(tx) {
		return
	}

	// 合约交易回收更多的令牌=手续费+额外token
	gasUsed := tx.GetGasUsed()
	if tx.Class() == config.Wallet_tx_type_deposit_in ||
		tx.Class() == config.Wallet_tx_type_deposit_out ||
		tx.Class() == config.Wallet_tx_type_contract {
		gasUsed += limiter.ContractExtraUsed
	}

	//engine.Log.Error("减少limit:%d   交易class:%d  交易hash：%s", calcTxToken(tx), tx.Class(), hex.EncodeToString(*tx.GetHash()))
	limiter.RpcReqLimiter.RecoverN(limiter_handletx, gasUsed) //回收令牌
}

func DelCacheTxAndLimit(key string, tx TxItr) {
	DelCacheTx(key)
	DelTxLimit(tx)
}

func checkReward(tx TxItr) bool {
	return tx.Class() == config.Wallet_tx_type_mining
}

func calcTxToken(tx TxItr) uint64 {
	var token uint64

	switch tx.Class() {
	case config.Wallet_tx_type_contract:
		token = tx.(*Tx_Contract).GetGasUsed()
		break
	default:
		token = config.DefaultTxToken
	}

	return token
}

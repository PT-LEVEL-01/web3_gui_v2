package mining

import (
	"encoding/hex"
	"github.com/pkg/errors"
	"math/big"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/mining/cache"
	"web3_gui/chain/rpc/limiter"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

var addtxlock sync.Mutex

func (tm *TransactionManager) AddTx(txItr TxItr) error {
	//if (txItr.Class() == config.Wallet_tx_type_light_in) ||
	//	(txItr.Class() == config.Wallet_tx_type_community_in) {
	//	return nil
	//}
	utils.Log.Info().Hex("TransactionManager AddTx:", *txItr.GetHash()).Send()

	isAddTxSign := false
	defer func() {
		//交易失败，释放nonce占用
		if !isAddTxSign {
			tm.witnessBackup.chain.Balance.DelLockTx(txItr)
			engine.Log.Warn("TransactionManager AddTx:%s failed", hex.EncodeToString(*txItr.GetHash()))
		}
	}()

	if !tm.CheckEnableAddTx(txItr) {
		engine.Log.Error("TransactionManager Disable AddTx:%s ", hex.EncodeToString(*txItr.GetHash()))
		return errors.New("DisableTx")
	}

	if CheckOutOfMemory() {
		engine.Log.Error("TransactionManager AddTx:%s, CheckOutOfMemory: %d", hex.EncodeToString(*txItr.GetHash()), config.Wallet_Memory_percentage_max)
		return errors.New("out of memory")
	}

	//交易限速
	config.GetTxRate(txItr.Class(), true)

	//验证GasUsed
	gasUsed, err := GetTxGasUsed(txItr)
	if err != nil {
		engine.Log.Error("TransactionManager AddTx:%s, check GasUsed Error: %s", hex.EncodeToString(*txItr.GetHash()), err.Error())
		return errors.Wrap(err, "check gas used")
	}
	txItr.SetGasUsed(gasUsed)

	//Gas限流
	if !AddTxLimit(txItr) {
		//engine.Log.Error("当前限流器 使用：%d 剩余：%d 交易gasUsed:%d", limiter.RpcReqLimiter.Len(limiter_handletx), limiter.RpcReqLimiter.Surplus(limiter_handletx), GetTxGasUsed(txItr))
		//engine.Log.Error("当前限流器 使用：%d 剩余：%d 交易gasUsed:%d", limiter.RpcReqLimiter.Len(limiter_handletx), limiter.RpcReqLimiter.Surplus(limiter_handletx), gasUsed)
		engine.Log.Error("TransactionManager AddTx:%s, 当前限流器 使用：%d 剩余：%d 交易gasUsed:%d", hex.EncodeToString(*txItr.GetHash()), limiter.RpcReqLimiter.Len(limiter_handletx), limiter.RpcReqLimiter.Surplus(limiter_handletx), gasUsed)
		return errors.New("tx limited")
	}

	//验证签名
	err = txItr.CheckSign()
	if err != nil {
		//engine.Log.Warn("Failed to verify transaction signature %s %s", hex.EncodeToString(*txItr.GetHash()), err.Error())
		engine.Log.Error("TransactionManager AddTx:%s, Failed to verify transaction signature %s", hex.EncodeToString(*txItr.GetHash()), err.Error())
		DelTxLimit(txItr)
		return errors.Wrap(err, "sign")
	}

	//检查地址冻结
	if txItr.CheckAddressFrozen() {
		engine.Log.Error("TransactionManager AddTx:%s, address is frozen", hex.EncodeToString(*txItr.GetHash()))
		return errors.New("address is frozen")
	}

	//检查地址绑定状态
	if err = txItr.CheckAddressBind(); err != nil {
		engine.Log.Error("TransactionManager AddTx:%s, CheckAddressBind", hex.EncodeToString(*txItr.GetHash()))
		return err
	}

	if !TxCheckCond.WaitCheck(tm.witnessBackup.chain, txItr) {
		DelTxLimit(txItr)
		//engine.Log.Error("AddTx TxCheckCond.WaitCheck,check cache is exist 52")
		engine.Log.Error("TransactionManager AddTx:%s, TxCheckCond.WaitCheck,check cache is exist 52", hex.EncodeToString(*txItr.GetHash()))
		return errors.New("check tx existed")
	}

	addtxlock.Lock()
	defer addtxlock.Unlock()

	//engine.Log.Error("添加limit:%d   交易class:%d 交易hash：%s", calcTxToken(txItr), txItr.Class(), hex.EncodeToString(*txItr.GetHash()))

	txhashkey := config.BuildBlockTx(*txItr.GetHash())
	hk := utils.Bytes2string(txhashkey)

	if cache.TransCache.Exists(hk) {
		DelTxLimit(txItr)
		//engine.Log.Error("AddTx TxCheckCond.Exists,check cache is exist 65")
		engine.Log.Error("TransactionManager AddTx:%s, TxCheckCond.WaitCheck,check cache is exist 65", hex.EncodeToString(*txItr.GetHash()))
		return errors.New("cache tx existed")
	}

	bs, err := txItr.Proto()
	if err != nil {
		//engine.Log.Error("tx proto fail:%s", err.Error())
		engine.Log.Error("TransactionManager AddTx:%s, tx proto fail:%s", hex.EncodeToString(*txItr.GetHash()), err.Error())
		DelTxLimit(txItr)
		return errors.Wrap(err, "proto")
	}

	isAddTxSign = true

	//存入缓存
	ok := cache.TransCache.Set(hk, *bs)
	if ok {
		engine.Log.Warn("TransactionManager AddTx:%s, cache set existed", hex.EncodeToString(*txItr.GetHash()))
	}

	//engine.Log.Info("tx：%s，存入缓存", hex.EncodeToString(h))

	tm.AddTxSignal(txItr)
	return nil
}

// 检查是否允许交易类型
func (tm *TransactionManager) CheckEnableAddTx(txItr TxItr) bool {
	if config.DisableCommunityTx {
		if txItr.Class() == config.Wallet_tx_type_vote_in { //质押社区
			if tx, ok := txItr.(*Tx_vote_in); ok {
				if tx.VoteType == VOTE_TYPE_community {
					return false
				}
			}
		}

		if txItr.Class() == config.Wallet_tx_type_vote_out { //取消质押社区
			if tx, ok := txItr.(*Tx_vote_out); ok {
				if tx.VoteType == VOTE_TYPE_community {
					return false
				}
			}
		}
	}

	if config.DisableLightTx {
		if txItr.Class() == config.Wallet_tx_type_vote_in { //质押轻节点/投票
			if tx, ok := txItr.(*Tx_vote_in); ok {
				if tx.VoteType == VOTE_TYPE_light || tx.VoteType == VOTE_TYPE_vote {
					return false
				}
			}
		}

		if txItr.Class() == config.Wallet_tx_type_vote_out { //取消质押轻节点/取消投票
			if tx, ok := txItr.(*Tx_vote_out); ok {
				if tx.VoteType == VOTE_TYPE_light || tx.VoteType == VOTE_TYPE_vote {
					return false
				}
			}
		}
	}
	return true
}

func (tm *TransactionManager) AddTxSignal(txItr TxItr) bool {
	spend := txItr.GetSpend()
	txbs := txItr.Serialize()

	div1 := new(big.Int).Mul(big.NewInt(int64(txItr.GetGas())), big.NewInt(100000000))
	div2 := big.NewInt(int64(len(*txbs)))

	ratioValue := new(big.Int).Div(div1, div2)
	ratio := TransactionRatio{
		tx:      txItr,              //交易
		size:    uint64(len(*txbs)), //交易总大小
		gas:     txItr.GetGas(),     //手续费
		Ratio:   ratioValue,         //价值比
		gasUsed: txItr.GetGasUsed(),
	}
	//if txItr.Class() == config.Wallet_tx_type_voting_reward {
	//	ratio.spendLock = spend
	//} else {
	//	ratio.spend = spend
	//}
	ratio.spend = spend

	tm.tempTxLock.Lock()
	tm.tempTx = append(tm.tempTx, ratio)
	tm.tempTxLock.Unlock()
	select {
	case tm.tempTxsignal <- false:
	default:
	}

	return true
}

func DelCacheTx(key string) {
	addtxlock.Lock()
	defer addtxlock.Unlock()

	cache.TransCache.GetAndDel(key)
}

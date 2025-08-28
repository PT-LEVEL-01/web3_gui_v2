package mining

import (
	"sync"

	"github.com/shirou/gopsutil/v3/mem"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
)

type TxCheck struct {
	*sync.Cond
	IsImport bool
}

var TxCheckCond = &TxCheck{sync.NewCond(&sync.Mutex{}), false}

func (t *TxCheck) WaitCheck(chain *Chain, tx TxItr) bool {
	t.L.Lock()
	defer t.L.Unlock()

	if t.IsImport {
		t.Wait()
	}

	if t.checkTxExistsForCache(chain, tx) || t.checkTxExistsForDb(tx) {
		return false
	}

	return true
}

func (t *TxCheck) SetImportTag() {
	t.IsImport = true
}

func (t *TxCheck) ResetImportTag() {
	t.broadcastAllowCheck()
}

func (t *TxCheck) broadcastAllowCheck() {
	t.L.Lock()
	defer t.L.Unlock()

	t.IsImport = false
	t.Broadcast()
}

func (t *TxCheck) checkTxExistsForCache(chain *Chain, tx TxItr) bool {
	return chain.TransactionManager.unpackedTransaction.ExistTxByAddrTxid(tx)
}

func (t *TxCheck) checkTxExistsForDb(tx TxItr) bool {
	hs := tx.GetHash()
	txhashkey := config.BuildBlockTx(*hs)
	exist, _ := db.LevelDB.CheckHashExist(txhashkey)
	notImport, _ := db.LevelDB.CheckHashExist(config.BuildTxNotImport(*hs))

	return exist && !notImport
}

/*
*
检查内存是否满了
*/
func CheckOutOfMemory() bool {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent > config.Wallet_Memory_percentage_max
}

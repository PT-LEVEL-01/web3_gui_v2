package mining

import (
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"sync"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/chain/mining/snapshot"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/utils"
)

const (
	AccountSnapshotName       = "snapshot_account"
	AccountSnapshotbalanceKey = "snapshot_account_balance_"
	AccountSnapshotNonceKey   = "snapshot_account_nonce_"
)

var (
	AccountSnap         *AccountSnapshot
	accountSnapshotLock = sync.RWMutex{}
)

const (
	AccountSnap_Balance = iota
	AccountSnap_Nonce
)

type AccountSnapshot go_protos.AccountSnapshot

// 初始化account snapshot
func InitAccountSnap() {
	AccountSnap = &AccountSnapshot{
		Balance: make(map[string]uint64),
		Nonce:   make(map[string]uint64),
	}

	snapshot.Add(AccountSnap)
}

func (a *AccountSnapshot) SnapshotName() string {
	return AccountSnapshotName
}

// 更新当前快照
func (s *AccountSnapshot) Update(t int, vals map[string]uint64) error {
	accountSnapshotLock.Lock()
	defer accountSnapshotLock.Unlock()

	switch t {
	case AccountSnap_Balance:
		for k, v := range vals {
			s.Balance[k] = v
		}
	case AccountSnap_Nonce:
		for k, v := range vals {
			s.Nonce[k] = v
		}
	}

	return nil
}

func (a *AccountSnapshot) SnapshotSerialize() ([]byte, error) {
	st := go_protos.AccountSnapshot(*a)
	return st.Marshal()
}

func (a *AccountSnapshot) SnapshotDeSerialize(bs []byte) error {
	t := go_protos.AccountSnapshot(*a)
	err := t.Unmarshal(bs)
	if err != nil {
		return err
	}

	st := AccountSnapshot(t)
	st.reset()

	if err = a.restoreBalance(); err != nil {
		return err
	}

	if err = a.restoreNonce(); err != nil {
		return err
	}

	return nil
}

// 快照扩展存储
func (s *AccountSnapshot) SnapshotExtStorage() ([]db2.KVPair, error) {
	pairs := make([]db2.KVPair, 0)
	for k, v := range s.Balance {
		pairs = append(pairs, db2.KVPair{
			Key:   append([]byte(AccountSnapshotbalanceKey), []byte(k)...),
			Value: utils.Uint64ToBytes(v),
		})
	}

	for k, v := range s.Nonce {
		pairs = append(pairs, db2.KVPair{
			Key:   append([]byte(AccountSnapshotNonceKey), []byte(k)...),
			Value: utils.Uint64ToBytes(v),
		})
	}

	return pairs, nil
}

// 快照保存回调
func (s *AccountSnapshot) SnapshotCallback(isSave bool) {
	accountSnapshotLock.Lock()
	defer accountSnapshotLock.Unlock()

	if !isSave {
		return
	}

	s.reset()
}

func (s *AccountSnapshot) reset() {
	s.Balance = make(map[string]uint64)
	s.Nonce = make(map[string]uint64)
}

// 从快照db中加载地址信息
func (a *AccountSnapshot) restoreBalance() error {
	balances := make(map[string]uint64)
	r := util.BytesPrefix([]byte(AccountSnapshotbalanceKey))
	it := db.LevelDB.NewIteratorWithOption(r, nil)
	for it.Next() {
		key := it.Key()
		onekey := key[len([]byte(AccountSnapshotbalanceKey))+2:]
		balances[utils.Bytes2string(onekey)] = utils.BytesToUint64(it.Value())
	}
	it.Release()
	err := it.Error()
	if err != nil {
		return err
	}

	a.setBalance(balances)
	return nil
}

func (a *AccountSnapshot) setBalance(balances map[string]uint64) {
	db.SetNotSpendBalances(balances)
}

// 从快照db中加载地址信息
func (a *AccountSnapshot) restoreNonce() error {
	nonces := make(map[string]big.Int)
	r := util.BytesPrefix([]byte(AccountSnapshotNonceKey))
	it := db.LevelDB.NewIteratorWithOption(r, nil)
	for it.Next() {
		key := it.Key()
		onekey := key[len([]byte(AccountSnapshotNonceKey))+2:]
		nonces[utils.Bytes2string(onekey)] = *new(big.Int).SetBytes(it.Value())
	}
	it.Release()
	err := it.Error()
	if err != nil {
		return err
	}

	a.setNonce(nonces)
	return nil
}

func (a *AccountSnapshot) setNonce(nonces map[string]big.Int) {
	//SetNonces(nonces)
}

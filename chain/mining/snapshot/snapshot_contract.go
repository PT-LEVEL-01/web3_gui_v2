package snapshot

import (
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
)

const (
	ContractSnapshotName = "snapshot_contract"
	ContractSnapshotKey  = "snapshot_contract_"
	MultiLoadCount       = 100
)

var (
	ContractSnap         *ContractSnapshot
	contractSnapshotLock = sync.RWMutex{}
)

type ContractSnapshot go_protos.ContractSnapshot

// 初始化contract snapshot
func InitContractSnap() {
	origin := &go_protos.ContractSnapshotInfo{
		ContractAccountData: make(map[string][]byte),
		DestructSet:         make(map[string]bool),
	}

	current := &go_protos.ContractSnapshotInfo{
		ContractAccountData: make(map[string][]byte),
		DestructSet:         make(map[string]bool),
	}

	ContractSnap = &ContractSnapshot{
		Origin:    origin,
		Current:   current,
		Contracts: make(map[string]bool),
	}

	Add(ContractSnap)
}

// 更新当前快照
func (s *ContractSnapshot) Update(contractAddr crypto.AddressCoin, obj *go_protos.StateObject, height uint64) error {
	contractSnapshotLock.Lock()
	defer contractSnapshotLock.Unlock()

	addrStr := contractAddr.B58String()

	//合约销毁则删除快照记录
	if len(obj.Code) == 0 || obj.Suicide {
		s.Current.DestructSet[addrStr] = true
		delete(s.Current.ContractAccountData, addrStr)
		return nil
	}

	v, err := proto.Marshal(obj)
	if err != nil {
		engine.Log.Error("序列化合约对象失败,%s", err.Error())
		return err
	}

	s.Current.ContractAccountData[addrStr] = v
	s.Current.BlockHeight = height

	//engine.Log.Info("源快照：%d，当前快照:%d", len(s.Origin.ContractAccountList), len(s.Current.ContractAccountList))
	return nil
}

// 合并快照到源
func (s *ContractSnapshot) merge() {
	//当前快照标记了删除的合约，则也删除源快照里的记录
	for addr := range s.Current.DestructSet {
		delete(s.Origin.ContractAccountData, addr)
		s.Origin.DestructSet[addr] = true
	}

	//更新当前快照到源快照中
	for k, v := range s.Current.ContractAccountData {
		s.Origin.ContractAccountData[k] = v
	}

	s.Origin.BlockHeight = s.Current.BlockHeight
}

// 构建快照
func (s *ContractSnapshot) BuildSnap() error {
	contractSnapshotLock.Lock()
	defer contractSnapshotLock.Unlock()

	s.merge()
	s.resetCurrent()

	return nil
}

// 快照扩展存储
func (s *ContractSnapshot) SnapshotExtStorage() ([]db2.KVPair, error) {
	if err := s.BuildSnap(); err != nil {
		return nil, err
	}

	contractPairs := make([]db2.KVPair, 0)
	for k, v := range s.Origin.ContractAccountData {
		key := crypto.AddressFromB58String(k)
		contractPairs = append(contractPairs, db2.KVPair{
			Key:   append([]byte(ContractSnapshotKey), key...),
			Value: v,
		})
	}
	return contractPairs, nil
}

// 快照保存回调
func (s *ContractSnapshot) SnapshotCallback(isSave bool) {
	contractSnapshotLock.Lock()
	defer contractSnapshotLock.Unlock()

	if !isSave {
		return
	}
	engine.Log.Info("存储虚拟机合约快照，区块高度：%d，源快照合约数：%d，当前快照合约数：%d", s.BlockHeight, len(s.Origin.ContractAccountData), len(s.Current.ContractAccountData))

	s.BlockHeight = s.Origin.BlockHeight

	//如果有删除的合约，则从磁盘删除
	if len(s.Origin.DestructSet) > 0 {
		keys := make([][]byte, 0)
		for k := range s.Origin.DestructSet {
			keys = append(keys, append([]byte(ContractSnapshotKey), crypto.AddressFromB58String(k)...))
		}
		db.LevelDB.Del(keys...)
	}

	s.resetOrigin()
}

// 重置当前快照
func (s *ContractSnapshot) resetCurrent() {
	s.Current = &go_protos.ContractSnapshotInfo{
		BlockHeight:         0,
		ContractAccountData: make(map[string][]byte),
		DestructSet:         make(map[string]bool),
	}
}

// 重置源快照
func (s *ContractSnapshot) resetOrigin() {
	s.Origin = &go_protos.ContractSnapshotInfo{
		BlockHeight:         0,
		ContractAccountData: make(map[string][]byte),
		DestructSet:         make(map[string]bool),
	}
}

func (s *ContractSnapshot) SnapshotName() string {
	return ContractSnapshotName
}

func (s *ContractSnapshot) SnapshotSerialize() ([]byte, error) {
	startAt := time.Now()
	st := go_protos.ContractSnapshot(*s)
	b, err := st.Marshal()
	engine.Log.Info(">>>>>> Snapshot[序列化] 合约序列化耗时(ms):%v,字节:%d", time.Now().Sub(startAt), len(b))
	return b, err
}

func (s *ContractSnapshot) SnapshotDeSerialize(bs []byte) error {
	startAt := time.Now()
	t := go_protos.ContractSnapshot(*s)
	err := t.Unmarshal(bs)
	if err != nil {
		return err
	}

	contractSnapshotLock.Lock()
	defer contractSnapshotLock.Unlock()

	st := ContractSnapshot(t)
	st.BlockHeight = st.Origin.BlockHeight
	st.resetOrigin()

	if err = restoreContracts(); err != nil {
		return err
	}

	s = &st
	engine.Log.Info(">>>>>> Snapshot[反序列化] 合约反序列化耗时(ms):%v,字节:%d", time.Now().Sub(startAt), len(bs))
	return nil
}

// 从快照db中加载合约对象
func restoreContracts() error {
	r := util.BytesPrefix([]byte(ContractSnapshotKey))
	it := db.LevelDB.NewIteratorWithOption(r, nil)
	for it.Next() {
		key := it.Key()
		addr := crypto.AddressCoin(key[len([]byte(ContractSnapshotKey))+2:])
		onekey := append([]byte(config.DBKEY_CONTRACT_OBJECT), addr...)
		if err := db.LevelTempDB.Set(onekey, it.Value()); err != nil {
			engine.Log.Error("保存合约账户对象失败: err:%s", err.Error())
			return err
		}
	}
	it.Release()
	err := it.Error()
	return err
}

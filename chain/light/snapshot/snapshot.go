package lightsnapshot

import (
	"errors"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

// 全局快照管理器
var (
	// 快照碎片集
	lightsnapshot = sync.Map{}
	// 快照起始高度
	LightSnapshotStartHeight = uint64(0)
	// 轻节点保证
	//lightSnapshotInternal = uint64(config.Mining_group_max * config.Witness_backup_group_new)
	lightSnapshotInternal = uint64(10)
)

// 快照管理器接口
type LightSnapshotMgr interface {
	LightSnapshotName() string               // 快照碎片对象名称
	LightSnapshotSerialize() ([]byte, error) // 序列化内存对象
	LightSnapshotDeSerialize([]byte) error   // 反序列化,还原内存对象
}

// 添加对象
func Add(o LightSnapshotMgr) error {
	if _, ok := lightsnapshot.LoadOrStore(o.LightSnapshotName(), o); ok {
		return errors.New("light snapshoter name is exist")
	}

	return nil
}

// 保存快照
// 开启事务
// 内存快照
// 增量快照合并
// 清空增量快照
func Save(height uint64) error {
	// 该高度已经快照了
	if Height() >= height || !(height%lightSnapshotInternal == 0) {
		return nil
	}
	startAt := config.TimeNow()

	snapshotPairs := make([]db2.KVPair, 0)

	// 快照高度
	snapshotPairs = append(snapshotPairs, db2.KVPair{
		Key:   config.DBKEY_snapshot_height,
		Value: utils.Uint64ToBytes(height),
	})

	var haserr error
	lightsnapshot.Range(func(key, value any) bool {
		o := value.(LightSnapshotMgr)
		objbytes, err := o.LightSnapshotSerialize()
		if err != nil {
			haserr = err
			return false
		}

		snapshotPairs = append(snapshotPairs, db2.KVPair{
			Key:   append(config.DBKEY_snapshot, []byte(o.LightSnapshotName())...),
			Value: objbytes,
		})
		return true
	})

	if haserr != nil {
		return haserr
	}
	if err := db.LevelTempDB.StoreSnap(snapshotPairs); err != nil {
		return err
	}

	// 成功则清空增量
	db.LevelTempDB.ResetSnap()

	engine.Log.Info(">>>>>> Light Snapshot[保存] 快照成功,高度:%d", height)
	engine.Log.Info(">>>>>> Light Snapshot[诊断] Seq:%d, 内存快照+增量快照合并+清空增量快照耗时:%v", db.LevelDB.GetSeq(), config.TimeNow().Sub(startAt))

	//isSave = true
	return nil
}

// 加载快照到内存
func Load() error {
	var haserr error
	lightsnapshot.Range(func(key, value any) bool {
		o := value.(LightSnapshotMgr)
		objbytes, err := db.LevelTempDB.Get(append(config.DBKEY_snapshot, []byte(o.LightSnapshotName())...))
		if err != nil {
			haserr = err
			return false
		}

		if len(objbytes) == 0 { // 快照碎片为空,继续下一个
			return true
		}

		if err := o.LightSnapshotDeSerialize(objbytes); err != nil {
			haserr = err
			return false
		}
		return true
	})
	return haserr
}

// 有效快照高度
func Height() uint64 {
	val, err := db.LevelTempDB.Get(config.DBKEY_snapshot_height)
	if err != nil {
		return 0
	}

	return utils.BytesToUint64(val)
}

// 快照起始高度
func StartHeightAt() {
	LightSnapshotStartHeight = Height()
}

package snapshot

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
	snapshot = sync.Map{}
	// 快照起始高度
	SnapshotStartHeight = uint64(0)
)

// 快照管理器接口
type SnapshotMgr interface {
	SnapshotName() string               // 快照碎片对象名称
	SnapshotSerialize() ([]byte, error) // 序列化内存对象
	SnapshotDeSerialize([]byte) error   // 反序列化,还原内存对象
	//SnapshotExtStorage() ([]db2.KVPair, error) // 快照源扩展存储
	//SnapshotCallback(bool)                     // 快照存储回调
}

// 添加对象
func Add(o SnapshotMgr) error {
	if _, ok := snapshot.LoadOrStore(o.SnapshotName(), o); ok {
		return errors.New("snapshoter name is exist")
	}

	return nil
}

// 保存快照
// 开启事务
// 内存快照
// 增量快照合并
// 清空增量快照
func Save(height uint64) error {
	// 间隔快照
	if height-Height() < config.SnapshotMinInterval && Height() > config.SnapshotMinInterval {
		return nil
	}

	// 快照高度<快照起始高度+新的备用见证人组数量*挖矿组最多成员
	// 开始启动后的首次快照至少需要跑完60个组
	if SnapshotStartHeight != 0 && height < SnapshotStartHeight+uint64(config.Witness_backup_group_new*config.Mining_group_max) {
		return nil
	}

	// 该高度已经快照了
	if Height() >= height {
		return nil
	}
	startAt := config.TimeNow()
	//isSave := false
	//defer func() {
	//	snapshot.Range(func(key, value any) bool {
	//		o := value.(SnapshotMgr)
	//		o.SnapshotCallback(isSave)
	//		return true
	//	})
	//}()

	snapshotPairs := make([]db2.KVPair, 0)

	// 快照高度
	snapshotPairs = append(snapshotPairs, db2.KVPair{
		Key:   config.DBKEY_snapshot_height,
		Value: utils.Uint64ToBytes(height),
	})

	var haserr error
	snapshot.Range(func(key, value any) bool {
		o := value.(SnapshotMgr)
		//pairs, err := o.SnapshotExtStorage()
		//if err != nil {
		//	haserr = err
		//	return false
		//}

		//if len(pairs) > 0 {
		//	snapshotPairs = append(snapshotPairs, pairs...)
		//}

		objbytes, err := o.SnapshotSerialize()
		if err != nil {
			haserr = err
			return false
		}

		snapshotPairs = append(snapshotPairs, db2.KVPair{
			Key:   append(config.DBKEY_snapshot, []byte(o.SnapshotName())...),
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
	//if err := db.LevelDB.MSet(snapshotPairs...); err != nil {
	//	return err
	//}

	engine.Log.Info(">>>>>> Snapshot[保存] 快照成功,高度:%d", height)
	engine.Log.Info(">>>>>> Snapshot[诊断] Seq:%d, 内存快照+增量快照合并+清空增量快照耗时:%v", db.LevelDB.GetSeq(), config.TimeNow().Sub(startAt))

	//isSave = true
	return nil
}

// 加载快照到内存
func Load() error {
	var haserr error
	snapshot.Range(func(key, value any) bool {
		o := value.(SnapshotMgr)
		objbytes, err := db.LevelTempDB.Get(append(config.DBKEY_snapshot, []byte(o.SnapshotName())...))
		if err != nil {
			haserr = err
			return false
		}

		if len(objbytes) == 0 { // 快照碎片为空,继续下一个
			return true
		}

		if err := o.SnapshotDeSerialize(objbytes); err != nil {
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
	SnapshotStartHeight = Height()
}

// 导出快照
func Export() ([]byte, error) {
	// TODO
	return nil, nil
}

// 导入快照
func Import(data []byte) error {
	// TODO
	return nil
}

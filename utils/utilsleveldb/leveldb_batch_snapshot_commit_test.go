package utilsleveldb

import (
	"testing"
)

func TestBatchSnapshotCommit(t *testing.T) {
	// startBatchSnapshotCommit()
}

func startBatchSnapshotCommit() {
	cleanLeveldb()
	createleveldb()
	leveldbExample_BatchSnapshotCommit()
	closeLeveldb()
}

/*
1
*/
func leveldbExample_BatchSnapshotCommit() {
	// dbkey, _ = NewLeveldbKey(utils.Uint64ToBytesByBigEndian(60))
	// num := 1000000
	// for i := 0; i < num; i++ {
	// 	key, _ = NewLeveldbKey(utils.Uint64ToBytesByBigEndian(uint64(i)))
	// 	ldb.SaveMap(*dbkey, *key, utils.Uint64ToBytesByBigEndian(uint64(i)))
	// }

	// sn, _ := ldb.GetDB().GetSnapshot()
	// sn.

}

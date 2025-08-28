package leveldb

// import (
// 	"sync"

// 	"web3_gui/utils"
// )

// var Once_ConnLevelDB sync.Once

// var LevelDB *utilsleveldb.LevelDB

// //链接leveldb
// func InitDB(name string) (err error) {
// 	Once_ConnLevelDB.Do(func() {
// 		LevelDB, err = utils.CreateLevelDB(name)
// 		if err != nil {
// 			return
// 		}
// 	})
// 	return
// }

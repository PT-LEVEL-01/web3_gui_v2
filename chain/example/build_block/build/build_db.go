package build

import (
	"path/filepath"
	db "web3_gui/chain/db/leveldb"
	"web3_gui/utils"
)

func BuildDB(n int, dirName string, dbName string) *db.LevelDB {
	err := utils.CheckCreateDir(dirName)
	if err != nil {
		panic(err)
	}
	dbPath := filepath.Join(dirName, dbName)
	dbOne, err := InitDB(dbPath)
	if err != nil {
		panic(err)
	}
	return dbOne
}

// 链接leveldb
func InitDB(namePath string) (*db.LevelDB, error) {
	err := utils.CheckCreateDir(namePath)
	if err != nil {
		panic(err)
	}
	levelDB, err := db.CreateLevelDB(namePath)
	if err != nil {
		return nil, err
	}
	return levelDB, nil
}

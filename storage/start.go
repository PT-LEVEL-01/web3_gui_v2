package storage

import (
	"math/big"
	"sync"
	"web3_gui/im/db"
	"web3_gui/libp2parea/v2"
	"web3_gui/utils"
)

var Node *libp2parea.Node

var StServer *StorageServer

var StClient *StorageClient

var genIDLock = new(sync.Mutex)
var GenID *big.Int

/*
获取自增长ID
*/
func GetGenID() ([]byte, utils.ERROR) {
	genIDLock.Lock()
	defer genIDLock.Unlock()
	temp := new(big.Int).Add(GenID, big.NewInt(1))
	ERR := db.StorageServer_SaveGenID(temp.Bytes())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	GenID = temp
	return temp.Bytes(), utils.NewErrorSuccess()
}

/*
启动存储服务器
*/
func StartStorageServer(area *libp2parea.Node) utils.ERROR {
	Node = area

	RegisterRPC(area)

	var ERR utils.ERROR
	StServer, ERR = CreateStorageServer()
	if !ERR.CheckSuccess() {
		return ERR
	}

	StClient, ERR = CreateStorageClient()
	if !ERR.CheckSuccess() {
		return ERR
	}
	//初始化自增长ID到内存
	number, err := db.StorageServer_GetGenID()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	GenID = new(big.Int).SetBytes(number)

	InitStorageServerMultcast()
	InitStorageServerSetup()
	RegisterHandlers()
	return utils.NewErrorSuccess()
}

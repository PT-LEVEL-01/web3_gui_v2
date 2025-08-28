package config

import (
	"crypto/sha256"
	"path/filepath"
	"web3_gui/chain/db/leveldb"
	"web3_gui/libp2parea/adapter"
)

const (
	AddrPre      = "ICOM"
	KeyPwd       = "123456789"
	Peer_total   = 5
	RootDir      = "E:\\test\\temp"
	WalletDBName = "data"
)

var (
	KeystoreDirPath = filepath.Join(RootDir, "keystores")
	WalletDirPath   = filepath.Join(RootDir, "wallet")
	AreaName        = sha256.Sum256([]byte("test"))
	Areas           = make([]*libp2parea.Area, 0, Peer_total)
	Ldb             = &leveldb.LevelDB{}
)

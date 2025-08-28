package mining

import (
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/keystore/adapter/crypto"
)

func FindWitnessName(addr crypto.AddressCoin) string {
	value, err := db.LevelTempDB.Find(append([]byte(config.WitnessAddr), addr...))
	if err != nil {
		return ""
	}
	return string(*value)
}

//func FindWitnessAddr(name string) *crypto.AddressCoin {
//	value, err := db.LevelTempDB.Find(append([]byte(config.WitnessName), []byte(name)...))
//	if err != nil {
//		return nil
//	}
//	addr := crypto.AddressCoin(*value)
//	return &addr
//}

//func SaveWitnessName(addr crypto.AddressCoin, name string) {
//	bs := []byte(name)
//	addrBs := []byte(addr)
//	db.LevelTempDB.Save(append([]byte(config.WitnessAddr), addr...), &bs)
//	db.LevelTempDB.Save(append([]byte(config.WitnessName), bs...), &addrBs)
//}
//func DelWitnessName(name string) {
//	addr := FindWitnessAddr(name)
//	db.LevelTempDB.Remove(append([]byte(config.WitnessName), []byte(name)...))
//	if addr != nil {
//		db.LevelTempDB.Remove(append([]byte(config.WitnessAddr), *addr...))
//	}
//}

func SaveWitnessName(addr crypto.AddressCoin, name string) {
	bs := []byte(name)
	db.LevelTempDB.Save(append([]byte(config.WitnessAddr), addr...), &bs)
}

func DelWitnessName(addr crypto.AddressCoin) {
	db.LevelTempDB.Remove(append([]byte(config.WitnessAddr), addr...))
}

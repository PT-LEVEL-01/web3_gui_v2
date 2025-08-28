package utilsleveldb

import (
	"fmt"
	"testing"
	"web3_gui/utils"
)

func TestLevelDBConfig(t *testing.T) {
	//startTestLeveldbConfig()
	//joinKey()
}

func joinKey() {
	joinKey, err := NewLeveldbKeyJoin([]byte{1}, []byte{2})
	fmt.Println("join key:", joinKey, err)
}

func startTestLeveldbConfig() {
	key1, ERR := BuildLeveldbKey(utils.Uint16ToBytesByBigEndian(1))
	if !ERR.CheckSuccess() {
		fmt.Println("构建key错误", ERR.String())
		return
	}
	key2, ERR := BuildLeveldbKey(utils.Uint16ToBytesByBigEndian(2))
	if !ERR.CheckSuccess() {
		fmt.Println("构建key错误", ERR.String())
		return
	}
	keyJoin := append(key1.key, key2.key...)
	keys, ERR := LeveldbParseKeyMore(keyJoin)
	if !ERR.CheckSuccess() {
		fmt.Println("构建key错误", ERR.String())
		return
	}

	fmt.Println("解析的key", keys)
}

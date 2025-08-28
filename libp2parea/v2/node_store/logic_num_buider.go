package nodeStore

import (
	"math/big"
	"sync"
)

type LogicNumBuider struct {
	lock  *sync.RWMutex
	id    *[]byte
	level uint
	ids   []*[]byte
}

func NewLogicNumBuider(id []byte, level uint) *LogicNumBuider {
	return &LogicNumBuider{
		lock:  new(sync.RWMutex),
		id:    &id,
		level: level,
		ids:   make([]*[]byte, 0),
	}
}

/*
得到每个节点网络的网络号，不包括本节点
@id        *utils.Multihash    要计算的id
@level     int                 深度
*/
func (this *LogicNumBuider) GetNodeNetworkNum() []*[]byte {
	this.lock.RLock()
	//if this.idStr != "" && this.idStr == hex.EncodeToString(*this.id) {
	if len(this.ids) > 0 {
		//utils.Log.Info().Str("已经存在", "").Send()
		this.lock.RUnlock()
		return this.ids
	}
	this.lock.RUnlock()

	this.lock.Lock()
	//utils.Log.Info().Str("不存在", "").Send()
	//this.idStr = hex.EncodeToString(*this.id)

	root := new(big.Int).SetBytes(*this.id)

	this.ids = make([]*[]byte, 0)
	for i := 0; i < int(this.level); i++ {
		//---------------------------------
		//将后面的i位置零
		//---------------------------------
		//		startInt := new(big.Int).Lsh(new(big.Int).Rsh(root, uint(i)), uint(i))
		//---------------------------------
		//第i位取反
		//---------------------------------
		networkNum := new(big.Int).Xor(root, new(big.Int).Lsh(big.NewInt(1), uint(i)))

		mhbs := networkNum.Bytes()

		this.ids = append(this.ids, &mhbs)
	}
	this.lock.Unlock()

	return this.ids
}

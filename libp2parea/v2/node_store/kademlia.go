package nodeStore

import (
	"bytes"
	"github.com/rs/zerolog"
	"math/big"
	"sync"
	"web3_gui/utils"
)

type Kademlia struct {
	level          uint            //id位数量
	bucket         [][]byte        //
	logicNumBuider *LogicNumBuider //
	lock           *sync.RWMutex   //
	log            *zerolog.Logger //日志
}

func NewKademlia(id []byte, level uint, log *zerolog.Logger) *Kademlia {
	lb := NewLogicNumBuider(id, level)
	k := &Kademlia{
		level:          level,
		bucket:         make([][]byte, 0, level),
		logicNumBuider: lb,
		lock:           new(sync.RWMutex),
		log:            log,
	}
	return k
}

/*
添加一个id
@return    bool    是否是自己需要的节点
*/
func (this *Kademlia) AddId(id []byte) (ok bool, removeIDs [][]byte) {
	// utils.Log.Info().Msgf("AddId: 1111111111111")
	if len(this.bucket) <= 0 {
		for i := 0; i < cap(this.bucket); i++ {
			this.bucket = append(this.bucket, id)
		}
		ok = true
		return
	}
	// utils.Log.Info().Msgf("AddId: 1111111111111")
	//非逻辑节点不要添加
	netIDs := this.logicNumBuider.GetNodeNetworkNum()

	delId := make([][]byte, 0)
	for i, one := range this.bucket {
		// utils.Log.Info().Msgf("AddId: 1111111111111")
		kl := NewBucket(2)
		kl.Add(new(big.Int).SetBytes(one))
		kl.Add(new(big.Int).SetBytes(id))
		nearId := kl.Get(new(big.Int).SetBytes(*netIDs[i]))
		//		fmt.Println(hex.EncodeToString(nearId[0].Bytes()))

		nearIdBs := nearId[0].Bytes()
		nearIdNewBs := utils.FullHighPositionZero(&nearIdBs, int(this.level/8))
		// utils.Log.Info().Msgf("AddId: 1111111111111")
		// if hex.EncodeToString(*one) == hex.EncodeToString(nearId[0].Bytes()) {
		if bytes.Equal(one, *nearIdNewBs) {
			// utils.Log.Info().Msgf("AddId: 1111111111111")
			continue
		}
		//this.Log.Info().Interface("要删除的id", one).Send()
		//		fmt.Println("删除的节点id", i, one, "替换", node.IdInfo.Id)
		delId = append(delId, one)
		//		netNodes[i] = node.IdInfo.Id
		this.bucket[i] = id
		ok = true
		// utils.Log.Info().Msgf("AddId: 1111111111111")
	}
	// utils.Log.Info().Msgf("AddId: 1111111111111")
	//找到删除的节点
	removeIDs = make([][]byte, 0)
	for _, one := range delId {
		// utils.Log.Info().Msgf("AddId: 1111111111111")
		find := false
		for _, netOne := range this.bucket {
			if bytes.Equal(one, netOne) {
				find = true
				break
			}
		}
		if !find {
			removeIDs = append(removeIDs, one)
		}
	}
	// utils.Log.Info().Msgf("AddId: 1111111111111")
	return
}

/*
删除一个id
*/
func (this *Kademlia) RemoveId(id []byte) {
	netIDs := this.logicNumBuider.GetNodeNetworkNum()
	for i, one := range this.bucket {
		if bytes.Equal(one, id) {
			continue
		}
		ids := this.GetIds()
		kl := NewBucket(len(ids))
		for _, one := range ids {
			kl.Add(new(big.Int).SetBytes(one))
		}
		nearId := kl.Get(new(big.Int).SetBytes(*netIDs[i]))
		idAddr := nearId[1].Bytes()
		nearIdNewBs := utils.FullHighPositionZero(&idAddr, int(this.level/8))
		this.bucket[i] = *nearIdNewBs
	}
}

/*
获取所有id
*/
func (this *Kademlia) GetIds() [][]byte {
	if len(this.bucket) <= 0 {
		// utils.Log.Info().Msgf("GetIds 000000000")
		return make([][]byte, 0)
	}
	//去重复
	m := make(map[string][]byte)
	for _, one := range this.bucket {
		// m[hex.EncodeToString(one)] = one
		m[utils.Bytes2string(one)] = one
	}
	//组装成数组
	ids := make([][]byte, 0)
	for _, v := range m {
		// utils.Log.Info().Msgf("GetIds 1111111111111")
		ids = append(ids, v)
	}
	// utils.Log.Info().Msgf("GetIds 222222222222")
	return ids
}

/*
通过下标获取id
*/
func (this *Kademlia) GetIndex(index int) []byte {
	return this.bucket[index]
}

/*
从新设置日志库
*/
func (this *Kademlia) SetLog(log *zerolog.Logger) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.log = log
}

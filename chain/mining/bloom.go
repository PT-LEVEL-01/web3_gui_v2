package mining

import (
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"hash"
	"sync"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/crypto"
)

const (
	//bloom过滤器字节长度
	BloomByteLength = 256

	BloomBitLength = 8 * BloomByteLength
)

type Bloom [BloomByteLength]byte

// 单个交易的所有事件
type EventInfos []*go_protos.ContractEventInfo

// 多个交易的所有事件
type EventMInfos []EventInfos

// 单交易创建bloom
func CreateBloom(infos EventInfos) Bloom {
	buf := make([]byte, 6)
	var bin Bloom
	//遍历单个交易中所有的事件
	for _, log := range infos {
		//添加合约地址到过滤器中
		addr := crypto.AddressFromB58String(log.ContractAddress)
		bin.add(addr[:], buf)
		//添加合约函数签名到过滤器中
		bin.add([]byte(log.Topic), buf)
		for k, d := range log.EventData {
			//除去最后一位的都是indexed参数，添加进过滤器中
			if k < len(log.EventData)-1 {
				bin.add([]byte(d), buf)
			}
		}
	}
	return bin
}

// 多交易创建bloom
func CreateMBloom(mInfos EventMInfos) Bloom {
	buf := make([]byte, 6)
	var bin Bloom
	//遍历多个交易
	for _, info := range mInfos {
		//遍历单个交易的所有事件
		for _, log := range info {
			addr := crypto.AddressFromB58String(log.ContractAddress)
			bin.add(addr[:], buf)
			bin.add([]byte(log.Topic), buf)
			for k, d := range log.EventData {
				if k < len(log.EventData)-1 {
					bin.add([]byte(d), buf)
				}
			}
		}
	}
	return bin
}

// 添加数据进bloom
func (b *Bloom) add(d []byte, buf []byte) {
	//获取数据所对应bloom的kv对，并添加进bloom中
	i1, v1, i2, v2, i3, v3 := bloomValues(d, buf)
	b[i1] |= v1
	b[i2] |= v2
	b[i3] |= v3
}

func (b Bloom) Bytes() []byte {
	return b[:]
}

// 检查过滤器中是否存在topic
func (b Bloom) Check(topic []byte) bool {
	i1, v1, i2, v2, i3, v3 := bloomValues(topic, make([]byte, 6))
	return v1 == v1&b[i1] &&
		v2 == v2&b[i2] &&
		v3 == v3&b[i3]
}

// 获取给定数据data在bloom过滤器中设置的字节值(index-value对)
func bloomValues(data []byte, hashbuf []byte) (uint, byte, uint, byte, uint, byte) {
	sha := hasherPool.Get().(KeccakState)
	sha.Reset()
	//写入数据用于生成hash
	sha.Write(data)
	//取生成出来的hash的前6个字节
	sha.Read(hashbuf)
	hasherPool.Put(sha)

	//1,3,5字节用于bloom的字节值，模7位移后最大左移7位，值为128
	v1 := byte(1 << (hashbuf[1] & 0x7))
	v2 := byte(1 << (hashbuf[3] & 0x7))
	v3 := byte(1 << (hashbuf[5] & 0x7))

	//0,2,4字节用于bloom的索引字节位置
	i1 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf)&0x7ff)>>3) - 1
	i2 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[2:])&0x7ff)>>3) - 1
	i3 := BloomByteLength - uint((binary.BigEndian.Uint16(hashbuf[4:])&0x7ff)>>3) - 1

	return i1, v1, i2, v2, i3, v3
}

var hasherPool = sync.Pool{
	New: func() interface{} { return sha3.NewLegacyKeccak256() },
}

type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

type bytesBacked interface {
	Bytes() []byte
}

// 在过滤器bin中查找topic是否存在
func BloomLookup(bin Bloom, topic bytesBacked) bool {
	return bin.Check(topic.Bytes())
}

// 字节转bloom
func BytesToBloom(bloomBs []byte) *Bloom {
	if len(bloomBs) != BloomByteLength {
		return new(Bloom)
	}

	var b Bloom
	copy(b[:], bloomBs)
	return &b
}

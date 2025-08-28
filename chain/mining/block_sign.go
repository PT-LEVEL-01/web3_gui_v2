package mining

import (
	"bytes"
	"errors"
	"web3_gui/chain/config"

	"golang.org/x/crypto/ed25519"
	"web3_gui/libp2parea/adapter/engine"
)

// const (
// 	blockHeadSignSep byte = ':'
// 	checkSignRatio        = float64(2) / 3
// )

var errBlockHeadSign = errors.New("blockHeadSign decode fail")

// var CurrSignBlock SignBlock

// type SignBlock struct {
// 	GroupHeight uint64
// 	BlockHeight uint64
// 	BlockHash   []byte
// }

// func (s SignBlock) IsEmpty() bool {
// 	return reflect.DeepEqual(s, SignBlock{})
// }

// func CleanCurrSignBlock() {
// 	CurrSignBlock = SignBlock{}
// }

/*
*
编码hash，sign
*/
// func EncodeBlockSign(hash, sign []byte) []byte {
// 	buf := make([]byte, len(hash)+len(sign)+2+1)
// 	pos := 0
// 	binary.BigEndian.PutUint16(buf[pos:], uint16(len(hash)))
// 	pos += 2

// 	copy(buf[pos:], hash)
// 	pos += len(hash)

// 	buf[pos] = blockHeadSignSep
// 	pos++

// 	copy(buf[pos:], sign)

// 	return buf
// }

/*
*
解码hash，sign
*/
// func DecodeBlockSign(buf []byte) ([]byte, []byte, error) {
// 	pos := 0
// 	if pos+2 > len(buf) {
// 		return nil, nil, errBlockHeadSign
// 	}

// 	prefixLen := int(binary.BigEndian.Uint16(buf[pos:]))
// 	pos += 2

// 	if prefixLen+pos > len(buf) {
// 		return nil, nil, errBlockHeadSign
// 	}
// 	hash := buf[pos : pos+prefixLen]
// 	pos += prefixLen

// 	if buf[pos] != blockHeadSignSep {
// 		return nil, nil, errBlockHeadSign
// 	}
// 	pos++
// 	sign := buf[pos:]
// 	return hash, sign, nil
// }

func IsSign() bool {
	return GetCurrWitnessBackupCount() > 1
}

func GetCurrWitnessBackupCount() int {
	currWitnessGroup := GetLongChain().WitnessChain.WitnessGroup
	currWitness := currWitnessGroup.Witness[len(currWitnessGroup.Witness)-1]

	return len(currWitness.WitnessBigGroup.Witnesses)
}

/*
*
是否要签名
*/
// func (this *BlockHead) IsSign() bool {
// 	return IsSign()
// }

/*
*
初始化其他见证人签名
*/
func (this *BlockHead) InitExtSign() {
	this.ExtSign = make([][]byte, 0)
}

/*
*
是否存在sign
*/
func (this *BlockHead) IsExistSign(sign []byte) bool {
	for _, v := range this.ExtSign {
		if bytes.Equal(v, sign) {
			return true
		}
	}
	return false
}

/*
*
设置extSign
*/
func (this *BlockHead) SetExtSign(sign []byte) bool {
	if this.IsExistSign(sign) {
		return false
	}
	this.ExtSign = append(this.ExtSign, sign)
	return true
}

/*
*
验证签名extSign
*/
func (this *BlockHead) CheckExtSign(witness []*Witness) bool {
	//验证签名数量
	extSignTotal := len(this.ExtSign)

	//验证时剔除ExtSign
	tmp := *this
	blockHeadCopy := &tmp
	blockHeadCopy.InitExtSign()
	bs := blockHeadCopy.Serialize()
	//根据见证人排序验证签名
	index := 0
	for _, v := range witness {
		if index > extSignTotal-1 {
			break
		}
		pkey := ed25519.PublicKey(v.Puk)
		if ed25519.Verify(pkey, *bs, this.ExtSign[index]) {
			if index == extSignTotal-1 {
				break
			}
			index++
		}
	}

	//验证成功签名数量
	if index+1+1 >= config.BftMajorityPrinciple(len(witness)) {
		return true
	}
	engine.Log.Info("导入区块签名个数量错误：签名总个数:%d 成功:%d 最少:%d", extSignTotal, index, config.BftMajorityPrinciple(len(witness))-1)

	return false
}

/*
*
验证单个签名extSign
*/
func (this *BlockHead) CheckExtSignOne(puk, sign []byte) bool {
	//检查签名是否正确
	bs := this.Serialize()
	pkey := ed25519.PublicKey(puk)
	if !ed25519.Verify(pkey, *bs, sign) {
		return false
	}
	return true
}

/*
*
添加其他见证人签名
*/
// func (this *BlockHead) SetSigns(s ...[]byte) {
// 	// if !this.IsSign() {
// 	// 	return
// 	// }

// 	this.ExtSign = append(this.ExtSign, s...)
// }

/*
*
验证extsign是否满足重构
*/
// func (this *BlockHead) CheckExtSign() bool {
// 	total := GetCurrWitnessBackupCount()
// 	if total == 1 {
// 		return true
// 	}

// 	return len(this.ExtSign) >= int(checkSignRatio*(float64(total)))
// }

/*
*
广播区块签名
*/
// func MulticastBlockSign(bhVO *BlockHeadVO) {
// 	if bhVO.BH.IsSign() {
// 		AddBlockToCache(bhVO)

// 		bs, err := bhVO.Proto()
// 		if err != nil {
// 			return
// 		}

// 		whiltlistNodes := Area.NodeManager.GetWhiltListNodes()

// 		Area.BroadcastsAll(1, config.MSGID_multicast_blockhead_sign, whiltlistNodes, nil, nil, bs)
// 		return
// 	}

// 	SignBuildBlock(bhVO)
// }

/*
*
签名构建区块
*/
// func SignBuildBlock(bhvo *BlockHeadVO) {
// 	if bhvo.BH.IsSign() {
// 		if !bhvo.BH.CheckExtSign() {
// 			return
// 		}
// 		DelBlockToCache(&bhvo.BH.Hash)
// 		bhvo.BH.Hash = nil
// 		bhvo.BH.BuildSign(bhvo.BH.Witness)
// 		bhvo.BH.BuildBlockHash()
// 	}

// 	engine.Log.Info("=== build block Success === group height:%d block height:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
// 	engine.Log.Info("=== build block Success === Block hash %s", hex.EncodeToString(bhvo.BH.Hash))
// 	engine.Log.Info("=== build block Success === pre Block hash %s", hex.EncodeToString(bhvo.BH.Previousblockhash))

// 	bhvo.FromBroadcast = true

// 	//广播区块
// 	UniformityMulticastBlock(bhvo)
// }

// type SignBlockExpire struct {
// 	sync.Locker
// 	bhMap map[uint64]time.Time
// }

// var CheckSignExipre = createSignExipre()

// const sign_expire = 5 * time.Second

// func createSignExipre() *SignBlockExpire {
// 	return &SignBlockExpire{
// 		new(sync.RWMutex), make(map[uint64]time.Time),
// 	}
// }

// func (s *SignBlockExpire) Check(h uint64) bool {
// 	t, ok := s.bhMap[h]
// 	if ok && t.Add(sign_expire).After(config.TimeNow()) {
// 		return false
// 	}

// 	return true
// }

// func (s *SignBlockExpire) Add(h uint64) bool {
// 	s.Lock()
// 	defer s.Unlock()

// 	if !s.Check(h) {
// 		return false
// 	}

// 	s.bhMap[h] = config.TimeNow()
// 	s.Del(h - 1)
// 	return true
// }

// func (s *SignBlockExpire) Del(h uint64) {
// 	delete(s.bhMap, h)
// }

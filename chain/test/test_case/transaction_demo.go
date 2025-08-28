package main

import (
	"math/big"
	"web3_gui/chain/config"

	"golang.org/x/crypto/ed25519"
)

// 基于转账交易demo
// 基于转账交易demo
// 基于转账交易demo
// 基于转账交易demo
// 基于转账交易demo
// 基于转账交易demo

type AddressCoinT []byte

/*
转账交易
*/
type Tx_Pay struct {
	TxBase
}

/*
交易
*/
type TxBase struct {
	Hash       []byte  `json:"h"`     //本交易hash，不参与区块hash，只用来保存
	Type       uint64  `json:"t"`     //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
	Vin_total  uint64  `json:"vin_t"` //输入交易数量
	Vin        []*Vin  `json:"vin"`   //交易输入
	Vout_total uint64  `json:"vot_t"` //输出交易数量
	Vout       []*Vout `json:"vot"`   //交易输出
	Gas        uint64  `json:"g"`     //交易手续费，此字段不参与交易hash
	LockHeight uint64  `json:"l_h"`   //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
	Payload    []byte  `json:"p"`     //备注信息
	BlockHash  []byte  `json:"bh"`    //本交易属于的区块hash，不参与区块hash，只用来保存
	GasUsed    uint64  //gasused
}

/*
UTXO输入
*/
type Vin struct {
	// Txid []byte `json:"txid"` //UTXO 前一个交易的id
	// // TxidStr string `json:"-"`    //
	// Vout uint64 `json:"vout"` //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	Puk   []byte  `json:"puk"` //公钥
	Nonce big.Int `json:"n"`   //转出账户nonce
	// PukStr       string             `json:"-"`    //缓存公钥字符串
	PukIsSelf int          `json:"-"` //缓存此公钥是否属于自己钱包。0=未检查;1=不属于自己;2=是属于自己;
	PukToAddr AddressCoinT `json:"-"` //缓存此公钥的地址
	// PukToAddrStr string             `json:"-"`    //缓存此公钥的地址字符串
	Sign []byte `json:"sign"` //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	//	VoutSign []byte `json:"voutsign"` //对本交易的输出签名
}

type VinVO struct {
	// Txid string `json:"txid"` //UTXO 前一个交易的id
	// Vout uint64 `json:"vout"` //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	Addr  string `json:"addr"`
	Puk   string `json:"puk"`  //公钥
	Nonce string `json:"n"`    //转出账户nonce
	Sign  string `json:"sign"` //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	//	VoutSign []byte `json:"voutsign"` //对本交易的输出签名
}

/*
UTXO输出
*/
type Vout struct {
	Value        uint64       `json:"value"`         //输出金额 = 实际金额 * 100000000
	Address      AddressCoinT `json:"address"`       //钱包地址
	FrozenHeight uint64       `json:"frozen_height"` //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
	AddrIsSelf   int          `json:"-"`             //缓存地址是否属于自己钱包。0=未检查;1=不属于自己;2=是属于自己;
	AddrStr      string       `json:"-"`             //缓存地址字符串
	// Txids        []byte             `json:"txid"`          //本输出被使用后的交易id
}

type VoutVO struct {
	Value        uint64 `json:"value"`         //输出金额 = 实际金额 * 100000000
	Address      string `json:"address"`       //钱包地址
	FrozenHeight uint64 `json:"frozen_height"` //冻结高度。在冻结高度以下，未花费的交易余额不能使用
	// Txids        string `json:"txid"`          //本输出被使用后的交易id
}

var srcaddress = "MMSAwFL1NwuMi471dA665v31czNCr15HjM494" // 当前交易发起地址
var address = "MMSGsuykNJswP19VWrzDuQqgaGJLZsuZJ1XH4"    // 当前交易接受地址
var amount = 10000000000                                 // 当前交易金额
var gas = 100000                                         // 当前交易手续费
var frozen_height = 7                                    // 当前交易冻结高低（第几个块后交易生效）
var pwd = "xhy19liu21@"                                  // 当前交易接受地址（密码）

func main() {

	// 当前内存是否超过设定的值
	// 判断地址是否正确
	// 钱包中查找地址，判断地址是否属于本钱包

	// 通过当前交易发起地址获取 余额 判断钱够不够

	// 判断当前地址未上链的交易数量

	// 获取nonce (这个参数是以此累加 防止同一个地址交易计算sign的时候出现重复)
	// chain.GetBalance().FindNonce(item.Addr)
	nonce := new(big.Int).SetUint64(uint64(1))

	// 从钱包的结构体 Wallet.addrMap字段可以获取
	// 当前地址的公钥 不是钱包的公钥
	puk := []byte("fdfjiauf0wuf09823u9fjds9ajf")

	// 构建交易输入
	vins := make([]*Vin, 0)
	vin := Vin{
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, // 当前地址的公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        uint64(amount),        //输出金额 = 实际金额 * 100000000
		Address:      AddressCoinT(address), //钱包地址
		FrozenHeight: uint64(frozen_height), //
	}
	vouts = append(vouts, &vout)

	var pay *Tx_Pay

	// 当前链的高度
	currentHeight := 10

	// 锁定高度
	Wallet_tx_lockHeight := 300

	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay,                                //交易类型
			Vin_total:  uint64(len(vins)),                                        //输入交易数量
			Vin:        vins,                                                     //交易输入
			Vout_total: uint64(len(vouts)),                                       //输出交易数量
			Vout:       vouts,                                                    //交易输出
			Gas:        uint64(gas),                                              //交易手续费
			LockHeight: uint64(currentHeight) + uint64(Wallet_tx_lockHeight) + i, //锁定高度
			//Payload:    commentbs,                                       // 备注消息
		}
		pay = &Tx_Pay{
			TxBase: base,
		}

		//给输出签名，防篡改
		for i, one := range pay.Vin {

			_, prk, err := GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return
			}
			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))
			sign := GetSign(&prk, uint64(i))
			pay.Vin[i].Sign = *sign
		}

		// 创建交易hash获得交易id
		//pay.BuildHash()

		// 判断当前hash是否已经存在
		//if pay.CheckHashExist() {
		//	pay = nil
		//	continue
		//} else {
		//	break
		//}

	}

	// 最后广播该交易到各个节点
	// ..........

}

/*
	构建hash值得到交易id
*/
//func (this *Tx_Pay) BuildHash() {
//	if this.Hash != nil && len(this.Hash) > 0 {
//		return
//	}
//
//	bs := this.Serialize()
//	id := make([]byte, 8)
//	binary.PutUvarint(id, config.Wallet_tx_type_pay)
//	// jsonBs, _ := this.Json()
//
//	// fmt.Println("序列化输出 111", string(*jsonBs))
//	// fmt.Println("序列化输出 222", len(*bs), hex.EncodeToString(*bs))
//	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
//}

/*
获取签名
*/
func GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {

	// 序列化本笔交易里面的指定字段
	//signDst := GetSignSerialize(nil, vinIndex)

	// 然后加密返回
	//sign := keystore.Sign(*key, *signDst)
	sign := []byte("234234")
	return &sign
}

/*
签名
*/
func Sign(prk ed25519.PrivateKey, content []byte) []byte {
	if len(prk) == 0 {
		return nil
	}
	return ed25519.Sign(prk, content)
}

/*
	获取本交易用作签名的序列化
	[上一个交易GetVoutSignSerialize()返回]+[本交易类型]+[本交易输入总数]+[本交易输入index]+
	[本交易输出总数]+[vouts序列化]+[锁定区块高度]
	@voutBs    *[]byte    上一个交易GetVoutSignSerialize()返回
*/
//func (this *TxBase) GetSignSerialize(voutBs *[]byte, vinIndex uint64) *[]byte {
//	if vinIndex > uint64(len(this.Vin)) {
//		return nil
//	}
//
//	voutBssLenght := 0
//	voutBss := make([]*[]byte, 0, len(this.Vout))
//	for _, one := range this.Vout {
//		voutBsOne := one.Serialize()
//		voutBss = append(voutBss, voutBsOne)
//		voutBssLenght += len(*voutBsOne)
//	}
//
//	var bs []byte
//	if voutBs == nil {
//		bs = make([]byte, 0, 8+8+8+8+voutBssLenght+8+len(this.Payload))
//	} else {
//		bs = make([]byte, 0, len(*voutBs)+8+8+8+8+voutBssLenght+8+len(this.Payload))
//		bs = append(bs, *voutBs...)
//	}
//
//	bs = append(bs, utils.Uint64ToBytes(this.Type)...)
//	bs = append(bs, utils.Uint64ToBytes(this.Vin_total)...)
//	bs = append(bs, utils.Uint64ToBytes(vinIndex)...)
//	bs = append(bs, utils.Uint64ToBytes(this.Vout_total)...)
//	for _, one := range voutBss {
//		bs = append(bs, *one...)
//	}
//	bs = append(bs, utils.Uint64ToBytes(this.Gas)...)
//	bs = append(bs, utils.Uint64ToBytes(this.LockHeight)...)
//	bs = append(bs, this.Payload...)
//	return &bs
//
//}

/*
通过公钥获取密钥
*/
func GetKeyByPuk(puk []byte, password string) (rand []byte, prk []byte, err error) {

	// 具体函数功能 Decrypt()
	// 通过密码+钱包Seed+Salt 利用cbc解密出 seed
	// 判断密码是否正确 然后获取key和code 经过如下函数 获得私钥
	// newKey, _, err = HkdfChainCodeNew(key, code, addrInfo.Index)
	//		if err != nil {
	//			return nil, nil, err
	//		}
	//rand = *newKey
	//puk, prk, err = ed25519.GenerateKey(bytes.NewBuffer(rand))

	return
}

/*
	获取hkdf链编码
	@master    []byte    随机数
	@salt      []byte    盐
	@index     uint64    索引，棘轮数
*/
//func HkdfChainCodeNew(master, salt []byte, index uint64) (*[]byte, *[]byte, error) {
//	if index > 100 {
//		return HkdfChainCodeNewV3(master, salt, index)
//	}
//	// fmt.Println("获取hkdf链编码:", hex.EncodeToString(master), hex.EncodeToString(salt), index)
//	hkdf := hkdf.New(sha256.New, master, salt, nil)
//	hashSeed := make([]byte, 64)
//	for i := uint64(0); i < index; i++ {
//		// fmt.Println("for获取hkdf链编码:", i)
//		n, err := io.ReadFull(hkdf, hashSeed)
//		if n != len(hashSeed) {
//			fmt.Println("hkdf read error:", n)
//			return nil, nil, errors.New("hkdf chain read hash fail")
//		}
//		if err != nil {
//			return nil, nil, err
//		}
//	}
//
//	key := hashSeed[:32]
//	chainCode := hashSeed[32:]
//	return &key, &chainCode, nil
//}

/*
	使用密码解密种子，获得私钥和链编码
	@return    ok    bool    密码是否正确
	@return    key   []byte  生成私钥的随机数
	@return    code  []byte  链编码
*/
//func (* Wallet)Decrypt(pwdbs [32]byte) (ok bool, key, code []byte, err error) {
//	//密码取hash
//
//	if this.Seed != nil && len(this.Seed) > 0 && (this.Key == nil || len(this.Key) <= 0) {
//		//先用密码解密种子
//		seedBs, err := DecryptCBC(this.Seed, pwdbs[:], Salt)
//		if err != nil {
//			return false, nil, nil, err
//		}
//		//判断密码是否正确
//		chackHash := sha256.Sum256(seedBs)
//		if !bytes.Equal(chackHash[:], this.CheckHash) {
//			return false, nil, nil, ERROR_password_fail
//		}
//
//		// hash := sha256.New
//		// key := &[32]byte{}
//		// hkdf := hkdf.New(hash, seedBs, salt, nil)
//		// _, err = io.ReadFull(hkdf, key[:])
//		// if err != nil {
//		// 	return false, nil, nil, err
//		// }
//		// code := &[32]byte{}
//		// _, err = io.ReadFull(hkdf, code[:])
//		// if err != nil {
//		// 	return false, nil, nil, err
//		// }
//		key, code, err := BuildKeyBySeed(&seedBs, Salt)
//		if err != nil {
//			return false, nil, nil, err
//		}
//
//		// keySec, err := crypto.EncryptCBC(key[:], pwdbs[:], salt)
//		// if err != nil {
//		// 	return false, nil, nil, err
//		// }
//		// codeSec, err := crypto.EncryptCBC(code[:], pwdbs[:], salt)
//		// if err != nil {
//		// 	return false, nil, nil, err
//		// }
//
//		// this.Key = keySec
//		// this.ChainCode = codeSec
//		// this.IV = salt
//		return true, *key, *code, nil
//	}
//
//	//先用密码解密key和链编码
//	keyBs, err := DecryptCBC(this.Key, pwdbs[:], this.IV)
//	if err != nil {
//		return false, nil, nil, ERROR_password_fail
//	}
//	codeBs, err := DecryptCBC(this.ChainCode, pwdbs[:], this.IV)
//	if err != nil {
//		return false, nil, nil, ERROR_password_fail
//	}
//
//	//验证密码是否正确
//	checkHash := append(keyBs, codeBs...)
//	h := sha256.New()
//	n, err := h.Write(checkHash)
//	if n != len(checkHash) {
//		//hash 写入失败
//		return false, nil, nil, errors.New("hash Write failure")
//	}
//	if err != nil {
//		return false, nil, nil, err
//	}
//	checkHash = h.Sum(pwdbs[:])
//	// checkHash = sha256.Sum256(checkHash)[:]
//	if !bytes.Equal(checkHash, this.CheckHash) {
//		return false, nil, nil, nil
//	}
//	return true, keyBs, codeBs, nil
//}

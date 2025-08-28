package mining

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"github.com/gogo/protobuf/proto"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
// BlockHead_Hash              = "Hash"
// BlockHead_Height            = "Height"
// BlockHead_MerkleRoot        = "MerkleRoot"
// BlockHead_Previousblockhash = "Previousblockhash"
// BlockHead_Nextblockhash     = "Nextblockhash"
// BlockHead_Sign              = "sign"
// BlockHead_Tx                = "Tx"
// BlockHead_Time              = "Time"
)

//var (

//	//	headBlock     = new(sync.Map)              //保存对应区块高度的区块头hash。key:uint64=区块高度;value:*[]byte=区块高度对应区块头hash;
//	lastBlockHead *BlockHead                   //最高区块
//	preBlockHead  *BlockHead                   //最高区块的上一个区块
//	syncBlock     = make(chan *BlockHeadVO, 1) //连续导入区块
//)

///*
//	从数据库加载区块
//*/
//func LoadBlockInDB() {

//}

/*
区块头
*/
type BlockHead struct {
	Hash              []byte             `json:"H"`   //区块头hash
	Height            uint64             `json:"Ht"`  //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
	GroupHeight       uint64             `json:"GH"`  //矿工组高度
	GroupHeightGrowth uint64             `json:"GHG"` //组高度增长量。默认0为自动计算增长量（兼容之前的区块）,最少增量为1
	Previousblockhash []byte             `json:"Pbh"` //上一个区块头hash
	Nextblockhash     []byte             `json:"Nbh"` //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
	NTx               uint64             `json:"NTx"` //交易数量
	MerkleRoot        []byte             `json:"M"`   //交易默克尔树根hash
	Tx                [][]byte           `json:"Tx"`  //本区块包含的交易id
	Time              int64              `json:"T"`   //出块时间，unix时间戳
	Witness           crypto.AddressCoin `json:"W"`   //此块见证人地址
	Sign              []byte             `json:"s"`   //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	ExtSign           [][]byte           `json:"Es"`  //其他见证人签名集合
	// Nonce             uint64             `json:"nonce"`             //随机数，用以调整当前区块头hash
}

/*
构建默克尔树根
*/
func (this *BlockHead) BuildMerkleRoot() {
	this.MerkleRoot = utils.BuildMerkleRoot(this.Tx)
}

/*
将需要hash的字段序列化，不包括Sign字段
*/
func (this *BlockHead) Serialize() *[]byte {
	length := 0
	for _, one := range this.Tx {
		length += len(one)
	}

	for _, one := range this.ExtSign {
		length += len(one)
	}

	length += 8 + 8 + 8 + 8 + len(this.Previousblockhash) + len(this.MerkleRoot) + len(this.Witness)
	if this.GroupHeightGrowth != 0 {
		length += 8
	}
	bs := make([]byte, 0, length)

	bs = append(bs, utils.Uint64ToBytes(this.Height)...)
	bs = append(bs, utils.Uint64ToBytes(this.GroupHeight)...)
	if this.GroupHeightGrowth != 0 {
		bs = append(bs, utils.Uint64ToBytes(this.GroupHeightGrowth)...)
	}
	bs = append(bs, this.Previousblockhash...)
	bs = append(bs, utils.Uint64ToBytes(this.NTx)...)
	bs = append(bs, this.MerkleRoot...)
	for _, one := range this.Tx {
		bs = append(bs, one...)
	}
	bs = append(bs, utils.Uint64ToBytes(uint64(this.Time))...)
	bs = append(bs, this.Witness...)

	for _, one := range this.ExtSign {
		bs = append(bs, one...)
	}
	return &bs

	//-----------------
	// buf := bytes.NewBuffer(nil)
	// buf.Write(utils.Uint64ToBytes(this.Height))
	// buf.Write(utils.Uint64ToBytes(this.GroupHeight))
	// if this.GroupHeightGrowth != 0 {
	// 	buf.Write(utils.Uint64ToBytes(this.GroupHeightGrowth))
	// }
	// buf.Write(this.Previousblockhash)

	// buf.Write(utils.Uint64ToBytes(this.NTx))
	// buf.Write(this.MerkleRoot)
	// for _, one := range this.Tx {
	// 	buf.Write(one)
	// }
	// buf.Write(utils.Uint64ToBytes(uint64(this.Time)))
	// buf.Write(this.Witness)
	// bs := buf.Bytes()
	// return &bs
}

/*
区块签名
*/
func (this *BlockHead) BuildSign(addr crypto.AddressCoin, key keystore.KeystoreInterface) {
	bs := this.Serialize()
	_, prk, _, err := key.GetKeyByAddr(addr, config.Wallet_keystore_default_pwd)
	if err != nil {
		return
	}
	signBs := keystore.Sign(prk, *bs)

	this.Sign = signBs
}

/*
检查区块头合法性
*/
func (this *BlockHead) CheckBlockHead(puk []byte) bool {
	//utils.Log.Info().Hex("检查公钥", puk).Send()
	if this.Height < config.Mining_block_start_height_jump {
		return true
	}
	//检查签名是否正确
	bs := this.Serialize()
	pkey := ed25519.PublicKey(puk)
	if !ed25519.Verify(pkey, *bs, this.Sign) {
		return false
	}

	old := this.Hash
	this.BuildBlockHash()
	if !bytes.Equal(old, this.Hash) {
		return false
	}
	if this.Height <= 1 {
		return true
	}

	return true

}

/*
	寻找幸运数字
	@zoroes        uint64       难度，前导零数量
	@stopSignal    chan bool    停止信号 true=已经找到；false=未找到，被终止；
*/
// func (this *BlockHead) FindNonce(zoroes uint64, stopSignal chan bool) chan bool {
// 	// fmt.Println("start 开始工作，寻找区块高度", this.Height, "幸运数字。请等待...")
// 	result := make(chan bool, 1)

// 	//TODO 测试区块分叉使用，发布版本可以删除
// 	// this.Nonce = uint64(utils.GetRandNum(20000))

// 	stop := false
// 	for !stop {
// 		this.Nonce++
// 		this.BuildHash()
// 		if utils.CheckNonce(this.Hash, zoroes) {
// 			result <- true
// 			// fmt.Println("end 停止工作，找到幸运数字", this.Height)
// 			return result
// 		}
// 		select {
// 		case <-stopSignal:
// 			// fmt.Println("end 停止工作，因外部中断", this.Height)
// 			// close(stopSignal)
// 			stop = true
// 		default:
// 		}
// 	}
// 	result <- false
// 	return result
// }

/*
构建区块头hash
*/
func (this *BlockHead) BuildBlockHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}

	buf := bytes.NewBuffer(*this.Serialize())
	// buf.Write(utils.Uint64ToBytes(this.Height))
	// buf.Write(utils.Uint64ToBytes(this.GroupHeight))
	// buf.Write(this.Previousblockhash)

	// buf.Write(utils.Uint64ToBytes(this.NTx))
	// buf.Write(this.MerkleRoot)
	// for _, one := range this.Tx {
	// 	buf.Write(one)
	// }
	// buf.Write(utils.Uint64ToBytes(uint64(this.Time)))
	// buf.Write(this.Witness)
	// buf.Write(this.Sign)
	bs := buf.Bytes()

	this.Hash = utils.Hash_SHA3_256(bs)
}

/*
构建区块头hash
*/
func (this *BlockHead) CheckHashExist() (bool, error) {
	bhashkey := config.BuildBlockHead(this.Hash)
	return db.LevelDB.CheckHashExist(bhashkey)
}

/*
	保存到本地磁盘
*/
// func (this *BlockHead) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, nil
// }

/*
保存到本地磁盘
*/
func (this *BlockHead) Proto() (*[]byte, error) {
	bhp := go_protos.BlockHead{
		Hash:              this.Hash,
		Height:            this.Height,
		GroupHeight:       this.GroupHeight,
		GroupHeightGrowth: this.GroupHeightGrowth,
		Previousblockhash: this.Previousblockhash,
		Nextblockhash:     this.Nextblockhash,
		NTx:               this.NTx,
		MerkleRoot:        this.MerkleRoot,
		Tx:                this.Tx,
		Time:              this.Time,
		Witness:           this.Witness,
		Sign:              this.Sign,
		ExtSign:           this.ExtSign,
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
解析区块头
*/
func ParseBlockHeadProto(bs *[]byte) (*BlockHead, error) {
	if bs == nil {
		return nil, nil
	}
	bhp := new(go_protos.BlockHead)
	err := proto.Unmarshal(*bs, bhp)
	if err != nil {
		return nil, err
	}
	bh := BlockHead{
		Hash:              bhp.Hash,
		Height:            bhp.Height,
		GroupHeight:       bhp.GroupHeight,
		GroupHeightGrowth: bhp.GroupHeightGrowth,
		Previousblockhash: bhp.Previousblockhash,
		Nextblockhash:     bhp.Nextblockhash,
		NTx:               bhp.NTx,
		MerkleRoot:        bhp.MerkleRoot,
		Tx:                bhp.Tx,
		Time:              bhp.Time,
		Witness:           bhp.Witness,
		Sign:              bhp.Sign,
		ExtSign:           bhp.ExtSign,
	}
	return &bh, nil
}

/*
	解析区块头
*/
// func ParseBlockHead(bs *[]byte) (*BlockHead, error) {
// 	bh := new(BlockHead)
// 	// var jso = jsoniter.ConfigCompatibleWithStandardLibrary
// 	// err := json.Unmarshal(*bs, bh)

// 	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 	decoder.UseNumber()
// 	err := decoder.Decode(bh)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return bh, nil
// }

type BlockHeadVO struct {
	FromBroadcast   bool       `json:"-"`   //是否来自于广播的区块
	StaretBlockHash []byte     `json:"sbh"` //创始区块hash
	BH              *BlockHead `json:"bh"`  //区块
	Txs             []TxItr    `json:"txs"` //交易明细
}

/*
	json格式化
*/
// func (this *BlockHeadVO) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, nil
// }

/*
格式化
*/
func (this *BlockHeadVO) Proto() (*[]byte, error) {
	bh := go_protos.BlockHead{
		Hash:              this.BH.Hash,
		Height:            this.BH.Height,
		GroupHeight:       this.BH.GroupHeight,
		GroupHeightGrowth: this.BH.GroupHeightGrowth,
		Previousblockhash: this.BH.Previousblockhash,
		Nextblockhash:     this.BH.Nextblockhash,
		NTx:               this.BH.NTx,
		MerkleRoot:        this.BH.MerkleRoot,
		Tx:                this.BH.Tx,
		Time:              this.BH.Time,
		Witness:           this.BH.Witness,
		Sign:              this.BH.Sign,
		ExtSign:           this.BH.ExtSign,
	}

	bhat := go_protos.BlockHeadAndTxs{
		StaretBlockHash: this.StaretBlockHash,
		Bh:              &bh,
		TxBs:            make([][]byte, 0),
	}
	for _, one := range this.Txs {
		bs, err := one.Proto()
		if err != nil {
			return nil, err
		}
		bhat.TxBs = append(bhat.TxBs, *bs)
	}
	bs, err := bhat.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, nil
	// return bhat.Marshal()
}

const startBlockHashLength = 6

/*
验证区块合法性
*/
func (this *BlockHeadVO) Verify(sbh []byte) bool {
	if sbh == nil || len(sbh) < startBlockHashLength {
		engine.Log.Info("Illegal block start block hash")
		return false
	}
	if !bytes.Equal(sbh, config.StartBlockHash) {
		engine.Log.Info("Illegal block start block hash %s", hex.EncodeToString(sbh[:startBlockHashLength]))
		return false
	}
	for i, one := range this.Txs {
		one.BuildHash()
		if !bytes.Equal(this.BH.Tx[i], *one.GetHash()) {
			//区块不合法
			engine.Log.Info("Illegal block")
			return false
		}
	}
	return true

}

/*
创建
*/
func CreateBlockHeadVO(sbh []byte, bh *BlockHead, txs []TxItr) *BlockHeadVO {
	bhvo := BlockHeadVO{
		StaretBlockHash: sbh, //
		BH:              bh,  //
		Txs:             txs, //交易明细
	}
	return &bhvo
}

type BlockHeadVOParse struct {
	StaretBlockHash []byte        `json:"sbh"` //创始区块hash
	BH              *BlockHead    `json:"bh"`  //区块
	Txs             []interface{} `json:"txs"` //交易明细
	// BM  *BackupMiners `json:"bm"`  //见证人投票结果
}

/*
	解析区块头
*/
// func ParseBlockHeadVO(bs *[]byte) (*BlockHeadVO, error) {
// 	bh := new(BlockHeadVOParse)

// 	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 	decoder.UseNumber()
// 	err := decoder.Decode(bh)

// 	if err != nil {
// 		return nil, err
// 	}

// 	txitrs := make([]TxItr, 0)
// 	for _, one := range bh.Txs {

// 		oneMap := one.(map[string]interface{})
// 		bs, err := json.Marshal(oneMap)
// 		if err != nil {
// 			return nil, err
// 		}
// 		txitr, err := ParseTxBase(0, &bs)
// 		if err != nil {
// 			return nil, err
// 		}
// 		txitrs = append(txitrs, txitr)
// 	}
// 	bhvo := BlockHeadVO{
// 		StaretBlockHash: bh.StaretBlockHash, //
// 		BH:              bh.BH,              //区块
// 		Txs:             txitrs,             //交易明细
// 	}
// 	return &bhvo, nil
// }
/*
	解析区块头
*/
func ParseBlockHeadVOProto(bs *[]byte) (*BlockHeadVO, error) {
	if bs == nil {
		return nil, nil
	}
	bhatp := new(go_protos.BlockHeadAndTxs)
	err := proto.Unmarshal(*bs, bhatp)
	if err != nil {
		return nil, err
	}

	txs := make([]TxItr, 0)
	for i, one := range bhatp.TxBs {
		tx, err := ParseTxBaseProto(ParseTxClass(bhatp.Bh.Tx[i]), &one)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	if bhatp.Bh == nil {
		return nil, errors.New("BlockHeadVO bh is nil")
	}

	bh := BlockHead{
		Hash:              bhatp.Bh.Hash,
		Height:            bhatp.Bh.Height,
		GroupHeight:       bhatp.Bh.GroupHeight,
		GroupHeightGrowth: bhatp.Bh.GroupHeightGrowth,
		Previousblockhash: bhatp.Bh.Previousblockhash,
		Nextblockhash:     bhatp.Bh.Nextblockhash,
		NTx:               bhatp.Bh.NTx,
		MerkleRoot:        bhatp.Bh.MerkleRoot,
		Tx:                bhatp.Bh.Tx,
		Time:              bhatp.Bh.Time,
		Witness:           bhatp.Bh.Witness,
		Sign:              bhatp.Bh.Sign,
		ExtSign:           bhatp.Bh.ExtSign,
	}

	return CreateBlockHeadVO(bhatp.StaretBlockHash, &bh, txs), nil

}

func ParseBlockHeadVOProtoMore(bs *[]byte) ([]*BlockHeadVO, error) {
	if bs == nil {
		return nil, nil
	}
	bhsatp := new(go_protos.RepeatedBlockHeadAndTxs)
	err := proto.Unmarshal(*bs, bhsatp)
	if err != nil {
		return nil, err
	}

	bhs := make([]*BlockHeadVO, 0)

	for _, bh := range bhsatp.Bhat {
		bhB, err := bh.Marshal()
		if err != nil {
			return nil, err
		}

		b, err := ParseBlockHeadVOProto(&bhB)
		if err != nil {
			return nil, err
		}

		bhs = append(bhs, b)
	}

	return bhs, err
}

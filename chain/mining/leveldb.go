package mining

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/protos/go_protos"
	chainbootConfig "web3_gui/chain_boot/config"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/adapter/engine"
	pgo_protos "web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/utils"
)

/*
通过区块hash从数据库加载区块头
*/
func LoadBlockHeadByHash(hash *[]byte) (*BlockHead, error) {
	bhashkey := config.BuildBlockHead(*hash)
	bh, err := db.LevelDB.Find(bhashkey)
	if err != nil {
		return nil, err
	}
	if bh == nil || len(*bh) == 0 {
		return nil, config.ERROR_db_not_exist
	}
	return ParseBlockHeadProto(bh)
}

/*
通过区块高度，查询一个区块头信息
*/
func LoadBlockHeadByHeight(height uint64) *BlockHead {
	//bhash, err := db.LevelDB.Find([]byte(config.BlockHeight + strconv.Itoa(int(height))))
	bhash, err := db.LevelDB.Find(append(config.BlockHeight, []byte(strconv.Itoa(int(height)))...))
	if err != nil {
		return nil
	}
	if bhash == nil || len(*bhash) == 0 {
		return nil
	}
	bh, err := LoadBlockHeadByHash(bhash)
	// bs, err := db.Find(*bhash)
	// if err != nil {
	// 	return nil
	// }
	// bh, err := ParseBlockHead(bs)
	if err != nil {
		return nil
	}
	return bh
}

//func LoadBlockHeadByHeightMore(sHeight, eHeight uint64) (list []*BlockHead, err error) {
//	for h := sHeight; h < eHeight; h++ {
//		bh := LoadBlockHeadByHeight(h)
//		if bh == nil {
//			break
//		}
//		list = append(list, bh)
//		if bh.Nextblockhash == nil || len(bh.Nextblockhash) <= 0 {
//			break
//		}
//	}
//	return list, nil
//}

/*
LoadBlockHeadByHeightMore通过区块高度区间，区块头信息
@Description:
@param sHeight 开始高度
@param count 查询总条数
@return list
@return err
*/
func LoadBlockHeadByHeightMore(sHeight, count uint64, sBhash []byte) (list []*BlockHead, err error) {
	//iter := db.LevelDB.NewIterator()
	findHash := sBhash
	for i := uint64(0); i < count; i++ {
		//先找根据高度找到区块头hash
		//hkey := db.LevelDB.GetEncodeKVKey(findHash)
		bhashkey := config.BuildBlockHead(findHash)
		bh, _ := db.LevelDB.Find(bhashkey)
		//bh := iter.Find(hkey)

		// var bkey []byte
		// if bhash != nil && len(bhash) > 0 {
		// 	//验证开始的高度区块hash是否正确
		// 	if sHeight && !bytes.Equal(bhash, sBhash) {
		// 		break
		// 	}
		// 	//再根据区块头hash找到区块头信息
		// 	bkey = db.LevelDB.GetEncodeKVKey(bhash)
		// } else if h == sHeight {
		// 	//engine.Log.Info("根据高度没找到区块hash，直接通过hash去找 error sHeight:%d", sHeight)
		// 	bkey = db.LevelDB.GetEncodeKVKey(sBhash)
		// } else {
		// 	break
		// }

		//再根据区块头hash找到区块头信息
		// bh := iter.Find(bkey)
		if bh == nil || len(*bh) == 0 {
			break
		}
		bhvo, er := ParseBlockHeadProto(bh)
		if er != nil {
			err = er
			break
		}
		if sHeight != bhvo.Height {
			//engine.Log.Info("高度不匹配 error sHeight:%d", sHeight)
			break
		}
		list = append(list, bhvo)
		if bhvo.Nextblockhash == nil || len(bhvo.Nextblockhash) == 0 {
			break
		}
		sHeight++
		findHash = bhvo.Nextblockhash
	}
	//iter.Release()
	if err != nil {
		return
	}
	//err = iter.Error()
	return
}

/*
查询某高度的区块hash
*/
func LoadBlockHashByHeight(height uint64) *[]byte {
	//bhash, err := db.LevelDB.Find([]byte(config.BlockHeight + strconv.Itoa(int(height))))
	bhash, err := db.LevelDB.Find(append(config.BlockHeight, []byte(strconv.Itoa(int(height)))...))
	if err != nil {
		return nil
	}
	if bhash == nil || len(*bhash) == 0 {
		return nil
	}
	return bhash
}

/*
查询数据库和解析交易
*/
func LoadTxBase(txid []byte) (TxItr, error) {
	var err error
	var txItr TxItr
	ok := false
	//是否启用缓存
	txhashkey := config.BuildBlockTx(txid)
	if config.EnableCache {
		//先判断缓存中是否存在
		// txItr, ok = TxCache.FindTxInCache(hex.EncodeToString(txid))
		txItr, ok = TxCache.FindTxInCache(txhashkey)
	}
	if !ok {
		// engine.Log.Error("未命中缓存 FindTxInCache")
		var bs *[]byte
		bs, err = db.LevelDB.Find(txhashkey)
		if err != nil {
			// engine.Log.Info("find block error:%s", err.Error())
			return nil, err
		}
		if bs == nil || len(*bs) == 0 {
			// engine.Log.Info("find block error:%s", config.ERROR_db_not_exist.Error())
			return nil, config.ERROR_db_not_exist
		}
		txItr, err = ParseTxBaseProto(ParseTxClass(txid), bs)
		if err != nil {
			// engine.Log.Info("find block error:%s", err.Error())
			return nil, err
		}
	}
	// engine.Log.Info("find block error:%v", err)
	return txItr, err
}

/*
从数据库中加载一个区块
*/
func loadBlockForDB(bhash *[]byte) (*BlockHead, []TxItr, error) {

	hB, err := LoadBlockHeadByHash(bhash)
	if err != nil {
		return nil, nil, err
	}
	txItrs := make([]TxItr, 0)
	for _, one := range hB.Tx {
		// txItr, err := FindTxBase(one, hex.EncodeToString(one))
		txItr, err := LoadTxBase(one)

		// txBs, err := db.Find(one)
		if err != nil {
			// fmt.Println("3333", err)
			return nil, nil, err
		}
		// txItr, err := ParseTxBase(ParseTxClass(one), txBs)
		txItrs = append(txItrs, txItr)
	}

	return hB, txItrs, nil
}

/*
从数据库加载一整个区块，包括区块中的所有交易
*/
func LoadBlockHeadVOByHash(hash *[]byte) (*BlockHeadVO, error) {
	bh, txs, err := loadBlockForDB(hash)
	if err != nil {
		return nil, err
	}

	bhvo := new(BlockHeadVO)
	bhvo.Txs = make([]TxItr, 0)

	bhvo.BH = bh
	bhvo.Txs = txs
	return bhvo, nil
}

/*
获取社区节点起始高度
*/
func GetCommunityAddrStartHeight() {

}

/*
获取社区投票锁定奖励是否存在
*/
func ExistCommunityVoteRewardFrozen(addr *crypto.AddressCoin) bool {
	dbkey := config.BuildDBKeyCommunityAddrFrozen(*addr)
	ok, err := db.LevelTempDB.CheckHashExist(*dbkey)
	if err != nil {
		return false
	}
	return ok
}

/*
获取社区投票锁定奖励
*/
func GetCommunityVoteRewardFrozen(addr *crypto.AddressCoin) uint64 {
	dbkey := config.BuildDBKeyCommunityAddrFrozen(*addr)
	valueBs, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return 0
	}
	return utils.BytesToUint64(*valueBs)
}

/*
设置社区投票锁定奖励
*/
func SetCommunityVoteRewardFrozen(addr *crypto.AddressCoin, value uint64) {
	dbkey := config.BuildDBKeyCommunityAddrFrozen(*addr)
	valueNewBs := utils.Uint64ToBytes(value)
	db.LevelTempDB.Save(*dbkey, &valueNewBs)
}

/*
设置轻节点获取奖励
*/
func SetLightNodeVoteReward(lightAddr, communityAddr *crypto.AddressCoin, value uint64) {
	reward, err := GetLightNodeVoteReward(lightAddr, communityAddr)
	if err != nil {
		engine.Log.Warn("SetLightNodeVoteReward err:%v", err)
		return
	}
	dbkey := config.BuildDBKeyLightVoteReward(*lightAddr, *communityAddr)
	valueNewBs := utils.Uint64ToBytes(value + reward)
	db.LevelTempDB.Save(*dbkey, &valueNewBs)
}

/*
获取轻节点获取奖励
*/
func GetLightNodeVoteReward(lightAddr, communityAddr *crypto.AddressCoin) (uint64, error) {
	dbkey := config.BuildDBKeyLightVoteReward(*lightAddr, *communityAddr)
	reward, err := db.LevelTempDB.Get(*dbkey)
	if err != nil {
		return 0, err
	}
	return utils.BytesToUint64(reward), nil
}

/*
查询当前高度某地址是否是社区节点
*/
func ExistCommunityAddr(addr *crypto.AddressCoin) bool {
	dbkey := config.BuildDBKeyCommunityAddr(*addr)
	ok, err := db.LevelTempDB.CheckHashExist(*dbkey)
	// _, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return false
	}
	return ok
}

/*
设置一个地址为社区节点
*/
func SetCommunityAddr(addr *crypto.AddressCoin) {
	dbkey := config.BuildDBKeyCommunityAddr(*addr)
	err := db.LevelTempDB.Save(*dbkey, nil)
	if err != nil {

	}
}

/*
删除一个社区节点地址
*/
func RemoveCommunityAddr(addr *crypto.AddressCoin) {
	dbkey := config.BuildDBKeyCommunityAddr(*addr)
	err := db.LevelTempDB.Remove(*dbkey)
	if err != nil {

	}
}

var notSpendBalanceLock = new(sync.RWMutex)
var notSpendBalance = make(map[string]uint64)

/*
获取一个地址的可用余额
*/
func GetNotSpendBalance(addr *crypto.AddressCoin) (*TxItem, uint64) {
	dbkey := config.BuildAddrValue(*addr)
	item := TxItem{
		Addr:  addr,
		Value: 0,
	}
	notSpendBalanceLock.Lock()
	value, _ := notSpendBalance[utils.Bytes2string(dbkey)]
	item.Value = value
	notSpendBalanceLock.Unlock()
	return &item, value

	//dbkey := config.BuildAddrValue(*addr)
	//valueBs, err := db.LevelTempDB.Find(dbkey)
	//if err != nil {
	//	return nil, 0
	//}
	//value := utils.BytesToUint64(*valueBs)
	//item := TxItem{
	//	Addr:  addr,
	//	Value: value,
	//}
	//return &item, value
}

/*
*
批量查询账户余额 缓存
*/
func GetNotSpendBalances(addrs []crypto.AddressCoin) []TxItem {
	if len(addrs) == 0 {
		return nil
	}

	notSpendBalanceLock.Lock()
	defer notSpendBalanceLock.Unlock()

	r := make([]TxItem, len(addrs))
	for k, v := range addrs {
		dbkey := config.BuildAddrValue(v)
		r[k].Addr = &addrs[k]
		value, _ := notSpendBalance[utils.Bytes2string(dbkey)]
		r[k].Value = value
	}

	return r
}

/*
*
批量设置余额 缓存
*/
func SetNotSpendBalances(balances map[string]uint64) {
	if len(balances) == 0 {
		return
	}

	//大余额的地址集合
	setBigAddrs := make([]db2.KVPair, 0)

	notSpendBalanceLock.Lock()

	for k, v := range balances {
		addr := crypto.AddressCoin(k)
		if v > config.Mining_coin_total {
			engine.Log.Info("大余额:%d", v)
			setBigAddrs = append(setBigAddrs, db2.KVPair{
				Key:   config.BuildAddrValueBig(addr),
				Value: nil,
			})
		}

		notSpendBalance[utils.Bytes2string(config.BuildAddrValue(addr))] = v
	}

	notSpendBalanceLock.Unlock()

	LedisMultiSaves(setBigAddrs)
}

/*
设置一个地址的可用余额
*/
func SetNotSpendBalance(addr *crypto.AddressCoin, value uint64) error {
	if value > config.Mining_coin_total {
		engine.Log.Info("大余额:%d", value)
		SetAddrValueBig(addr)
	}

	dbkey := config.BuildAddrValue(*addr)
	notSpendBalanceLock.Lock()
	notSpendBalance[utils.Bytes2string(dbkey)] = value
	notSpendBalanceLock.Unlock()
	return nil

	valueNewBs := utils.Uint64ToBytes(value)
	return db.LevelTempDB.Save(dbkey, &valueNewBs)
}

/*
	从本地钱包中查询一个有足够余额的地址，返回地址及余额
*/
// func FindNotSpendBalance(amount uint64) (*TxItem, uint64) {
// 	addrs := keystore.GetAddrAll()
// 	for _, one := range addrs {
// 		item, value := GetNotSpendBalance(&one.Addr)
// 		if amount <= value {
// 			return item, value
// 		}
// 	}
// 	return nil, 0
// }

/*
查询一个地址给谁投票了
*/
func GetVoteAddr(addr *crypto.AddressCoin) *crypto.AddressCoin {
	dbkey := config.BuildDBKeyVoteAddr(*addr)
	bs, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return nil
	}
	if bs == nil || len(*bs) <= 0 {
		return nil
	}
	voteAddr := crypto.AddressCoin(*bs)
	return &voteAddr
}

/*
设置一个地址给谁投票了
*/
func SetVoteAddr(addr, voteAddr *crypto.AddressCoin) error {
	dbkey := config.BuildDBKeyVoteAddr(*addr)
	bs := []byte(*voteAddr)
	return db.LevelTempDB.Save(*dbkey, &bs)
}

/*
删除一个地址给谁投票了
*/
func RemoveVoteAddr(addr *crypto.AddressCoin) error {
	dbkey := config.BuildDBKeyVoteAddr(*addr)
	return db.LevelTempDB.Remove(*dbkey)
}

/*
查询见证人押金
*/
func GetDepositWitnessAddr(addr *crypto.AddressCoin) uint64 {
	dbkey := config.BuildDBKeyDepositWitnessAddr(*addr)
	valueBs, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return 0
	}
	value := utils.BytesToUint64(*valueBs)
	return value
}

/*
设置见证人押金
*/
func SetDepositWitnessAddr(addr *crypto.AddressCoin, value uint64) error {
	dbkey := config.BuildDBKeyDepositWitnessAddr(*addr)
	valueNewBs := utils.Uint64ToBytes(value)
	return db.LevelTempDB.Save(*dbkey, &valueNewBs)
}

/*
删除见证人押金
*/
func RemoveDepositWitnessAddr(addr *crypto.AddressCoin) error {
	dbkey := config.BuildDBKeyDepositWitnessAddr(*addr)
	return db.LevelTempDB.Remove(*dbkey)
}

/*
查询轻节点押金
*/
func GetDepositLightAddr(addr *crypto.AddressCoin) uint64 {
	dbkey := config.BuildDBKeyDepositLightAddr(*addr)
	valueBs, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return 0
	}
	value := utils.BytesToUint64(*valueBs)
	return value
}

/*
设置轻节点押金
*/
func SetDepositLightAddr(addr *crypto.AddressCoin, value uint64) error {
	dbkey := config.BuildDBKeyDepositLightAddr(*addr)
	valueNewBs := utils.Uint64ToBytes(value)
	return db.LevelTempDB.Save(*dbkey, &valueNewBs)
}

/*
删除轻节点押金
*/
func RemoveDepositLightAddr(addr *crypto.AddressCoin) error {
	dbkey := config.BuildDBKeyDepositLightAddr(*addr)
	return db.LevelTempDB.Remove(*dbkey)
}

/*
查询社区节点押金
*/
func GetDepositCommunityAddr(addr *crypto.AddressCoin) uint64 {
	dbkey := config.BuildDBKeyDepositCommunityAddr(*addr)
	valueBs, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return 0
	}
	value := utils.BytesToUint64(*valueBs)
	return value
}

/*
设置社区节点押金
*/
func SetDepositCommunityAddr(addr *crypto.AddressCoin, value uint64) error {
	dbkey := config.BuildDBKeyDepositCommunityAddr(*addr)
	valueNewBs := utils.Uint64ToBytes(value)
	return db.LevelTempDB.Save(*dbkey, &valueNewBs)
}

/*
删除社区节点押金
*/
func RemoveDepositCommunityAddr(addr *crypto.AddressCoin) error {
	dbkey := config.BuildDBKeyDepositCommunityAddr(*addr)
	return db.LevelTempDB.Remove(*dbkey)
}

/*
查询轻节点投票总金额
*/
func GetDepositLightVoteValue(lightAddr, communityAddr *crypto.AddressCoin) uint64 {
	dbkey := config.BuildDBKeyDepositLightVoteValue(*lightAddr, *communityAddr)
	valueBs, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return 0
	}
	value := utils.BytesToUint64(*valueBs)
	return value
}

/*
设置轻节点投票总金额
*/
func SetDepositLightVoteValue(lightAddr, communityAddr *crypto.AddressCoin, value uint64) error {
	dbkey := config.BuildDBKeyDepositLightVoteValue(*lightAddr, *communityAddr)
	valueNewBs := utils.Uint64ToBytes(value)
	return db.LevelTempDB.Save(*dbkey, &valueNewBs)
}

/*
删除轻节点投票总金额
*/
func RemoveDepositLightVoteValue(lightAddr, communityAddr *crypto.AddressCoin) error {
	dbkey := config.BuildDBKeyDepositLightVoteValue(*lightAddr, *communityAddr)
	return db.LevelTempDB.Remove(*dbkey)
}

var (
	addrNonceLock = new(sync.RWMutex)
	addrNonce     = make(map[string]big.Int)
)

//region nonce cache

/*
*
查询地址的nonce cache
*/
func GetAddrNonceOld(addr *crypto.AddressCoin) (big.Int, error) {
	addrNonceLock.Lock()
	defer addrNonceLock.Unlock()

	value, _ := addrNonce[utils.Bytes2string(*addr)]

	return value, nil
}

/*
GetAddrNonceMore
@Description: 查询多个地址的nonce
@param addrStrs
@return map[string]uint64
*/
func GetAddrNonceMore(addrStrs []interface{}) map[string]uint64 {
	addrNonceLock.Lock()
	defer addrNonceLock.Unlock()
	addrNonceMap := make(map[string]uint64, len(addrStrs))
	for i, _ := range addrStrs {
		addrStr, ok := addrStrs[i].(string)
		if !ok {
			addrNonceMap[addrStr] = 0
			continue
		}
		addrMul := crypto.AddressFromB58String(addrStr)
		if addrMul == nil || len(addrMul) == 0 {
			addrNonceMap[addrStr] = 0
			continue
		}
		value, _ := addrNonce[utils.Bytes2string(addrMul)]
		addrNonceMap[addrStr] = value.Uint64()
	}
	return addrNonceMap
}

/*
@Description: 查询多个地址的nonce
@param addrStrs
@return map[string]uint64
*/
func GetAddressNonceMore(addrs []coin_address.AddressCoin) ([]uint64, utils.ERROR) {
	addrNonceLock.Lock()
	defer addrNonceLock.Unlock()
	nonces := make([]uint64, 0, len(addrs))
	for _, one := range addrs {
		value, ok := addrNonce[utils.Bytes2string(one)]
		if !ok {
			return nil, utils.NewErrorBus(chainbootConfig.ERROR_CODE_CHAIN_address_not_exist, "")
		}
		nonces = append(nonces, value.Uint64())
	}
	return nonces, utils.NewErrorSuccess()
}

/*
设置地址的nonce cache
*/
func SetAddrNonceOld(addr *crypto.AddressCoin, value *big.Int) error {
	addrNonceLock.Lock()
	defer addrNonceLock.Unlock()

	addrNonce[utils.Bytes2string(*addr)] = *value

	return nil
}

//endregion

/*
查询地址的nonce
*/
func GetAddrNonce(addr *crypto.AddressCoin) (big.Int, error) {
	dbkey := config.BuildDBKeyAddrNonce(*addr)
	valueBs, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return big.Int{}, err
	}
	nonce := new(big.Int).SetBytes(*valueBs)
	return *nonce, nil
}

/*
设置地址的nonce
*/
func SetAddrNonce(addr *crypto.AddressCoin, value *big.Int) error {
	dbkey := config.BuildDBKeyAddrNonce(*addr)
	valueNewBs := value.Bytes()
	return db.LevelTempDB.Save(*dbkey, &valueNewBs)
}

/*
设置一个地址的可用余额Token
*/
func SetTokenNotSpendBalance(txid *[]byte, addr *crypto.AddressCoin, value *big.Int) error {
	dbkey := config.BuildDBKeyTokenAddrValue(*txid, *addr)
	bs := value.Bytes()
	return db.LevelTempDB.Save(*dbkey, &bs)
}

/*
获取一个地址的可用余额Token
*/
func getTokenNotSpendBalance(txid *[]byte, addr *crypto.AddressCoin) *big.Int {
	dbkey := config.BuildDBKeyTokenAddrValue(*txid, *addr)
	valueBs, err := db.LevelTempDB.Find(*dbkey)
	if err != nil {
		return big.NewInt(0)
	}

	return new(big.Int).SetBytes(*valueBs)
}

/*
更新一个地址的可用余额Token
*/
func updateTokenNotSpendBalance(key *[]byte, value *big.Int) error {
	newValue := big.NewInt(0)
	valueBs, err := db.LevelDB.Find(*key)
	if err == nil {
		newValue.SetBytes(*valueBs)
	}

	newValue.Add(newValue, value)
	bs := newValue.Bytes()
	return db.LevelTempDB.Save(*key, &bs)
}

var frozenBalanceLock = new(sync.RWMutex)
var frozenBalance = make(map[string]uint64)

/*
查询地址锁定余额
*/
func GetAddrFrozenValue(addr *crypto.AddressCoin) uint64 {
	dbkey := config.BuildAddrFrozen(*addr)
	valueBs, err := db.LevelTempDB.Find(dbkey)
	if err != nil {
		return 0
	}
	value := utils.BytesToUint64(*valueBs)
	return value
}

/*
设置地址锁定余额
*/
func SetAddrFrozenValue(addr *crypto.AddressCoin, value uint64) error {
	dbkey := config.BuildAddrFrozen(*addr)
	valueNewBs := utils.Uint64ToBytes(value)
	return db.LevelTempDB.Save(dbkey, &valueNewBs)
}

/*
删除地址锁定余额
*/
func RemoveAddrFrozenValue(addr *crypto.AddressCoin) error {
	dbkey := config.BuildAddrFrozen(*addr)
	return db.LevelTempDB.Remove(dbkey)
}

/*
	查询地址锁定余额
*/
// func GetLightVoteLockValue(lightAddr, communityAddr *crypto.AddressCoin) uint64 {
// 	dbkey := config.BuildDBKeyAddrVoteLock(*lightAddr, *communityAddr)
// 	valueBs, err := db.LevelTempDB.Find(*dbkey)
// 	if err != nil {
// 		return 0
// 	}
// 	value := utils.BytesToUint64(*valueBs)
// 	return value
// }

// /*
// 	设置地址锁定余额
// */
// func SetLightVoteLockValue(lightAddr, communityAddr *crypto.AddressCoin, value uint64) error {
// 	dbkey := config.BuildDBKeyAddrVoteLock(*lightAddr, *communityAddr)
// 	valueNewBs := utils.Uint64ToBytes(value)
// 	return db.LevelTempDB.Save(*dbkey, &valueNewBs)
// }

// /*
// 	删除地址锁定余额
// */
// func RemoveLightVoteLockValue(lightAddr, communityAddr *crypto.AddressCoin) error {
// 	dbkey := config.BuildDBKeyAddrVoteLock(*lightAddr, *communityAddr)
// 	return db.LevelTempDB.Remove(*dbkey)
// }

/*
是否在列表中
@return    bool    true=在列表中;false=不在列表中;
*/
func GetAddrValueBig(addr *crypto.AddressCoin) bool {
	dbkey := config.BuildAddrValueBig(*addr)
	ok, err := db.LevelTempDB.CheckHashExist(dbkey)
	if err != nil {
		return true
	}
	return ok
}

/*
设置地址
*/
func SetAddrValueBig(addr *crypto.AddressCoin) error {
	dbkey := config.BuildAddrValueBig(*addr)
	return db.LevelTempDB.Save(dbkey, nil)
}

//-------------------------------
/*
	给一个地址添加NFT
*/
func AddAddrNFT(owner *crypto.AddressCoin, nftID []byte) error {
	db := db.LevelTempDB.GetDB()
	dbkey := config.BuildAddrToNft(*owner)
	_, err := db.HSet(dbkey, nftID, nil)
	return err
}

/*
给一个地址删除NFT
*/
func RemoveAddrNFT(owner *crypto.AddressCoin, nftID []byte) error {
	db := db.LevelTempDB.GetDB()
	dbkey := config.BuildAddrToNft(*owner)
	_, err := db.HDel(dbkey, nftID)
	return err
}

/*
查询一个地址的所有NFT
*/
func GetAddrNFT(owner *crypto.AddressCoin) ([]*Tx_nft, error) {
	dbConn := db.LevelTempDB.GetDB()
	dbkey := config.BuildAddrToNft(*owner)
	kv, err := dbConn.HGetAll(dbkey)
	if err != nil {
		return nil, err
	}

	nfts := make([]*Tx_nft, 0, len(kv))
	for _, one := range kv {
		// engine.Log.Info("field:%s value:%s", hex.EncodeToString(one.Field), hex.EncodeToString(one.Value))
		bs, err := db.LevelDB.Find(one.Field)
		if err != nil {
			return nil, err
		}
		txItr, err := ParseTxBaseProto(ParseTxClass(one.Field), bs)
		if err != nil {
			return nil, err
		}
		nfts = append(nfts, txItr.(*Tx_nft))
	}
	return nfts, nil
}

//--------------------------
/*
	保存一个NFT属于哪个地址
*/
func AddNFTOwner(nftID []byte, owner *crypto.AddressCoin) error {
	dbkey := config.BuildNftOwner(nftID)
	ownerBs := []byte(*owner)
	return db.LevelTempDB.Save(dbkey, &ownerBs)
}

/*
删除一个NFT属于哪个地址
*/
func RemoveNFTOwner(nftID []byte) error {
	dbkey := config.BuildNftOwner(nftID)
	return db.LevelTempDB.Remove(dbkey)
}

/*
查询一个NFT属于哪个地址
*/
func FindNFTOwner(nftID []byte) (*crypto.AddressCoin, error) {
	dbkey := config.BuildNftOwner(nftID)
	bs, err := db.LevelTempDB.Find(dbkey)
	if err != nil {
		return nil, err
	}
	owner := crypto.AddressCoin(*bs)
	return &owner, nil
}

// 保存合约源代码
func SetContractSource(source []byte, address crypto.AddressCoin) error {
	dbKey := config.BuildContractSource(address)
	return db.LevelDB.Save(dbKey, &source)
}

// 获取合约源代码
func GetContractSource(address crypto.AddressCoin) ([]byte, error) {
	dbKey := config.BuildContractSource(address)
	value, err := db.LevelDB.Find(dbKey)
	if err != nil {
		return nil, err
	}
	return *value, nil
}

// 保存合约abi
func SetContractAbi(abi []byte, address crypto.AddressCoin) error {
	dbKey := config.BuildContractAbi(address)
	return db.LevelDB.Save(dbKey, &abi)
}

// 获取合约abi
func GetContractAbi(address crypto.AddressCoin) ([]byte, error) {
	dbKey := config.BuildContractAbi(address)
	value, err := db.LevelDB.Find(dbKey)
	if err != nil {
		return nil, err
	}
	return *value, nil
}

// 保存合约bin
func SetContractBin(bin []byte, address crypto.AddressCoin) error {
	dbKey := config.BuildContractBin(address)
	return db.LevelDB.Save(dbKey, &bin)
}

// 获取合约bin
func GetContractBin(address crypto.AddressCoin) ([]byte, error) {
	dbKey := config.BuildContractBin(address)
	value, err := db.LevelDB.Find(dbKey)
	if err != nil {
		return nil, err
	}
	return *value, nil
}

// 获取合约交易的gasused
func GetContractTxGasUsed(txHash []byte) uint64 {
	dbKey := append([]byte(config.DBKEY_TxCONTRACT_GASUSED), txHash...)
	value, err := db.LevelDB.Find(dbKey)
	if err != nil {
		return 0
	}
	return utils.BytesToUint64(*value)
}

// 保存特定类型合约
func SetSpecialContract(addr, contract crypto.AddressCoin, class uint64) error {
	dbkey := append([]byte(config.DBKEY_SPECIAL_CONTRACT), addr...)
	_, err := db.LevelDB.HSet(dbkey, utils.Uint64ToBytes(class), contract)
	return err
}

// 获取特定类型合约
func GetSpecialContract(addr crypto.AddressCoin, class uint64) (crypto.AddressCoin, error) {
	dbkey := append([]byte(config.DBKEY_SPECIAL_CONTRACT), addr...)
	value, err := db.LevelDB.HGet(dbkey, utils.Uint64ToBytes(class))
	if err != nil {
		return nil, err
	}
	return crypto.AddressCoin(value), nil
}

// 搜索erc20合约
func SearchErc20Contract(keyWord string) ([]map[string]interface{}, error) {
	list := []map[string]interface{}{}
	r := util.BytesPrefix([]byte(config.DBKEY_CONTRACT_OBJECT))
	it := db.LevelTempDB.NewIteratorWithOption(r, nil)
	for it.Next() {
		key := it.Key()
		key, _ = db.LevelTempDB.GetDB().DecodeKVKey(key)

		obj := new(go_protos.StateObject)
		proto.Unmarshal(it.Value(), obj)
		if len(obj.Code) > 0 && !obj.Suicide {

			//有效合约,再验证是否是ERC20
			start := len([]byte(config.DBKEY_CONTRACT_OBJECT))
			contractAddr := crypto.AddressCoin(key[start:])
			name, symbol, checkRes := precompiled.IsErc20(config.Area.Keystore.GetCoinbase().Addr.B58String(), contractAddr.B58String())
			if checkRes {
				if name == keyWord || contractAddr.B58String() == keyWord {
					list = append(list, map[string]interface{}{
						"name":    name,
						"symbol":  symbol,
						"address": contractAddr.B58String(),
					})
				}

			}
		}
	}
	it.Release()
	err := it.Error()
	return list, err
}

func SetMessageMulticast(hash []byte, msg *pgo_protos.MessageMulticast) error {
	dbkey := append([]byte(config.DBKEY_Message_Multicast), hash...)
	content, err := msg.Marshal()
	if err != nil {
		fmt.Println("############## err111", err.Error())
		return err
	}
	err = db.LevelTempDB.Save(dbkey, &content)
	if err != nil {
		fmt.Println("############## err222 -- ", err.Error(), len(dbkey))
		return err
	}
	return err
}

func GetMessageMulticast(hash []byte) (*pgo_protos.MessageMulticast, error) {
	dbkey := append([]byte(config.DBKEY_Message_Multicast), hash...)
	bs, err := db.LevelTempDB.Get(dbkey)

	mmp := new(pgo_protos.MessageMulticast)
	if err != nil {
		return mmp, err
	}

	if bs == nil {
		return nil, errors.New("not found")
	}

	err = proto.Unmarshal(bs, mmp)
	if err != nil {
		return mmp, err
	}
	return mmp, nil
}

func GetFaucetTime(address string) int64 {
	//dbkey := []byte(config.FAUCET_ADDRESS + address)
	dbkey := append(config.FAUCET_ADDRESS, []byte(address)...)
	bs, err := db.LevelDB.Get(dbkey)
	if err != nil {
		return 0
	}
	t := utils.BytesToInt64(bs)
	return t
}
func SetFaucetTime(address string) error {
	//dbkey := []byte(config.FAUCET_ADDRESS + address)
	dbkey := append(config.FAUCET_ADDRESS, []byte(address)...)
	t := utils.Int64ToBytes(config.TimeNow().Unix())

	err := db.LevelDB.Set(dbkey, t)
	return err
}

// 通过区块高度保存区块bloom
func SaveBlockHeadBloom(height uint64, bloom []byte) error {
	dbkey := append(config.DBKEY_blockhead_bloom, Uint64ToBytes(height)...)
	return db.LevelDB.Set(dbkey, bloom)
}

// 通过区块高度获取区块bloom
func GetBlockHeadBloom(height uint64) ([]byte, error) {
	dbkey := append(config.DBKEY_blockhead_bloom, Uint64ToBytes(height)...)
	return db.LevelDB.Get(dbkey)
}

// 过滤区块头Bloom包含主题的区块集合
func FilterBlockHeadBloomTopic(start, end uint64, topic []byte) []uint64 {
	startKey := append(config.DBKEY_blockhead_bloom, Uint64ToBytes(start)...)
	endKey := append(config.DBKEY_blockhead_bloom, Uint64ToBytes(end)...)
	s := db.LevelDB.GetEncodeKVKey(startKey)
	e := db.LevelDB.GetEncodeKVKey(endKey)

	kvs := db.LevelDB.GetSnap().GetInRange(s, e, db2.RangeClose)

	r := make([]uint64, 0)
	for k := range kvs {
		if BytesToBloom(kvs[k]).Check(topic) {
			key, _ := db.LevelDB.GetDB().DecodeKVKey(db.LevelDB.GetSnap().DecodeKey(k))
			key = key[len(config.DBKEY_blockhead_bloom):]
			r = append(r, BytesToUint64(key))
		}
	}
	sort.SliceStable(r, func(i, j int) bool {
		return r[i] < r[j]
	})

	return r
}

func Uint64ToBytes(n uint64) []byte {
	bytesBuffer := bytes.NewBuffer(nil)
	binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}

func BytesToUint64(b []byte) uint64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint64
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return tmp
}

// 获取交易bloom
func GetBloomByTx(txid []byte) []byte {
	key := append(config.DBKEY_tx_bloom, txid...)
	bs, err := db.LevelDB.Find(key)
	if err != nil {
		return nil
	}
	return *bs
}

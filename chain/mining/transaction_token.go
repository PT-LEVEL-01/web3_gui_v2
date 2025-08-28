package mining

import (
	"encoding/hex"
	"math/big"
	"sort"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/crypto"
)

/*
保存token详细信息
*/
func SaveTokenInfo(owner crypto.AddressCoin, tokenid []byte, name, symbol string, accuracy uint64, supply *big.Int) error {
	tokeninfo := &go_protos.TokenInfo{
		TokenId:       tokenid,
		TokenName:     name,
		TokenSymbol:   symbol,
		TokenSupply:   supply.Bytes(),
		TokenAccuracy: accuracy,
		Owner:         owner,
	}

	bs, err := tokeninfo.Marshal()
	if err != nil {
		return err
	}
	return db.LevelTempDB.Save(config.BuildKeyForPublishToken(tokenid), &bs)
}

/*
查询token信息信息
*/
func FindTokenInfo(tokenid []byte) (*go_protos.TokenInfo, error) {
	bs, err := db.LevelTempDB.Find(config.BuildKeyForPublishToken(tokenid))
	if err != nil {
		return nil, err
	}
	tokeninfo := &go_protos.TokenInfo{}
	err = tokeninfo.Unmarshal(*bs)
	if err != nil {
		return nil, err
	}

	return tokeninfo, err
}

/*
查询token列表
*/
func ListTokenInfos() ([]*go_protos.TokenInfo, error) {
	startKey := config.TokenInfo
	endKey := make([]byte, len(startKey))
	copy(endKey, startKey)
	for i := len(endKey) - 1; i >= 0; i-- {
		endKey[i] += 1
		if endKey[i] != 0 {
			break
		}
	}

	s := db.LevelDB.GetEncodeKVKey(startKey)
	e := db.LevelDB.GetEncodeKVKey(endKey)

	kvs := db.LevelDB.GetSnap().GetInRange(s, e, db2.RangeROpen)
	keys := make([]string, 0)
	for k, _ := range kvs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tokeninfos := []*go_protos.TokenInfo{}
	for _, key := range keys {
		if bs, ok := kvs[key]; ok {
			tokeninfo := &go_protos.TokenInfo{}
			if err := tokeninfo.Unmarshal(bs); err == nil {
				tokeninfos = append(tokeninfos, tokeninfo)
			}
		}
	}

	return tokeninfos, nil
}

/*
查询token信息信息
*/
func ToTokenInfoV0(tokeninfo *go_protos.TokenInfo) *TokenInfoV0 {
	supply := new(big.Int).SetBytes(tokeninfo.TokenSupply).Text(10)
	owner := crypto.AddressCoin(tokeninfo.Owner)
	return &TokenInfoV0{
		Owner:    owner.B58String(),
		TokenId:  hex.EncodeToString(tokeninfo.TokenId),
		Name:     tokeninfo.TokenName,
		Symbol:   tokeninfo.TokenSymbol,
		Accuracy: tokeninfo.TokenAccuracy,
		Supply:   supply,
	}
}

/*
Token地址金额
*/
type AddrToken struct {
	TokenId []byte
	Addr    crypto.AddressCoin
	Value   *big.Int
}

/*
token信息
*/
type TokenInfoV0 struct {
	Owner    string //拥有者
	TokenId  string //发布交易地址
	Name     string //名称（全称）
	Symbol   string //单位
	Accuracy uint64 //小数点之后位数
	Supply   string //发行总量
}

/*
添加token的收益
*/
func CountToken(txItemCounts *map[string]*map[string]*big.Int) {
	for txidStr, itemValue := range *txItemCounts {
		txid := []byte(txidStr)
		for addrStr, value := range *itemValue {
			addr := crypto.AddressCoin([]byte(addrStr))
			oldvalue := getTokenNotSpendBalance(&txid, &addr)
			newValue := new(big.Int).Set(oldvalue)
			newValue.Add(newValue, value)
			SetTokenNotSpendBalance(&txid, &addr, newValue)
		}
	}
}

package db

import (
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
保存商品数量和价格
*/
func Goods_SaveGoods(finfo *model.FilePrice) utils.ERROR {
	bs, err := finfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	key, ERR := utilsleveldb.BuildLeveldbKey(finfo.Hash)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return config.LevelDB.SaveMap(*config.DBKEY_sharebox_price_list, *key, *bs, nil)
}

/*
查询文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Goods_FindFilePrice(fileHash []byte) (*model.FilePrice, utils.ERROR) {
	key, ERR := utilsleveldb.BuildLeveldbKey(fileHash)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(*config.DBKEY_sharebox_price_list, *key)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	filePrice, err := model.ParseFilePrice(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return filePrice, utils.NewErrorSuccess()
}

/*
查询多个文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Goods_FindFilePriceMore(fileHash [][]byte) ([]*model.FilePrice, utils.ERROR) {

	keys := make([]utilsleveldb.LeveldbKey, 0, len(fileHash))
	for _, one := range fileHash {
		key, ERR := utilsleveldb.BuildLeveldbKey(one)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		keys = append(keys, *key)
	}

	items, err := config.LevelDB.FindMapByKeys(*config.DBKEY_sharebox_price_list, keys...)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	prices := make([]*model.FilePrice, 0, len(fileHash))
	for _, one := range items {
		filePrice, err := model.ParseFilePrice(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		prices = append(prices, filePrice)
	}
	return prices, utils.NewErrorSuccess()
}

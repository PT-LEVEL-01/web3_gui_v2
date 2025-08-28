package db

import (
	"web3_gui/chain_boot/object_beans"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
保存或修改商品价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Sharebox_server_SaveOrUpdateGoodsPrice(goods *object_beans.OrderShareboxGoods) utils.ERROR {
	goodsOld, ERR := Sharebox_server_FindGoods(goods.GoodsId)
	if ERR.CheckFail() {
		return ERR
	}
	if goodsOld == nil {
		goods.Class = object_beans.CLASS_SHAREBOX_goods_file
		goodsOld = goods
	}
	goodsOld.Price = goods.Price
	bs, err := goodsOld.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	key, ERR := utilsleveldb.BuildLeveldbKey(goodsOld.GetId())
	if !ERR.CheckSuccess() {
		return ERR
	}
	return config.LevelDB.SaveMap(*config.DBKEY_ORDERS_goods, *key, *bs, nil)
}

/*
查询商品
@goodsId    []byte]        商品id
*/
func Sharebox_server_FindGoods(goodsId []byte) (*object_beans.OrderShareboxGoods, utils.ERROR) {
	key, ERR := utilsleveldb.BuildLeveldbKey(goodsId)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(*config.DBKEY_ORDERS_goods, *key)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	itr, err := object_beans.ParseOrderShareboxGoods(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	order := itr.(*object_beans.OrderShareboxGoods)
	return order, ERR
}

/*
查询多个商品
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Sharebox_server_FindGoodsMore(fileHash [][]byte) ([]*object_beans.OrderShareboxGoods, utils.ERROR) {
	keys := make([]utilsleveldb.LeveldbKey, 0, len(fileHash))
	for _, one := range fileHash {
		key, ERR := utilsleveldb.BuildLeveldbKey(one)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		keys = append(keys, *key)
	}

	items, err := config.LevelDB.FindMapByKeys(*config.DBKEY_ORDERS_goods, keys...)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	prices := make([]*object_beans.OrderShareboxGoods, 0, len(fileHash))
	for _, one := range items {
		itr, err := object_beans.ParseOrderShareboxGoods(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		goods := itr.(*object_beans.OrderShareboxGoods)
		prices = append(prices, goods)
	}
	return prices, utils.NewErrorSuccess()
}

/*
保存未支付订单
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Sharebox_server_SaveOrderUnpaid(orderInfo *object_beans.OrderShareboxOrder) utils.ERROR {
	bs, err := orderInfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	key, ERR := utilsleveldb.BuildLeveldbKey(orderInfo.GetId())
	if !ERR.CheckSuccess() {
		return ERR
	}
	return config.LevelDB.SaveMap(*config.DBKEY_SHAREBOX_server_order_unpaid, *key, *bs, nil)
}

/*
查询文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Sharebox_server_FindOrderUnpaid(id []byte) (*object_beans.OrderShareboxOrder, utils.ERROR) {
	key, ERR := utilsleveldb.BuildLeveldbKey(id)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(*config.DBKEY_SHAREBOX_server_order_unpaid, *key)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	itr, err := object_beans.ParseOrderShareboxOrder(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	order := itr.(*object_beans.OrderShareboxOrder)
	return order, ERR
}

/*
保存未支付订单为已经支付订单
@id    []byte    订单id
*/
func Sharebox_server_SaveOrderPaid(orderId []byte, txHash, txBs *[]byte) utils.ERROR {
	//查询未支付订单
	order, ERR := Sharebox_server_FindOrderUnpaid(orderId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	order.TxHash = *txHash
	order.ChainTx = *txBs
	orderBs, err := order.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	cache := utilsleveldb.NewCache()
	//删除未支付订单
	key, ERR := utilsleveldb.BuildLeveldbKey(orderId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	cache.Set_Remove(config.DBKEY_SHAREBOX_server_order_unpaid, key)
	//保存已经支付订单
	cache.Set_Save(config.DBKEY_SHAREBOX_server_order_paid, key, orderBs)

	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
保存未支付订单为过期未支付订单
@id    []byte    订单id
*/
func Sharebox_server_SaveOrderUnpaidOvertime(orderId []byte) utils.ERROR {
	//查询未支付订单
	order, ERR := Sharebox_server_FindOrderUnpaid(orderId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	orderBs, err := order.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	cache := utilsleveldb.NewCache()
	//删除未支付订单
	key, ERR := utilsleveldb.BuildLeveldbKey(orderId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	cache.Set_Remove(config.DBKEY_SHAREBOX_server_order_unpaid, key)
	//保存已经支付订单
	cache.Set_Save(config.DBKEY_SHAREBOX_server_order_unpaid_overtime, key, orderBs)

	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
查询多个文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Sharebox_server_FindFilePriceMore(fileHash [][]byte) ([]*model.FilePrice, utils.ERROR) {

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

/*
查询文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
//func Sharebox_FindFilePrice(fileHash []byte) (*model.FilePrice, utils.ERROR) {
//	key, ERR := utilsleveldb.BuildLeveldbKey(fileHash)
//	if !ERR.CheckSuccess() {
//		return nil, ERR
//	}
//	bs, err := config.LevelDB.FindMap(config.DBKEY_sharebox_price_list, *key)
//	if err != nil {
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	if bs == nil {
//		return nil, utils.NewErrorSuccess()
//	}
//	filePrice, err := model.ParseFilePrice(bs)
//	if err != nil {
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	return filePrice, utils.NewErrorSuccess()
//}

/*
查询多个文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
//func Sharebox_FindFilePriceMore(fileHash [][]byte) ([]*model.FilePrice, utils.ERROR) {
//
//	keys := make([]utilsleveldb.LeveldbKey, 0, len(fileHash))
//	for _, one := range fileHash {
//		key, ERR := utilsleveldb.BuildLeveldbKey(one)
//		if !ERR.CheckSuccess() {
//			return nil, ERR
//		}
//		keys = append(keys, *key)
//	}
//
//	items, err := config.LevelDB.FindMapByKeys(config.DBKEY_sharebox_price_list, keys...)
//	if err != nil {
//		return nil, utils.NewErrorSysSelf(err)
//	}
//	prices := make([]*model.FilePrice, 0, len(fileHash))
//	for _, one := range items {
//		filePrice, err := model.ParseFilePrice(one.Value)
//		if err != nil {
//			return nil, utils.NewErrorSysSelf(err)
//		}
//		prices = append(prices, filePrice)
//	}
//	return prices, utils.NewErrorSuccess()
//}

/*
保存文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
//func Sharebox_SaveFilePrice(finfo *model.FilePrice) utils.ERROR {
//	bs, err := finfo.Proto()
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	key, ERR := utilsleveldb.BuildLeveldbKey(finfo.Hash)
//	if !ERR.CheckSuccess() {
//		return ERR
//	}
//	return config.LevelDB.SaveMap(config.DBKEY_sharebox_price_list, *key, *bs, nil)
//}

/*
保存文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
//func Sharebox_GetFilePriceRange(startFileHash []byte, limit uint64) ([]model.FilePrice, utils.ERROR) {
//	items, ERR := config.LevelDB.FindMapAllToListRange(config.DBKEY_sharebox_price_list, startFileHash, limit, false)
//	if ERR.CheckFail() {
//		return nil, ERR
//	}
//	filePrices := make([]model.FilePrice, 0, len(items))
//	for _, one := range items {
//		filePrice, err := model.ParseFilePrice(one.Value)
//		if err != nil {
//			filePrices = append(filePrices, *filePrice)
//		}
//	}
//	return filePrices, utils.NewErrorSuccess()
//}

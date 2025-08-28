package db

import (
	"encoding/hex"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"sync"
	"web3_gui/config"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

var proxyServerInfoLock = new(sync.RWMutex)

/*
查询自己提供离线服务设置信息
@return    uint64    单价
@return    bool      服务是否打开
*/
func ImProxyServer_GetProxyInfoSelf(addrSelf nodeStore.AddressNet, lock bool) (*model.StorageServerInfo, utils.ERROR) {
	if lock {
		proxyServerInfoLock.RLock()
		defer proxyServerInfoLock.RUnlock()
	}
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_info_self)
	item, err := config.LevelDB.Find(*dbKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if item == nil || item.Value == nil || len(item.Value) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	userinfo, err := model.ParseStorageServerInfo(item.Value)
	return userinfo, utils.NewErrorSysSelf(err)
}

/*
设置自己提供离线服务信息
@price       uint64    单价
@open        bool      是否打开
*/
func ImProxyServer_SetProxyInfoSelf(addrSelf nodeStore.AddressNet, info *model.StorageServerInfo) utils.ERROR {
	proxyServerInfoLock.Lock()
	defer proxyServerInfoLock.Unlock()
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_info_self)
	bs, err := info.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return config.LevelDB.Save(*dbKey, bs)
}

/*
获取存储服务器正在服务的订单id列表
*/
func ImProxyServer_GetInUseOrdersIDs(addrSelf nodeStore.AddressNet) ([]*model.OrderForm, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_ordersID_inuse)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_orders)

	keys := make([]utilsleveldb.LeveldbKey, 0)
	dbitem, err := config.LevelDB.FindMapAllToList(*dbKey1)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			return nil, utils.NewErrorSuccess()
		}
		return nil, utils.NewErrorSysSelf(err)
	}
	for _, one := range dbitem {
		keys = append(keys, one.Key)
	}
	items, err := config.LevelDB.FindMapByKeys(*dbKey2, keys...)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	orders := make([]*model.OrderForm, 0, len(items))
	for _, one := range items {
		of, err := model.ParseOrderForm(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		orders = append(orders, of)
	}
	return orders, utils.NewErrorSuccess()
}

/*
查询未支付的订单
*/
func ImProxyServer_GetNotPayOrders(addrSelf nodeStore.AddressNet) ([]*model.OrderForm, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_ordersID_not_pay)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_orders)

	keys := make([]utilsleveldb.LeveldbKey, 0)
	dbitem, err := config.LevelDB.FindMapAllToList(*dbKey1)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			return nil, utils.NewErrorSuccess()
		}
		return nil, utils.NewErrorSysSelf(err)
	}
	for _, one := range dbitem {
		keys = append(keys, one.Key)
	}
	items, err := config.LevelDB.FindMapByKeys(*dbKey2, keys...)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	orders := make([]*model.OrderForm, 0, len(items))
	for _, one := range items {
		of, err := model.ParseOrderForm(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		orders = append(orders, of)
	}
	return orders, utils.NewErrorSuccess()
}

/*
保存未支付订单
*/
func ImProxyServer_SaveOrderFormNotPay(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	proxyServerInfoLock.Lock()
	defer proxyServerInfoLock.Unlock()
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_orders)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_ordersID_not_pay)
	dbKey3 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_info_self)
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(form.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	cache := utilsleveldb.NewCache()
	cache.Set_Save(dbKey1, numberDBKey, bs)
	cache.Set_Save(dbKey2, numberDBKey, nil)
	//查询服务器信息
	info, ERR := ImProxyServer_GetProxyInfoSelf(addrSelf, false)
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Msgf("修改已经售卖的空间:%d %d", info.Sold, form.SpaceTotal)
	//修改已经使用的空间
	info.Sold += form.SpaceTotal
	//utils.Log.Info().Msgf("修改已经售卖的空间:%d %d", info.Sold, form.SpaceTotal)
	bs, err = info.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//保存新的服务器信息
	cache.Base_Save(dbKey3, bs)
	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
保存已支付订单
*/
func ImProxyServer_SaveOrderFormInUse(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	proxyServerInfoLock.Lock()
	defer proxyServerInfoLock.Unlock()
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_orders)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_ordersID_inuse)
	dbKey3 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_info_self)
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(form.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}

	cache := utilsleveldb.NewCache()
	//保存订单
	cache.Set_Save(dbKey1, numberDBKey, bs)
	//保存到正在服务期的订单列表
	cache.Set_Save(dbKey2, numberDBKey, nil)
	//查询服务器信息
	info, ERR := ImProxyServer_GetProxyInfoSelf(addrSelf, false)
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Msgf("修改已经售卖的空间:%d %d", info.Sold, form.SpaceTotal)
	//修改已经使用的空间
	info.Sold += form.SpaceTotal
	//utils.Log.Info().Msgf("修改已经售卖的空间:%d %d", info.Sold, form.SpaceTotal)
	bs, err = info.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//保存新的服务器信息
	cache.Base_Save(dbKey3, bs)
	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
删除一个已支付订单
*/
func ImProxyServer_DelOrderFormInUse(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_orders)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_ordersID_inuse)
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(form.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}

	cache := utilsleveldb.NewCache()
	cache.Set_Save(dbKey1, numberDBKey, bs)
	cache.Set_Save(dbKey2, numberDBKey, nil)
	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
将一个未支付订单删除并保存到已支付订单中
*/
func imProxyServer_MoveOrder(addrSelf nodeStore.AddressNet, order *model.OrderForm, fromDbKey, toDbKey *utilsleveldb.LeveldbKey) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_orders)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *fromDbKey)
	dbKey3 := utilsleveldb.JoinDbKey(*addrSelfKey, *toDbKey)
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(order.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := order.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	cache := utilsleveldb.NewCache()
	//刷新订单
	cache.Set_Save(dbKey1, numberDBKey, bs)
	//从未支付订单中删除
	cache.Set_Remove(dbKey2, numberDBKey)
	//保存到已支付订单列表
	cache.Set_Save(dbKey3, numberDBKey, &order.Number)
	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
将一个未支付订单删除并保存到已支付订单中
*/
func ImProxyServer_MoveOrderToInUse(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	return imProxyServer_MoveOrder(addrSelf, form, config.DBKEY_improxy_server_ordersID_not_pay, config.DBKEY_improxy_server_ordersID_inuse)
}

/*
将一个未支付订单删除并保存到，过期未支付订单中
*/
func ImProxyServer_MoveOrderToNotPayTimeout(addrSelf nodeStore.AddressNet, order *model.OrderForm) utils.ERROR {
	return imProxyServer_MoveOrder(addrSelf, order, config.DBKEY_improxy_server_ordersID_not_pay, config.DBKEY_improxy_server_ordersID_not_pay_timeout)
}

/*
将一个已支付订单删除，并保存到服务已到期订单中
同时，没有续费订单，要修改自己的空间信息
*/
func ImProxyServer_MoveOrderToInUseTimeout(addrSelf nodeStore.AddressNet, order, renewalOrder *model.OrderForm) utils.ERROR {
	proxyServerInfoLock.Lock()
	defer proxyServerInfoLock.Unlock()
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_ordersID_inuse)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_ordersID_inuse_timeout)
	dbKey3 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_info_self)
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(order.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	cache := utilsleveldb.NewCache()
	//从正在服务的订单中删除
	cache.Set_Remove(dbKey1, numberDBKey)
	//保存到服务到期列表中
	cache.Set_Save(dbKey2, numberDBKey, nil)
	//查询服务器信息
	info, ERR := ImProxyServer_GetProxyInfoSelf(addrSelf, false)
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Msgf("修改已经售卖的空间:%d %d", info.Sold, form.SpaceTotal)
	//减去订单的空间
	info.Sold -= order.SpaceTotal
	//utils.Log.Info().Msgf("修改已经售卖的空间:%d %d", info.Sold, form.SpaceTotal)
	//有续费订单
	if renewalOrder != nil {
		renewalNumberDBKey, ERR := utilsleveldb.BuildLeveldbKey(renewalOrder.Number)
		if !ERR.CheckSuccess() {
			return ERR
		}
		//把续费订单保存到正在服务期的订单列表中
		cache.Set_Save(dbKey1, renewalNumberDBKey, nil)
		//加上续费订单的空间
		info.Sold += renewalOrder.SpaceTotal
	}
	bs, err := info.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//保存新的服务器信息
	cache.Base_Save(dbKey3, bs)
	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

//
///*
//保存代理节点接收的消息
//*/
//func ImProxyServer_SaveDatachainRecv(itr imdatachain.DataChainProxyItr) error {
//	addrToKey, err := utilsleveldb.BuildLeveldbKey(itr.GetAddrTo())
//	if err != nil {
//		return err
//	}
//	addrKey, err := utilsleveldb.NewLeveldbKeyJoin(itr.GetAddrTo(), itr.GetAddrFrom())
//	if err != nil {
//		return err
//	}
//	idKey, err := utilsleveldb.BuildLeveldbKey(itr.GetID())
//	if err != nil {
//		return err
//	}
//	//检查是否重复
//
//	bs, err := config.LevelDB.FindMapInMapByKeyIn(config.DBKEY_improxy_server_datachain_nolink_id, *addrKey, *idKey)
//	if err != nil {
//		if err.Error() == leveldb.ErrNotFound.Error() {
//		} else {
//			return err
//		}
//	}
//	if bs != nil || len(bs) != 0 {
//		//记录已经存在。因为没用事务的原因，可能存在数据链ID已经保存，而消息内容未保存的情况，并且只有保存最后一条记录的情况才存在。所以查询一下
//		_, _, endIndex, err := config.LevelDB.FindMapInListTotal(config.DBKEY_improxy_server_datachain_nolink, *addrToKey)
//		if err != nil {
//			return err
//		}
//		_, err = config.LevelDB.FindMapInListByIndex(config.DBKEY_improxy_server_datachain_nolink, *addrToKey, endIndex)
//		if err != nil {
//			if err.Error() == leveldb.ErrNotFound.Error() {
//				//重新保存消息内容
//				nbs, err := itr.Proto()
//				if err != nil {
//					return err
//				}
//				_, err = config.LevelDB.SaveMapInList(false, config.DBKEY_improxy_server_datachain_nolink, *addrToKey, *nbs)
//				if err != nil {
//					return err
//				}
//				return nil
//			}
//			return err
//		}
//		return nil
//	}
//
//	//先保存数据链id
//	err = config.LevelDB.SaveMapInMap(config.DBKEY_improxy_server_datachain_nolink_id, *addrKey, *idKey, nil)
//	if err != nil {
//		return err
//	}
//	//再保存消息内容
//	nbs, err := itr.Proto()
//	if err != nil {
//		return err
//	}
//	bs, err = config.LevelDB.SaveMapInList(false, config.DBKEY_improxy_server_datachain_nolink, *addrToKey, *nbs)
//	if err != nil {
//		utils.Log.Error().Msgf("保存消息错误，只保存了一半:%s", err.Error())
//		return err
//	}
//	return nil
//}

/*
获取所有用户已经使用的空间大小
*/
func ImProxyServer_GetAllUserSpaceUseSize(addrSelf nodeStore.AddressNet) (map[string]uint64, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_user_spaceSize)
	dbItem, err := config.LevelDB.FindMapAllToList(*dbKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	m := make(map[string]uint64)
	for _, one := range dbItem {
		key, ERR := one.Key.BaseKey()
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		value := utils.BytesToUint64(one.Value)
		m[utils.Bytes2string(key)] = value
	}
	return m, utils.NewErrorSuccess()
}

/*
保存代理节点接收的消息
*/
func ImProxyServer_SaveDatachainNoLink(addrSelf nodeStore.AddressNet, proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	bs1 := append(proxyItr.GetAddrFrom().GetAddr(), proxyItr.GetAddrTo().GetAddr()...)
	bs1 = append(bs1, proxyItr.GetBase().SendIndex.Bytes()...)
	utils.Log.Info().Msgf("保存未上链消息:%s", hex.EncodeToString(bs1))
	//batch := new(leveldb.Batch)
	//属于谁的数据
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	//谁的离线消息
	addrToKey, ERR := utilsleveldb.BuildLeveldbKey(proxyItr.GetAddrTo().GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	//dbKey1 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_server_datachain_nolink_id, *addrSelfKey)
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_datachain_nolink, *addrToKey)
	//发送者
	addrFromKey, ERR := utilsleveldb.BuildLeveldbKey(proxyItr.GetAddrFrom().GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	//消息内容
	bs, err := proxyItr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	proxyBase := proxyItr.GetBase()
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(*dbKey, *addrFromKey, proxyBase.SendIndex.Bytes(), *bs, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//err = SaveBatch(batch)
	//if err != nil {
	//	return utils.NewErrorSysSelf(err)
	//}
	return utils.NewErrorSuccess()
}

/*
查询代理节点接收的消息的发送index
*/
func ImProxyServer_FindDatachainNoLinkSendIndex(addrSelf, addrFrom, addrTo nodeStore.AddressNet) (*big.Int, utils.ERROR) {
	//属于谁的数据
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//谁的离线消息
	addrToKey, ERR := utilsleveldb.BuildLeveldbKey(addrTo.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//dbKey1 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_server_datachain_nolink_id, *addrSelfKey)
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_datachain_nolink, *addrToKey)
	//发送者
	addrFromKey, ERR := utilsleveldb.BuildLeveldbKey(addrFrom.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//开始查询,查询index最大的一条记录
	items, ERR := config.LevelDB.FindMapInListRangeByKeyIn(*dbKey, *addrFromKey, nil, 1, false)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if len(items) != 1 {
		return big.NewInt(0), utils.NewErrorSuccess()
	}
	return new(big.Int).SetBytes(items[0].Index), utils.NewErrorSuccess()
}

/*
查询代理节点接收的消息的发送index
*/
func ImProxyServer_FindDatachainNoLinkRange(addrSelf, addrTo nodeStore.AddressNet, limit uint64) ([]imdatachain.DataChainProxyItr, utils.ERROR) {
	//属于谁的数据
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//谁的离线消息
	addrToKey, ERR := utilsleveldb.BuildLeveldbKey(addrTo.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//dbKey1 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_server_datachain_nolink_id, *addrSelfKey)
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_datachain_nolink, *addrToKey)
	//发送者
	//addrFromKey, ERR := utilsleveldb.BuildLeveldbKey(addrFrom)
	//if !ERR.CheckSuccess() {
	//	return nil, ERR
	//}
	//开始查询,查询index最大的一条记录
	items, ERR := config.LevelDB.FindMapInListRangeByKeyOut(*dbKey, nil, limit, true)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	itrs := make([]imdatachain.DataChainProxyItr, 0, len(items))
	for _, one := range items {
		//utils.Log.Info().Msgf("查询的信息:%+v", one.Value)
		itr, ERR := imdatachain.ParseDataChain(one.Value)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		//utils.Log.Info().Msgf("解析后:%+v", itr)
		itrs = append(itrs, itr)
	}
	return itrs, utils.NewErrorSuccess()
}

/*
删除代理节点接收的消息
*/
func ImProxyServer_RemoveDatachainNoLinkBySendIndex(addrSelf, addrFrom, addrTo nodeStore.AddressNet, sendIndex []byte, batch *leveldb.Batch) utils.ERROR {
	bs := append(addrFrom.GetAddr(), addrTo.GetAddr()...)
	bs = append(bs, sendIndex...)
	utils.Log.Info().Msgf("删除未上链消息:%s", hex.EncodeToString(bs))
	//属于谁的数据
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	//谁的离线消息
	addrToKey, ERR := utilsleveldb.BuildLeveldbKey(addrTo.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	//dbKey1 := utilsleveldb.JoinDbKey(config.DBKEY_improxy_server_datachain_nolink_id, *addrSelfKey)
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_datachain_nolink, *addrToKey)
	//发送者
	addrFromKey, ERR := utilsleveldb.BuildLeveldbKey(addrFrom.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	//删除
	ERR = config.LevelDB.RemoveMapInListByIndex(*dbKey, *addrFromKey, sendIndex, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
保存发送给代理节点用户的sendIndex
*/
func ImProxyServer_SaveDatachainSendIndex(addrSelf, addrFrom, addrTo nodeStore.AddressNet, sendIndex []byte, batch *leveldb.Batch) utils.ERROR {
	//属于谁的数据
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	//谁的离线消息
	addrToKey, ERR := utilsleveldb.BuildLeveldbKey(addrTo.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_datachain_sendIndex, *addrToKey)
	//发送者
	addrFromKey, ERR := utilsleveldb.BuildLeveldbKey(addrFrom.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存
	ERR = config.LevelDB.SaveMap(*dbKey, *addrFromKey, sendIndex, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询发送给代理节点用户的sendIndex
*/
func ImProxyServer_FindDatachainSendIndex(addrSelf, addrFrom, addrTo nodeStore.AddressNet) (*big.Int, utils.ERROR) {
	//属于谁的数据
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//谁的离线消息
	addrToKey, ERR := utilsleveldb.BuildLeveldbKey(addrTo.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_server_datachain_sendIndex, *addrToKey)
	//发送者
	addrFromKey, ERR := utilsleveldb.BuildLeveldbKey(addrFrom.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//查询
	sendIndexBs, err := config.LevelDB.FindMap(*dbKey, *addrFromKey)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		return nil, ERR
	}
	if sendIndexBs == nil || len(sendIndexBs) == 0 {
		return big.NewInt(0), utils.NewErrorSuccess()
	}
	return new(big.Int).SetBytes(sendIndexBs), utils.NewErrorSuccess()
}

package db

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"sync"
	"web3_gui/config"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
添加或修改IM服务器列表
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func ImProxyClient_AddImProxy(addrSelf nodeStore.AddressNet, sinfo *model.StorageServerInfo) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_list)
	bs, err := sinfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	key, ERR := utilsleveldb.BuildLeveldbKey(sinfo.Addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	return config.LevelDB.SaveMap(*dbKey, *key, *bs, nil)
}

/*
通过服务器地址，获取所有提供离线服务的节点信息
*/
func ImProxyClient_GetProxyListByAddrs(addrSelf nodeStore.AddressNet, sAddr ...nodeStore.AddressNet) ([]model.StorageServerInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_list)
	keys := make([]utilsleveldb.LeveldbKey, 0)
	for _, one := range sAddr {
		keyOne, _ := utilsleveldb.BuildLeveldbKey(one.GetAddr())
		keys = append(keys, *keyOne)
	}
	items, err := config.LevelDB.FindMapByKeys(*dbKey, keys...)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	sinfos := make([]model.StorageServerInfo, 0)
	for _, one := range items {
		sinfo, err := model.ParseStorageServerInfo(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		sinfos = append(sinfos, *sinfo)
	}
	return sinfos, utils.NewErrorSysSelf(err)
}

/*
获取所有提供离线服务的节点信息
*/
func ImProxyClient_GetProxyList(addrSelf nodeStore.AddressNet) ([]model.StorageServerInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_list)
	items, err := config.LevelDB.FindMapAllToList(*dbKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	sinfos := make([]model.StorageServerInfo, 0)
	for _, one := range items {
		sinfo, err := model.ParseStorageServerInfo(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		sinfos = append(sinfos, *sinfo)
	}
	return sinfos, utils.NewErrorSysSelf(err)
}

/*
保存一个订单
*/
func imProxyClient_SaveOrder(addrSelf nodeStore.AddressNet, order *model.OrderForm, dbKey *utilsleveldb.LeveldbKey) utils.ERROR {
	//utils.Log.Info().Hex("要保存的dbkey", dbKey.Byte()).Send()
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_orders)
	dbKey2 := addrSelfKey.JoinKey(dbKey) //utilsleveldb.JoinDbKey(*addrSelfKey, *dbKey)
	//utils.Log.Info().Msgf("保存未支付订单key:%s", hex.EncodeToString(dbKey2.Byte()))
	bs, err := order.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(order.Number) //utilsleveldb.NewLeveldbKeyJoin(order.ServerAddr.GetAddr(), order.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	cache := utilsleveldb.NewCache()
	//保存订单
	cache.Set_Save(dbKey1, numberDBKey, bs)
	//客户端正在服务的订单列表
	cache.Set_Save(dbKey2, numberDBKey, &order.Number)
	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存未支付订单:%s", hex.EncodeToString(order.Number))
	return utils.NewErrorSuccess()
}

/*
加载订单
*/
func imProxyClient_LoadOrder(addrSelf nodeStore.AddressNet, dbKey *utilsleveldb.LeveldbKey) ([]model.OrderForm, utils.ERROR) {
	//utils.Log.Info().Hex("要加载的dbkey", dbKey.Byte()).Send()
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *dbKey)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_orders)
	//utils.Log.Info().Hex("未支付订单key", dbKey1.Byte()).Send()

	items, ERR := config.LevelDB.Cache_Set_FindRange(dbKey1, nil, 0, true)
	//items, ERR := config.LevelDB.FindMapAllToListRange(*dbKey1, startNumberKey, limit, false)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Int("未支付订单 返回数量", len(items)).Send()
	keys := make([]*utilsleveldb.LeveldbKey, 0, len(items))
	for _, one := range items {
		//utils.Log.Info().Hex("订单 key", one.Value).Send()
		keyOne, ERR := utilsleveldb.BuildLeveldbKey(one.Value)
		if ERR.CheckFail() {
			return nil, ERR
		}
		keys = append(keys, keyOne)
	}

	items, ERR = config.LevelDB.Cache_Set_FindMore(dbKey2, keys, true)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Int("未支付订单 返回数量", len(items)).Send()

	//items, err := config.LevelDB.FindMapByKeys(*dbKey2, keys...)
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	orders := make([]model.OrderForm, 0, len(items))
	for _, one := range items {
		orderOne, err := model.ParseOrderForm(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		orders = append(orders, *orderOne)
	}
	return orders, utils.NewErrorSuccess()
}

/*
将一个未支付订单删除并保存到已支付订单中
*/
func imProxyClient_MoveOrder(addrSelf nodeStore.AddressNet, order *model.OrderForm, fromDbKey, toDbKey *utilsleveldb.LeveldbKey) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_orders)
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
保存一个未支付订单
*/
func ImProxyClient_SaveOrderFormNotPay(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	return imProxyClient_SaveOrder(addrSelf, form, config.DBKEY_improxy_client_ordersID_not_pay)
}

/*
加载未过期的未支付订单
*/
func ImProxyClient_LoadOrderFormNotPay(addrSelf nodeStore.AddressNet) ([]model.OrderForm, utils.ERROR) {
	return imProxyClient_LoadOrder(addrSelf, config.DBKEY_improxy_client_ordersID_not_pay)
}

/*
加载支付后等待上链的订单
*/
func ImProxyClient_LoadOrderFormNotOnChain(addrSelf nodeStore.AddressNet) ([]model.OrderForm, utils.ERROR) {
	return imProxyClient_LoadOrder(addrSelf, config.DBKEY_improxy_client_ordersID_pay_notOnChain)
}

/*
保存一个已支付订单
*/
func ImProxyClient_SaveOrderFormInUse(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	return imProxyClient_SaveOrder(addrSelf, form, config.DBKEY_improxy_client_ordersID_inuse)
}

/*
加载未过期的已支付订单
*/
func ImProxyClient_LoadOrderFormInUse(addrSelf nodeStore.AddressNet) ([]model.OrderForm, utils.ERROR) {
	return imProxyClient_LoadOrder(addrSelf, config.DBKEY_improxy_client_ordersID_inuse)
}

/*
将一个未支付订单移动到已支付但未上链列表中
*/
func ImProxyClient_MoveOrderToNotOnChain(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	return imProxyClient_MoveOrder(addrSelf, form, config.DBKEY_improxy_client_ordersID_not_pay, config.DBKEY_improxy_client_ordersID_pay_notOnChain)
}

/*
移动一个支付未上链的订单到已支付并上链列表中
*/
func ImProxyClient_MoveOrderNotOnChainToInUse(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	return imProxyClient_MoveOrder(addrSelf, form, config.DBKEY_improxy_client_ordersID_pay_notOnChain, config.DBKEY_improxy_client_ordersID_inuse)
}

/*
移动一个已支付未上链的订单到未支付列表中
*/
func ImProxyClient_MoveOrderNotOnChainToNotPay(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	return imProxyClient_MoveOrder(addrSelf, form, config.DBKEY_improxy_client_ordersID_pay_notOnChain, config.DBKEY_improxy_client_ordersID_not_pay)
}

/*
移动一个未支付订单到已支付并上链列表中
*/
func ImProxyClient_MoveOrderToInUse(addrSelf nodeStore.AddressNet, form *model.OrderForm) utils.ERROR {
	return imProxyClient_MoveOrder(addrSelf, form, config.DBKEY_improxy_client_ordersID_not_pay, config.DBKEY_improxy_client_ordersID_inuse)
}

/*
未支付订单移动过期未支付并上链列表中
*/
func ImProxyClient_MoveOrderToNotPayTimeout(addrSelf nodeStore.AddressNet, order *model.OrderForm) utils.ERROR {
	return imProxyClient_MoveOrder(addrSelf, order, config.DBKEY_improxy_client_ordersID_not_pay, config.DBKEY_improxy_client_ordersID_timeout)
}

/*
将一个已支付订单移动到服务已到期订单中
同时把续费订单放入正在服务的订单列表中
*/
func ImProxyClient_MoveOrderToInUseTimeout(addrSelf nodeStore.AddressNet, order *model.OrderForm) utils.ERROR {
	return imProxyClient_MoveOrder(addrSelf, order, config.DBKEY_improxy_client_ordersID_inuse, config.DBKEY_improxy_client_ordersID_timeout)
}

/*
加载过期订单，包括已支付和未支付的过期订单
*/
func ImProxyClient_LoadOrderTimeout(addrSelf nodeStore.AddressNet, startNumberKey []byte, limit uint64) ([]model.OrderForm, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_ordersID_timeout)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_client_orders)
	//utils.Log.Info().Hex("超时订单key", dbKey1.Byte()).Send()
	var startKey *utilsleveldb.LeveldbKey
	if len(startNumberKey) > 0 {
		startKey = utilsleveldb.NewLeveldbKey(startNumberKey)
	}
	items, ERR := config.LevelDB.Cache_Set_FindRange(dbKey1, startKey, limit, true)
	//items, ERR := config.LevelDB.FindMapAllToListRange(*dbKey1, startNumberKey, limit, false)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Int("超时订单 返回数量", len(items)).Send()
	keys := make([]*utilsleveldb.LeveldbKey, 0, len(items))
	for _, one := range items {
		//utils.Log.Info().Hex("订单 key", one.Value).Send()
		keyOne, ERR := utilsleveldb.BuildLeveldbKey(one.Value)
		if ERR.CheckFail() {
			return nil, ERR
		}
		keys = append(keys, keyOne)
	}

	items, ERR = config.LevelDB.Cache_Set_FindMore(dbKey2, keys, true)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Int("超时订单 返回数量", len(items)).Send()

	//items, err := config.LevelDB.FindMapByKeys(*dbKey2, keys...)
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	orders := make([]model.OrderForm, 0, len(items))
	for _, one := range items {
		orderOne, err := model.ParseOrderForm(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		orders = append(orders, *orderOne)
	}
	return orders, utils.NewErrorSuccess()
}

/*
保存快照高度
@index    []byte    快照高度
@return    []byte    记录最后的index
@return    error     错误
*/
func ImProxyClient_SaveShot(addrSelf, addr nodeStore.AddressNet, index []byte, batch *leveldb.Batch) utils.ERROR {
	//utils.Log.Info().Msgf("加载快照高度:%+v", addr)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_shot_index)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = config.LevelDB.SaveMap(*dbKey, *addrKey, index, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//index := itr.GetIndex()
	return utils.NewErrorSuccess()
}

/*
加载自己聊天记录的快照高度
@return    []byte    快照高度
@return    []byte    记录最后的index
@return    error     错误
*/
func ImProxyClient_LoadShot(addrSelf, addr nodeStore.AddressNet) ([]byte, imdatachain.DataChainProxyItr, utils.ERROR) {
	//utils.Log.Info().Msgf("加载快照高度:%+v", addr)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_shot_index)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	indexBs, err := config.LevelDB.FindMap(*dbKey, *addrKey)
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	itr, ERR := ImProxyClient_FindDataChainLast(addrSelf, addr)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("加载快照高度 错误:%s", ERR.String())
		return nil, nil, ERR
	}
	//index := itr.GetIndex()
	return indexBs, itr, utils.NewErrorSuccess()
}

/*
加载自己聊天记录的快照高度
@return    imdatachain.DataChainProxyItr    被查询用户最后一条消息
@return    error     错误
*/
func ImProxyClient_FindDataChainLast(addrSelf, addr nodeStore.AddressNet) (imdatachain.DataChainProxyItr, utils.ERROR) {
	//utils.Log.Info().Msgf("查询用户数据链:%+v", config.DBKEY_improxy_user_datachain)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain)

	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("查询用户数据链 错误:%s", ERR.String())
		return nil, ERR
	}
	//查询用户数据库最后的index
	//注意：因为保存时考虑性能未使用事务，则可能存在出错时，只保存了数据链记录，而未保存记录唯一id的情况。
	_, _, endIndex, ERR := config.LevelDB.FindMapInListTotal(*dbKey, *addrKey)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("查询用户数据链 错误:%s", ERR.String())
		return nil, ERR
	}
	//utils.Log.Info().Msgf("查询用户数据链:%+v %+v %+v", dbKey, *addrKey, endIndex)
	if endIndex == nil || len(endIndex) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	//查询最后一条记录的唯一id
	return ImProxyClient_FindDataChainByIndex(addrSelf, addr, endIndex)
}

var ImProxyClient_SaveDataChainMoreLock = new(sync.Mutex)

/*
保存多条数据链
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_SaveDataChainMore(addrSelf nodeStore.AddressNet, batch *leveldb.Batch, itrs ...imdatachain.DataChainProxyItr) utils.ERROR {
	//utils.Log.Info().Msgf("保存数据链")

	if len(itrs) == 0 {
		return utils.NewErrorSuccess()
	}
	var proxyClientAddr nodeStore.AddressNet
	var preItr imdatachain.DataChainProxyItr
	for _, itrOne := range itrs {
		//utils.Log.Info().Msgf("保存数据链:%+v", itrOne)
		//验证消息hash
		if !itrOne.CheckHash() {
			//ReplyError(Area, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_IM_check_hash_fail, "")
			return utils.NewErrorBus(config.ERROR_CODE_IM_check_hash_fail, "")
		}
		//utils.Log.Info().Msgf("保存数据链")
		//接收的消息必须是连续的
		if preItr != nil {
			if !bytes.Equal(preItr.GetHash(), itrOne.GetPreHash()) {
				//ReplyError(Area, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_IM_check_hash_fail, "")
				return utils.NewErrorBus(config.ERROR_CODE_IM_check_hash_fail, "")
			}
			//utils.Log.Info().Msgf("保存数据链")
			//判断index是连续增长的
			preIndex := preItr.GetIndex()
			indexOne := itrOne.GetIndex()
			if new(big.Int).Add(&preIndex, big.NewInt(1)).Cmp(&indexOne) != 0 {
				//index不连续
				return utils.NewErrorBus(config.ERROR_CODE_IM_index_discontinuity, "")
			}
			//本次保存的所有记录，为同一人记录
			if !bytes.Equal(proxyClientAddr.GetAddr(), itrOne.GetProxyClientAddr().GetAddr()) {
				return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_user_different, "")
			}
			//utils.Log.Info().Msgf("保存数据链")
		}
		proxyClientAddr = itrOne.GetProxyClientAddr()
		preItr = itrOne
	}
	//utils.Log.Info().Msgf("保存数据链:%s", proxyClientAddr.B58String())
	ImProxyClient_SaveDataChainMoreLock.Lock()
	defer ImProxyClient_SaveDataChainMoreLock.Unlock()
	//判断新添加的记录，和本地数据链是否连续
	preItr, ERR := ImProxyClient_FindDataChainLast(addrSelf, proxyClientAddr)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//utils.Log.Info().Msgf("保存数据链")
	firstIndex := itrs[0].GetIndex()
	//utils.Log.Info().Msgf("保存数据链:%s %+v", firstIndex.String(), preItr)
	//utils.Log.Info().Msgf("保存数据链，本条记录index:%s %+v", firstIndex.String(), preItr)
	//if preItr != nil {
	//	utils.Log.Info().Msgf("保存数据链，对比hash是否连续:%+v %+v", preItr.GetHash(), itrs[0].GetPreHash())
	//}
	//utils.Log.Info().Msgf("哪个是空指针:%+v %+v %+v", firstIndex, preItr, itrs)
	//验证消息前置hash
	if firstIndex.Cmp(big.NewInt(1)) != 0 && !bytes.Equal(preItr.GetHash(), itrs[0].GetPreHash()) {
		//utils.Log.Info().Msgf("保存数据链")
		//检查消息index大小
		serverIndex := preItr.GetIndex()
		//utils.Log.Info().Msgf("保存数据链")
		localIndex := itrs[0].GetIndex()
		//utils.Log.Info().Msgf("保存数据链")
		if serverIndex.Cmp(&localIndex) >= 0 {
			//utils.Log.Info().Msgf("保存数据链")
			return utils.NewErrorBus(config.ERROR_CODE_IM_index_too_small, "")
		} else {
			return utils.NewErrorBus(config.ERROR_CODE_IM_index_too_big, "")
		}
	}
	//utils.Log.Info().Msgf("保存数据链")
	//batch := new(leveldb.Batch)
	for _, one := range itrs {
		//utils.Log.Info().Msgf("保存数据链")
		ERR := saveDataChain(addrSelf, one, batch)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	//err := config.LevelDB.SaveBatch(batch)
	//utils.Log.Info().Msgf("保存数据链")
	//return utils.NewErrorSysSelf(err)
	return utils.NewErrorSuccess()
}

/*
加载自己聊天记录的快照高度
@return    []byte    保存记录的index
@return    error     错误
*/
func saveDataChain(addrSelf nodeStore.AddressNet, itr imdatachain.DataChainProxyItr, batch *leveldb.Batch) utils.ERROR {
	//utils.Log.Info().Msgf("保存消息:%+v", itr)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_id)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain)
	dbKey3 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_sendid)

	bs, err := itr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("消息序列化:%+v", bs)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(itr.GetProxyClientAddr().GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	idKey, ERR := utilsleveldb.BuildLeveldbKey(itr.GetID())
	if !ERR.CheckSuccess() {
		return ERR
	}

	index := itr.GetIndex()
	//utils.Log.Info().Msgf("保存数据链索引ID:%+v %+v %+v %+v %s", dbKey1, *addrKey, *idKey, index.Bytes(), hex.EncodeToString(itr.GetID()))
	//保存这条消息的唯一ID，用于去重
	ERR = config.LevelDB.SaveMapInMap(*dbKey1, *addrKey, *idKey, index.Bytes(), batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存这条消息的sendID，用于去重
	if itr.GetSendID() != nil && len(itr.GetSendID()) != 0 {
		sendIdKey, ERR := utilsleveldb.BuildLeveldbKey(itr.GetSendID())
		if !ERR.CheckSuccess() {
			return ERR
		}
		ERR = config.LevelDB.SaveMapInMap(*dbKey3, *addrKey, *sendIdKey, index.Bytes(), batch)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	//utils.Log.Info().Msgf("保存消息唯一ID:%+v index:%+v", itr.GetID(), index.Bytes())
	//utils.Log.Info().Msgf("保存消息:%+v %+v %+v", *dbKey2, *addrKey, index.Bytes())
	//保存消息内容
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(*dbKey2, *addrKey, index.Bytes(), *bs, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询一条数据链记录是否存在
@addrSelf  nodeStore.AddressNet    避免账号污染，属于哪个账号的数据
@addr      nodeStore.AddressNet    代理节点保存多个用户的数据链，此参数用于区分属于谁的数据链
@id        []byte                  数据链id
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_FindDataChainByID(addrSelf nodeStore.AddressNet, addr nodeStore.AddressNet, id []byte) (imdatachain.DataChainProxyItr, utils.ERROR) {
	//utils.Log.Info().Msgf("通过ID查询数据链记录:%+v %+v", addr, id)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_id)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain)

	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	idKey, ERR := utilsleveldb.BuildLeveldbKey(id)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//utils.Log.Info().Msgf("通过ID查询数据链记录:%+v %+v %+v", dbKey1, *addrKey, *idKey)
	index, err := config.LevelDB.FindMapInMapByKeyIn(*dbKey1, *addrKey, *idKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if index == nil {
		//utils.Log.Info().Msgf("未找到记录:%+v %+v", addr, id)
		return nil, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("查询数据链记录id:%+v %+v %+v", dbKey2, *addrKey, *index)
	//注意：因为保存时考虑性能未使用事务，则可能存在出错时，只保存了数据链id，而未保存数据内容的情况。
	item, ERR := config.LevelDB.FindMapInListByIndex(*dbKey2, *addrKey, *index)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if item == nil {
		//utils.Log.Info().Msgf("未找到记录:%+v %+v", addr, id)
		return nil, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("查询数据链数据:%+v %+v %+v", config.DBKEY_improxy_user_datachain, *addrKey, index)
	itr, ERR := imdatachain.ParseDataChain(item.Value)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return itr, utils.NewErrorSuccess()
}

/*
查询一条数据链记录是否存在
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_FindDataChainBySendID(addrSelf nodeStore.AddressNet, addr nodeStore.AddressNet, sendId []byte) (bool, utils.ERROR) {
	//utils.Log.Info().Msgf("通过ID查询数据链记录:%+v %+v", addr, id)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_sendid)
	//dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, config.DBKEY_improxy_user_datachain)

	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	sendIdKey, ERR := utilsleveldb.BuildLeveldbKey(sendId)
	if !ERR.CheckSuccess() {
		return false, ERR
	}
	//utils.Log.Info().Msgf("通过ID查询数据链记录:%+v %+v %+v", config.DBKEY_improxy_user_datachain_id, *addrKey, *idKey)
	index, err := config.LevelDB.FindMapInMapByKeyIn(*dbKey1, *addrKey, *sendIdKey)
	if err != nil {
		return false, utils.NewErrorSysSelf(err)
	}
	if index == nil {
		//utils.Log.Info().Msgf("未找到记录:%+v %+v", addr, id)
		return false, utils.NewErrorSuccess()
	}
	return true, utils.NewErrorSuccess()
}

/*
通过index查询一条记录
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_FindDataChainByIndex(addrSelf nodeStore.AddressNet, addr nodeStore.AddressNet,
	index []byte) (imdatachain.DataChainProxyItr, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	item, ERR := config.LevelDB.FindMapInListByIndex(*dbKey, *addrKey, index)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if item == nil {
		return nil, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("解析的数据:%+v", item)
	itr, ERR := imdatachain.ParseDataChain(item.Value)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return itr, utils.NewErrorSuccess()
}

/*
修改一条记录的状态
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_UpdateDataChainStatusByID(addrSelf nodeStore.AddressNet, addr nodeStore.AddressNet, index []byte,
	status uint8, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}

	item, ERR := config.LevelDB.FindMapInListByIndex(*dbKey, *addrKey, index)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if item == nil {
		return utils.NewErrorSuccess()
	}
	itr, ERR := imdatachain.ParseDataChain(item.Value)
	if !ERR.CheckSuccess() {
		return ERR
	}
	itr.SetStatus(status)
	bs, err := itr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(*dbKey, *addrKey, index, *bs, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
按条件查询数据链
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_FindDataChainRange(addrSelf nodeStore.AddressNet, addr nodeStore.AddressNet, startIndexBs []byte,
	limit uint64) ([]imdatachain.DataChainProxyItr, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbItems, ERR := config.LevelDB.FindMapInListRangeByKeyIn(*dbKey, *addrKey, startIndexBs, limit, true)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	itrs := make([]imdatachain.DataChainProxyItr, 0, len(dbItems))
	for _, one := range dbItems {
		itr, ERR := imdatachain.ParseDataChain(one.Value)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		itrs = append(itrs, itr)
	}
	return itrs, utils.NewErrorSuccess()
}

var ImProxyClient_SaveDataChainSendFailLock = new(sync.Mutex)

/*
保存一条发送失败的消息记录
@return    []byte    保存记录的index
@return    error     错误
*/
func ImProxyClient_SaveDataChainSendFail(addrSelf nodeStore.AddressNet, itr imdatachain.DataChainProxyItr, batch *leveldb.Batch) utils.ERROR {
	//utils.Log.Info().Msgf("保存发送失败的消息记录：%+v", itr)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误：%s", ERR.String())
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_send_fail)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_send_fail_id)
	bs, err := itr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存的信息:%+v", bs)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(itr.GetAddrTo().GetAddr())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误：%s", ERR.String())
		return ERR
	}
	idKey, ERR := utilsleveldb.BuildLeveldbKey(itr.GetID())
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误：%s", ERR.String())
		return ERR
	}
	//batch := new(leveldb.Batch)
	ImProxyClient_SaveDataChainSendFailLock.Lock()
	defer ImProxyClient_SaveDataChainSendFailLock.Unlock()
	//查询数据库最后一个索引，用来推测下一个索引
	_, _, endIndex, ERR := config.LevelDB.FindMapInListTotal(*dbKey1, *addrKey)
	if !ERR.CheckSuccess() {
		return ERR
	}
	indexBs := new(big.Int).Add(new(big.Int).SetBytes(endIndex), big.NewInt(1)).Bytes()
	//构建地址和消息index索引组合的数据包
	bss := append([][]byte{itr.GetAddrTo().GetAddr()}, indexBs)
	valueBs, err := model.BytesProto(bss, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//指定索引并保存数据链消息
	//utils.Log.Info().Msgf("保存一个未发送消息:%+v %+v", addrKey, indexBs)
	ERR = config.LevelDB.SaveOrUpdateMapInListByIndex(*dbKey1, *addrKey, indexBs, *bs, batch)
	//_, ERR = config.LevelDB.SaveMapInList(config.DBKEY_improxy_user_datachain_send_fail, *addrKey, *bs, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//utils.Log.Info().Msgf("保存一个未发送索引:%+v %+v", *idKey, valueBs)
	//保存索引，方便删除
	ERR = config.LevelDB.SaveMap(*dbKey2, *idKey, valueBs, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//err = config.LevelDB.SaveBatch(batch)
	//if err != nil {
	//	return utils.NewErrorSysSelf(err)
	//}
	return utils.NewErrorSuccess()
}

/*
按条件查询数据链
@return    []imdatachain.DataChainProxyItr    查询的记录内容
@return    [][]byte                           数据库的index
@return    error                              错误
*/
func ImProxyClient_FindDataChainSendFailRange(addrSelf nodeStore.AddressNet,
	addr nodeStore.AddressNet, limit uint64) ([]imdatachain.DataChainProxyItr, [][]byte, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_send_fail)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(addr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	dbItems, ERR := config.LevelDB.FindMapInListRangeByKeyIn(*dbKey, *addrKey, nil, limit, true)
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	itrs := make([]imdatachain.DataChainProxyItr, 0, len(dbItems))
	indexs := make([][]byte, 0, len(dbItems))
	for _, one := range dbItems {
		//utils.Log.Info().Msgf("查询的信息:%+v", one.Value)
		itr, ERR := imdatachain.ParseDataChain(one.Value)
		if !ERR.CheckSuccess() {
			return nil, nil, ERR
		}
		//utils.Log.Info().Msgf("解析后:%+v", itr)
		itrs = append(itrs, itr)
		indexs = append(indexs, one.Index)
	}
	return itrs, indexs, utils.NewErrorSuccess()
}

/*
删除多条未发送成功的记录
@ids       [][]byte    要删除的消息id
@return    error       错误
*/
func ImProxyClient_DelDataChainSendFail(addrSelf nodeStore.AddressNet, proxyIds [][]byte) utils.ERROR {
	//utils.Log.Info().Msgf("删除记录:%+v", proxyIds)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey1 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_send_fail)
	dbKey2 := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_user_datachain_send_fail_id)
	idKeys := make([]utilsleveldb.LeveldbKey, 0)
	for _, one := range proxyIds {
		oneKey, ERR := utilsleveldb.BuildLeveldbKey(one)
		if !ERR.CheckSuccess() {
			return ERR
		}
		idKeys = append(idKeys, *oneKey)
	}
	items, err := config.LevelDB.FindMapByKeys(*dbKey2, idKeys...)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if items == nil || len(items) == 0 {
		//utils.Log.Error().Msgf("本次未删除任何数据")
		return utils.NewErrorSuccess()
	}

	addrKeys := make([]utilsleveldb.LeveldbKey, 0, len(items))
	indexs := make([][]byte, 0, len(items))
	for _, one := range items {
		//utils.Log.Info().Msgf("查询到的地址和索引组合:%+v", one.Value)
		list, _, err := model.ParseBytes(one.Value)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		addrKey, ERR := utilsleveldb.BuildLeveldbKey(list[0])
		if !ERR.CheckSuccess() {
			return ERR
		}
		//utils.Log.Info().Msgf("地址和索引组合:&+v", list)
		addrKeys = append(addrKeys, *addrKey)
		indexs = append(indexs, list[1])
	}

	batch := new(leveldb.Batch)
	//删除索引
	for _, one := range idKeys {
		//utils.Log.Info().Msgf("删除一个未发送索引:%+v", one.Byte())
		err = config.LevelDB.RemoveMapByKey(*dbKey2, one, batch)
		if err != nil {
			utils.Log.Error().Msgf("删除索引 错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
	}
	//删除数据
	for i, one := range addrKeys {
		index := indexs[i]
		//utils.Log.Info().Msgf("删除一个未发送消息:%+v %+v", one, index)
		ERR := config.LevelDB.RemoveMapInListByIndex(*dbKey1, one, index, batch)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	err = config.LevelDB.SaveBatch(batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//
	//addrKey, _ := addrKeys[0].BaseKey()
	//addr := nodeStore.AddressNet(addrKey)
	//itrs, _, ERR := ImProxyClient_FindDataChainSendFailRange(addr, 0)
	//if !ERR.CheckSuccess() {
	//	return ERR
	//}
	//utils.Log.Info().Msgf("查询的结果数量:%d", len(itrs))
	return utils.NewErrorSuccess()
}

/*
保存用户基本信息
*/
func ImProxyClient_SaveUserinfo(addrSelf nodeStore.AddressNet, userinfo *model.UserInfo) utils.ERROR {
	utils.Log.Info().Msgf("保存用户信息:%+v", userinfo)
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_userinfo)
	bs, err := userinfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(userinfo.Addr.GetAddr())
	if err != nil {
		return ERR
	}
	ERR = config.LevelDB.SaveMap(*dbKey, *addrKey, *bs, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询用户基本信息
*/
func ImProxyClient_FindUserinfo(addrSelf nodeStore.AddressNet, userAddr nodeStore.AddressNet) (*model.UserInfo, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, *config.DBKEY_improxy_userinfo)
	addrKey, ERR := utilsleveldb.BuildLeveldbKey(userAddr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(*dbKey, *addrKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	userinfo, err := model.ParseUserInfo(&bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return userinfo, utils.NewErrorSuccess()
}

/*
查询用户加密公钥
*/
func ImProxyClient_FindUserDhPuk(addrSelf nodeStore.AddressNet, userAddr nodeStore.AddressNet) (*keystore.Key, utils.ERROR) {
	//好友列表中查询
	friend, ERR := FindUserListByAddr(addrSelf, userAddr)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}

	if friend == nil || friend.GroupDHPuk == nil || len(friend.GroupSignPuk) == 0 {
		//utils.Log.Info().Msgf("好友列表为空")
		//好友信息中查询
		friend, ERR = ImProxyClient_FindUserinfo(addrSelf, userAddr)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
	}
	//if friend == nil {
	//	//好友申请列表中查
	//	friend, ERR = FindUserListByAddr(config.DBKEY_apply_remote_userlist, addrSelf, userAddr)
	//	if !ERR.CheckSuccess() {
	//		return nil, ERR
	//	}
	//}
	if friend == nil || friend.GroupDHPuk == nil || len(friend.GroupDHPuk) == 0 {
		//utils.Log.Info().Msgf("好友信息为空")
		return nil, utils.NewErrorBus(config.ERROR_CODE_IM_dh_not_exist, "")
	}
	dhPuk := keystore.Key{}
	copy(dhPuk[:], friend.GroupDHPuk)
	return &dhPuk, utils.NewErrorSuccess()
}

/*
保存发送者的index
*/
func ImProxyClient_SaveSendIndex(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet, addrFrom,
	addrTo nodeStore.AddressNet, indexBs []byte, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, key)
	keyDB, ERR := utilsleveldb.NewLeveldbKeyJoin(addrFrom.GetAddr(), addrTo.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	return config.LevelDB.SaveMap(*dbKey, *keyDB, indexBs, batch)
}

/*
查询发送者的index
*/
func ImProxyClient_FindSendIndex(key utilsleveldb.LeveldbKey, addrSelf nodeStore.AddressNet,
	addrFrom, addrTo nodeStore.AddressNet) (*big.Int, utils.ERROR) {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, key)
	keyDB, ERR := utilsleveldb.NewLeveldbKeyJoin(addrFrom.GetAddr(), addrTo.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(*dbKey, *keyDB)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return big.NewInt(0), utils.NewErrorSuccess()
	}
	index := new(big.Int).SetBytes(bs)
	return index, utils.NewErrorSuccess()
}

/*
删除发送者的index
*/
func ImProxyClient_RemoveSendIndex(key utilsleveldb.LeveldbKey, addrSelf *nodeStore.AddressNet,
	addrFrom, addrTo nodeStore.AddressNet, batch *leveldb.Batch) utils.ERROR {
	addrSelfKey, ERR := utilsleveldb.BuildLeveldbKey(addrSelf.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dbKey := utilsleveldb.JoinDbKey(*addrSelfKey, key)
	keyDB, ERR := utilsleveldb.NewLeveldbKeyJoin(addrFrom.GetAddr(), addrTo.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	err := config.LevelDB.RemoveMapByKey(*dbKey, *keyDB, batch)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

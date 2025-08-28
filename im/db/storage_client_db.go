package db

import (
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
添加或修改存储提供者服务器列表
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func StorageClient_AddStorageServer(sinfo *model.StorageServerInfo) utils.ERROR {
	bs, err := sinfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	key, ERR := utilsleveldb.BuildLeveldbKey(sinfo.Addr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	return config.LevelDB.SaveMap(*config.DBKEY_storage_serverlist, *key, *bs, nil)
}

/*
通过服务器地址，获取所有提供离线服务的节点信息
*/
func StorageClient_GetStorageServerListByAddrs(sAddr ...nodeStore.AddressNet) ([]model.StorageServerInfo, error) {
	keys := make([]utilsleveldb.LeveldbKey, 0)
	for _, one := range sAddr {
		keyOne, _ := utilsleveldb.BuildLeveldbKey(one.GetAddr())
		keys = append(keys, *keyOne)
	}
	items, err := config.LevelDB.FindMapByKeys(*config.DBKEY_storage_serverlist, keys...)
	if err != nil {
		return nil, err
	}
	sinfos := make([]model.StorageServerInfo, 0)
	for _, one := range items {
		sinfo, err := model.ParseStorageServerInfo(one.Value)
		if err != nil {
			return nil, err
		}
		sinfos = append(sinfos, *sinfo)
	}
	return sinfos, err
}

/*
获取所有提供离线服务的节点信息
*/
func StorageClient_GetStorageServerList() ([]model.StorageServerInfo, error) {
	items, err := config.LevelDB.FindMapAllToList(*config.DBKEY_storage_serverlist)
	if err != nil {
		return nil, err
	}
	sinfos := make([]model.StorageServerInfo, 0)
	for _, one := range items {
		sinfo, err := model.ParseStorageServerInfo(one.Value)
		if err != nil {
			return nil, err
		}
		sinfos = append(sinfos, *sinfo)
	}
	return sinfos, err
}

/*
保存一个未支付订单
*/
func StorageClient_SaveOrderFormNotPay(form *model.OrderForm) utils.ERROR {
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.NewLeveldbKeyJoin(form.ServerAddr.GetAddr(), form.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//开启事务
	err = config.LevelDB.OpenTransaction()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	defer func() {
		if ERR.CheckSuccess() {
			//事务提交
			err = config.LevelDB.Commit()
			if err != nil {
				config.LevelDB.Discard()
				utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
				return
			}
			return
		}
		//事务回滚
		config.LevelDB.Discard()
	}()
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_client_orders, *numberDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_client_ordersID_not_pay, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
保存一个已支付订单
*/
func StorageClient_SaveOrderFormInUse(form *model.OrderForm) utils.ERROR {
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.NewLeveldbKeyJoin(form.ServerAddr.GetAddr(), form.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//开启事务
	err = config.LevelDB.OpenTransaction()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	defer func() {
		if ERR.CheckSuccess() {
			//事务提交
			err = config.LevelDB.Commit()
			if err != nil {
				config.LevelDB.Discard()
				utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
				return
			}
			return
		}
		//事务回滚
		config.LevelDB.Discard()
	}()
	//保存订单
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_client_orders, *numberDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//客户端正在服务的订单列表
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_client_ordersID_inuse, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
删除一个超过服务期的订单
*/
func StorageClient_DelOrderFormInUse(order, renewalOrder *model.OrderForm) utils.ERROR {
	numberDBKey, ERR := utilsleveldb.NewLeveldbKeyJoin(order.ServerAddr.GetAddr(), order.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//开启事务
	err := config.LevelDB.OpenTransaction()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	defer func() {
		if ERR.CheckSuccess() {
			//事务提交
			err = config.LevelDB.Commit()
			if err != nil {
				config.LevelDB.Discard()
				utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
				return
			}
			return
		}
		//事务回滚
		config.LevelDB.Discard()
	}()
	//从正在服务的列表中删除
	err = config.LevelDB.RemoveMapByKey_Transaction(*config.DBKEY_storage_client_ordersID_inuse, *numberDBKey)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		return ERR
	}
	//保存到超期的列表中
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_client_ordersID_inuse_timeout, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//有续费订单
	if renewalOrder != nil {
		renewalNumberDBKey, ERR := utilsleveldb.NewLeveldbKeyJoin(renewalOrder.ServerAddr.GetAddr(), renewalOrder.Number)
		if !ERR.CheckSuccess() {
			return ERR
		}
		//把续费订单保存到正在服务期的订单列表中
		ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_client_ordersID_inuse, *renewalNumberDBKey, nil)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	return utils.NewErrorSuccess()
}

/*
加载未过期的已支付订单
*/
func StorageClient_LoadOrderFormInUse() ([]model.OrderForm, error) {
	items, err := config.LevelDB.FindMapAllToList(*config.DBKEY_storage_client_ordersID_inuse)
	if err != nil {
		return nil, err
	}
	keys := make([]utilsleveldb.LeveldbKey, 0)
	for _, one := range items {
		keys = append(keys, one.Key)
	}
	items, err = config.LevelDB.FindMapByKeys(*config.DBKEY_storage_client_orders, keys...)
	if err != nil {
		return nil, err
	}
	orders := make([]model.OrderForm, 0)
	for _, one := range items {
		orderOne, err := model.ParseOrderForm(one.Value)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *orderOne)
	}
	return orders, nil
}

/*
加载未过期的未支付订单
*/
func StorageClient_LoadOrderFormNotPay() ([]model.OrderForm, error) {
	items, err := config.LevelDB.FindMapAllToList(*config.DBKEY_storage_client_ordersID_inuse)
	if err != nil {
		return nil, err
	}
	keys := make([]utilsleveldb.LeveldbKey, 0)
	for _, one := range items {
		keys = append(keys, one.Key)
	}
	items, err = config.LevelDB.FindMapByKeys(*config.DBKEY_storage_client_orders, keys...)
	if err != nil {
		return nil, err
	}
	orders := make([]model.OrderForm, 0)
	for _, one := range items {
		orderOne, err := model.ParseOrderForm(one.Value)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *orderOne)
	}
	return orders, nil
}

/*
保存加密文件的密码，默认密码是文件的真实hash值
@plainHash    []byte    文件加密后的hash
@pwd          []byte    密码
*/
func StorageClient_SaveFilePwd(plainHash, pwd []byte) utils.ERROR {
	key, ERR := utilsleveldb.BuildLeveldbKey(plainHash)
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = config.LevelDB.SaveMap(*config.DBKEY_storage_client_file_pwd, *key, pwd, nil)
	//if err != nil {
	//	return ERR
	//}
	return ERR
}

/*
查询加密文件的密码
@plainHash    []byte    文件加密后的hash
*/
func StorageClient_FindFilePwd(plainHash []byte) ([]byte, utils.ERROR) {
	key, ERR := utilsleveldb.BuildLeveldbKey(plainHash)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	value, err := config.LevelDB.FindMap(*config.DBKEY_storage_client_file_pwd, *key)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return value, utils.NewErrorSuccess()
}

package db

import (
	"web3_gui/chain_boot/object_beans"
	"web3_gui/config"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
保存未支付订单
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Sharebox_Client_SaveOrderUnpaid(orderInfo *object_beans.OrderShareboxOrder) utils.ERROR {
	bs, err := orderInfo.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	key, ERR := utilsleveldb.BuildLeveldbKey(orderInfo.GetId())
	if !ERR.CheckSuccess() {
		return ERR
	}
	return config.LevelDB.SaveMap(*config.DBKEY_SHAREBOX_client_order_unpaid, *key, *bs, nil)
}

/*
查询文件的价格
@nickName    string                  昵称
@price       uint64                  单价
@addr        nodeStore.AddressNet    节点地址
@count       uint64                  次数
*/
func Sharebox_Client_FindOrderUnpaid(id []byte) (*object_beans.OrderShareboxOrder, utils.ERROR) {
	key, ERR := utilsleveldb.BuildLeveldbKey(id)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(*config.DBKEY_SHAREBOX_client_order_unpaid, *key)
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
func Sharebox_Client_SaveOrderPaid(orderId []byte, txHash, txBs *[]byte) utils.ERROR {
	//查询未支付订单
	order, ERR := Sharebox_Client_FindOrderUnpaid(orderId)
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
	cache.Set_Remove(config.DBKEY_SHAREBOX_client_order_unpaid, key)
	//保存已经支付订单
	cache.Set_Save(config.DBKEY_SHAREBOX_client_order_paid, key, orderBs)

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
func Sharebox_Client_SaveOrderUnpaidOvertime(orderId []byte) utils.ERROR {
	//查询未支付订单
	order, ERR := Sharebox_server_FindOrderUnpaid(orderId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if order == nil {
		return utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
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
	cache.Set_Remove(config.DBKEY_SHAREBOX_client_order_unpaid, key)
	//保存已经支付订单
	cache.Set_Save(config.DBKEY_SHAREBOX_client_order_unpaid_overtime, key, orderBs)
	err = config.LevelDB.Cache_CommitCache(cache)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

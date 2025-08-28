package db

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"slices"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
获取数据库中全局自增长ID
*/
func StorageServer_GetGenID() ([]byte, error) {
	dbItem, err := config.LevelDB.Find(*config.DBKEY_storage_server_GenID)
	if err != nil {
		return nil, err
	}
	if dbItem == nil {
		return nil, nil
	}
	return dbItem.Value, nil
}

/*
保存数据库中全局自增长ID
*/
func StorageServer_SaveGenID(number []byte) utils.ERROR {
	return config.LevelDB.Save(*config.DBKEY_storage_server_GenID, &number)
}

/*
设置自己提供存储服务信息
@info    *config.StorageServerInfo    信息
*/
func StorageServer_SetServerInfo(info *model.StorageServerInfo) utils.ERROR {
	bs, err := info.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return config.LevelDB.Save(*config.DBKEY_storage_serverinfo, bs)
}

/*
查询自己提供存储服务设置信息
@return    *config.StorageServerInfo    信息
*/
func StorageServer_GetServerInfo() (*model.StorageServerInfo, error) {
	item, err := config.LevelDB.Find(*config.DBKEY_storage_serverinfo)
	if err != nil {
		return nil, err
	}
	if item == nil || item.Value == nil || len(item.Value) == 0 {
		return nil, nil
	}
	return model.ParseStorageServerInfo(item.Value)
}

/*
获取存储服务器正在服务的订单id列表
*/
func StorageServer_GetStorageServerInUseOrdersIDs() ([]*model.OrderForm, error) {
	keys := make([]utilsleveldb.LeveldbKey, 0)
	dbitem, err := config.LevelDB.FindMapAllToList(*config.DBKEY_storage_server_ordersID_inuse)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			return nil, nil
		}
		return nil, err
	}
	for _, one := range dbitem {
		keys = append(keys, one.Key)
	}
	items, err := config.LevelDB.FindMapByKeys(*config.DBKEY_storage_server_orders, keys...)
	if err != nil {
		return nil, err
	}
	orders := make([]*model.OrderForm, 0, len(items))
	for _, one := range items {
		of, err := model.ParseOrderForm(one.Value)
		if err != nil {
			return nil, err
		}
		orders = append(orders, of)
	}
	return orders, nil
}

/*
保存一个未支付订单
*/
func StorageServer_SaveOrderFormNotPay(form *model.OrderForm) utils.ERROR {
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(form.Number)
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
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_orders, *numberDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_ordersID_not_pay, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
保存一个已支付订单
*/
func StorageServer_SaveOrderFormInUse(form *model.OrderForm) utils.ERROR {
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(form.Number)
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
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_orders, *numberDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存到正在服务期的订单列表
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_ordersID_inuse, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//查询服务器信息
	info, err := StorageServer_GetServerInfo()
	if err != nil {
		return utils.NewErrorSysSelf(err)
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
	ERR = config.LevelDB.Save_Transaction(*config.DBKEY_storage_serverinfo, bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
删除一个已支付订单
*/
func StorageServer_DelOrderFormInUse(form *model.OrderForm) utils.ERROR {
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(form.Number)
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
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_orders, *numberDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_ordersID_inuse, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
将一个未支付订单删除并保存到已支付订单中
*/
func StorageServer_MoveOrderToInUse(form *model.OrderForm) utils.ERROR {
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(form.Number)
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := form.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
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
	//刷新订单
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_orders, *numberDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//从未支付订单中删除
	err = config.LevelDB.RemoveMapByKey_Transaction(*config.DBKEY_storage_server_ordersID_not_pay, *numberDBKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//保存到已支付订单列表
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_ordersID_inuse, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
将一个未支付订单删除并保存到，过期未支付订单中
*/
func StorageServer_MoveOrderToNotPayTimeout(number []byte) utils.ERROR {
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(number)
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
	//从未支付订单中删除
	err = config.LevelDB.RemoveMapByKey_Transaction(*config.DBKEY_storage_server_ordersID_not_pay, *numberDBKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//保存到已支付订单列表
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_ordersID_not_pay_timeout, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
将一个已支付订单删除并保存到，服务已到期订单中
同时把续费订单放入正在服务的订单列表中
*/
func StorageServer_MoveOrderToInUseTimeout(order, renewalOrder *model.OrderForm) utils.ERROR {
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(order.Number)
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
	//从正在服务的订单中删除
	err = config.LevelDB.RemoveMapByKey_Transaction(*config.DBKEY_storage_server_ordersID_inuse, *numberDBKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//保存到服务到期列表中
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_ordersID_inuse_timeout, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//查询服务器信息
	info, err := StorageServer_GetServerInfo()
	if err != nil {
		return utils.NewErrorSysSelf(err)
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
		ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_ordersID_inuse, *renewalNumberDBKey, nil)
		if !ERR.CheckSuccess() {
			return ERR
		}
		//加上续费订单的空间
		info.Sold += renewalOrder.SpaceTotal
	}
	bs, err := info.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//保存新的服务器信息
	ERR = config.LevelDB.Save_Transaction(*config.DBKEY_storage_serverinfo, bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询一个用户的顶层目录
*/
func StorageServer_GetUserTopDirIndex(userAddr nodeStore.AddressNet) (*model.DirectoryIndex, utils.ERROR) {
	uAddrDBKey, ERR := utilsleveldb.BuildLeveldbKey(userAddr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(*config.DBKEY_storage_server_user_top_dir_index, *uAddrDBKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("查询到用户的顶层文件夹:%+v", bs)
	dirDBKey, ERR := utilsleveldb.BuildLeveldbKey(bs)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err = config.LevelDB.FindMap(*config.DBKEY_storage_server_dir_index, *dirDBKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	dirIndex, err := model.ParseDirectoryIndex(bs)
	//utils.Log.Info().Msgf("查询到用户的顶层文件夹:%+v %+v", bs, dirIndex)
	return dirIndex, utils.NewErrorSysSelf(err)
}

/*
保存一个用户的顶层目录
*/
func StorageServer_SaveUserTopDirIndex(userAddr nodeStore.AddressNet, dirIndex model.DirectoryIndex) utils.ERROR {
	uAddrDBKey, ERR := utilsleveldb.BuildLeveldbKey(userAddr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	dirIDDBKey, ERR := utilsleveldb.BuildLeveldbKey(dirIndex.ID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := dirIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
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
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_user_top_dir_index, *uAddrDBKey, dirIndex.ID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_dir_index, *dirIDDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询一个用户的目录
*/
func StorageServer_GetDirIndex(dirID []byte) (*model.DirectoryIndex, utils.ERROR) {
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(dirID)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	bs, err := config.LevelDB.FindMap(*config.DBKEY_storage_server_dir_index, *numberDBKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	di, err := model.ParseDirectoryIndex(bs)
	return di, utils.NewErrorSysSelf(err)
}

/*
查询多个用户的目录
*/
func StorageServer_GetDirIndexMore(dirID ...[]byte) ([]*model.DirectoryIndex, utils.ERROR) {
	dirIDkey := make([]utilsleveldb.LeveldbKey, 0)
	for _, one := range dirID {
		numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(one)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		dirIDkey = append(dirIDkey, *numberDBKey)
	}
	items, err := config.LevelDB.FindMapByKeys(*config.DBKEY_storage_server_dir_index, dirIDkey...)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			return nil, utils.NewErrorSuccess()
		}
		return nil, utils.NewErrorSysSelf(err)
	}
	dirs := make([]*model.DirectoryIndex, 0)
	for _, one := range items {
		dirOne, err := model.ParseDirectoryIndex(one.Value)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		dirs = append(dirs, dirOne)
	}
	return dirs, utils.NewErrorSuccess()
}

/*
保存一个用户的文件夹
*/
func StorageServer_SaveUserDirIndex(dirIndex model.DirectoryIndex) utils.ERROR {
	parentDBKey, ERR := utilsleveldb.BuildLeveldbKey(dirIndex.ParentID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	newDirIDkey, ERR := utilsleveldb.BuildLeveldbKey(dirIndex.ID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := dirIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//err = config.LevelDB.SaveMap(config.DBKEY_storage_server_dir_index, *newDirIDkey, *bs)
	//if err != nil {
	//	return err
	//}

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
	//保存新文件夹
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_dir_index, *newDirIDkey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//查询父文件夹
	parentDirBs, err := config.LevelDB.FindMap(*config.DBKEY_storage_server_dir_index, *parentDBKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("查询新文件夹信息bs:%+v", parentDirBs)
	newDIr, err := model.ParseDirectoryIndex(parentDirBs)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("查询新文件夹信息:%+v", newDIr)
	newDIr.DirsID = append(newDIr.DirsID, dirIndex.ID)
	//保存添加文件夹后的文件夹
	newDIrBs, err := newDIr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存新文件夹信息:%+v %+v", newDIr, newDIrBs)
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_dir_index, *parentDBKey, *newDIrBs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
保存一个用户正在上传的文件索引
*/
func StorageServer_SaveFileIndex(uAddr nodeStore.AddressNet, fileIndex model.FileIndex) utils.ERROR {
	//utils.Log.Info().Msgf("卡在哪里了")
	userAddrKey, ERR := utilsleveldb.BuildLeveldbKey(uAddr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := fileIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(fileIndex.Hash)
	if err != nil {
		return ERR
	}
	//utils.Log.Info().Msgf("卡在哪里了")

	//batch := new(leveldb.Batch)

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
	//utils.Log.Info().Msgf("卡在哪里了")
	//查询用户已经使用空间大小
	useSpaceSize := uint64(0)
	sizeBs, err := config.LevelDB.FindMap(*config.DBKEY_storage_server_user_spaceSize, *userAddrKey)
	//utils.Log.Info().Msgf("卡在哪里了")
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if sizeBs == nil {
		//没有记录
		sizeBs = utils.Uint64ToBytes(0)
	}
	//utils.Log.Info().Msgf("卡在哪里了")
	useSpaceSize = utils.BytesToUint64(sizeBs)
	useSpaceSize += fileIndex.FileSize
	sizeBs = utils.Uint64ToBytes(useSpaceSize)
	//保存用户已使用空间
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_user_spaceSize, *userAddrKey, sizeBs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存文件索引
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_file_index, *numberDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存正在上传中的文件索引
	ERR = config.LevelDB.SaveMapInMap_Transaction(*config.DBKEY_storage_server_upload_file_index, *userAddrKey, *numberDBKey, nil)
	//err = config.LevelDB.SaveMap_Transaction(config.DBKEY_storage_server_upload_file_index, *numberDBKey, nil)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
保存一个索引到指定文件夹
*/
func StorageServer_SaveFileIndexToDir(uAddr nodeStore.AddressNet, parentDirID []byte, fileIndex model.FileIndex) utils.ERROR {
	//utils.Log.Info().Msgf("卡在哪里了")
	userAddrKey, ERR := utilsleveldb.BuildLeveldbKey(uAddr.GetAddr())
	if !ERR.CheckSuccess() {
		return ERR
	}
	bs, err := fileIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(fileIndex.Hash)
	if !ERR.CheckSuccess() {
		return ERR
	}
	parentDirIDkey, ERR := utilsleveldb.BuildLeveldbKey(parentDirID)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("移动一个文件:%s %d", err.Error(), len(parentDirID))
		return ERR
	}
	//utils.Log.Info().Msgf("卡在哪里了")
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
	//utils.Log.Info().Msgf("卡在哪里了")
	//查询用户已经使用空间大小
	useSpaceSize := uint64(0)
	sizeBs, err := config.LevelDB.FindMap(*config.DBKEY_storage_server_user_spaceSize, *userAddrKey)
	//utils.Log.Info().Msgf("卡在哪里了")
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if sizeBs == nil {
		//没有记录
		sizeBs = utils.Uint64ToBytes(0)
	}
	//utils.Log.Info().Msgf("卡在哪里了")
	useSpaceSize = utils.BytesToUint64(sizeBs)
	useSpaceSize += fileIndex.FileSize
	sizeBs = utils.Uint64ToBytes(useSpaceSize)
	//保存用户已使用空间
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_user_spaceSize, *userAddrKey, sizeBs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存文件索引
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_file_index, *numberDBKey, *bs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存正在上传中的文件索引
	//err = config.LevelDB.SaveMapInMap_Transaction(config.DBKEY_storage_server_upload_file_index, *userAddrKey, *numberDBKey, nil)
	////err = config.LevelDB.SaveMap_Transaction(config.DBKEY_storage_server_upload_file_index, *numberDBKey, nil)
	//if err != nil {
	//	//事务回滚
	//	config.LevelDB.Discard()
	//	return err
	//}
	//查父文件夹
	dirBs, err := config.LevelDB.FindMap(*config.DBKEY_storage_server_dir_index, *parentDirIDkey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("查询新文件夹信息bs:%+v", bs)
	newDIr, err := model.ParseDirectoryIndex(dirBs)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("查询新文件夹信息:%+v", newDIr)
	newDIr.FilesID = append(newDIr.FilesID, fileIndex.Hash)
	//保存删除文件后的文件夹
	newDIrBs, err := newDIr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存新文件夹信息:%+v %+v", parentDirID, newDIr)
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_dir_index, *parentDirIDkey, *newDIrBs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
移动一个文件
*/
func StorageServer_MoveFileIndex(uAddr nodeStore.AddressNet, oldDirID, newDirID, fileID []byte) utils.ERROR {
	//utils.Log.Info().Msgf("移动一个文件:%v %v %v", oldDirID, newDirID, fileID)
	var ERR utils.ERROR
	var err error
	var oldDirIDkey *utilsleveldb.LeveldbKey
	if oldDirID != nil && len(oldDirID) != 0 {
		oldDirIDkey, ERR = utilsleveldb.BuildLeveldbKey(oldDirID)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	newDirIDkey, ERR := utilsleveldb.BuildLeveldbKey(newDirID)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("移动一个文件:%s %d", err.Error(), len(newDirID))
		return ERR
	}
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(fileID)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("移动一个文件:%s %d", err.Error(), len(fileID))
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
	//查文件
	bs, err := config.LevelDB.FindMap(*config.DBKEY_storage_server_file_index, *numberDBKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	fileIndex, err := model.ParseFileIndex(bs)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("查询到的文件信息:%+v", fileIndex)
	//修改所属文件夹
	for i, one := range fileIndex.UserAddr {
		if bytes.Equal(one, uAddr.GetAddr()) {
			fileIndex.DirID[i] = newDirID
			break
		}
	}
	//fileIndex.DirID = newDirID
	bs1, err := fileIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存文件信息:%+v", fileIndex)
	//保存新的文件
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_file_index, *numberDBKey, *bs1)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//查旧文件夹
	if oldDirID != nil && len(oldDirID) != 0 {
		bs, err = config.LevelDB.FindMap(*config.DBKEY_storage_server_dir_index, *oldDirIDkey)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		oldDIr, err := model.ParseDirectoryIndex(bs)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		for i, one := range oldDIr.FilesID {
			if bytes.Equal(one, fileID) {
				temp := oldDIr.FilesID[:i]
				temp = append(temp, oldDIr.FilesID[i+1:]...)
				oldDIr.FilesID = temp
				break
			}
		}
		//保存删除文件后的文件夹
		oldDirBs, err := oldDIr.Proto()
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_dir_index, *oldDirIDkey, *oldDirBs)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}

	//查新文件夹
	bs, err = config.LevelDB.FindMap(*config.DBKEY_storage_server_dir_index, *newDirIDkey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("查询新文件夹信息bs:%+v", bs)
	newDIr, err := model.ParseDirectoryIndex(bs)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("查询新文件夹信息:%+v", newDIr)
	newDIr.FilesID = append(newDIr.FilesID, fileID)
	//保存删除文件后的文件夹
	newDIrBs, err := newDIr.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存新文件夹信息:%+v %+v", newDIr, newDIrBs)
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_dir_index, *newDirIDkey, *newDIrBs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
查询多个文件
*/
func StorageServer_GetFileIndexMore(fileIDs ...[]byte) ([]*model.FileIndex, utils.ERROR) {
	fileIDkey := make([]utilsleveldb.LeveldbKey, 0)
	for _, one := range fileIDs {
		numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(one)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		fileIDkey = append(fileIDkey, *numberDBKey)
	}
	//
	items, err := config.LevelDB.FindMapByKeys(*config.DBKEY_storage_server_file_index, fileIDkey...)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			return nil, utils.NewErrorSuccess()
		}
		return nil, utils.NewErrorSysSelf(err)
	}
	files := make([]*model.FileIndex, 0)
	for _, one := range items {
		dirOne, err := model.ParseFileIndex(one.Value)
		//utils.Log.Info().Msgf("查询到的文件索引:%+v", dirOne)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		files = append(files, dirOne)
	}
	return files, utils.NewErrorSuccess()
}

/*
递归查询多个文件夹中的文件
@return    []model.DirectoryIndex    所有文件夹列表
@return    []model.FileIndex         所有文件列表
@return    [][]byte                  所有文件夹ID
@return    [][]byte                  所有文件ID
@return    error                     错误
*/
func StorageServer_GetDirIndexRecursion(dirIDs [][]byte) ([]*model.DirectoryIndex, []*model.FileIndex, [][]byte, [][]byte, utils.ERROR) {
	//递归查询文件夹
	dirIndexAll := make([]*model.DirectoryIndex, 0)
	fileIndexAll := make([]*model.FileIndex, 0)
	dirAllIDs := make([][]byte, 0)
	//dirAllIDs = append(dirAllIDs, dirIDs...)
	fileIDs := make([][]byte, 0)
	for {
		dirIndexs, ERR := StorageServer_GetDirIndexMore(dirIDs...)
		if !ERR.CheckSuccess() {
			return nil, nil, nil, nil, ERR
		}
		dirIDs = make([][]byte, 0)
		for _, one := range dirIndexs {
			fileIndexs, ERR := StorageServer_GetFileIndexMore(one.FilesID...)
			if !ERR.CheckSuccess() {
				return nil, nil, nil, nil, ERR
			}
			one.Files = fileIndexs
			for _, fileOne := range fileIndexs {
				fileIDs = append(fileIDs, fileOne.Hash)
				fileIndexAll = append(fileIndexAll, fileOne)
			}
			dirIDs = append(dirIDs, one.DirsID...)
			dirIndexAll = append(dirIndexAll, one)
			dirAllIDs = append(dirAllIDs, one.ID)
		}
		if len(dirIDs) == 0 {
			break
		}
	}
	return dirIndexAll, fileIndexAll, dirAllIDs, fileIDs, utils.NewErrorSuccess()
}

/*
删除文件和文件夹，前提是这些文件和文件夹都在同一父文件夹中
@uAddr      nodeStore.AddressNet    用户地址
@dirIDs     [][]byte                修改的文件索引
@fileIDs    [][]byte                删除的文件索引
*/
func StorageServer_DelDirAndFileIndex(uAddr nodeStore.AddressNet, dirIDs, fileIDs [][]byte) ([]model.FileIndex, []model.FileIndex, utils.ERROR) {
	//开启事务
	err := config.LevelDB.OpenTransaction()
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	updateFileIndexs, delFileIndexs, ERR := StorageServer_delDirAndFileIndex_transaction(uAddr, dirIDs, fileIDs)
	if ERR.CheckSuccess() {
		err = config.LevelDB.Commit()
		if err != nil {
			config.LevelDB.Discard()
			utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
			return nil, nil, utils.NewErrorSysSelf(err)
		}
		return updateFileIndexs, delFileIndexs, utils.NewErrorSysSelf(err)
	}
	config.LevelDB.Discard()
	return updateFileIndexs, delFileIndexs, utils.NewErrorSuccess()
}

/*
删除一个用户的所有文件和文件夹
@uAddr      nodeStore.AddressNet    用户地址
@dirIDs     [][]byte                修改的文件索引
@fileIDs    [][]byte                删除的文件索引
*/
func StorageServer_DelUserDirAndFileAll(userAddr nodeStore.AddressNet) ([]model.FileIndex, []model.FileIndex, utils.ERROR) {
	//开启事务
	err := config.LevelDB.OpenTransaction()
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	//查询用户顶层目录
	dir, ERR := StorageServer_GetUserTopDirIndex(userAddr)
	if !ERR.CheckSuccess() {
		config.LevelDB.Discard()
		return nil, nil, ERR
	}
	utils.Log.Info().Msgf("查询到用户顶层目录:%+v", dir)
	//迭代查询文件夹中的文件
	_, _, dirIDs, fileIDs, ERR := StorageServer_GetDirIndexRecursion([][]byte{dir.ID})
	if !ERR.CheckSuccess() {
		config.LevelDB.Discard()
		return nil, nil, ERR
	}
	utils.Log.Info().Msgf("迭代查询文件夹中的文件:%+v %+v", dirIDs, fileIDs)
	//因为文件夹和文件不在同一个父文件夹中，所以分开删除
	//先删除文件
	updateFileIndexs, delFileIndexs, ERR := StorageServer_delDirAndFileIndex_transaction(userAddr, nil, fileIDs)
	if !ERR.CheckSuccess() {
		config.LevelDB.Discard()
		return nil, nil, ERR
	}
	//再删除文件夹
	_, _, ERR = StorageServer_delDirAndFileIndex_transaction(userAddr, dirIDs, nil)
	if !ERR.CheckSuccess() {
		config.LevelDB.Discard()
		return nil, nil, ERR
	}
	err = config.LevelDB.Commit()
	if err != nil {
		config.LevelDB.Discard()
		utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	utils.Log.Info().Msgf("删除的文件:%+v", delFileIndexs)
	return updateFileIndexs, delFileIndexs, ERR
}

/*
删除文件和文件夹，前提是这些文件和文件夹都在同一父文件夹中
@uAddr      nodeStore.AddressNet    用户地址
@dirIDs     [][]byte                修改的文件索引
@fileIDs    [][]byte                删除的文件索引
*/
func StorageServer_delDirAndFileIndex_transaction(uAddr nodeStore.AddressNet, dirIDs, fileIDs [][]byte) ([]model.FileIndex, []model.FileIndex, utils.ERROR) {
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	userAddrKey, ERR := utilsleveldb.BuildLeveldbKey(uAddr.GetAddr())
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}

	//查询待删除的文件夹，判断他们的父文件夹是否同一个
	delDirIndexs, ERR := StorageServer_GetDirIndexMore(dirIDs...)
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	//查询待删除的文件,判断他们的父文件夹是否同一个
	fileIndexs, ERR := StorageServer_GetFileIndexMore(fileIDs...)
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	var parentID []byte
	//判断他们的父文件夹是否同一个
	if len(delDirIndexs) > 0 {
		parentID = delDirIndexs[0].ParentID
	} else if len(fileIndexs) > 0 {
		for i, one := range fileIndexs[0].UserAddr {
			if bytes.Equal(one, uAddr.GetAddr()) {
				parentID = fileIndexs[0].DirID[i]
				break
			}
		}
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//没有待删除的文件，则返回成功
	if parentID == nil || len(parentID) == 0 {
		return nil, nil, utils.NewErrorSuccess()
	}
	//判断所有文件的父文件夹是否同一个
	for _, fileIndex := range fileIndexs {
		have := false
		for i, one := range fileIndex.UserAddr {
			if bytes.Equal(one, uAddr.GetAddr()) && bytes.Equal(fileIndex.DirID[i], parentID) {
				have = true
				break
			}
		}
		if !have {
			return nil, nil, utils.NewErrorBus(config.ERROR_CODE_storage_del_dirAndFile_NotSameFolder, "")
		}
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//判断所有文件夹的父文件夹是否同一个
	for _, dirIndex := range delDirIndexs {
		if !bytes.Equal(dirIndex.ParentID, parentID) {
			return nil, nil, utils.NewErrorBus(config.ERROR_CODE_storage_del_dirAndFile_NotSameFolder, "")
		}
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//查询这个父文件夹
	parentDir, ERR := StorageServer_GetDirIndex(parentID)
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	//判断父文件夹权限
	if !bytes.Equal(parentDir.UAddr.GetAddr(), uAddr.GetAddr()) {
		return nil, nil, utils.NewErrorBus(config.ERROR_CODE_storage_auth_file_No_permission, "")
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//删除子文件夹
	parentDir.DirsID = slices.DeleteFunc(parentDir.DirsID, func(bs []byte) bool {
		for _, one := range dirIDs {
			if bytes.Equal(one, bs) {
				return true
			}
		}
		return false
	})
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//删除子文件
	parentDir.FilesID = slices.DeleteFunc(parentDir.FilesID, func(bs []byte) bool {
		for _, one := range fileIDs {
			if bytes.Equal(one, bs) {
				return true
			}
		}
		return false
	})
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//修改这个父文件夹
	parentIDKey, ERR := utilsleveldb.BuildLeveldbKey(parentID)
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	bs, err := parentDir.Proto()
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_dir_index, *parentIDKey, *bs)
	if !ERR.CheckSuccess() {
		//事务回滚
		return nil, nil, ERR
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//递归查询文件夹下的子文件夹和文件
	delDirIndexs, fileIndexAll, _, _, ERR := StorageServer_GetDirIndexRecursion(dirIDs)
	if !ERR.CheckSuccess() {
		return nil, nil, ERR
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//待删除的文件。当用户引用数量为0时，这个文件被彻底删除
	delFileIndexs := make([]model.FileIndex, 0)
	//待修改的文件。减少所属用户后需要刷新数据库保存
	updateFileIndexs := make([]model.FileIndex, 0)
	for _, one := range append(fileIndexAll, fileIndexs...) {
		if !one.SubUser(uAddr) {
			return nil, nil, utils.NewErrorBus(config.ERROR_CODE_storage_auth_file_No_permission, "")
		}
		if len(one.UserAddr) == 0 {
			delFileIndexs = append(delFileIndexs, *one)
		} else {
			updateFileIndexs = append(updateFileIndexs, *one)
		}
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//删除文件夹，循环删除文件夹索引
	for _, one := range dirIDs {
		dirIDKey, ERR := utilsleveldb.BuildLeveldbKey(one)
		if !ERR.CheckSuccess() {
			return nil, nil, ERR
		}
		err = config.LevelDB.RemoveMapByKey_Transaction(*config.DBKEY_storage_server_dir_index, *dirIDKey)
		if err != nil {
			//事务回滚
			return nil, nil, utils.NewErrorSysSelf(err)
		}
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//删除文件索引
	for _, one := range delFileIndexs {
		fileIDKey, ERR := utilsleveldb.BuildLeveldbKey(one.Hash)
		if !ERR.CheckSuccess() {
			return nil, nil, ERR
		}
		err = config.LevelDB.RemoveMapByKey_Transaction(*config.DBKEY_storage_server_file_index, *fileIDKey)
		if err != nil {
			//事务回滚
			return nil, nil, utils.NewErrorSysSelf(err)
		}
		err = config.LevelDB.RemoveMapInMapByKeyIn_Transaction(*config.DBKEY_storage_server_upload_file_index, *userAddrKey, *fileIDKey)
		if err != nil {
			//事务回滚
			return nil, nil, utils.NewErrorSysSelf(err)
		}
	}
	utils.Log.Info().Msgf("删除文件和文件夹:%+v %+v", dirIDs, fileIDs)
	//修改文件索引
	for _, one := range updateFileIndexs {
		fileIDKey, ERR := utilsleveldb.BuildLeveldbKey(one.Hash)
		if !ERR.CheckSuccess() {
			return nil, nil, ERR
		}
		bs, err := one.Proto()
		if err != nil {
			return nil, nil, utils.NewErrorSysSelf(err)
		}
		ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_file_index, *fileIDKey, *bs)
		//err = config.LevelDB.RemoveMapByKey_Transaction(config.DBKEY_storage_server_file_index, *fileIDKey)
		if !ERR.CheckSuccess() {
			//事务回滚
			return nil, nil, ERR
		}
	}

	utils.Log.Info().Msgf("删除的所有文件夹:%d %+v", len(delDirIndexs), delDirIndexs)
	utils.Log.Info().Msgf("删除的所有文件:%d %+v", len(delFileIndexs), delFileIndexs)

	return updateFileIndexs, delFileIndexs, utils.NewErrorSysSelf(err)
}

/*
获取所有用户已经使用的空间大小
*/
func StorageServer_GetAllUserSpaceUseSize() (map[string]uint64, utils.ERROR) {
	dbItem, err := config.LevelDB.FindMapAllToList(*config.DBKEY_storage_server_user_spaceSize)
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
修改文件夹名称
*/
func StorageServer_UpdateDirName(uAddr nodeStore.AddressNet, dirID []byte, newName string) utils.ERROR {
	var ERR utils.ERROR
	newDirIDkey, ERR := utilsleveldb.BuildLeveldbKey(dirID)
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
	//查询文件夹
	var dirIndexList []*model.DirectoryIndex
	dirIndexList, ERR = StorageServer_GetDirIndexMore(dirID)
	if ERR.CheckFail() {
		return ERR
	}
	if len(dirIndexList) == 0 {
		ERR = utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
		return ERR
	}
	dirIndex := dirIndexList[0]
	//判断是否所属用户
	if !bytes.Equal(dirIndex.UAddr.GetAddr(), uAddr.GetAddr()) {
		ERR = utils.NewErrorBus(config.ERROR_CODE_storage_auth_file_No_permission, "")
		return ERR
	}
	dirIndex.Name = newName
	newDIrBs, err := dirIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存新文件夹信息:%+v %+v", newDIr, newDIrBs)
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_dir_index, *newDirIDkey, *newDIrBs)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
修改文件名称
*/
func StorageServer_UpdateFileName(uAddr nodeStore.AddressNet, fileID []byte, newName string) utils.ERROR {
	//utils.Log.Info().Msgf("卡在哪里了")
	var ERR utils.ERROR
	numberDBKey, ERR := utilsleveldb.BuildLeveldbKey(fileID)
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("移动一个文件:%s %d", err.Error(), len(fileID))
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
	//查询文件索引
	filesIndex, ERR := StorageServer_GetFileIndexMore(fileID)
	if ERR.CheckFail() {
		return ERR
	}
	if len(filesIndex) == 0 {
		ERR = utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
		return ERR
	}
	fileIndex := filesIndex[0]
	for i, one := range fileIndex.UserAddr {
		if bytes.Equal(one, uAddr.GetAddr()) {
			fileIndex.Name[i] = newName
			break
		}
	}
	bs1, err := fileIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存文件信息:%+v", fileIndex)
	//保存新的文件
	ERR = config.LevelDB.SaveMap_Transaction(*config.DBKEY_storage_server_file_index, *numberDBKey, *bs1)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

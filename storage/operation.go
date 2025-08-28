package storage

import (
	"bytes"
	"github.com/gogo/protobuf/proto"
	"time"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
客户端向存储提供商获取一个租用空间的订单
@serverAddr    uint64     //存储提供商地址
@spaceTotal    uint64     //购买空间数量
@useTime       uint64     //空间使用时间 单位：1天
*/
func ClientGetOrders(serverAddr nodeStore.AddressNet, spaceTotal, useTime uint64) (*model.OrderForm, utils.ERROR) {
	//utils.Log.Info().Msgf("下单")
	order := &model.OrderForm{
		UserAddr:   *Node.GetNetId(), //这里填别人的地址，相当于赠送给别人
		SpaceTotal: spaceTotal,
		UseTime:    useTime,
	}
	bs, err := order.Proto()
	//utils.Log.Info().Msgf("本次购买空间:%d %d %+v", spaceTotal, order.SpaceTotal, bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("下单")
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_getorders, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取订单网络错误:%s", ERR.String())
		return nil, ERR
	}
	//utils.Log.Info().Msgf("下单")
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	//utils.Log.Info().Msgf("下单")
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("下单")
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//utils.Log.Info().Msgf("下单")
	//返回成功
	order, err = model.ParseOrderForm(nr.Data)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("返回订单:%+v", order)
	if order.TotalPrice > 0 {
		ERR := StClient.AddOrdersUnpaid(order)
		return order, ERR
	}
	//utils.Log.Info().Msgf("下单")
	ERR = StClient.AddOrders(order)
	//免支付订单
	return order, ERR
}

/*
获取续费订单
@serverAddr    uint64     //存储提供商地址
@spaceTotal    uint64     //购买空间数量
@useTime       uint64     //空间使用时间 单位：1天
*/
func ClientGetRenewalOrders(preNumber []byte, serverAddr nodeStore.AddressNet, spaceTotal, useTime uint64) (*model.OrderForm, utils.ERROR) {
	order := &model.OrderForm{
		PreNumber:  preNumber,
		SpaceTotal: spaceTotal,
		UseTime:    useTime,
	}
	bs, err := order.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_getRenewalOrders, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取订单网络错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}

	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	order, err = model.ParseOrderForm(nr.Data)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	if order.TotalPrice > 0 {
		ERR := db.StorageClient_SaveOrderFormNotPay(order)
		return order, ERR
	}
	ERR = db.StorageClient_SaveOrderFormInUse(order)
	return order, ERR
}

/*
获取自己可用空间
*/
func GetFreeSpace(serverAddr nodeStore.AddressNet) (uint64, uint64, utils.ERROR) {
	np := utils.NewNetParams(config.NET_protocol_version_v1, nil)
	bs, err := np.Proto()
	if err != nil {
		return 0, 0, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_getFreeSpace, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取订单网络错误:%s", ERR.String())
		return 0, 0, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return 0, 0, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return 0, 0, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return 0, 0, nr.ConvertERROR()
	}
	//返回成功
	spacesTotal := utils.BytesToUint64(nr.Data[:8])
	spacesUse := utils.BytesToUint64(nr.Data[8:16])
	return spacesTotal, spacesUse, utils.NewErrorSuccess()
}

/*
获取文件列表
*/
func GetFileList(serverAddr nodeStore.AddressNet, dirID []byte) (*model.DirectoryIndex, utils.ERROR) {
	np := utils.NewNetParams(config.NET_protocol_version_v1, dirID)
	bs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_getFileList, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("获取订单网络错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	dIndex, err := model.ParseDirectoryIndexMore(nr.Data)
	return dIndex, utils.NewErrorSysSelf(err)
}

/*
发送上传文件的索引
*/
func SendUploadFileIndex(serverAddr nodeStore.AddressNet, fileIndex *model.FileIndex) utils.ERROR {
	bs, err := fileIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_upload_fileindex, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("上传文件索引错误:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
	//nr, err := model.ParseNetResult(nr.Data)
	//if err != nil {
	//	utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
	//	return utils.NewErrorSysSelf(err)
	//}
	//if nr.Code == config.NET_CODE_storage_Insufficient_user_space {
	//	//用户空间不足
	//	return config.ERROR_Insufficient_user_space_size
	//}
	//return nil
}

/*
发送文件保存到数据库完成
*/
func SendUploadFileSaveDB(userAddr nodeStore.AddressNet, ERR utils.ERROR) utils.ERROR {
	//utils.Log.Info().Msgf("通知客户端已经保存到数据库:%s", ERR.String())
	bs, err := ERR.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_upload_savedb_finish, &userAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("上传文件索引错误:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
}

/*
创建新文件夹
*/
func SendCreateNewDir(serverAddr nodeStore.AddressNet, parentDirID []byte, newDirName string) utils.ERROR {
	dirIndex := model.DirectoryIndex{
		ParentID: parentDirID, //父文件夹ID
		Name:     newDirName,  //文件夹名称
	}
	bs, err := dirIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_upload_create_dir, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("创建新文件夹错误:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
}

/*
删除多个文件和文件夹
*/
func SendDelDirAndFile(serverAddr nodeStore.AddressNet, dirIDs, fileIDs [][]byte) utils.ERROR {
	bss := go_protos.Bytes{
		List:  dirIDs,
		List2: fileIDs,
	}
	bs, err := bss.Marshal()
	if err != nil {
		utils.Log.Info().Msgf("删除文件失败:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, bs)
	nbs, err := np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_upload_delete_dir_and_file, &serverAddr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("删除文件错误:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
}

/*
去服务器循环查询目录中包含的目录及文件列表
*/
func ParseFilesWithin(serverAddr nodeStore.AddressNet, dirIDs [][]byte) ([]*model.DirectoryIndex, utils.ERROR) {
	bss := go_protos.Bytes{
		List: dirIDs,
	}
	bs, err := bss.Marshal()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, bs)
	nbs, err := np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_download_select_dir_fileindex, &serverAddr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("查询文件夹中的文件列表错误:%s", ERR.String())
		return nil, ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nil, nr.ConvertERROR()
	}
	//返回成功
	err = proto.Unmarshal(nr.Data, &bss)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	dirs := make([]*model.DirectoryIndex, 0)
	for _, one := range bss.List {
		dirIndex, err := model.ParseDirectoryIndexMore(one)
		if err != nil {
			utils.Log.Info().Msgf("查询文件夹中的文件列表错误:%s", err.Error())
			return nil, utils.NewErrorSysSelf(err)
		}
		dirs = append(dirs, dirIndex)
	}

	//utils.Log.Info().Msgf("查询并返回的文件夹:%+v", dirs)
	//给所有的文件夹找父文件夹
	topDirs := make([]*model.DirectoryIndex, 0)
	haveParent := false
	for i := len(dirs); i > 0; i-- {
		haveParent = false
		dirChild := dirs[i-1]
		//utils.Log.Info().Msgf("本次查找的子文件夹:%+v", dirChild)
		for j := i - 1; j >= 0; j-- {
			dirParent := dirs[j]
			//utils.Log.Info().Msgf("本次查找的父文件夹:%+v", dirParent)
			//对比是否属于这个父文件夹
			for _, idOne := range dirParent.DirsID {
				if bytes.Equal(idOne, dirChild.ID) {
					//utils.Log.Info().Msgf("找到了父亲文件夹:%+v %+v", idOne, dirChild.ID)
					if dirParent.Dirs == nil {
						dirParent.Dirs = make([]*model.DirectoryIndex, 0)
					}
					dirParent.Dirs = append(dirParent.Dirs, dirChild)
					dirChild.ParentDir = dirParent
					haveParent = true
					break
				}
			}
			if haveParent {
				break
			}
		}
		//如果没有找到父母文件夹，则属于顶层文件夹
		if !haveParent {
			topDirs = append(topDirs, dirChild)
		}
	}
	//utils.Log.Info().Msgf("整理好的文件夹:%+v", topDirs)
	return topDirs, utils.NewErrorSuccess()
}

/*
申请下载文件
*/
func SendApplyDownload(serverAddr nodeStore.AddressNet, fileIndex model.FileIndex) utils.ERROR {
	bs, err := fileIndex.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_download_apply, &serverAddr, bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("申请下载文件:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
}

/*
修改文件或文件夹名称
*/
func SendUpdateDirAndFile(serverAddr nodeStore.AddressNet, dirID, fileID []byte, newName string) utils.ERROR {
	//这里借用了此对象为参数序列化
	di := model.DirectoryIndex{
		ID:       fileID,  //文件ID
		ParentID: dirID,   //文件夹ID
		Name:     newName, //新名称
	}
	bs, err := di.Proto()
	if err != nil {
		utils.Log.Info().Msgf("修改文件失败:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *bs)
	nbs, err := np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_STORAGE_upload_update_name, &serverAddr, nbs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("删除文件 错误:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	//返回成功
	return utils.NewErrorSuccess()
}

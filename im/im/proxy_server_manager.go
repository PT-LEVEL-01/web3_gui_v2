package im

import (
	"bytes"
	"encoding/hex"
	"github.com/oklog/ulid/v2"
	"sync"
	"time"
	chainConfig "web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain_boot/chain_plus"
	"web3_gui/chain_orders"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/im/subscription"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/storage"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

var StaticProxyServerManager *ProxyServerManager

type ProxyServerManager struct {
	lock                 *sync.RWMutex               //
	AuthManager          *storage.AuthManager        //权限管理
	DBC                  *storage.DBCollections      //数据库管理
	UnpaidOrders         map[string]*model.OrderForm //未支付订单.key:string=用户地址;value:*model.OrderForm=订单;
	UnpaidOrdersId       map[string]*model.OrderForm //未支付订单.key:string=订单id;value:*model.OrderForm=订单;
	RenewalUnpaidOrder   map[string]*model.OrderForm //续费未支付订单.key:string=用户地址;value:*model.OrderForm=订单;
	RenewalUnpaidOrderId map[string]*model.OrderForm //续费未支付订单.key:string=订单id;value:*model.OrderForm=订单;
	Orders               map[string]*model.OrderForm //已支付订单.key:string=订单id;value:*model.OrderForm=订单;
	SellingLock          uint64                      //未支付订单锁定空间大小
	groupKnitManager     *GroupKnitManager           //客户端没有代理的时候，只有自己作为群服务器
	gsmm                 *GroupSyncMinorManager      //
	syncServer           *SyncServer                 //服务器同步数据链管理器
	syncCount            *chain_plus.ChainSyncCount  //区块链同步统计程序
}

func CreateProxyServerManager(db *utilsleveldb.LevelDB) (*ProxyServerManager, utils.ERROR) {
	//dbc := CreateDBCollections()
	//storageServerInfo, err := db.ImProxyServer_GetServerInfo()
	//if err != nil {
	//	utils.Log.Error().Msgf("广播云存储提供者节点时查询本节点信息错误:%s", err.Error())
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	//if storageServerInfo != nil {
	//	ERR := dbc.AddDBone(storageServerInfo.Directory...)
	//	if !ERR.CheckSuccess() {
	//		return nil, ERR
	//	}
	//}

	gsmm, ERR := NewGroupSyncMinorManager()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}

	//
	auth := storage.CreateAuthManager("")
	ss := ProxyServerManager{
		lock:        new(sync.RWMutex), //
		AuthManager: auth,
		//DBC:             dbc,
		UnpaidOrders:         make(map[string]*model.OrderForm), //未支付订单
		UnpaidOrdersId:       make(map[string]*model.OrderForm), //
		RenewalUnpaidOrder:   make(map[string]*model.OrderForm), //续费订单
		RenewalUnpaidOrderId: make(map[string]*model.OrderForm), //
		Orders:               make(map[string]*model.OrderForm), //已支付订单
		gsmm:                 gsmm,                              //
		syncServer:           NewSyncServer(),                   //
	}
	ERR = ss.LoadOrdersAndInit()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	ERR = ss.LoadUserSpacesUse()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	ss.syncCount = chain_plus.NewChainSyncCount(db, config.DBKEY_improxy_server_chain_count_height, ss.countBlockOne)
	return &ss, utils.NewErrorSuccess()
}

/*
查询一个用户的未支付订单
*/
func (this *ProxyServerManager) FindUnpaidOrders(userAddr nodeStore.AddressNet) *model.OrderForm {
	//有未支付的订单，先支付
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	unpaidOrders, ok := this.UnpaidOrders[utils.Bytes2string(userAddr.GetAddr())]
	if !ok {
		return nil
	}
	return unpaidOrders
}

/*
创建订单
*/
func (this *ProxyServerManager) CreateOrders(userAddr nodeStore.AddressNet, spaceTotal utils.Byte, useTime uint64) (*model.OrderForm, utils.ERROR) {
	//要等拉取到最高区块，才能接收订单
	if !this.syncCount.GetChainPullFinish() {
		return nil, utils.NewErrorBus(config.ERROR_CODE_order_chain_not_finish, "")
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	//查询一个用户的未支付订单
	unpaidOrders, ok := this.UnpaidOrders[utils.Bytes2string(userAddr.GetAddr())]
	if ok {
		return unpaidOrders, utils.NewErrorBus(config.ERROR_CODE_order_not_pay, "")
	}
	unpaidOrders, ok = this.RenewalUnpaidOrder[utils.Bytes2string(userAddr.GetAddr())]
	if ok {
		return unpaidOrders, utils.NewErrorBus(config.ERROR_CODE_order_not_pay, "")
	}

	storageServerInfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*Node.GetNetId(), true)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("广播云存储提供者节点时查询本节点信息错误:%s", ERR.String())
		return nil, ERR
	}

	//用户空闲空间限制
	spaceTotalPay, _, remain := this.AuthManager.QueryUserSpaces(userAddr)
	if spaceTotal+remain/utils.GB > utils.Byte(storageServerInfo.UserFreelimit) {
		remainGB := remain / utils.GB
		if remainGB%utils.GB > 0 {
			remainGB += 1
		}
		spaceTotal = utils.Byte(storageServerInfo.UserFreelimit) - remain/utils.GB
		//utils.Log.Info().Msgf("购买空间数量:%d %d %d %d", spaceTotal, spaceTotal, remain, storageServerInfo.UserFreelimit)
	}
	if spaceTotal == 0 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_over_free_space_limit, "")
	}
	//检查用户可以购买的空间总量
	if (spaceTotalPay/utils.GB)+spaceTotal > utils.Byte(storageServerInfo.UserCanTotal) {
		spaceTotal = utils.Byte(storageServerInfo.UserCanTotal) - spaceTotalPay
		//utils.Log.Info().Msgf("购买空间数量:%d %d %d %d", spaceTotal, spaceTotalPay, spaceTotal, storageServerInfo.UserCanTotal)
	}
	if spaceTotal == 0 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_over_pay_space_limit, "")
	}
	//不能超过最大时间
	if useTime > storageServerInfo.UseTimeMax {
		useTime = storageServerInfo.UseTimeMax
	}
	total := utils.Byte(storageServerInfo.PriceUnit) * spaceTotal * utils.Byte(useTime)
	//number, err := GetGenID()
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	addrSelf := Node.GetNetId()
	addrCoin := Node.Keystore.GetCoinAddrAll()
	orders := model.OrderForm{
		Number:         ulid.Make().Bytes(),      //订单编号
		UserAddr:       userAddr,                 //消费者地址
		ServerAddr:     *addrSelf,                //服务器地址
		ServerAddrCoin: addrCoin[0].Addr.Bytes(), //
		SpaceTotal:     uint64(spaceTotal),       //购买空间数量 单位：1G
		UseTime:        uint64(useTime),          //空间使用时间 单位：1天
		TotalPrice:     uint64(total),            //订单总金额
		//ChainTx    []byte               //区块链上的交易
		//TxHash     []byte               //已经上链的交易hash
		CreateTime: time.Now().Unix(), //订单创建时间
		//TimeOut    int64                //订单过期时间
		PayLockBlockHeight: this.syncCount.GetPullHeight() + chain_orders.Order_Overtime_Height,
	}
	if orders.TotalPrice > 0 {
		utils.Log.Error().Hex("创建订单", orders.Number).Send()
		ERR := db.ImProxyServer_SaveOrderFormNotPay(*Node.GetNetId(), &orders)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		this.SellingLock += orders.SpaceTotal
		this.UnpaidOrdersId[utils.Bytes2string(orders.Number)] = &orders
		this.UnpaidOrders[utils.Bytes2string(orders.UserAddr.GetAddr())] = &orders
		return &orders, utils.NewErrorSuccess()
	}
	//免支付订单
	this.Orders[utils.Bytes2string(orders.Number)] = &orders
	//过期时间是服务期，根据链端出块时间计算高度
	orders.PayLockBlockHeight = (uint64(useTime) * 24 * 60 * 60) / uint64(chainConfig.Mining_block_time/time.Second)
	//orders.TimeOut = time.Now().Unix() + (int64(useTime) * 24 * 60 * 60)
	ERR = db.ImProxyServer_SaveOrderFormInUse(*Node.GetNetId(), &orders)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	utils.Log.Info().Msgf("用户添加空间:%s %d", orders.UserAddr.B58String(), orders.SpaceTotal*uint64(utils.GB))
	this.AuthManager.AddPurchaseSpace(orders.UserAddr, utils.Byte(orders.SpaceTotal)*utils.GB)
	InitUserTopDirIndex(orders.UserAddr)
	return &orders, utils.NewErrorSuccess()
}

/*
创建续费订单
@preNumber     []byte    续费的订单ID
@spaceTotal    uint64    要续费的空间数量，只能缩减，不能增加
@useTime       []byte    续费的时间
*/
func (this *ProxyServerManager) CreateRenewalOrders(preNumber []byte, spaceTotal, useTime uint64) (*model.OrderForm, utils.ERROR) {
	//要等拉取到最高区块，才能接收订单
	if !this.syncCount.GetChainPullFinish() {
		return nil, utils.NewErrorBus(config.ERROR_CODE_order_chain_not_finish, "")
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	oldOrder, ok := this.Orders[utils.Bytes2string(preNumber)]
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_orders_not_exist, "")
	}
	//查询一个用户的未支付订单
	unpaidOrders, ok := this.UnpaidOrders[utils.Bytes2string(oldOrder.UserAddr.GetAddr())]
	if ok {
		return unpaidOrders, utils.NewErrorBus(config.ERROR_CODE_order_not_pay, "")
	}
	unpaidOrders, ok = this.RenewalUnpaidOrder[utils.Bytes2string(oldOrder.UserAddr.GetAddr())]
	if ok {
		return unpaidOrders, utils.NewErrorBus(config.ERROR_CODE_order_not_pay, "")
	}
	storageServerInfo, ERR := db.ImProxyServer_GetProxyInfoSelf(*Node.GetNetId(), true)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("广播云存储提供者节点时查询本节点信息错误:%s", ERR.String())
		return nil, ERR
	}
	//服务到期关闭
	if storageServerInfo.RenewalTime == 0 {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_Service_expiration_and_closure, "")
	}
	//判断订单到期
	pullHeight := this.syncCount.GetPullHeight()
	if oldOrder.PayLockBlockHeight <= pullHeight {
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_orders_overtime, "")
	}
	//未到续费时间
	if oldOrder.PayLockBlockHeight > pullHeight && (oldOrder.PayLockBlockHeight-pullHeight) >
		uint64(storageServerInfo.RenewalTime*24*60*60/uint64(chainConfig.Mining_block_time/time.Second)) {
		//未到续费时间
		return nil, utils.NewErrorBus(config.ERROR_CODE_storage_orders_not_overtime, "") // config.ERROR_storage_orders_not_overtime
	}
	//不能超过最大时间
	if useTime > storageServerInfo.UseTimeMax {
		useTime = storageServerInfo.UseTimeMax
	}
	//续费空间只能减少，不能增加
	if spaceTotal > oldOrder.SpaceTotal {
		//return nil, config.ERROR_storage_user_renewal_space_total_max
		spaceTotal = oldOrder.SpaceTotal
	}

	total := storageServerInfo.PriceUnit * spaceTotal * useTime
	//newNumber, err := GetGenID()
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	addrSelf := Node.GetNetId()
	orders := model.OrderForm{
		Number:     ulid.Make().Bytes(), //订单编号，全局自增长ID
		PreNumber:  preNumber,           //
		UserAddr:   oldOrder.UserAddr,   //消费者地址
		ServerAddr: *addrSelf,           //服务器地址
		SpaceTotal: spaceTotal,          //购买空间数量 单位：1G
		UseTime:    useTime,             //空间使用时间 单位：1天
		TotalPrice: total,               //订单总金额
		//ChainTx    []byte               //区块链上的交易
		//TxHash     []byte               //已经上链的交易hash
		CreateTime: time.Now().Unix(), //订单创建时间
		//TimeOut    int64                //订单过期时间
		PayLockBlockHeight: this.syncCount.GetPullHeight() + chain_orders.Order_Overtime_Height,
	}
	if orders.TotalPrice > 0 {
		ERR := db.ImProxyServer_SaveOrderFormNotPay(*Node.GetNetId(), &orders)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		this.UnpaidOrdersId[utils.Bytes2string(orders.Number)] = &orders
		this.RenewalUnpaidOrder[utils.Bytes2string(orders.UserAddr.GetAddr())] = &orders
		return &orders, ERR
	}
	//免支付订单
	this.Orders[utils.Bytes2string(orders.Number)] = &orders
	//过期时间是服务期
	orders.PayLockBlockHeight = (uint64(useTime) * 24 * 60 * 60) / uint64(chainConfig.Mining_block_time/time.Second)
	//orders.TimeOut = time.Now().Unix() + (int64(useTime) * 3 * 60)
	ERR = db.ImProxyServer_SaveOrderFormInUse(*Node.GetNetId(), &orders)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	this.AuthManager.AddPurchaseSpace(orders.UserAddr, utils.Byte(orders.SpaceTotal)*utils.GB)
	ERR = InitUserTopDirIndex(orders.UserAddr)
	return &orders, ERR
}

/*
清理用户上传的文件，以达到空间要求
按照上传时间顺序，清理最新上传的文件
*/
func (this *ProxyServerManager) cleanUserFiles() utils.ERROR {
	//utils.Log.Info().Msgf("开始清理使用空间不够的用户文件")
	//查询用户购买了多少空间
	userList := this.AuthManager.QueryUserList()
	for _, one := range userList {
		spacesTotal, spacesUse, _ := this.AuthManager.QueryUserSpaces(one)
		//utils.Log.Info().Msgf("检查这一用户空间:%s %d", one.B58String(), spacesTotal)
		if spacesTotal == 0 {
			utils.Log.Info().Msgf("清除所有数据:%s", one.B58String())
			//删除全部文件
			//_, delFileIndexs, ERR := db.ImProxyServer_DelUserDirAndFileAll(one)
			//if !ERR.CheckSuccess() {
			//	utils.Log.Error().Msgf("删除用户全部文件错误:%s", ERR.String())
			//	continue
			//}
			//删除数据
			//for _, fileOne := range delFileIndexs {
			//	for _, one := range fileOne.Chunks {
			//		utils.Log.Info().Msgf("删除块数据:%s", hex.EncodeToString(one))
			//		this.DBC.DelChunk(one)
			//	}
			//}
			//删除用户
			this.AuthManager.DelUser(one)
		} else if spacesTotal < spacesUse {
			utils.Log.Info().Msgf("清除部分文件:%s", one.B58String())
			//清理部分文件
			//dir, ERR := db.ImProxyServer_GetUserTopDirIndex(one)
			//if !ERR.CheckSuccess() {
			//	utils.Log.Error().Msgf("查询用户的顶层目录错误:%s", ERR.String())
			//	continue
			//}
			//迭代查询文件夹中的文件
			//_, files, _, _, ERR := db.ImProxyServer_GetDirIndexRecursion([][]byte{dir.ID})
			//if !ERR.CheckSuccess() {
			//	utils.Log.Error().Msgf("查询用户的顶层目录错误:%s", ERR.String())
			//	continue
			//}
			//按照上传时间排序，优先删除最新上传的文件
			//newFiles := make([]*model.FileIndex, 0, len(files))
			//for _, fileOne := range files {
			//	fileOne.FilterUser(one)
			//	newFiles = append(newFiles, fileOne)
			//}
			//slices.SortFunc(newFiles, func(a, b *model.FileIndex) int {
			//	return cmp.Compare(a.Time[0], b.Time[0])
			//})
			//计算释放的空间
			//sizeTotal := uint64(0)
			//要删除的文件
			//delFileIDs := make([][]byte, 0)
			//for _, fileOne := range newFiles {
			//	//utils.Log.Info().Msgf("")
			//	sizeTotal += fileOne.FileSize
			//	delFileIDs = append(delFileIDs, fileOne.ID)
			//	if spacesTotal >= spacesUse-sizeTotal {
			//		break
			//	}
			//}
			//删除用户文件
			//_, delFileIndexs, ERR := db.ImProxyServer_DelDirAndFileIndex(one, nil, delFileIDs)
			//if !ERR.CheckSuccess() {
			//	utils.Log.Error().Msgf("删除用户全部文件错误:%s", ERR.String())
			//	continue
			//}
			//删除数据
			//for _, fileOne := range delFileIndexs {
			//	for _, one := range fileOne.Chunks {
			//		this.DBC.DelChunk(one)
			//	}
			//}
			//this.AuthManager.SubUseSpace(one, sizeTotal)
		}
	}
	return utils.NewErrorSuccess()
}

/*
加载数据库订单并且初始化对象
*/
func (this *ProxyServerManager) LoadOrdersAndInit() utils.ERROR {
	//加载已经支付的订单
	orders, ERR := db.ImProxyServer_GetInUseOrdersIDs(*Node.GetNetId())
	if ERR.CheckFail() {
		return ERR
	}
	//加载未支付订单
	orders, ERR = db.ImProxyServer_GetNotPayOrders(*Node.GetNetId())
	if ERR.CheckFail() {
		return ERR
	}
	//先放入已支付订单集合，首次订单和续费订单放一起
	this.lock.Lock()
	defer this.lock.Unlock()
	for _, one := range orders {
		this.Orders[utils.Bytes2string(one.Number)] = one
	}
	//根据订单给用户分配空间
	for _, one := range this.Orders {
		//避免因为购买空间和续费空间订单同时存在，导致重复分配空间
		if len(one.PreNumber) > 0 {
			//续费订单，检查前置订单是否还在
			if _, ok := this.Orders[utils.Bytes2string(one.PreNumber)]; ok {
				//还在
				continue
			}
		}
		this.AuthManager.AddPurchaseSpace(one.UserAddr, utils.Byte(one.SpaceTotal)*utils.GB)
	}
	for _, one := range orders {
		//this.AuthManager.AddPurchaseSpace(one.UserAddr, utils.Byte(one.SpaceTotal)*utils.GB)
		if len(one.PreNumber) > 0 {
			//续费订单
			this.RenewalUnpaidOrderId[utils.Bytes2string(one.Number)] = one
			this.RenewalUnpaidOrder[utils.Bytes2string(one.UserAddr.GetAddr())] = one
		} else {
			//全新订单
			this.UnpaidOrdersId[utils.Bytes2string(one.Number)] = one
			this.UnpaidOrders[utils.Bytes2string(one.UserAddr.GetAddr())] = one
		}
	}
	return utils.NewErrorSuccess()
}

/*
加载数据库中所有用户已经使用的存储空间
*/
func (this *ProxyServerManager) LoadUserSpacesUse() utils.ERROR {
	m, ERR := db.ImProxyServer_GetAllUserSpaceUseSize(*Node.GetNetId())
	if !ERR.CheckSuccess() {
		return ERR
	}
	for k, v := range m {
		addr := nodeStore.NewAddressNet([]byte(k))
		this.AuthManager.AddUseSpace(*addr, utils.Byte(v))
	}
	return utils.NewErrorSuccess()
}

/*
初始化用户顶层目录
*/
func InitUserTopDirIndex(userAddr nodeStore.AddressNet) utils.ERROR {
	//dirIndex, ERR := db.ImProxyServer_GetUserTopDirIndex(userAddr)
	//if !ERR.CheckSuccess() {
	//	return ERR
	//}
	//if dirIndex != nil {
	//	return utils.NewErrorSuccess()
	//}
	////id, err := GetGenID()
	////if err != nil {
	////	return err
	////}
	//dirIndex = &model.DirectoryIndex{
	//	ID: ulid.Make().Bytes(), //文件夹唯一ID采用全局自增长ID
	//	//Name    string           //文件夹名称
	//	//Dirs    []DirectoryIndex //文件夹中包含文件夹
	//	//Files   []FileIndex      //文件列表
	//	//DirsID  [][]byte         //包含文件夹的ID
	//	//FilesID [][]byte         //包含文件的ID
	//	UAddr: userAddr, //
	//}
	//return db.ImProxyServer_SaveUserTopDirIndex(userAddr, *dirIndex)
	return utils.NewErrorSuccess()
}

/*
查询用户是否在本节点代理服务列表中，并返回用户信息
*/
func (this *ProxyServerManager) QueryUserSpaces(userAddr nodeStore.AddressNet) (utils.Byte, utils.Byte, utils.Byte) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	utils.Log.Info().Msgf("查询用户空间:%s", userAddr.B58String())
	//查询用户剩余空间
	spacesTotal, spacesUse, remain := this.AuthManager.QueryUserSpaces(userAddr)
	return spacesTotal, spacesUse, remain
	//if remain <= 0 {
	//	utils.Log.Info().Msgf("用户剩余空间为0:%s", userAddr.B58String())
	//	return 0, utils.NewErrorSuccess()
	//}
	//剩余空间大于0，则代码是自己的代理节点
	//_, ok := this.UserMap[utils.Bytes2string(userAddr)]
	//if !ok {
	//	return nil, utils.NewErrorSuccess()
	//}
	//userinfo, ERR := db.ImProxyClient_FindUserinfo(userAddr)
	//if !ERR.CheckSuccess() {
	//	return nil, ERR
	//}
	//return userinfo, utils.NewErrorSuccess()
}

/*
检查并保存客户端的离线消息
*/
func (this *ProxyServerManager) CheckSaveDataChainNolink(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	//是代理消息，先检查用户是否在代理列表中
	_, _, remain := this.QueryUserSpaces(proxyItr.GetAddrTo())
	if remain == 0 {
		//不在代理用户列表中
		//ReplyError(Node, config.NET_protocol_version_v1, message, replyMsgID, config.ERROR_CODE_IM_not_proxy, "")
		return utils.NewErrorBus(config.ERROR_CODE_IM_not_proxy, "")
	}
	//保存代理消息
	return this.syncServer.checkSaveDataChainNolink(proxyItr)
}

/*
移动订单到已支付列表
*/
func (this *ProxyServerManager) MoveOrderToInUse(order *model.OrderForm) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.UnpaidOrdersId[utils.Bytes2string(order.Number)]
	if !ok {
		return utils.NewErrorSuccess()
	}
	ERR := db.ImProxyServer_MoveOrderToInUse(*Node.GetNetId(), order)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Str("错误", ERR.String()).Send()
		return ERR
	}
	delete(this.UnpaidOrdersId, utils.Bytes2string(order.Number))
	delete(this.UnpaidOrders, utils.Bytes2string(order.UserAddr.GetAddr()))
	this.Orders[utils.Bytes2string(order.Number)] = order
	return utils.NewErrorSuccess()
}

/*
移动续费订单到已支付列表
*/
func (this *ProxyServerManager) MoveRenewalOrderToInUse(order *model.OrderForm) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.RenewalUnpaidOrderId[utils.Bytes2string(order.Number)]
	if !ok {
		return utils.NewErrorSuccess()
	}
	ERR := db.ImProxyServer_MoveOrderToInUse(*Node.GetNetId(), order)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Str("错误", ERR.String()).Send()
		return ERR
	}
	delete(this.RenewalUnpaidOrderId, utils.Bytes2string(order.Number))
	delete(this.RenewalUnpaidOrder, utils.Bytes2string(order.UserAddr.GetAddr()))
	this.Orders[utils.Bytes2string(order.Number)] = order
	return utils.NewErrorSuccess()
}

/*
已上链区块统计
*/
func (this *ProxyServerManager) countBlockOne(bhvo *mining.BlockHeadVO) {
	for _, txOne := range bhvo.Txs {
		orderId, txPay := chain_orders.ParseTxOrderId(txOne)
		if orderId == nil {
			continue
		}
		//统计第一次购买订单
		ERR := this.CountOrderPay(*orderId, txPay)
		if ERR.CheckFail() {
			continue
		}
		//统计续费订单
		ERR = this.CountRenewalOrderPay(*orderId, txPay)
		if ERR.CheckFail() {
			continue
		}

		//给前端推送已支付订单
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_chain_payOrder_server,
			Content: hex.EncodeToString(*orderId)}
		subscription.AddSubscriptionMsg(&msgInfo)
	}
	//清理过期未支付订单
	this.CleanOvertimeNotPayOrder(bhvo.BH.Height)
	//清理已经支付，过期订单
	this.CleanOvertimeOrder(bhvo.BH.Height)
	//清理用户上传的文件，以达到空间要求
	this.cleanUserFiles()
}

/*
统计订单支付
*/
func (this *ProxyServerManager) CountOrderPay(orderId []byte, txPay *mining.Tx_Pay) utils.ERROR {
	//查找订单
	this.lock.Lock()
	orderFrom, ok := this.UnpaidOrdersId[utils.Bytes2string(orderId)]
	this.lock.Unlock()
	if !ok {
		return utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
	}
	//判断地址和费用是否给够
	haveAddr := false
	for _, one := range txPay.Vout {
		if bytes.Equal(one.Address, orderFrom.ServerAddrCoin) && one.Value >= orderFrom.TotalPrice {
			haveAddr = true
			break
		}
	}
	if !haveAddr {
		return utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
	}
	txBs, err := txPay.Proto()
	if err != nil {
		utils.Log.Error().Str("错误", err.Error()).Send()
		return utils.NewErrorSysSelf(err)
	}
	orderFrom.TxHash = *txPay.GetHash()
	orderFrom.ChainTx = *txBs
	//新订单，增加空间
	this.AuthManager.AddPurchaseSpace(orderFrom.UserAddr, utils.Byte(orderFrom.SpaceTotal)*utils.GB)
	ERR := InitUserTopDirIndex(orderFrom.UserAddr)
	if ERR.CheckFail() {
		utils.Log.Error().Str("错误", ERR.String()).Send()
		return ERR
	}
	//订单保存到已支付列表
	//删除未支付订单，更新库存
	ERR = this.MoveOrderToInUse(orderFrom)
	if ERR.CheckFail() {
		return ERR
	}
	utils.Log.Info().Hex("已支付订单", orderFrom.Number).Send()
	return ERR
}

/*
统计续费订单支付
*/
func (this *ProxyServerManager) CountRenewalOrderPay(orderId []byte, txPay *mining.Tx_Pay) utils.ERROR {
	//查找订单
	this.lock.Lock()
	orderFrom, ok := this.RenewalUnpaidOrderId[utils.Bytes2string(orderId)]
	this.lock.Unlock()
	if !ok {
		return utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
	}
	//判断地址和费用是否给够
	haveAddr := false
	for _, one := range txPay.Vout {
		if bytes.Equal(one.Address, orderFrom.ServerAddrCoin) && one.Value >= orderFrom.TotalPrice {
			haveAddr = true
			break
		}
	}
	if !haveAddr {
		return utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
	}
	txBs, err := txPay.Proto()
	if err != nil {
		utils.Log.Error().Str("错误", err.Error()).Send()
		return utils.NewErrorSysSelf(err)
	}
	orderFrom.TxHash = *txPay.GetHash()
	orderFrom.ChainTx = *txBs
	//订单保存到已支付列表
	//删除未支付订单，更新库存
	ERR := this.MoveRenewalOrderToInUse(orderFrom)
	if ERR.CheckFail() {
		return ERR
	}
	utils.Log.Info().Hex("已支付续费订单", orderFrom.Number).Send()
	return ERR
}

/*
清理过期未支付订单
*/
func (this *ProxyServerManager) CleanOvertimeNotPayOrder(pullHeight uint64) {
	//定期检查已支付订单服务期是否超时
	this.lock.Lock()
	defer this.lock.Unlock()
	for orderKey, one := range this.UnpaidOrders {
		//未超时
		if one.PayLockBlockHeight > pullHeight {
			continue
		}
		//utils.Log.Info().Msgf("删除服务器过期订单:%+v", one)
		ERR := db.ImProxyServer_MoveOrderToNotPayTimeout(*Node.GetNetId(), one)
		if !ERR.CheckSuccess() {
			//utils.Log.Error().Msgf("删除超期订单失败:%s", ERR.String())
			continue
		}
		//删除过期订单
		delete(this.UnpaidOrders, orderKey)
		delete(this.UnpaidOrdersId, utils.Bytes2string(one.Number))
		utils.Log.Info().Hex("超时未支付订单", one.Number).Send()
	}
	//续费未支付订单
	for orderKey, one := range this.RenewalUnpaidOrder {
		//未超时
		if one.PayLockBlockHeight > pullHeight {
			continue
		}
		//utils.Log.Info().Msgf("删除服务器过期订单:%+v", one)
		ERR := db.ImProxyServer_MoveOrderToNotPayTimeout(*Node.GetNetId(), one)
		if !ERR.CheckSuccess() {
			//utils.Log.Error().Msgf("删除超期订单失败:%s", ERR.String())
			continue
		}
		//删除过期订单
		delete(this.RenewalUnpaidOrder, orderKey)
		delete(this.RenewalUnpaidOrderId, utils.Bytes2string(one.Number))
		utils.Log.Info().Hex("续费超时未支付订单", one.Number).Send()
	}
}

/*
定时清理过期的订单
*/
func (this *ProxyServerManager) CleanOvertimeOrder(pullHeight uint64) {
	//定期检查已支付订单服务期是否超时
	this.lock.Lock()
	defer this.lock.Unlock()
	for orderKey, one := range this.Orders {
		//未超时
		if one.PayLockBlockHeight > pullHeight {
			continue
		}
		//超过服务时间，查看是否有未支付订单，有则等一等
		have := false
		for _, uOone := range this.UnpaidOrders {
			if bytes.Equal(uOone.PreNumber, one.Number) {
				have = true
				break
			}
		}
		//有则等一等
		if have {
			continue
		}

		var renewalOrder *model.OrderForm
		//检查是否有续费订单
		for _, orderOne := range this.RenewalUnpaidOrder {
			if bytes.Equal(orderOne.PreNumber, one.Number) {
				renewalOrder = orderOne
				break
			}
		}
		//有续费订单，等一等
		if renewalOrder != nil {
			continue
		}
		//没有续费，直接删除
		//utils.Log.Info().Msgf("删除服务器过期订单:%+v", one)
		ERR := db.ImProxyServer_MoveOrderToInUseTimeout(*Node.GetNetId(), one, renewalOrder)
		if !ERR.CheckSuccess() {
			//utils.Log.Error().Msgf("删除超期订单失败:%s", ERR.String())
			continue
		}
		//删除过期订单
		delete(this.Orders, orderKey)
		//utils.Log.Info().Msgf("订单过期，空间减少:%s %d", one.UserAddr, utils.Byte(one.SpaceTotal)*utils.GB)
		this.AuthManager.SubPurchaseSpace(one.UserAddr, utils.Byte(one.SpaceTotal)*utils.GB)
		utils.Log.Info().Hex("服务到期订单", one.Number).Send()
	}
}

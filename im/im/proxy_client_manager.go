package im

import (
	"bytes"
	"encoding/hex"
	"slices"
	"sort"
	"sync"
	"time"
	"web3_gui/chain/mining"
	"web3_gui/chain_boot/chain_plus"
	"web3_gui/chain_orders"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/im/subscription"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

var StaticProxyClientManager *ProxyClientManager

type ProxyClientManager struct {
	lock *sync.RWMutex //
	//AuthManager  map[string]*storage.AuthManager //权限管理。key:string=服务器地址;value:*AuthManager=权限管理;
	UnpaidOrders     map[string]*model.OrderForm //未支付订单.key:string=订单编号;value:*model.OrderForm=订单;
	OrdersNotOnChain map[string]*model.OrderForm //已支付但未上链订单.key:string=订单编号;value:*model.OrderForm=订单;
	Orders           map[string]*model.OrderForm //已支付订单.key:string=订单编号;value:*model.OrderForm=订单;
	//RenewalOrder map[string]*model.OrderForm     //续费订单

	ProxyMap           map[string]*SyncClient               //代理服务器列表。value:*SyncClient=同步客户端;
	ParserClient       *ClientParser                        //客户端数据链解析器
	uploadChan         chan imdatachain.DataChainProxyItr   //有新消息上传
	downloadChan       chan []imdatachain.DataChainProxyItr //有新消息下载来了
	downloadNoLinkChan chan []imdatachain.DataChainProxyItr //有新消息下载来了

	RetryUploadOnce        *utils.Once //同一时间只能有一个程序在执行上传消息
	RetrySendDataChainOnce *sync.Map   //一个用户同一时间只能有一个程序在执行发送消息。key:string=好友地址;value:*utils.Once=正在执行的发送程序;

	groupKnitManager   *GroupKnitManager                  //客户端没有代理的时候，只有自己作为群服务器
	groupParserManager *GroupParserManager                //自己加入的所有群聊天解析器
	groupOperationChan chan imdatachain.DataChainProxyItr //群操作管道
	syncCount          *chain_plus.ChainSyncCount         //区块链同步统计程序
}

func CreateProxyClientManager(db *utilsleveldb.LevelDB) (*ProxyClientManager, utils.ERROR) {
	groupOperationChan := make(chan imdatachain.DataChainProxyItr, 100)
	gkm, ERR := NewGroupKnitManager()
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("创建客户端群服务器 错误:%s", ERR.String())
		return nil, ERR
	}
	gpm, ERR := NewGroupParserManager(groupOperationChan)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("创建客户端群解析器 错误:%s", ERR.String())
		return nil, ERR
	}
	ss := ProxyClientManager{
		lock: new(sync.RWMutex), //
		//AuthManager:  make(map[string]*storage.AuthManager), //
		UnpaidOrders:     make(map[string]*model.OrderForm), //未支付订单
		OrdersNotOnChain: make(map[string]*model.OrderForm), //已支付但未上链订单
		Orders:           make(map[string]*model.OrderForm), //已支付订单
		//RenewalOrder: make(map[string]*model.OrderForm),     //续费订单

		ProxyMap:               make(map[string]*SyncClient),                     //
		uploadChan:             make(chan imdatachain.DataChainProxyItr, 100),    //有新消息上传
		downloadChan:           make(chan []imdatachain.DataChainProxyItr, 1000), //有新消息下载来了
		downloadNoLinkChan:     make(chan []imdatachain.DataChainProxyItr, 1000), //有新消息下载来了
		RetryUploadOnce:        utils.NewOnce(),                                  //
		RetrySendDataChainOnce: new(sync.Map),                                    //
		groupKnitManager:       gkm,                                              //客户端没有代理的时候，只有自己作为群服务器
		groupParserManager:     gpm,                                              //自己加入的所有群聊天解析器
		groupOperationChan:     groupOperationChan,                               //群操作管道
	}
	cp, ERR := NewClientParser(*Node.GetNetId(), ss.uploadChan, ss.downloadChan, ss.downloadNoLinkChan, ss.groupOperationChan)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("创建数据链客户端解析器 错误:%s", ERR.String())
		return nil, ERR
	}
	ss.ParserClient = cp

	//加载客户端订单
	ERR = ss.LoadStorageClientOriders()
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("加载客户端订单 错误:%s", ERR.String())
		return nil, ERR
	}

	//
	go ss.LoopUploadAndSendImDataChain()
	go ss.LoopListenGroupOperation()
	go ss.LooploadDataChainSendFail()
	ss.syncCount = chain_plus.NewChainSyncCount(db, config.DBKEY_improxy_server_chain_count_height, ss.countBlockOne)
	return &ss, utils.NewErrorSuccess()
}

/*
获取订单中的云存储服务器列表
*/
func (this *ProxyClientManager) GetOrderServerList() ([]model.StorageServerInfo, utils.ERROR) {
	addrMaps := make(map[string]nodeStore.AddressNet)
	this.lock.RLock()
	for _, one := range this.Orders {
		addrMaps[utils.Bytes2string(one.ServerAddr.GetAddr())] = one.ServerAddr
	}
	this.lock.RUnlock()
	addrs := make([]nodeStore.AddressNet, 0, len(addrMaps))
	for _, one := range addrMaps {
		addrs = append(addrs, one)
	}
	list, ERR := db.ImProxyClient_GetProxyListByAddrs(*Node.GetNetId(), addrs...)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return list, utils.NewErrorSuccess()
}

/*
本地查询用户在一个提供商的容量
@return    utils.Byte    购买的总容量
@return    utils.Byte    已使用容量
@return    utils.Byte    剩余容量
*/
func (this *ProxyClientManager) QueryUserSpacesLocal(addr nodeStore.AddressNet) (utils.Byte, utils.Byte, utils.Byte) {
	total := uint64(0)
	this.lock.RLock()
	for _, one := range this.UnpaidOrders {
		if !bytes.Equal(addr.GetAddr(), one.ServerAddr.GetAddr()) {
			continue
		}
		//判断续费订单，续费订单的前置订单未超时的情况，避免重复统计容量
		if one.PreNumber != nil && len(one.PreNumber) > 0 {
			_, ok := this.UnpaidOrders[utils.Bytes2string(one.PreNumber)]
			if ok {
				continue
			}
		}
		total = total + (uint64(utils.GB) * one.SpaceTotal)
	}
	this.lock.RUnlock()
	return utils.Byte(total), utils.Byte(0), utils.Byte(0)
}

/*
获取我的代理节点列表
*/
func (this *ProxyClientManager) GetProxyList() []nodeStore.AddressNet {
	addrs := make([]nodeStore.AddressNet, 0)
	this.lock.RLock()
	for _, v := range this.ProxyMap {
		addrs = append(addrs, v.ServerAddr)
	}
	this.lock.RUnlock()
	return addrs
}

/*
获取我的代理节点详细信息列表
*/
//func (this *ProxyClientManager) GetProxyInfoList() ([]model.ImProxy, error) {
//	addrs := this.GetProxyList()
//	list, err := db.ImProxyClient_GetProxyListByAddrs(addrs...)
//	if err != nil {
//		return nil, err
//	}
//	return list, nil
//}

/*
获取活跃的订单列表
*/
func (this *ProxyClientManager) GetOrdersListRange(startNumber []byte, limit uint64) []*model.OrderFormVO {
	//pullHeight := this.syncCount.GetPullHeight()
	this.lock.RLock()
	defer this.lock.RUnlock()
	orders := make([]*model.OrderFormVO, 0, len(this.Orders))
	//未支付订单
	for _, one := range this.UnpaidOrders {
		orderVO := one.ConverVO()
		orderVO.Status = config.ORDER_status_not_pay
		orders = append(orders, orderVO)
	}
	//已支付，但未上链订单
	for _, one := range this.OrdersNotOnChain {
		orderVO := one.ConverVO()
		orderVO.Status = config.ORDER_status_pay_not_onchain
		orders = append(orders, orderVO)
	}
	//已支付订单
	for _, one := range this.Orders {
		orderVO := one.ConverVO()
		orderVO.Status = config.ORDER_status_pay_onchain
		orders = append(orders, orderVO)
	}
	//utils.Log.Info().Uint64("查询过期订单", limit).Hex("订单号", startNumber).Send()
	//过期订单
	ordersOvertime, ERR := db.ImProxyClient_LoadOrderTimeout(*Node.GetNetId(), startNumber, limit)
	if ERR.CheckFail() {
		utils.Log.Error().Msgf("查询过期订单失败:%s", ERR.String())
		return nil
	}
	//utils.Log.Info().Int("数量", len(ordersOvertime)).Send()
	for _, one := range ordersOvertime {
		orderVO := one.ConverVO()
		orderVO.Status = config.ORDER_status_not_pay_overtime
		orders = append(orders, orderVO)
	}
	return orders
}

/*
添加未支付订单
*/
//func (this *ProxyClientManager) AddOrdersUnpaid(order *model.OrderForm) utils.ERROR {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	this.UnpaidOrders[utils.Bytes2string(order.Number)] = order
//	//this.UnpaidOrders = append(this.UnpaidOrders, order)
//	return utils.NewErrorSuccess()
//}

/*
添加订单
当支付费用为0时，直接添加到已支付账单列表
*/
func (this *ProxyClientManager) AddOrders(order *model.OrderForm) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	if order.TotalPrice > 0 {
		//保存到未支付订单列表数据库
		ERR := db.ImProxyClient_SaveOrderFormNotPay(*Node.GetNetId(), order)
		if !ERR.CheckSuccess() {
			return ERR
		}
		this.UnpaidOrders[utils.Bytes2string(order.Number)] = order
		utils.Log.Info().Hex("创建订单", order.Number).Send()
		return utils.NewErrorSuccess()
	}
	//免支付的订单
	ERR := db.ImProxyClient_SaveOrderFormInUse(*Node.GetNetId(), order)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//添加到已经支付订单列表
	this.Orders[utils.Bytes2string(order.Number)] = order
	utils.Log.Info().Hex("创建订单", order.Number).Send()
	return utils.NewErrorSuccess()
}

/*
设置订单等待上链
*/
func (this *ProxyClientManager) SetOrderWaitOnChain(orderId []byte, lockHeight uint64) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	order, ok := this.UnpaidOrders[utils.Bytes2string(orderId)]
	if !ok {
		return utils.NewErrorSuccess()
	}
	order.LockHeightOnChain = lockHeight
	utils.Log.Info().Msgf("设置为已支付的订单 net:%s coin:%s", order.ServerAddr.B58String(), order.ServerAddrCoin.B58String())
	ERR := db.ImProxyClient_MoveOrderToNotOnChain(*Node.GetNetId(), order)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//删除未支付订单
	delete(this.UnpaidOrders, utils.Bytes2string(order.Number))
	//添加到已经支付但未上链订单列表
	this.OrdersNotOnChain[utils.Bytes2string(order.Number)] = order
	return utils.NewErrorSuccess()
}

/*
移动未支付订单到已支付订单中
*/
func (this *ProxyClientManager) MoveOrders(order *model.OrderForm) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.OrdersNotOnChain[utils.Bytes2string(order.Number)]
	if ok {
		ERR := db.ImProxyClient_MoveOrderNotOnChainToInUse(*Node.GetNetId(), order)
		if !ERR.CheckSuccess() {
			return ERR
		}
		//删除未支付订单
		delete(this.OrdersNotOnChain, utils.Bytes2string(order.Number))
		//添加到已经支付订单列表
		this.Orders[utils.Bytes2string(order.Number)] = order
		return utils.NewErrorSuccess()
	}
	_, ok = this.UnpaidOrders[utils.Bytes2string(order.Number)]
	if !ok {
		return utils.NewErrorSuccess()
	}
	ERR := db.ImProxyClient_MoveOrderToInUse(*Node.GetNetId(), order)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//删除未支付订单
	delete(this.UnpaidOrders, utils.Bytes2string(order.Number))
	//添加到已经支付订单列表
	this.Orders[utils.Bytes2string(order.Number)] = order
	return utils.NewErrorSuccess()
}

//
///*
//添加续费订单
//当支付费用为0时，直接添加到已支付账单列表
//*/
//func (this *ProxyClientManager) AddRenewalOrders(order *model.OrderForm) utils.ERROR {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	if order.TotalPrice > 0 {
//		ERR := db.ImProxyClient_SaveOrderFormInUse(*Node.GetNetId(), order)
//		if !ERR.CheckSuccess() {
//			return ERR
//		}
//		this.UnpaidOrders[utils.Bytes2string(order.Number)] = order
//		return utils.NewErrorSuccess()
//	}
//	//免支付的订单
//	ERR := db.ImProxyClient_SaveOrderFormInUse(*Node.GetNetId(), order)
//	if !ERR.CheckSuccess() {
//		return ERR
//	}
//	this.Orders[utils.Bytes2string(order.Number)] = order
//	//this.RenewalOrder = append(this.RenewalOrder, order)
//	//auth, ok := this.AuthManager[utils.Bytes2string(order.ServerAddr.GetAddr())]
//	//if !ok {
//	//	auth = storage.CreateAuthManager(utils.Bytes2string(order.ServerAddr.GetAddr()))
//	//	sc := NewSyncClient(order.ServerAddr, *Node.GetNetId(), this.downloadChan, this.downloadNoLinkChan, this.ParserClient)
//	//	this.ProxyMap[utils.Bytes2string(order.ServerAddr.GetAddr())] = sc
//	//}
//	////找到续费关联订单
//	//var oldOrder *model.OrderForm
//	//for _, one := range this.Orders {
//	//	if bytes.Equal(one.ServerAddr.GetAddr(), order.ServerAddr.GetAddr()) && bytes.Equal(one.Number, order.PreNumber) {
//	//		oldOrder = one
//	//		break
//	//	}
//	//}
//	//if oldOrder != nil {
//	//	//添加到已经支付的续费订单列表
//	//	this.RenewalOrder[utils.Bytes2string(order.Number)] = order
//	//	//this.RenewalOrder = append(this.RenewalOrder, order)
//	//} else {
//	//	//添加到已经支付的订单列表
//	//	this.Orders[utils.Bytes2string(order.Number)] = order
//	//	//this.Orders = append(this.Orders, order)
//	//	////判断空间是否减少
//	//	//count := uint64(0)
//	//	//if order.SpaceTotal < oldOrder.SpaceTotal {
//	//	//	count = oldOrder.SpaceTotal - order.SpaceTotal
//	//	//}
//	//	auth.AddPurchaseSpace(order.ServerAddr, utils.Byte(order.SpaceTotal)*utils.GB)
//	//}
//	return utils.NewErrorSuccess()
//}

//
///*
//添加已支付续费订单
//当支付费用为0时，直接添加到已支付账单列表
//*/
//func (this *ProxyClientManager) AddRenewalOrders(order *model.OrderForm) utils.ERROR {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	ERR := db.ImProxyClient_SaveOrderFormInUse(order)
//	if !ERR.CheckSuccess() {
//		return ERR
//	}
//	this.RenewalOrder[utils.Bytes2string(order.Number)] = order
//	//this.RenewalOrder = append(this.RenewalOrder, order)
//	sc := NewSyncClient(order.ServerAddr, Node.GetNetId(), this.downloadChan)
//	this.ProxyMap[utils.Bytes2string(order.ServerAddr)] = sc
//	//auth, ok := this.AuthManager[utils.Bytes2string(order.ServerAddr)]
//	//if !ok {
//	//	auth = CreateAuthManager(utils.Bytes2string(order.ServerAddr))
//	//}
//	//找到续费关联订单
//	var oldOrder *model.OrderForm
//	for _, one := range this.Orders {
//		if bytes.Equal(one.ServerAddr, order.ServerAddr) && bytes.Equal(one.Number, order.PreNumber) {
//			oldOrder = one
//			break
//		}
//	}
//	if oldOrder != nil {
//		//添加到已经支付的续费订单列表
//		this.RenewalOrder[utils.Bytes2string(order.Number)] = order
//		//this.RenewalOrder = append(this.RenewalOrder, order)
//	} else {
//		//添加到已经支付的订单列表
//		this.Orders[utils.Bytes2string(order.Number)] = order
//		//this.Orders = append(this.Orders, order)
//		////判断空间是否减少
//		//count := uint64(0)
//		//if order.SpaceTotal < oldOrder.SpaceTotal {
//		//	count = oldOrder.SpaceTotal - order.SpaceTotal
//		//}
//		//auth.AddPurchaseSpace(order.ServerAddr, order.SpaceTotal*config.GB)
//		sc := NewSyncClient(order.ServerAddr, Node.GetNetId(), this.downloadChan)
//		this.ProxyMap[utils.Bytes2string(order.ServerAddr)] = sc
//	}
//
//	//代理地址保存到用户信息中
//	userinfo, err := db.GetSelfInfo()
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	userinfo.Proxy.Store(utils.Bytes2string(order.ServerAddr), &order.ServerAddr)
//	ERR = db.UpdateSelfInfo(userinfo)
//	if !ERR.CheckSuccess() {
//		return ERR
//	}
//	return utils.NewErrorSuccess()
//}

/*
移动未支付的续费订单到已支付续费订单中
*/
//func (this *ProxyClientManager) MoveRenewalOrders(order *model.OrderForm) utils.ERROR {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//
//	ERR := db.ImProxyClient_SaveOrderFormInUse(order)
//	if !ERR.CheckSuccess() {
//		return ERR
//	}
//
//	//找到并删除未支付订单
//	delete(this.UnpaidOrders, utils.Bytes2string(order.Number))
//
//	//添加到已经支付订单列表
//	this.Orders[utils.Bytes2string(order.Number)] = order
//	//this.Orders = append(this.Orders, order)
//	sc := NewSyncClient(order.ServerAddr, Node.GetNetId(), this.downloadChan)
//	this.ProxyMap[utils.Bytes2string(order.ServerAddr)] = sc
//	//代理地址保存到用户信息中
//	userinfo, err := db.GetSelfInfo()
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	userinfo.Proxy.Store(utils.Bytes2string(order.ServerAddr), &order.ServerAddr)
//	ERR = db.UpdateSelfInfo(userinfo)
//	if !ERR.CheckSuccess() {
//		return ERR
//	}
//	return utils.NewErrorSuccess()
//}

/*
定时清理过期的订单
包括过期未支付的订单、存储时间过期未续费的订单
*/
//func (this *ProxyClientManager) LoopClean() {
//	for range time.NewTicker(config.Clean_Interval).C {
//		//utils.Log.Info().Msgf("定期检查客户端订单")
//		now := time.Now().Unix()
//		//定期检查未支付订单
//
//		//定期检查已支付订单服务期是否超时
//		this.lock.Lock()
//		for orderKey, one := range this.Orders {
//			//未超时
//			if one.TimeOut > now {
//				continue
//			}
//			//超过服务时间，查看是否有未支付订单，有则等一等
//			have := false
//			for _, uOone := range this.UnpaidOrders {
//				if bytes.Equal(uOone.PreNumber, one.Number) {
//					have = true
//					break
//				}
//			}
//			if have {
//				continue
//			}
//
//			var renewalOrder *model.OrderForm
//			//检查是否有续费订单
//			//for i, orderOne := range this.RenewalOrder {
//			//	if bytes.Equal(orderOne.ServerAddr, one.ServerAddr) && bytes.Equal(orderOne.PreNumber, one.Number) {
//			//		renewalOrder = orderOne
//			//		//删除续费订单
//			//		temp := this.RenewalOrder[:i]
//			//		temp = append(temp, this.RenewalOrder[i+1:]...)
//			//		this.RenewalOrder = temp
//			//		break
//			//	}
//			//}
//			//检查是否有续费订单
//			for key, orderOne := range this.RenewalOrder {
//				if bytes.Equal(orderOne.PreNumber, one.Number) {
//					renewalOrder = orderOne
//					//删除续费订单
//					delete(this.RenewalOrder, key)
//					break
//				}
//			}
//			//auth, ok := this.AuthManager[utils.Bytes2string(one.ServerAddr)]
//			//if !ok {
//			//	auth = CreateAuthManager(utils.Bytes2string(one.ServerAddr))
//			//}
//			utils.Log.Info().Msgf("删除客户端过期订单:%+v", one)
//			//删除过期订单
//			delete(this.Orders, orderKey)
//			//temp := this.Orders[:orderIndex]
//			//temp = append(temp, this.Orders[orderIndex+1:]...)
//			//this.Orders = temp
//
//			ERR := db.ImProxyClient_DelOrderFormInUse(one, renewalOrder)
//			if !ERR.CheckSuccess() {
//				utils.Log.Error().Msgf("删除过期订单错误:%s", ERR.String())
//				continue
//			}
//
//			//没有续费，直接删除
//			if renewalOrder == nil {
//				key := utils.Bytes2string(one.ServerAddr)
//				pclient, ok := this.ProxyMap[key]
//				if ok {
//					//存在则调用销毁方法
//					pclient.Destroy()
//				}
//				delete(this.ProxyMap, key)
//
//				//代理地址保存到用户信息中
//				userinfo, err := db.GetSelfInfo()
//				if err != nil {
//					continue
//				}
//				userinfo.Proxy.Delete(key)
//				ERR = db.UpdateSelfInfo(userinfo)
//				if !ERR.CheckSuccess() {
//					continue
//				}
//
//				continue
//			}
//			//有续费，判断空间是否减少
//		}
//		this.lock.Unlock()
//	}
//}

/*
加载客户端订单
*/
func (this *ProxyClientManager) LoadStorageClientOriders() utils.ERROR {
	ordersInUse, ERR := db.ImProxyClient_LoadOrderFormInUse(*Node.GetNetId())
	if ERR.CheckFail() {
		return ERR
	}
	utils.Log.Info().Msgf("加载已支付订单 数量:%d", len(ordersInUse))
	ordersNotPay, ERR := db.ImProxyClient_LoadOrderFormNotPay(*Node.GetNetId())
	if ERR.CheckFail() {
		return ERR
	}
	utils.Log.Info().Msgf("加载未支付订单 数量:%d", len(ordersNotPay))
	ordersNotOnChain, ERR := db.ImProxyClient_LoadOrderFormNotOnChain(*Node.GetNetId())
	if ERR.CheckFail() {
		return ERR
	}
	utils.Log.Info().Msgf("加载已支付等待上链订单 数量:%d", len(ordersNotOnChain))

	this.lock.Lock()
	defer this.lock.Unlock()
	for _, one := range ordersInUse {
		utils.Log.Info().Msgf("加载已支付订单:%+v", one)
		this.Orders[utils.Bytes2string(one.Number)] = &one
	}
	for _, one := range ordersNotPay {
		utils.Log.Info().Msgf("加载未支付订单:%+v", one)
		this.UnpaidOrders[utils.Bytes2string(one.Number)] = &one
	}
	for _, one := range ordersNotOnChain {
		utils.Log.Info().Msgf("加载已支付待上链订单:%+v", one)
		this.OrdersNotOnChain[utils.Bytes2string(one.Number)] = &one
	}
	return utils.NewErrorSuccess()
}

/*
协程循环加载好友未发送成功的消息，并重新发送
*/
func (this *ProxyClientManager) LooploadDataChainSendFail() {
	sort.Ints(config.IM_RetrySend_interval)
	interval := 0
	last := 0
	if len(config.IM_RetrySend_interval) == 0 {
		interval = 5
	} else if len(config.IM_RetrySend_interval) == 1 {
		interval = config.IM_RetrySend_interval[0]
	} else {
		last = config.IM_RetrySend_interval[len(config.IM_RetrySend_interval)-1]
		interval = last - config.IM_RetrySend_interval[len(config.IM_RetrySend_interval)-2]
	}

	userMap := make(map[string]int)
	ticker := time.NewTicker(time.Minute * 30)
	count := 0
	for range ticker.C {
		utils.Log.Info().Msgf("时间到，开始重发消息")
		count++
		//先传给自己的代理节点
		if count%interval == 0 {
			this.uploadProxy(nil)
		}

		//查询好友列表
		userinfos, ERR := db.FindUserListAll(*Node.GetNetId())
		if ERR.CheckFail() {
			utils.Log.Error().Msgf("查询好友列表错误")
			continue
		}
		for _, userinfo := range userinfos {
			if userinfo.IsGroup {
				continue
			}
			//查询是否有未发送成功的消息
			itrs, _, ERR := db.ImProxyClient_FindDataChainSendFailRange(*Node.GetNetId(), userinfo.Addr, 1)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("发送消息失败:%s", ERR.String())
				return
			}

			addrKeyStr := utils.Bytes2string(userinfo.Addr.GetAddr())
			if itrs == nil || len(itrs) == 0 {
				delete(userMap, addrKeyStr)
				continue
			}
			count := userMap[addrKeyStr]
			count++
			userMap[addrKeyStr] = count

			if len(config.IM_RetrySend_interval) < 2 {
				if count%interval != 0 {
					continue
				}
			} else {
				if !slices.Contains(config.IM_RetrySend_interval, count) {
					if (count-last)%interval != 0 {
						continue
					}
				}
			}
			//重发消息
			this.RetrySendDataChain(*Node.GetNetId(), userinfo.Addr)
		}
		ticker.Reset(time.Minute)
	}
}

/*
协程循环监听发送通道
上传或者直接发送数据链记录
*/
func (this *ProxyClientManager) LoopUploadAndSendImDataChain() {
	for one := range this.uploadChan {
		this.checkUploadProxyOrSendRemote(one)
	}
}

/*
判断上传给自己的代理节点，还是直接发送给对方
*/
func (this *ProxyClientManager) checkUploadProxyOrSendRemote(one imdatachain.DataChainProxyItr) utils.ERROR {
	//先传给自己的代理节点
	have := this.uploadProxy(one)
	//是否发送给代理节点
	if have {
		utils.Log.Info().Msgf("有代理节点，发送给自己的代理节点")
		return utils.NewErrorSuccess()
	}
	if one.GetCmd() != config.IMPROXY_Command_server_forward {
		return utils.NewErrorSuccess()
	}
	//自己没有代理节点，直接发送
	this.RetrySendDataChain(*Node.GetNetId(), one.GetAddrTo())
	return utils.NewErrorSuccess()
}

/*
上传给自己的代理节点
*/
func (this *ProxyClientManager) uploadProxy(one imdatachain.DataChainProxyItr) bool {
	have := false
	for _, scOne := range this.ProxyMap {
		have = true
		scOne.UploadDataChain(one)
	}
	//是否发送给代理节点
	return have
}

/*
发送一条数据链
根据自己有没有代理节点，决定发送给自己的代理节点，还是发送给对方
*/
func (this *ProxyClientManager) uploadAndSendDataChain(one imdatachain.DataChainProxyItr) utils.ERROR {
	index := one.GetIndex()
	utils.Log.Info().Msgf("有新消息需要发送:%s %d", index.String(), one.GetCmd())
	//修改状态并通知前端
	if one.GetCmd() == config.IMPROXY_Command_server_forward {
		clientOne := one.GetClientItr()
		if clientOne.GetClientCmd() == config.IMPROXY_Command_client_sendText {
			//修改状态为失败
			clientBase := clientOne.(*imdatachain.DataChainSendText)
			messageOne := &model.MessageContent{
				Type:       config.MSG_type_text,          //消息类型
				FromIsSelf: true,                          //是否自己发出的
				From:       clientBase.AddrFrom,           //发送者
				To:         clientBase.AddrTo,             //接收者
				Content:    clientBase.Content,            //消息内容
				Time:       time.Now().Unix(),             //时间
				SendID:     one.GetID(),                   //
				QuoteID:    clientBase.QuoteID,            //
				State:      config.MSG_GUI_state_not_send, //
				//PullAndPushID: pullAndPushID,                 //
			}
			//更新状态
			ERR := db.UpdateSendMessageStateV2(one.GetAddrFrom(), one.GetAddrTo(), one.GetID(), messageOne.State)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("修改消息状态 错误:%s", ERR.String())
				return ERR
			}
			//给前端一个通知
			msgVO := messageOne.ConverVO()
			msgVO.Subscription = config.SUBSCRIPTION_type_msg
			msgVO.State = messageOne.State
			subscription.AddSubscriptionMsg(msgVO)
		}
	}

	//先传给自己的代理节点
	have := false
	for _, scOne := range this.ProxyMap {
		have = true
		scOne.UploadDataChain(one)
	}
	//是否发送给代理节点
	if have {
		utils.Log.Info().Msgf("有代理节点，发送给自己的代理节点")
		return utils.NewErrorSuccess()
	}
	var ERR utils.ERROR
	//utils.Log.Info().Msgf("自己没有代理节点，则自己发送")
	//自己没有代理节点，则自己发送
	for range 2 {
		//循环发送是因为，此时对方添加了代理节点而拒收消息，所以需要刷新用户代理节点信息并重新发送
		ERR = this.SendDataChain(one)
		if ERR.CheckSuccess() {
			break
		}
		//对方因设置了代理节点而拒收消息
		//刷新对方代理节点信息
		_, ERR2 := SearchUserProxy(one.GetAddrTo())
		if ERR2.CheckFail() {
			break
		}
	}
	engine.ResponseItrKey(config.FLOOD_key_addFriend, one.GetID(), ERR)
	engine.ResponseItrKey(config.FLOOD_key_agreeFriend, one.GetID(), ERR)
	//flood.ResponseItrKey(config.FLOOD_key_addFriend, one.GetID(), ERR)
	//flood.ResponseItrKey(config.FLOOD_key_agreeFriend, one.GetID(), ERR)

	//修改状态并通知前端
	if one.GetCmd() == config.IMPROXY_Command_server_forward {
		clientOne := one.GetClientItr()
		if clientOne.GetClientCmd() == config.IMPROXY_Command_client_sendText {
			//修改状态为失败
			clientBase := clientOne.(*imdatachain.DataChainSendText)
			messageOne := &model.MessageContent{
				Type:       config.MSG_type_text,         //消息类型
				FromIsSelf: true,                         //是否自己发出的
				From:       clientBase.AddrFrom,          //发送者
				To:         clientBase.AddrTo,            //接收者
				Content:    clientBase.Content,           //消息内容
				Time:       time.Now().Unix(),            //时间
				SendID:     one.GetID(),                  //
				QuoteID:    clientBase.QuoteID,           //
				State:      config.MSG_GUI_state_success, //
				//PullAndPushID: pullAndPushID,                 //
			}
			if ERR.CheckFail() {
				messageOne.State = config.MSG_GUI_state_fail
			} else {
				//发送成功
				messageOne.State = config.MSG_GUI_state_success
			}
			//更新状态，这里的ERR不能影响外面的ERR，外面的ERR后面还需要继续使用
			ERR := db.UpdateSendMessageStateV2(one.GetAddrFrom(), one.GetAddrTo(), one.GetID(), messageOne.State)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("修改消息状态 错误:%s", ERR.String())
				return ERR
			}
			//给前端一个通知
			msgVO := messageOne.ConverVO()
			msgVO.Subscription = config.SUBSCRIPTION_type_msg
			msgVO.State = messageOne.State
			subscription.AddSubscriptionMsg(msgVO)
		}
		if clientOne.GetClientCmd() == config.IMPROXY_Command_client_file {
			//修改状态为失败
			dataChainFile := clientOne.(*imdatachain.DatachainFile)
			key := config.DBKEY_improxy_user_message_history_fileHash_send
			messageOne, ERR := db.FindMessageHistoryByFileHash(*key, *Node.GetNetId(), dataChainFile.AddrTo, dataChainFile.Hash, dataChainFile.SendTime)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
				return ERR
			}
			//保存传送进度到数据库
			messageOne.TransProgress = int(float64(dataChainFile.BlockIndex+1) / float64(dataChainFile.BlockTotal) * 100)
			//utils.Log.Info().Msgf("保存传送进度到数据库:%d", messageOne.TransProgress)
			ERR = db.UpdateMessageHistoryByFileHash(*Node.GetNetId(), messageOne, nil)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
				return ERR
			}
			//图片传完的，拼接起来
			ERR = JoinImageBase64(messageOne)
			if ERR.CheckFail() {
				return ERR
			}
			//给前端一个通知
			msgVO := messageOne.ConverVO()
			msgVO.Subscription = config.SUBSCRIPTION_type_msg
			//msgVO.State = messageOne.State
			subscription.AddSubscriptionMsg(msgVO)
		}
	}
	//发送后的处理
	if !ERR.CheckSuccess() {
		//发送失败
		utils.Log.Error().Msgf("发送消息 错误:%s", ERR.String())
		return ERR
	} else {
		//发送成功
		//标记数据链为已经发送成功状态
		index := one.GetIndex()
		ERR = db.ImProxyClient_UpdateDataChainStatusByID(*Node.GetNetId(), one.GetAddrFrom(), index.Bytes(),
			config.IMPROXY_datachain_status_sendSuccess, nil)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("发送消息失败:%s", ERR.String())
			return ERR
		}
		//utils.Log.Info().Msgf("发送成功，删除未发送消息:%+v", one.GetID())
		//删除发送失败列表中的记录
		ERR = db.ImProxyClient_DelDataChainSendFail(*Node.GetNetId(), [][]byte{one.GetID()})
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("修改消息状态 错误:%s", ERR.String())
			return ERR
		}
		return utils.NewErrorSuccess()
	}
}

/*
发送一条数据链消息给好友节点
*/
func (this *ProxyClientManager) sendDataChainOne(one imdatachain.DataChainProxyItr) utils.ERROR {
	index := one.GetIndex()
	utils.Log.Info().Msgf("有新消息需要发送:%s %d", index.String(), one.GetCmd())
	//修改状态并通知前端
	if one.GetCmd() == config.IMPROXY_Command_server_forward {
		clientOne := one.GetClientItr()
		if clientOne.GetClientCmd() == config.IMPROXY_Command_client_sendText {
			//修改状态为失败
			clientBase := clientOne.(*imdatachain.DataChainSendText)
			messageOne := &model.MessageContent{
				Type:       config.MSG_type_text,          //消息类型
				FromIsSelf: true,                          //是否自己发出的
				From:       clientBase.AddrFrom,           //发送者
				To:         clientBase.AddrTo,             //接收者
				Content:    clientBase.Content,            //消息内容
				Time:       time.Now().Unix(),             //时间
				SendID:     one.GetID(),                   //
				QuoteID:    clientBase.QuoteID,            //
				State:      config.MSG_GUI_state_not_send, //
				//PullAndPushID: pullAndPushID,                 //
			}
			//更新状态
			ERR := db.UpdateSendMessageStateV2(one.GetAddrFrom(), one.GetAddrTo(), one.GetID(), messageOne.State)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("修改消息状态 错误:%s", ERR.String())
				return ERR
			}
			//给前端一个通知
			msgVO := messageOne.ConverVO()
			msgVO.Subscription = config.SUBSCRIPTION_type_msg
			msgVO.State = messageOne.State
			subscription.AddSubscriptionMsg(msgVO)
		}
	}

	var ERR utils.ERROR
	//utils.Log.Info().Msgf("自己没有代理节点，则自己发送")
	//自己发送
	for range 2 {
		//循环发送是因为，此时对方添加了代理节点而拒收消息，所以需要刷新用户代理节点信息并重新发送
		ERR = this.SendDataChain(one)
		if ERR.CheckSuccess() {
			break
		}
		//对方因设置了代理节点而拒收消息
		//刷新对方代理节点信息
		_, ERR2 := SearchUserProxy(one.GetAddrTo())
		if ERR2.CheckFail() {
			break
		}
	}

	engine.ResponseItrKey(config.FLOOD_key_addFriend, one.GetID(), ERR)
	engine.ResponseItrKey(config.FLOOD_key_agreeFriend, one.GetID(), ERR)
	//flood.ResponseItrKey(config.FLOOD_key_addFriend, one.GetID(), ERR)
	//flood.ResponseItrKey(config.FLOOD_key_agreeFriend, one.GetID(), ERR)

	//修改状态并通知前端
	if one.GetCmd() == config.IMPROXY_Command_server_forward {
		clientOne := one.GetClientItr()
		if clientOne.GetClientCmd() == config.IMPROXY_Command_client_sendText {
			//修改状态为失败
			clientBase := clientOne.(*imdatachain.DataChainSendText)
			messageOne := &model.MessageContent{
				Type:       config.MSG_type_text,         //消息类型
				FromIsSelf: true,                         //是否自己发出的
				From:       clientBase.AddrFrom,          //发送者
				To:         clientBase.AddrTo,            //接收者
				Content:    clientBase.Content,           //消息内容
				Time:       time.Now().Unix(),            //时间
				SendID:     one.GetID(),                  //
				QuoteID:    clientBase.QuoteID,           //
				State:      config.MSG_GUI_state_success, //
				//PullAndPushID: pullAndPushID,                 //
			}
			if ERR.CheckFail() {
				messageOne.State = config.MSG_GUI_state_fail
			} else {
				//发送成功
				messageOne.State = config.MSG_GUI_state_success
			}
			//更新状态，这里的ERR不能影响外面的ERR，外面的ERR后面还需要继续使用
			ERR := db.UpdateSendMessageStateV2(one.GetAddrFrom(), one.GetAddrTo(), one.GetID(), messageOne.State)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("修改消息状态 错误:%s", ERR.String())
				return ERR
			}
			//给前端一个通知
			msgVO := messageOne.ConverVO()
			msgVO.Subscription = config.SUBSCRIPTION_type_msg
			msgVO.State = messageOne.State
			subscription.AddSubscriptionMsg(msgVO)
		}
		if clientOne.GetClientCmd() == config.IMPROXY_Command_client_file {
			//修改状态为失败
			dataChainFile := clientOne.(*imdatachain.DatachainFile)
			key := config.DBKEY_improxy_user_message_history_fileHash_send
			messageOne, ERR := db.FindMessageHistoryByFileHash(*key, *Node.GetNetId(), dataChainFile.AddrTo, dataChainFile.Hash, dataChainFile.SendTime)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
				return ERR
			}
			//保存传送进度到数据库
			messageOne.TransProgress = int(float64(dataChainFile.BlockIndex+1) / float64(dataChainFile.BlockTotal) * 100)
			//utils.Log.Info().Msgf("保存传送进度到数据库:%d", messageOne.TransProgress)
			ERR = db.UpdateMessageHistoryByFileHash(*Node.GetNetId(), messageOne, nil)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
				return ERR
			}
			//图片传完的，拼接起来
			ERR = JoinImageBase64(messageOne)
			if ERR.CheckFail() {
				return ERR
			}
			//给前端一个通知
			msgVO := messageOne.ConverVO()
			msgVO.Subscription = config.SUBSCRIPTION_type_msg
			//msgVO.State = messageOne.State
			subscription.AddSubscriptionMsg(msgVO)
		}
	}
	//发送后的处理
	if !ERR.CheckSuccess() {
		//发送失败
		utils.Log.Error().Msgf("发送消息 错误:%s", ERR.String())
		return ERR
	} else {
		//发送成功
		//标记数据链为已经发送成功状态
		index := one.GetIndex()
		ERR = db.ImProxyClient_UpdateDataChainStatusByID(*Node.GetNetId(), one.GetAddrFrom(), index.Bytes(),
			config.IMPROXY_datachain_status_sendSuccess, nil)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("发送消息失败:%s", ERR.String())
			return ERR
		}
		//utils.Log.Info().Msgf("发送成功，删除未发送消息:%+v", one.GetID())
		//删除发送失败列表中的记录
		ERR = db.ImProxyClient_DelDataChainSendFail(*Node.GetNetId(), [][]byte{one.GetID()})
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("修改消息状态 错误:%s", ERR.String())
			return ERR
		}
		return utils.NewErrorSuccess()
	}
}

/*
自己没有代理节点，则直接发送给对方
判断对方是否有代理节点，有则优先发给对方代理节点
*/
func (this *ProxyClientManager) SendDataChain(itr imdatachain.DataChainProxyItr) utils.ERROR {
	index := itr.GetIndex()
	utils.Log.Info().Msgf("发送数据链:%s", index.String())
	addrTo := itr.GetAddrTo()
	if len(addrTo.GetAddr()) == 0 {
		//创建群消息没有发送者。
		return utils.NewErrorSuccess()
	}

	//发送数据链
	//utils.Log.Info().Msgf("自己发送，查询对方是否有代理节点")
	//先看对方是否有代理节点
	userinfo, ERR := FindAndSearchUserProxy(addrTo)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("查询对方代理节点 错误:%s", ERR.String())
		return ERR
	}
	utils.Log.Info().Msgf("发送数据链:%s", index.String())
	//utils.Log.Info().Msgf("用户信息:%+v", userinfo)
	//查询对方的代理节点
	proxys := make([]nodeStore.AddressNet, 0)
	userinfo.Proxy.Range(func(k, v interface{}) bool {
		key := k.(string)
		addr := nodeStore.NewAddressNet([]byte(key)) //nodeStore.AddressNet([]byte(key))
		proxys = append(proxys, *addr)
		return true
	})
	utils.Log.Info().Int("发送给代理节点数量", len(proxys)).Send()
	if len(proxys) == 0 {
		go SearchUserProxyInterval(addrTo)
	}

	//开始发送给多个代理节点，有一个成功就算成功
	have := false
	for _, proxy := range proxys {
		ERR = SendDatachainMsgToClientOrProxy(proxy, itr)
		if ERR.CheckSuccess() {
			have = true
		}
	}
	if have {
		return utils.NewErrorSuccess()
	}

	//只有转发消息，才需要发送出去
	if itr.GetCmd() != config.IMPROXY_Command_server_forward {
		return utils.NewErrorSuccess()
	}

	//判断对方是否有代理节点
	if len(proxys) > 0 {
		utils.Log.Info().Msgf("对方有代理节点")
		return ERR
	}
	utils.Log.Info().Msgf("发送数据链:%s", index.String())
	//utils.Log.Info().Msgf("直接发送给对方:%s", index.String())
	//对方没有代理节点，则直接发送消息
	ERR = SendDatachainMsgToClientOrProxy(itr.GetAddrTo(), itr)
	if ERR.CheckFail() {
		if ERR.Code == config.ERROR_CODE_IM_datachain_exist {
			return utils.NewErrorSuccess()
		}
		utils.Log.Info().Msgf("发送失败:%s", ERR.String())
		return ERR
	}
	utils.Log.Info().Msgf("发送数据链:%s", index.String())
	//utils.Log.Info().Msgf("直接发送给对方:%s", index.String())
	return utils.NewErrorSuccess()
}

/*
重新向某人发送消息
@addrFriend    nodeStore.AddressNet    好友地址
*/
func (this *ProxyClientManager) RetrySendDataChain(addrSelf, addrFriend nodeStore.AddressNet) {
	utils.Log.Info().Msgf("开始重发消息")
	once := utils.NewOnce()
	onceItr, ok := this.RetrySendDataChainOnce.LoadOrStore(utils.Bytes2string(addrFriend.GetAddr()), once)
	if ok {
		once = onceItr.(*utils.Once)
	}
	once.TryRun(func() {
		utils.Log.Info().Msgf("开始重发消息 ing")
		for {
			itrs, _, ERR := db.ImProxyClient_FindDataChainSendFailRange(*Node.GetNetId(), addrFriend, 0)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("发送消息失败:%s", ERR.String())
				return
			}
			if itrs == nil || len(itrs) == 0 {
				break
			}
			//先把消息状态改为正在发送
			for _, one := range itrs {
				utils.Log.Info().Msgf("开始重发消息:%+v", one.GetID())
				//先解密内容
				ERR = DecryptContent(one)
				if !ERR.CheckSuccess() {
					utils.Log.Info().Msgf("解密消息 错误:%s", ERR.String())
					return
				}
				clientItr := one.GetClientItr()
				//for _, clientItr := range one.GetClientItr() {
				clientItr.SetProxyItr(one)
				//}

				//推送给前端
				if one.GetCmd() == config.IMPROXY_Command_server_forward {
					clientItr := one.GetClientItr()
					if clientItr.GetClientCmd() == config.IMPROXY_Command_client_sendText {
						clientBase := clientItr.(*imdatachain.DataChainSendText)
						//utils.Log.Info().Msgf("修改消息状态:%s", clientBase.Content)
						//修改消息状态为待发送
						ERR = db.UpdateSendMessageStateV2(addrSelf, addrFriend, one.GetID(), config.MSG_GUI_state_not_send)
						if !ERR.CheckSuccess() {
							utils.Log.Info().Msgf("修改消息状态 错误:%s", ERR.String())
							return
						}
						//推送给前端
						messageOne := &model.MessageContent{
							Type:       config.MSG_type_text,            //消息类型
							FromIsSelf: true,                            //是否自己发出的
							From:       clientBase.AddrFrom,             //发送者
							To:         clientBase.AddrTo,               //接收者
							Content:    clientBase.Content,              //消息内容
							Time:       time.Now().Unix(),               //时间
							SendID:     clientItr.GetProxyItr().GetID(), //
							QuoteID:    clientBase.QuoteID,              //
							State:      config.MSG_GUI_state_not_send,   //
							//PullAndPushID: pullAndPushID,                 //
						}
						msgVO := messageOne.ConverVO()
						msgVO.Subscription = config.SUBSCRIPTION_type_msg
						msgVO.State = config.MSG_GUI_state_not_send
						subscription.AddSubscriptionMsg(msgVO)
					}
				}
			}

			failItrs := make([]imdatachain.DataChainProxyItr, 0)
			//发送出去
			for i, one := range itrs {
				ERR = this.sendDataChainOne(one)
				//ERR = this.uploadAndSendDataChain(one)
				if ERR.CheckFail() {
					utils.Log.Info().Msgf("发送消息失败:%s", ERR.String())
					failItrs = itrs[i:]
					break
				}
			}
			//把剩下未发送的消息状态改为失败，并推送给前端
			for _, one := range failItrs {
				//推送给前端
				if one.GetCmd() == config.IMPROXY_Command_server_forward {
					clientItr := one.GetClientItr()
					if clientItr.GetClientCmd() == config.IMPROXY_Command_client_sendText {
						clientBase := clientItr.(*imdatachain.DataChainSendText)
						//utils.Log.Info().Msgf("修改消息状态:%s", clientBase.Content)
						//修改消息状态为待发送
						ERR = db.UpdateSendMessageStateV2(addrSelf, addrFriend, one.GetID(), config.MSG_GUI_state_fail)
						if !ERR.CheckSuccess() {
							utils.Log.Info().Msgf("修改消息状态 错误:%s", ERR.String())
							return
						}
						//推送给前端
						messageOne := &model.MessageContent{
							Type:       config.MSG_type_text,            //消息类型
							FromIsSelf: true,                            //是否自己发出的
							From:       clientBase.AddrFrom,             //发送者
							To:         clientBase.AddrTo,               //接收者
							Content:    clientBase.Content,              //消息内容
							Time:       time.Now().Unix(),               //时间
							SendID:     clientItr.GetProxyItr().GetID(), //
							QuoteID:    clientBase.QuoteID,              //
							State:      config.MSG_GUI_state_fail,       //
							//PullAndPushID: pullAndPushID,                 //
						}
						msgVO := messageOne.ConverVO()
						msgVO.Subscription = config.SUBSCRIPTION_type_msg
						msgVO.State = config.MSG_GUI_state_fail
						subscription.AddSubscriptionMsg(msgVO)
					}
				}
			}
			//有失败，则跳出循环，否则会陷入死循环
			if len(failItrs) > 0 {
				return
			}
		}
	})
	utils.Log.Info().Msgf("开始重发消息 结束")
	return
}

/*
自己保存
*/
func (this *ProxyClientManager) SaveDataChain(itr imdatachain.DataChainProxyItr) utils.ERROR {
	//区分群消息和个人消息
	isGroupMsg := false
	switch itr.GetCmd() {
	case config.IMPROXY_Command_server_group_create:
		isGroupMsg = true
	}

	if isGroupMsg {
		proxyAddrs := this.GetProxyList()
		if len(proxyAddrs) == 0 {
			//没有代理，自己保存

		} else {
			//有代理，代理保存
		}
	} else {
		return this.ParserClient.SaveDataChain(itr)
	}

	return utils.NewErrorSuccess()
}

/*
自己循环监听群操作
*/
func (this *ProxyClientManager) LoopListenGroupOperation() {
	for one := range this.groupOperationChan {
		this.groupKnitManager.KnitDataChain(one)
		this.groupParserManager.SaveGroupDataChain(one)
		//this.groupParserManager.ParseDataChain(one)
	}
}

/*
强制托管群数据链
*/
func (this *ProxyClientManager) TrusteeshipKnitGroup(updateGroup *imdatachain.DataChainUpdateGroup) {
	this.groupKnitManager.KnitDataChain(updateGroup)
}

/*
解密内容
*/
func DecryptContent(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	proxyBase := proxyItr.GetBase()
	if proxyBase.Content == nil || len(proxyBase.Content) == 0 {
		//不需要解密
		return utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("开始解密:%d", proxyItr.GetCmd())
	//自己的私钥
	keystoreSelf, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	//keystoreSelf := Node.Keystore.GetDHKeyPair().KeyPair
	dhPrk := keystoreSelf.GetPrivateKey()
	//var ERR utils.ERROR
	var dhPuk *keystore.Key

	//utils.Log.Info().Msgf("解密消息:%d %+v", proxyItr.GetCmd(), proxyItr)
	addrRemote := proxyItr.GetAddrTo()
	switch proxyItr.GetCmd() {
	case config.IMPROXY_Command_server_init:
		puk := keystoreSelf.GetPublicKey()
		dhPuk = &puk
	case config.IMPROXY_Command_server_forward:
		addrRemote = proxyItr.GetAddrTo()
	case config.IMPROXY_Command_server_msglog_add: //添加消息
		addrRemote = proxyItr.GetAddrFrom()
		//优先用传来的公钥解密
		if proxyBase.DhPuk != nil && len(proxyBase.DhPuk) != 0 {
			utils.Log.Info().Msgf("优先用传来的公钥解密")
			//解析公钥信息
			dhPuk, ERR = config.ParseDhPukInfoV1(proxyBase.DhPuk)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("错误:%s", ERR.String())
				return ERR
			}
		}
	case config.IMPROXY_Command_server_msglog_del: //删除消息
		//addrRemote = proxyItr.GetAddrFrom()
	case config.IMPROXY_Command_server_group_create:
		//addrRemote = proxyItr.GetAddrTo()
	case config.IMPROXY_Command_server_group_members:
	case config.IMPROXY_Command_server_group_dissolve:
	}
	if dhPuk == nil {
		//utils.Log.Info().Msgf("查询公钥的地址:%s", addrRemote.B58String())
		dhPuk, ERR = db.ImProxyClient_FindUserDhPuk(*Node.GetNetId(), addrRemote)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		if dhPuk == nil {
			utils.Log.Info().Msgf("查询公钥未找到")
			return utils.NewErrorBus(config.ERROR_CODE_IM_dh_not_exist, "")
		}
	}

	//utils.Log.Info().Msgf("加密用公私密钥:%+v %+v", dhPrk, dhPuk)
	//生成共享密钥sharekey
	sharekey, err := keystore.KeyExchange(keystore.NewDHPair(dhPrk, *dhPuk))
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	//解密消息
	ERR = proxyItr.DecryptContent(sharekey[:])
	if ERR.CheckFail() {
		utils.Log.Info().Msgf("解密失败 使用公钥地址:%s", ERR.String())
	}
	return ERR
}

/*
已上链区块统计
*/
func (this *ProxyClientManager) countBlockOne(bhvo *mining.BlockHeadVO) {
	for _, txOne := range bhvo.Txs {
		orderId, txPay := chain_orders.ParseTxOrderId(txOne)
		if orderId == nil {
			continue
		}
		ERR := this.CountOrderPay(*orderId, txPay)
		if ERR.CheckFail() {
			continue
		}
		//给前端推送已支付订单
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_chain_payOrder_client,
			Content: hex.EncodeToString(*orderId)}
		subscription.AddSubscriptionMsg(&msgInfo)
	}
	//清理过期未上链订单
	this.CleanOvertimeNotOnChain(bhvo.BH.Height)
	//清理过期未支付订单
	this.CleanOvertimeNotPayOrder(bhvo.BH.Height)
	//清理已经支付，过期订单
	this.CleanOvertimeOrder(bhvo.BH.Height)
}

/*
统计订单支付
*/
func (this *ProxyClientManager) CountOrderPay(orderId []byte, txPay *mining.Tx_Pay) utils.ERROR {
	//查找订单
	this.lock.RLock()
	orderFrom, ok := this.OrdersNotOnChain[utils.Bytes2string(orderId)]
	if !ok {
		orderFrom, ok = this.UnpaidOrders[utils.Bytes2string(orderId)]
	}
	this.lock.RUnlock()
	if !ok {
		return utils.NewErrorBus(config.ERROR_CODE_Not_present, "")
	}
	//utils.Log.Info().Msgf("链上有已支付订单:%+v", orderFrom)
	//判断地址和费用是否给够
	haveAddr := false
	for _, one := range txPay.Vout {
		//utils.Log.Info().Msgf("链上有已支付订单%s %s %d %d", one.Address.B58String(), orderFrom.ServerAddrCoin.B58String(), one.Value, orderFrom.TotalPrice)
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
	ERR := this.MoveOrders(orderFrom)
	utils.Log.Info().Hex("已支付订单", orderFrom.Number).Send()
	return ERR
}

/*
统计续费订单支付
*/
func (this *ProxyClientManager) CountRenewalOrderPay(orderId []byte, txPay mining.TxItr) {

}

/*
清理过期未上链订单
*/
func (this *ProxyClientManager) CleanOvertimeNotOnChain(pullHeight uint64) {
	//定期检查已支付订单服务期是否超时
	this.lock.Lock()
	defer this.lock.Unlock()
	for orderKey, one := range this.OrdersNotOnChain {
		//utils.Log.Info().Msgf("检查支付未上链的交易:%d %d", one.LockHeightOnChain, pullHeight)
		//未超时
		if one.LockHeightOnChain > pullHeight {
			continue
		}
		//utils.Log.Info().Msgf("移动订单:%+v", one)
		ERR := db.ImProxyClient_MoveOrderNotOnChainToNotPay(*Node.GetNetId(), one)
		if !ERR.CheckSuccess() {
			//utils.Log.Error().Msgf("删除过期订单失败:%s", ERR.String())
			continue
		}
		//删除过期订单
		delete(this.OrdersNotOnChain, orderKey)
		this.UnpaidOrders[orderKey] = one
		utils.Log.Info().Hex("过期未上链订单", one.Number).Send()
	}
}

/*
清理过期未支付订单
*/
func (this *ProxyClientManager) CleanOvertimeNotPayOrder(pullHeight uint64) {
	//
	this.lock.Lock()
	defer this.lock.Unlock()
	for orderKey, one := range this.UnpaidOrders {
		//utils.Log.Info().Msgf("检查未支付过期订单:%d %d", one.PayLockBlockHeight, pullHeight)
		//未超时
		if one.PayLockBlockHeight > pullHeight {
			continue
		}
		//utils.Log.Info().Msgf("删除过期订单:%+v", one)
		ERR := db.ImProxyClient_MoveOrderToNotPayTimeout(*Node.GetNetId(), one)
		if !ERR.CheckSuccess() {
			//utils.Log.Error().Msgf("删除过期订单失败:%s", ERR.String())
			continue
		}
		//删除过期订单
		delete(this.UnpaidOrders, orderKey)
		utils.Log.Info().Hex("过期未支付订单", one.Number).Send()
	}
}

/*
定时清理过期的订单
*/
func (this *ProxyClientManager) CleanOvertimeOrder(pullHeight uint64) {
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
		for _, orderOne := range this.UnpaidOrders {
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
		ERR := db.ImProxyClient_MoveOrderToInUseTimeout(*Node.GetNetId(), one)
		if !ERR.CheckSuccess() {
			//utils.Log.Error().Msgf("删除超期订单失败:%s", ERR.String())
			continue
		}
		//删除过期订单
		delete(this.Orders, orderKey)
		utils.Log.Info().Hex("服务过期订单", one.Number).Send()
		//utils.Log.Info().Msgf("订单过期，空间减少:%s %d", one.UserAddr, utils.Byte(one.SpaceTotal)*utils.GB)
	}
}

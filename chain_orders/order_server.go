package chain_orders

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"sync"
	"time"
	chainConfig "web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain_boot/chain_plus"
	"web3_gui/chain_boot/object_beans"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/im/subscription"
	"web3_gui/utils"
)

var OrderServer *ChainOrderServer = NewChainOrderServer()

func Start() {
	//OrderServer = NewChainOrderServer()
	RegisterHandlers()
}

/*
注册一个类型的订单
*/
func RegisterGoodsClass(class uint64, gf GoodsFactory) {
	OrderServer.RegisterGoodsClass(class, gf)
}

type ChainOrderServer struct {
	syncCount    *chain_plus.ChainSyncCount //
	lock         *sync.RWMutex              //
	GoodsFactory map[uint64]GoodsFactory    //商品类型工厂。key:=商品类型;value:=工厂;
	SalesGoods   map[string]GoodsItr        //在售商品。key:=商品id;value:=;
	UnpaidOrders map[string]OrderItr        //未完成订单。key:=订单id;value:=;
}

func NewChainOrderServer() *ChainOrderServer {
	Order_Overtime_Height = uint64(Order_Overtime / chainConfig.Mining_block_time) //把订单超时时间换算为区块高度
	cos := ChainOrderServer{
		lock:         new(sync.RWMutex),
		GoodsFactory: make(map[uint64]GoodsFactory), //商品类型工厂。key:=商品类型;value:=工厂;
		SalesGoods:   make(map[string]GoodsItr),     //在售商品。key:=商品id;value:=;
		UnpaidOrders: make(map[string]OrderItr),     //未完成订单。key:=订单id;value:=;
	}
	cos.syncCount = chain_plus.NewChainSyncCount(nil, config.DBKEY_ChainCount_order_server, cos.countBlockOne)
	go cos.lazyStartChainCount()
	//RegisterHandlers()
	return &cos
}

/*
循环拉取区块记录
*/
func (this *ChainOrderServer) RegisterGoodsClass(class uint64, gf GoodsFactory) {
	this.lock.Lock()
	this.GoodsFactory[class] = gf
	this.lock.Unlock()
}

/*
循环拉取区块记录
*/
func (this *ChainOrderServer) lazyStartChainCount() {
	//等待链端准备好
	for {
		remoteHeight := this.syncCount.GetChainCurrentHeight()
		if remoteHeight == 0 {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	//延迟设置后，它自动启动
	this.syncCount.SetDB(config.LevelDB)
}

/*
统计区块
*/
func (this *ChainOrderServer) countBlockOne(bhvo *mining.BlockHeadVO) {
	for _, txOne := range bhvo.Txs {
		orderId, txPay := ParseTxOrderId(txOne)
		if orderId == nil {
			continue
		}

		//查找订单
		this.lock.Lock()
		orderFrom, ok := this.UnpaidOrders[utils.Bytes2string(*orderId)]
		this.lock.Unlock()
		if !ok {
			continue
		}
		//判断地址和费用是否给够
		haveAddr := false
		for _, one := range txPay.Vout {
			if bytes.Equal(one.Address, orderFrom.GetRecipient().GetAddr()) && one.Value >= orderFrom.GetPrice() {
				haveAddr = true
				break
			}
		}
		if !haveAddr {
			continue
		}
		//订单保存到已支付列表
		txBs, err := txOne.Proto()
		if err != nil {
			utils.Log.Error().Str("错误", err.Error()).Send()
			continue
		}
		ERR := db.Sharebox_server_SaveOrderPaid(orderFrom.GetId(), txOne.GetHash(), txBs)
		if ERR.CheckFail() {
			utils.Log.Error().Str("错误", ERR.String()).Send()
			continue
		}

		//删除未支付订单，更新库存
		this.lock.Lock()
		_, ok = this.UnpaidOrders[utils.Bytes2string(*orderId)]
		if ok {
			goods, ok := this.SalesGoods[utils.Bytes2string(orderFrom.GetGoodsId())]
			if ok {
				goods.UnLockOne()
			}
			delete(this.UnpaidOrders, utils.Bytes2string(*orderId))
		}
		this.lock.Unlock()

		//给前端推送已支付订单
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_chain_payOrder_server,
			Content: hex.EncodeToString(orderFrom.GetId())}
		subscription.AddSubscriptionMsg(&msgInfo)
	}
	//定时清理未支付订单
	if bhvo.BH.Height%1000 == 0 {
		this.lock.Lock()
		for k, one := range this.UnpaidOrders {
			if one.GetPayLockBlockHeight() < bhvo.BH.Height {
				ERR := db.Sharebox_Client_SaveOrderUnpaidOvertime([]byte(k))
				if ERR.CheckFail() {
					utils.Log.Error().Str("ERR", ERR.String()).Send()
					continue
				}
				delete(this.UnpaidOrders, k)
			}
		}
		this.lock.Unlock()
	}
}

/*
上架商品
@goodsId           []byte     商品id
@salesTotalMany    bool       虚拟商品，无限量供应
@total             uint64     上架数量
@price             uint64     价格
*/
func (this *ChainOrderServer) UpGoods(class uint64, goodsId []byte, salesTotalMany bool, total, price uint64) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	goodsFactory, ok := this.GoodsFactory[class]
	if !ok {
		utils.Log.Error().Uint64("未找到的商品类型", class).Send()
		return utils.NewErrorBus(config.ERROR_CODE_order_class_not_exist, strconv.Itoa(int(class)))
	}
	goodsInfo, ok := this.SalesGoods[utils.Bytes2string(goodsId)]
	if ok {
		goodsInfo.AddSalesTotal(goodsId, salesTotalMany, total, price)
	} else {
		goodsItr := goodsFactory()
		goodsItr.AddSalesTotal(goodsId, salesTotalMany, total, price)
		this.SalesGoods[utils.Bytes2string(goodsId)] = goodsItr
	}
	return utils.NewErrorSuccess()
}

/*
获取订单
*/
func (this *ChainOrderServer) GetOrder(goodsId []byte) (OrderItr, utils.ERROR) {
	//要等拉取到最高区块，才能接收订单
	if !this.syncCount.GetChainPullFinish() {
		return nil, utils.NewErrorBus(config.ERROR_CODE_order_chain_not_finish, "")
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	goods, ok := this.SalesGoods[utils.Bytes2string(goodsId)]
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_CODE_order_goodsId_noexist, "")
	}

	orderItr, ERR := goods.GetOrder(goodsId, this.syncCount.GetPullHeight())
	if ERR.CheckFail() {
		return nil, ERR
	}
	//放入未支付订单
	this.UnpaidOrders[utils.Bytes2string(orderItr.GetId())] = orderItr
	return orderItr, utils.NewErrorSuccess()
}

/*
解析交易中的订单id
*/
func ParseTxOrderId(txOne mining.TxItr) (*[]byte, *mining.Tx_Pay) {
	if txOne.Class() != chainConfig.Wallet_tx_type_pay {
		return nil, nil
	}
	txPay := txOne.(*mining.Tx_Pay)
	if txPay.Comment == nil || len(txPay.Comment) == 0 {
		return nil, nil
	}
	commonItr, ERR := object_beans.ParseObject(txPay.Comment)
	if ERR.CheckFail() {
		utils.Log.Error().Str("解析错误", ERR.String()).Send()
		return nil, nil
	}
	if commonItr.GetClass() != object_beans.CLASS_COMMON_order_number {
		return nil, nil
	}
	commonOrder := commonItr.(*object_beans.CommonOrder)
	return &commonOrder.Number, txPay
}

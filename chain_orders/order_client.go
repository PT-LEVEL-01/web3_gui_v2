package chain_orders

import (
	"bytes"
	"encoding/hex"
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

var OrderClientStatic *ChainOrderClient = NewChainOrderClient()

type ChainOrderClient struct {
	syncCount    *chain_plus.ChainSyncCount //
	lock         *sync.RWMutex              //
	UnpaidOrders map[string]OrderItr        //未完成订单。key:=订单id;value:=;
}

func NewChainOrderClient() *ChainOrderClient {
	Order_Overtime_Height = uint64(Order_Overtime / chainConfig.Mining_block_time) //把订单超时时间换算为区块高度
	cos := ChainOrderClient{
		lock:         new(sync.RWMutex),
		UnpaidOrders: make(map[string]OrderItr), //未完成订单。key:=订单id;value:=;
	}
	cos.syncCount = chain_plus.NewChainSyncCount(nil, config.DBKEY_ChainCount_order_client, cos.countBlockOne)
	go cos.lazyStartChainCount()
	//RegisterHandlers()
	return &cos
}

/*
循环拉取区块记录
*/
func (this *ChainOrderClient) lazyStartChainCount() {
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
func (this *ChainOrderClient) countBlockOne(bhvo *mining.BlockHeadVO) {
	for _, txOne := range bhvo.Txs {
		if txOne.Class() != chainConfig.Wallet_tx_type_pay {
			continue
		}
		txPay := txOne.(*mining.Tx_Pay)
		if txPay.Comment == nil || len(txPay.Comment) == 0 {
			continue
		}
		commonItr, ERR := object_beans.ParseObject(txPay.Comment)
		if ERR.CheckFail() {
			utils.Log.Error().Str("解析错误", ERR.String()).Send()
			continue
		}
		if commonItr.GetClass() != object_beans.CLASS_COMMON_order_number {
			continue
		}
		commonOrder := commonItr.(*object_beans.CommonOrder)
		//查找订单
		this.lock.Lock()
		orderFrom, ok := this.UnpaidOrders[utils.Bytes2string(commonOrder.Number)]
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
		//utils.Log.Info().Interface("交易详情", txOne).Send()
		//utils.Log.Info().Interface("交易hash", txOne.GetHash()).Send()
		ERR = db.Sharebox_Client_SaveOrderPaid(orderFrom.GetId(), txOne.GetHash(), txBs)
		if ERR.CheckFail() {
			utils.Log.Error().Str("错误", ERR.String()).Send()
			continue
		}

		//删除未支付订单，更新库存
		this.lock.Lock()
		delete(this.UnpaidOrders, utils.Bytes2string(commonOrder.Number))
		this.lock.Unlock()

		//给前端推送已支付订单
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_chain_payOrder_client,
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
注册订单
*/
func (this *ChainOrderClient) RegOrder(order OrderItr) utils.ERROR {
	this.lock.Lock()
	defer this.lock.Unlock()
	//放入未支付订单
	this.UnpaidOrders[utils.Bytes2string(order.GetId())] = order
	return utils.NewErrorSuccess()
}

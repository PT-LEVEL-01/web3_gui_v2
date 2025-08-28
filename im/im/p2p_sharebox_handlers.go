package im

import (
	"web3_gui/chain_boot/object_beans"
	"web3_gui/chain_orders"
	"web3_gui/config"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/utils"
)

/*
获取订单
*/
func Sharebox_GetOrder(message *message_center.MessageBase) {
	np, err := utils.ParseNetParams(message.Content)
	if err != nil {
		utils.Log.Info().Msgf("接收云存储空间提供者广播，解析参数出错:%s", err.Error())
		return
	}
	orderItr, err := object_beans.ParseOrderShareboxOrder(np.Data) //model.ParseOrderForm(np.Data)
	if err != nil {
		return
	}
	order := orderItr.(*object_beans.OrderShareboxOrder)
	orderItr, ERR := chain_orders.OrderServer.GetOrder(order.GoodsId)
	if ERR.CheckFail() {
		config.ReplyError(config.Node, config.NET_protocol_version_v1, message, ERR.Code, ERR.Msg)
		return
	}
	bs, err := orderItr.Proto()
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	config.ReplySuccess(config.Node, config.NET_protocol_version_v1, message, bs)
}

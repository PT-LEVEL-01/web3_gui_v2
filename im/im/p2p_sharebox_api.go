package im

import (
	"time"
	"web3_gui/chain_boot/object_beans"
	"web3_gui/config"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func GetOrder(remoteAddr nodeStore.AddressNet, goodsId []byte, price uint64) (*object_beans.OrderShareboxOrder, utils.ERROR) {
	//utils.Log.Info().Msgf("下单")
	order := &object_beans.OrderShareboxOrder{
		OrderBase: object_beans.OrderBase{
			GoodsId:  goodsId,
			UserAddr: *config.Node.GetNetId(), //这里填别人的地址，相当于赠送给别人
		},
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
	resultBS, ERR := config.Node.SendP2pMsgHEWaitRequest(config.MSGID_SHAREBOX_Order_getOrder, &remoteAddr, bs, time.Second*10)
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
	orderItr, err := object_beans.ParseOrderShareboxOrder(nr.Data)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	order = orderItr.(*object_beans.OrderShareboxOrder)

	//utils.Log.Info().Msgf("返回订单:%+v", order)
	return order, utils.NewErrorSuccess()
}

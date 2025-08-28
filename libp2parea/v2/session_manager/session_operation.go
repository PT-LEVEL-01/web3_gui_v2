package session_manager

import (
	"bytes"
	"time"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
建立连接后，第一件事就是注册自己节点。
向对方发送自己的节点信息
@try    bool    是否尝试连接
*/
func (this *SessionManager) exchangeNodeInfo_net(ss engine.Session) (*nodeStore.NodeInfo, utils.ERROR) {
	//this.Log.Info().Str("调用了注册方法", "111").Send()
	nodeSelf := *this.nodeManager.NodeSelf
	//this.Log.Info().Interface("RemoteMultiaddr", nodeSelf).Send()
	//this.Log.Info().Interface("GetRemoteMultiaddr", ss.GetRemoteMultiaddr()).Send()
	nodeSelf.RemoteMultiaddr = ss.GetRemoteMultiaddr()
	//this.Log.Info().Interface("节点信息", nodeSelf).Send()
	bs, err := nodeSelf.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.Version_1, *bs)
	bs, err = np.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	bs, ERR := ss.SendWait(config.MSGID_exchange_nodeinfo, bs, time.Minute)
	if ERR.CheckFail() {
		this.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}
	//this.Log.Info().Str("调用了注册方法", "111").Send()
	nr, err := utils.ParseNetResult(*bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	ERR = nr.ConvertERROR()
	if ERR.CheckFail() {
		this.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}
	nodeRemote, err := nodeStore.ParseNodeProto(nr.Data)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//域名称相同
	if !bytes.Equal(nodeRemote.AreaName, this.nodeManager.NodeSelf.AreaName) {
		this.Log.Error().Str("AreaName different", nodeSelf.RemoteMultiaddr.String()).Send()
	}
	//域名称不相同
	return nodeRemote, utils.NewErrorSuccess()
}

/*
建立连接后，第一件事就是注册自己节点。
向对方发送自己的节点信息
@try    bool    是否尝试连接
*/
func (this *SessionManager) registerNodeInfo_net(ss engine.Session) utils.ERROR {
	//this.Log.Info().Str("调用了注册方法", "111").Send()
	//nodeSelf := *this.nodeManager.NodeSelf
	//this.Log.Info().Interface("RemoteMultiaddr", nodeSelf).Send()
	//this.Log.Info().Interface("GetRemoteMultiaddr", ss.GetRemoteMultiaddr()).Send()
	//nodeSelf.RemoteMultiaddr = ss.GetRemoteMultiaddr()
	//this.Log.Info().Interface("节点信息", nodeSelf).Send()
	//bs, err := nodeSelf.Proto()
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	np := utils.NewNetParams(config.Version_1, nil)
	bs, err := np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	bs, ERR := ss.SendWait(config.MSGID_register_node, bs, time.Minute)
	if ERR.CheckFail() {
		this.Log.Error().Str("ERR", ERR.String()).Send()
		return ERR
	}
	//this.Log.Info().Str("调用了注册方法", "111").Send()
	nr, err := utils.ParseNetResult(*bs)
	if err != nil {
		//this.Log.Info().Str("调用了注册方法", "111").Err(err).Send()
		return utils.NewErrorSysSelf(err)
	}
	ERR = nr.ConvertERROR()
	if ERR.CheckFail() {
		if ERR.Code != config.ERROR_code_Repeated_connections {
			this.Log.Error().Str("ERR", ERR.String()).Send()
		}
		//this.Log.Error().Str("调用了注册方法", "111").Str("出错", ERR.String()).Send()
		return ERR
	}
	//this.Log.Info().Str("调用了注册方法", "111").Send()
	//nodeRemote, err := nodeStore.ParseNodeProto(nr.Data)
	//if err != nil {
	//	return  utils.NewErrorSysSelf(err)
	//}
	//域名称相同
	//if bytes.Equal(nodeRemote.AreaName, this.nodeManager.NodeSelf.AreaName) {
	//	return nodeRemote, utils.NewErrorSuccess()
	//}
	//域名称不相同
	return utils.NewErrorSuccess()
}

/*
从邻居节点获取自己相关的逻辑节点地址
@return    []nodeStore.NodeInfo    WAN网络下的节点
@return    []nodeStore.NodeInfo    LAN网络下的节点
*/
func (this *SessionManager) getNearSuperIP_net(ss engine.Session) ([]nodeStore.NodeInfo, []nodeStore.NodeInfo, utils.ERROR) {
	np := utils.NewNetParams(config.Version_1, nil)
	bs, err := np.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	bs, ERR := ss.SendWait(config.MSGID_getNearSuperIP, bs, time.Second*8)
	if ERR.CheckFail() {
		//this.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, nil, ERR
	}
	//this.Log.Info().Int("收到的返回大小", len(*bs)).Send()
	nr, err := utils.ParseNetResult(*bs)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	bsss, err := config.ParseByteList(&nr.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	nodesWAN, err := nodeStore.ParseNodesProto(&bsss[0][0])
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	nodesLAN, err := nodeStore.ParseNodesProto(&bsss[0][1])
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	return nodesWAN, nodesLAN, utils.NewErrorSuccess()
}

/*
从邻居节点获取指定域网络
*/
func (this *SessionManager) searchAreaIP_net(ss engine.Session) ([]nodeStore.NodeInfo, []nodeStore.NodeInfo, utils.ERROR) {

	np := utils.NewNetParams(config.Version_1, nil)
	bs, err := np.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	bs, ERR := ss.SendWait(config.MSGID_searchAreaIP, bs, time.Second*8)
	if ERR.CheckFail() {
		//this.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, nil, ERR
	}
	//this.Log.Info().Int("收到的返回大小", len(*bs)).Send()
	nr, err := utils.ParseNetResult(*bs)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	bsss, err := config.ParseByteList(&nr.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	nodesWAN, err := nodeStore.ParseNodesProto(&bsss[0][0])
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	nodesLAN, err := nodeStore.ParseNodesProto(&bsss[0][1])
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	return nodesWAN, nodesLAN, utils.NewErrorSuccess()
}

/*
从邻居获取逻辑域名称
@return    []nodeStore.NodeInfo    WAN网络中的节点信息
@return    []nodeStore.NodeInfo    LAN网络中的节点信息
*/
func (this *SessionManager) getLogicAreaName_net(ss engine.Session) ([]nodeStore.NodeInfo, []nodeStore.NodeInfo, utils.ERROR) {
	np := utils.NewNetParams(config.Version_1, nil)
	bs, err := np.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	bs, ERR := ss.SendWait(config.MSGID_getLogicAreaIP, bs, time.Second*8)
	if ERR.CheckFail() {
		//this.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, nil, ERR
	}
	//this.Log.Info().Int("收到的返回大小", len(*bs)).Send()
	nr, err := utils.ParseNetResult(*bs)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	bsss, err := config.ParseByteList(&nr.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	nodesWAN, err := nodeStore.ParseNodesProto(&bsss[0][0])
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	nodesLAN, err := nodeStore.ParseNodesProto(&bsss[0][1])
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	return nodesWAN, nodesLAN, utils.NewErrorSuccess()

	//np := utils.NewNetParams(config.Version_1, nil)
	//bs, err := np.Proto()
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	//bs, ERR := ss.SendWait(config.MSGID_getLogicAreaIP, bs, time.Second*8)
	//if ERR.CheckFail() {
	//	return nil, ERR
	//}
	//nr, err := utils.ParseNetResult(*bs)
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	//bsss, err := config.ParseByteList(&nr.Data)
	//if err != nil {
	//	return nil, utils.NewErrorSysSelf(err)
	//}
	//nodeInfoBss := bsss[0]
	//nodeInfos := make([]*nodeStore.NodeInfo, 0, len(nodeInfoBss))
	//for _, one := range nodeInfoBss {
	//	nodeInfo, err := nodeStore.ParseNodeProto(one)
	//	if err != nil {
	//		return nil, utils.NewErrorSysSelf(err)
	//	}
	//	nodeInfos = append(nodeInfos, nodeInfo)
	//}
	//return nodeInfos, utils.NewErrorSuccess()
}

/*
询问关闭连接
*/
func (this *SessionManager) askCloseConn_net(ss engine.Session) utils.ERROR {
	if ss == nil {
		return utils.NewErrorSuccess()
	}
	//nodeRemote := GetNodeInfo(ss)
	//if nodeRemote != nil {
	//	this.Log.Info().Str("询问关闭 self", this.nodeManager.NodeSelf.IdInfo.Id.B58String()).
	//		Str("询问关闭 remote", nodeRemote.IdInfo.Id.B58String()).Send()
	//} else {
	//	this.Log.Info().Str("询问关闭", "222222").Send()
	//}
	np := utils.NewNetParams(config.Version_1, nil)
	bs, err := np.Proto()
	if err != nil {
		//this.Log.Error().Err(err).Send()
		return utils.NewErrorSysSelf(err)
	}
	_, ERR := ss.SendWait(config.MSGID_ask_close_conn, bs, time.Second*8)
	if ERR.CheckFail() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
向邻居节点推送新节点
*/
func (this *SessionManager) sendNewNode_net(ss engine.Session, newNode *nodeStore.NodeInfo) utils.ERROR {
	//this.Log.Info().Str("推送新节点信息", newNode.IdInfo.Id.B58String()).Send()
	data, err := nodeStore.NewNodeInfoList([]nodeStore.NodeInfo{*newNode}).Proto()
	if err != nil {
		//this.Log.Error().Err(err).Send()
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.Version_1, *data)
	bs, err := np.Proto()
	if err != nil {
		//this.Log.Error().Err(err).Send()
		return utils.NewErrorSysSelf(err)
	}
	_, ERR := ss.SendWait(config.MSGID_send_new_node, bs, time.Second*8)
	if ERR.CheckFail() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

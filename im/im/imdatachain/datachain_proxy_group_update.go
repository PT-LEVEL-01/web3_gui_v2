package imdatachain

import (
	"web3_gui/config"
	"web3_gui/libp2parea/v2/node_store"
)

func init() {
	RegisterLog(config.IMPROXY_Command_server_group_update, ParseDataChainUpdateGroupFactory)
}

/*
修改一个群
*/
type DataChainUpdateGroup struct {
	DataChainGroupMember
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseDataChainUpdateGroupFactory(bs []byte) (DataChainProxyItr, error) {
	mamberGroup, err := ParseDataChainGroupMember(bs)
	if err != nil {
		return nil, err
	}
	createGroup := DataChainUpdateGroup{DataChainGroupMember: *mamberGroup}
	return &createGroup, nil
}

/*
修改群昵称和是否禁言、修改代理节点
*/
func NewDataChainUpdateGroup(addrSelf nodeStore.AddressNet, proxyAddr nodeStore.AddressNet, nickname string,
	createTime int64, shoutUp bool, groupId []byte) *DataChainUpdateGroup {
	proxyBase := NewDataChainProxyBase(config.IMPROXY_Command_server_group_update, addrSelf, nodeStore.AddressNet{}, nodeStore.AddressNet{})
	proxyBase.GroupID = groupId
	addFriend := DataChainGroupMember{
		DataChainProxyBase: *proxyBase,
		ProxyMajor:         proxyAddr,  //指定一个代理节点作为群数据链构建者
		ShoutUp:            shoutUp,    //是否禁言
		Nickname:           nickname,   //群名称
		CreateTime:         createTime, //创建时间
	}
	updateGroup := DataChainUpdateGroup{
		addFriend,
	}
	return &updateGroup
}

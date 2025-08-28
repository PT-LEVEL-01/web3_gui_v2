package multicast

import (
	"errors"

	nodeStore "web3_gui/libp2parea/v2/node_store"
)

// 20010 - 20020 组播消息号范围

// 消息号
const (
	MSGID_P2P_SEND_MULTICAST_MSG      = 20010 // 发送组播消息
	MSGID_P2P_SEND_MULTICAST_MSG_RECV = 20011 // 发送组播消息的返回
)

// 错误信息
var (
	Err_MULTICAST_MSG_NOT_INIT       = errors.New("multicast is not init")
	Err_MULTICAST_NODES_EMPTY        = errors.New("nodeIds is empty")
	Err_MULTICAST_MSGID_INVALID      = errors.New("msgid is invalid")
	Err_MULTICAST_COUNTRY_NODE_EMPTY = errors.New("country node is empty")
)

// 节点机器id结构信息
type NodeMachineInfo struct {
	NodeId    nodeStore.AddressNet // 节点地址
	MachineID string               // 机器id
}

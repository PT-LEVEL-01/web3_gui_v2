package config

import "web3_gui/libp2parea/v2/engine"

// 消息协议版本号
var (
	Version_neighbor       = engine.RegMsgIdExistPanic(1) //邻居节点消息，本消息不转发
	Version_multicast_head = engine.RegMsgIdExistPanic(2) //广播消息hash
	Version_multicast_body = engine.RegMsgIdExistPanic(3) //通过广播消息hash值查询邻居节点的广播消息内容
	Version_multicast_recv = engine.RegMsgIdExistPanic(4) //通过广播消息hash值查询邻居节点的广播消息内容_返回
	Version_p2p            = engine.RegMsgIdExistPanic(5) //点对点明文消息
	//Version_p2pHE           = engine.RegMsgIdExistPanic(6)  //点对点加密消息，无等待返回消息
	Version_p2pHE_wait = engine.RegMsgIdExistPanic(7) //点对点加密消息，需要等待返回消息
	//Version_p2pHE_wait_recv = engine.RegMsgIdExistPanic(8)  //点对点加密消息，返回错误通道
	MSGID_p2p_get_node_info = engine.RegMsgIdExistPanic(11) //获取对方的公钥
	MSGID_register_node     = engine.RegMsgIdExistPanic(12) //首次连接注册节点
	MSGID_searchAreaIP      = engine.RegMsgIdExistPanic(13) //查找自己的域网络
	MSGID_getLogicAreaIP    = engine.RegMsgIdExistPanic(14) //从邻居节点得到自己的逻辑域网络节点
	MSGID_ask_close_conn    = engine.RegMsgIdExistPanic(15) //询问关闭连接
	MSGID_getNearSuperIP    = engine.RegMsgIdExistPanic(16) //从邻居节点得到自己的逻辑节点
	MSGID_send_new_node     = engine.RegMsgIdExistPanic(17) //向邻居节点推送新节点
	MSGID_exchange_nodeinfo = engine.RegMsgIdExistPanic(18) //节点连接后，互相交换节点信息

	MSGID_MAX_ID = 700 // 最大消息id号限制，系统只能注册0-700之间的消息号消息，用户可以注册>=700消息号的消息
)

// 等待消息类名称前缀
var (
	Wait_major_p2p_sys_msg = engine.RegClassNameExistPanic([]byte{11}) //p2p网络协议内部使用等待类型
)

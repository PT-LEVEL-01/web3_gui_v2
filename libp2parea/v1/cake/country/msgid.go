package country

// 20001 - 20010 大区消息号范围

const (
	//---------------- 白名单设置列表 模块 --------------------------
	MSGID_P2P_GET_DATA      = 20001 //获取设置数据
	MSGID_P2P_GET_DATA_BACK = 20002 //获取设置数据返回

	MSGID_P2P_SEND_DATA      = 20003 //同步数据到其它节点
	MSGID_P2P_SEND_DATA_BACK = 20004 //同步数据到其它节点返回

	MSGID_P2P_GET_NODE_IDS      = 20005 //获取设置的大区信息
	MSGID_P2P_GET_NODE_IDS_RECV = 20006 //获取设置的大区信息返回
)

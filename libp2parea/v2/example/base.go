package example

import (
	"bytes"
	"path/filepath"
	"strconv"
	"sync"
	"web3_gui/keystore/v2"
	kconfig "web3_gui/keystore/v2/config"
	"web3_gui/libp2parea/v2"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/libp2parea/v2/session_manager"
	"web3_gui/utils"
)

func StartNodeOne(areaName [32]byte, pre string, n int, wg *sync.WaitGroup) (*libp2parea.Node, utils.ERROR) {
	//utils.Log.Info().Str("start", "11111111").Send()
	//日志文件路径
	logPath := filepath.Join(Path_log_root, "log"+strconv.Itoa(n)+".txt")
	//密钥库文件路径
	keyPath := filepath.Join(Path_keystore_root, "keystore"+strconv.Itoa(n)+".key")
	//数据库文件路径
	dbPath := filepath.Join(Path_db_root, "db"+strconv.Itoa(n))
	//初始化一个密钥库，并指定路径和前缀
	key1 := keystore.NewKeystoreSingle(keyPath, AddrPre)
	//加载路径中的密钥库
	ERR := key1.Load()
	if ERR.CheckFail() {
		//文件不存在
		if ERR.Code == kconfig.ERROR_code_wallet_file_not_exist {
			//创建一个新的密钥库
			ERR = key1.CreateRand(keyPwd, keyPwd, keyPwd, keyPwd)
		}
		if ERR.CheckFail() {
			utils.Log.Error().Str("创建key1错误", ERR.String()).Send()
			return nil, ERR
		}
	}
	//utils.Log.Info().Str("start", "11111111").Send()
	//创建一个域节点
	node, ERR := libp2parea.NewNode(areaName, pre, key1, keyPwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("创建area错误", ERR.String()).Send()
		return nil, ERR
	}
	//
	log, err := utils.NewLogDefaultFile(logPath)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	node.SetLog(log)
	//设置数据库路径
	node.SetLeveldbPath(dbPath)
	//设置发现节点的地址和端口
	node.SetDiscoverPeer([]string{"/ip4/" + ip + "/tcp/" + strconv.Itoa(int(basePort-1)) + "/ws"})
	node.SetDiscoverPeer([]string{"/ip4/" + ip + "/tcp/" + strconv.Itoa(int(basePort)) + "/ws"})
	node.SetDiscoverPeer([]string{"/ip4/" + ip + "/tcp/" + strconv.Itoa(int(basePort+1)) + "/ws"})
	//area.SetDiscoverPeer([]string{"/ip4/127.0.0.1/tcp/" + strconv.Itoa(int(basePort))})
	node.Log.Info().Int("启动节点", n).Send()
	//area.Log.Info().Str("start", "11111111").Send()
	//指定本地监听端口，并启动节点
	ERR = node.StartUP(basePort + uint16(n))
	if ERR.CheckFail() {
		node.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}
	node.Log.Info().Int("启动节点", n).Send()
	//area.Log.Info().Str("END", "11111111").Send()
	//等待节点组网完成，当首节点无其他节点可以连接时，尝试连接所有节点失败后组网完成。
	node.WaitAutonomyFinish()
	node.Log.Info().Int("节点组网完成", n).Send()
	wg.Done()
	return node, ERR
}

/*
计算各个节点的逻辑节点，并打印
*/
func BuildLogicNodes(nodes []*libp2parea.Node) {
	for _, one := range nodes {
		kad := nodeStore.NewKademlia(one.NodeManager.NodeSelf.IdInfo.Id.Data(), nodeStore.NodeIdLevel, one.Log)
		for _, j := range nodes {
			if bytes.Equal(j.NodeManager.NodeSelf.IdInfo.Id.Data(), one.NodeManager.NodeSelf.IdInfo.Id.Data()) {
				continue
			}
			//one.Log.Info().Str("添加id", j.NodeManager.NodeSelf.IdInfo.Id.B58String()).Send()
			kad.AddId(j.NodeManager.NodeSelf.IdInfo.Id.Data())
			//one.Log.Info().Bool("是否添加", ok).Send()
		}
		ids := kad.GetIds()
		one.Log.Info().Str("开始计算逻辑节点", "------------------- "+strconv.Itoa(len(ids))).Send()
		for _, j := range ids {
			//one.Log.Info().Int("地址数据段长度", len(j)).Send()
			addr, ERR := nodeStore.BuildAddrByData(one.NodeManager.NodeSelf.IdInfo.Id.GetPre(), j)
			if ERR.CheckFail() {
				panic(ERR.String())
			}
			one.Log.Info().Str("理论逻辑节点", addr.B58String()).Send()
		}
	}
}

/*
打印实际逻辑节点地址
*/
func PrintActualNodes(nodes []*libp2parea.Node) {
	for _, one := range nodes {
		nodeInfos := one.GetNodeManager().GetLogicNodesLAN()
		one.Log.Info().Str("实际逻辑节点地址 LAN", "------------- "+strconv.Itoa(len(nodeInfos))).Send()
		for _, nodeOne := range nodeInfos {
			one.Log.Info().Str("实际逻辑节点地址", nodeOne.IdInfo.Id.B58String()).Send()
		}

		nodeInfos = one.GetNodeManager().GetLogicNodesWAN()
		one.Log.Info().Str("实际逻辑节点地址 WAN", "------------- "+strconv.Itoa(len(nodeInfos))).Send()
		for _, nodeOne := range nodeInfos {
			one.Log.Info().Str("实际逻辑节点地址", nodeOne.IdInfo.Id.B58String()).Send()
		}
	}
}

/*
打印各个节点的所有连接
*/
func PrintSessionNodes(nodes []*libp2parea.Node) {
	for _, one := range nodes {
		sss := one.SessionEngine.GetSessionAll()
		one.Log.Info().Str("开始打印各个会话节点", "------------------- "+strconv.Itoa(len(sss))).Send()
		//one.Log.Info().Int("开始打印各个会话节点", len(sss)).Send()
		for _, ss := range sss {
			nodeRemote := session_manager.GetNodeInfo(ss)
			if nodeRemote == nil {
				one.Log.Info().Str("会话节点", "").Hex("sid", ss.GetId()).Uint8("类型", ss.GetConnType()).Send()
			} else {
				one.Log.Info().Str("会话节点", nodeRemote.IdInfo.Id.B58String()).Hex("sid", ss.GetId()).Uint8("类型", ss.GetConnType()).Send()
			}
		}
	}
}

/*
打印理论上逻辑域
*/
func PrintLogicAreaName(nodes []*libp2parea.Node) {
	for _, one := range nodes {
		kad := nodeStore.NewKademlia(one.NodeManager.NodeSelf.AreaName, nodeStore.NodeIdLevel, one.Log)
		for _, j := range nodes {
			if bytes.Equal(j.NodeManager.NodeSelf.AreaName, one.NodeManager.NodeSelf.AreaName) {
				continue
			}
			//one.Log.Info().Str("添加id", j.NodeManager.NodeSelf.IdInfo.Id.B58String()).Send()
			kad.AddId(j.NodeManager.NodeSelf.AreaName)
			//one.Log.Info().Bool("是否添加", ok).Send()
		}
		ids := kad.GetIds()
		one.Log.Info().Str("开始计算逻辑域名称", "------------------- "+strconv.Itoa(len(ids))).Send()
		for _, j := range ids {
			one.Log.Info().Hex("理论逻辑域名称", j).Send()
		}
	}
}

/*
打印实际逻辑域名称
*/
func PrintActualAreaname(nodes []*libp2parea.Node) {
	for _, one := range nodes {
		nodeInfos := one.GetNodeManager().AreaLogicAddrLAN.GetNodeInfos()
		one.Log.Info().Str("实际逻辑域名称 LAN", "------------- "+strconv.Itoa(len(nodeInfos))).Send()
		for _, nodeOne := range nodeInfos {
			one.Log.Info().Hex("实际逻辑域名称", nodeOne.AreaName).Send()
		}

		nodeInfos = one.GetNodeManager().AreaLogicAddrWAN.GetNodeInfos()
		one.Log.Info().Str("实际逻辑域名称 WAN", "------------- "+strconv.Itoa(len(nodeInfos))).Send()
		for _, nodeOne := range nodeInfos {
			one.Log.Info().Hex("实际逻辑域名称", nodeOne.AreaName).Send()
		}
	}
}

/*
打印各个节点的所有域节点连接
*/
func PrintSessionNodeAreaname(nodes []*libp2parea.Node) {
	for _, one := range nodes {
		sss := one.SessionEngine.GetSessionAll()
		one.Log.Info().Str("开始打印各个域节点", "------------------- "+strconv.Itoa(len(sss))).Send()
		//one.Log.Info().Int("开始打印各个会话节点", len(sss)).Send()
		for _, ss := range sss {
			nodeRemote := session_manager.GetNodeInfo(ss)
			if nodeRemote == nil {
				one.Log.Info().Str("会话节点域名称", "").Hex("sid", ss.GetId()).Uint8("类型", ss.GetConnType()).Send()
			} else {
				one.Log.Info().Hex("会话节点域名称", nodeRemote.AreaName).Hex("sid", ss.GetId()).Uint8("类型", ss.GetConnType()).Send()
			}
		}
	}
}

/*
发送消息
*/
func SendMsg(nodes []*libp2parea.Node) {
	for _, nodeSend := range nodes {
		nodeSend.Log.Info().Str("开始发送消息", "------------------- "+strconv.Itoa(len(nodes))).Send()
		//one.Log.Info().Int("开始打印各个会话节点", len(sss)).Send()
		for _, nodeRecv := range nodes {
			//先发送个明文消息
			ERR := hello_o(nodeSend, nodeRecv.GetNetId())
			if ERR.CheckFail() {
				nodeSend.Log.Error().Str("发送明文消息失败", ERR.String()).Send()
			}
		}
	}
}

/*
发送加密消息
*/
func SendMsgHE(nodes []*libp2parea.Node) {
	for _, nodeSend := range nodes {
		nodeSend.Log.Info().Str("开始发送加密消息", "------------------------------ "+strconv.Itoa(len(nodes))).Send()
		//one.Log.Info().Int("开始打印各个会话节点", len(sss)).Send()
		for _, nodeRecv := range nodes {
			//发送个加密消息
			ERR := HelloHE_o(nodeSend, nodeRecv.GetNetId())
			if ERR.CheckFail() {
				nodeSend.Log.Error().Str("发送加密消息失败", ERR.String()).Send()
			}
		}
	}
}

/*
重启单个节点
*/
func Restart(nodes []*libp2parea.Node, index int) {
	node := nodes[index]
	//node.Log.Info().Str("开始重启节点", "--------------------").Send()
	node.Destroy()
	wg := new(sync.WaitGroup)
	wg.Add(1)
	newNode, ERR := StartNodeOne(node.AreaName, node.NodeManager.NodeSelf.IdInfo.Id.GetPre(), index, wg)
	if ERR.CheckFail() {
		node.Log.Error().Str("启动节点出错", ERR.String()).Send()
		return
	}
	//node.Log.Info().Str("开始重启节点", "--------------------").Send()
	wg.Wait()
	nodes[index] = newNode
	//node.Log.Info().Str("开始重启节点", "--------------------").Send()
	newNode.WaitAutonomyFinish()
	//node.Log.Info().Str("开始重启节点", "--------------------").Send()
}

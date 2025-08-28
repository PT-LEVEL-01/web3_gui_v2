package libp2parea

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node/manager"
	"web3_gui/utils"
)

const (
	version = 1
)

type Auth struct {
	AreaName      [32]byte
	nodeManager   *nodeStore.NodeManager
	sessionEngine *engine.Engine
	vc            *manager.VnodeCenter
	isOnline      chan bool
}

func NewAuth(areaId [32]byte, nm *nodeStore.NodeManager, se *engine.Engine, vc *manager.VnodeCenter) *Auth {
	return &Auth{
		AreaName:      areaId,
		nodeManager:   nm,
		sessionEngine: se,
		vc:            vc,
	}
}

/*
+++++++++++++++++++++++++++++++++++++++++++++++++++++++
| version   | ctp        | size      | name           |
+++++++++++++++++++++++++++++++++++++++++++++++++++++++
| 版本       | 连接类型    | 数据长度    | 连接名称         |
+++++++++++++++++++++++++++++++++++++++++++++++++++++++
| 2 byte    | 2 byte     | 4 byte    |                |
+++++++++++++++++++++++++++++++++++++++++++++++++++++++

version：版本
	1：第一个版本

ctp：连接类型
	1：带name的连接
	2：不带name的连接

name：连接名称
	区分每一个客户端的名称

*/

/*
 * GetAreaName 获取区域名称
 *
 * @return areaName		string		区域名称
 */
func (this *Auth) GetAreaName() (areaName string) {
	return utils.Bytes2string(this.AreaName[:])
}

// 发送
// @name                 本机服务器的名称
// @setGodAddr			 是否设置对方为自己的上帝节点标识
// @return  remoteName   对方服务器的名称
func (this *Auth) SendKey(conn net.Conn, session engine.Session, name string, setGodAddr bool) (remoteName, machineID string, setGodTime int64, params interface{}, connectKey string, err error) {

	// utils.Log.Info().Msgf("%s SendKey 11111 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
	// start := time.Now()
	//设置此方法总共10秒钟内完成验证，否则超时。
	conn.SetDeadline(time.Now().Add(time.Second * 10))
	defer conn.SetDeadline(time.Time{})
	//向对方发送域名称
	if err = this.sendAreaName(conn); err != nil {
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", err
	}
	// utils.Log.Info().Msgf("%s SendKey 11111", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	//接收对方的areaname验证结果
	areaNameResult, err := this.recvAreaCheckResult(conn)
	if err != nil {
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", err
	}
	// utils.Log.Info().Msgf("%s SendKey 11111", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	if areaNameResult == engine.AreaNameResult_same {
		//域名称相同
		// utils.Log.Info().Msgf("域名称相同 %v", areaNameResult[0])
	} else {
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", engine.Error_different_netid
	}
	// utils.Log.Info().Msgf("SendKey spend 111111:%s", time.Now().Sub(start))
	//第一次连接，向对方发送自己的Node
	var curTime int64
	if setGodAddr {
		curTime = time.Now().UnixMilli() // 设置超级代理时，记录下设置的时间戳信息
	}
	node := &nodeStore.Node{
		IdInfo:       this.nodeManager.NodeSelf.IdInfo,
		IsSuper:      0, //自己是否是超级节点，对方会判断，这里只需要虚心的说自己不是超级节点
		Addr:         this.nodeManager.NodeSelf.Addr,
		TcpPort:      this.nodeManager.NodeSelf.TcpPort,
		Version:      config.Version_1,
		IsApp:        this.nodeManager.NodeSelf.IsApp,
		MachineID:    this.nodeManager.NodeSelf.MachineID,
		SetGod:       setGodAddr,
		MachineIDStr: this.nodeManager.NodeSelf.MachineIDStr,
		SetGodTime:   curTime,
		QuicPort:     this.nodeManager.NodeSelf.QuicPort,
	}
	if err = this.sendNodeInfo(conn, node); err != nil {
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", err
	}
	// utils.Log.Info().Msgf("%s SendKey 11111", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	// utils.Log.Info().Msgf("SendKey spend 222222:%s", time.Now().Sub(start))
	nodeCheckResult, err := this.recvNodeCheckResult(conn)
	if err != nil {
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", err
	}
	// utils.Log.Info().Msgf("%s SendKey 11111", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	if nodeCheckResult == config.NodeIdResult_self {
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", config.ERROR_conn_self
	}
	if nodeCheckResult == config.NodeIdResult_online {
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", config.ERROR_online
	}
	// utils.Log.Info().Msgf("%s SendKey 11111", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	//接收对方的Node
	node, err = this.recvNodeInfo(conn)
	if err != nil {
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", err
	}
	node.Lock = new(sync.RWMutex)
	node.Sessions = make([]engine.Session, 0)
	//检查这个链接是否已经存在
	remoteName = utils.Bytes2string(node.IdInfo.Id) //
	connectKey = fmt.Sprintf("%s_%s", remoteName, node.MachineIDStr)
	if this.sessionEngine.LoadOrStoreConnecting(connectKey) {
		return "", "", 0, nil, connectKey, config.Error_node_connecting
	}
	// if _, ok := this.sessionEngine.GetSession(remoteName); ok {
	// 	//send 这个链接已经存在
	// 	return "", nil, config.ERROR_conn_exists
	// }
	// utils.Log.Info().Msgf("SendKey spend 333333:%s", time.Now().Sub(start))
	// utils.Log.Info().Msgf("%s SendKey 11111", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	//获取对方ip地址
	host, portStr, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		this.sessionEngine.DeleteConnecting(connectKey)
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, connectKey, err
	}
	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		this.sessionEngine.DeleteConnecting(connectKey)
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, connectKey, err
	}
	node.Addr = host
	node.TcpPort = uint16(portInt)
	//能直接通过ip地址访问的节点，一定是超级节点。
	node.IsSuper = 1

	isSuperBs := make([]byte, 2)

	_, err = io.ReadFull(conn, isSuperBs)
	if err != nil {
		this.sessionEngine.DeleteConnecting(connectKey)
		// utils.Log.Info().Msgf("%s SendKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		// utils.Log.Info().Msgf("add super nodeid error: %s", node.IdInfo.Id.B58String())
		return "", "", 0, nil, connectKey, err
	}

	// utils.Log.Info().Msgf("%s SendKey 11111", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	// utils.Log.Info().Msgf("SendKey spend 444444:%s", time.Now().Sub(start))
	isSuperInt := binary.LittleEndian.Uint16(isSuperBs)
	if isSuperInt == 1 {
		this.nodeManager.NodeSelf.SetIsSuper(true)
	} else {
		this.nodeManager.NodeSelf.SetIsSuper(false)
	}
	// utils.Log.Info().Msgf("%s SendKey 33333 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
	return remoteName, node.MachineIDStr, node.SetGodTime, node, connectKey, nil

	// err = nil

	// ok := this.nodeManager.AddNode(*node)
	// utils.Log.Info().Msgf("add super nodeid self:%s targetID:%s %t", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), node.IdInfo.Id.B58String(), ok)
	// if !ok {
	// 	utils.Log.Info().Msgf("%s SendKey 22222", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	// 	// utils.Log.Info().Msgf("不需要的节点:%s", hex.EncodeToString(node.IdInfo.Id))
	// 	return remoteName, node, engine.Error_node_unwanted
	// }
	// this.vc.NoticeAddNode(node.IdInfo.Id)
	// utils.Log.Info().Msgf("%s SendKey 33333", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	// // utils.Log.Info().Msgf("add super nodeid: %s", node.IdInfo.Id.B58String())
	// return
}

// 接收
// name   自己的名称
// @return  remoteName   对方服务器的名称
func (this *Auth) RecvKey(conn net.Conn, name string) (remoteName, machineID string, setGodTime int64, params interface{}, connectKey string, err error) {
	//接收对方的Node
	sizebs := make([]byte, 2)
	_, err = io.ReadFull(conn, sizebs)
	if err != nil {
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", err
	}
	// utils.Log.Info().Msgf("RecvKey 222222:%s", time.Now().Sub(start))
	size := binary.LittleEndian.Uint16(sizebs)
	nodeBs := make([]byte, size)
	_, err = io.ReadFull(conn, nodeBs)
	if err != nil {
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		// utils.Log.Error().Msgf("接收对方node错误 33333")
		return "", "", 0, nil, "", err
	}

	node, err := nodeStore.ParseNodeProto(nodeBs)
	if err != nil {
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", err
	}
	node.Lock = new(sync.RWMutex)
	node.Sessions = make([]engine.Session, 0)
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	// utils.Log.Info().Msgf("打印自己node:%s", this.nodeManager.NodeSelf.IdInfo)

	//检查地址是不是安全地址
	//	if !nodeStore.CheckSafeAddr(node.IdInfo.Puk) {
	//		fmt.Println("000", errors.New("idinfo非安全地址"))
	//		return "", errors.New("idinfo非安全地址")
	//	}
	//验证s256生成的地址
	if !nodeStore.CheckIdInfo(node.IdInfo) {
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		//		utils.Log.Error().Msgf("222 %s", "非法的 idinfo")
		//非法的 idinfo
		return "", "", 0, nil, "", errors.New("Illegal idinfo")
	}
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	// if !isOldVersion {
	nodeCheckResult := byte(config.NodeIdResult_offline)
	//若对方网络地址和自己的一样，那么断开连接
	if bytes.Equal(node.IdInfo.Id, this.nodeManager.NodeSelf.IdInfo.Id) {
		// utils.Log.Error().Msgf("333 自己连接自己，断开连接 %s self:%s", node.IdInfo.Id.B58String(), this.nodeManager.NodeSelf.IdInfo.Id.B58String())
		//自己连接自己，断开连接
		// return "", config.ERROR_conn_self, false
		nodeCheckResult = byte(config.NodeIdResult_self)
	}
	if err = this.sendNodeCheckResult(conn, nodeCheckResult); err != nil {
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, "", err
	}
	// }
	// utils.Log.Info().Msgf("RecvKey 333333:%s", time.Now().Sub(start))
	// utils.Log.Info().Msgf("打印对方nodeid:%s %s", node.IdInfo.Id.B58String(), node.MachineID)
	this.nodeManager.NodeSelf.SetIsSuper(true)

	//检查这个链接是否已经存在
	remoteName = utils.Bytes2string(node.IdInfo.Id)
	connectKey = fmt.Sprintf("%s_%s", remoteName, node.MachineIDStr)
	if this.sessionEngine.LoadOrStoreConnecting(connectKey) {
		return "", "", 0, nil, connectKey, config.Error_node_connecting
	}
	// if ss, ok := this.sessionEngine.GetSession(remoteName); ok {
	// 	//		utils.Log.Error().Msgf("444 这个链接已经存在 %s", remoteName)
	// 	// if ss.GetRemoteHost() == conn.RemoteAddr().String() {
	// 	// }
	// 	ss.Close()
	// 	//这个链接已经存在
	// 	// err = config.ERROR_conn_exists
	// 	// return
	// }
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	//	utils.Log.Info().Msgf("检查这个地址是否在网络中已经存在 111")

	// //检查这个地址是否在网络中已经存在
	// mid := GetNodeMachineID(&node.IdInfo.Id)
	// utils.Log.Info().Msgf("检查这个地址是否在网络中已经存在 222 %d", mid)

	// if mid != "" && node.MachineID != mid {
	// 	utils.Log.Info().Msgf("不能用相同的节点地址连接到网络")
	// 	return "", errors.New("不能用相同的节点地址连接到网络")
	// }
	// utils.Log.Info().Msgf("这个地址在网络中不存在 %d", mid)

	//给对方发送自己的Node
	// bs := nodeStore.NodeSelf.Marshal()
	bs, err := this.nodeManager.NodeSelf.Proto()
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, uint16(len(bs)))
	_, err = buf.Write(bs)
	if err != nil {
		this.sessionEngine.DeleteConnecting(connectKey)
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, connectKey, err
	}
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		this.sessionEngine.DeleteConnecting(connectKey)
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, connectKey, err
	}
	// utils.Log.Info().Msgf("RecvKey 4444444:%s", time.Now().Sub(start))
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	//获取对方ip地址
	node.Addr, _, err = net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		this.sessionEngine.DeleteConnecting(connectKey)
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		return "", "", 0, nil, connectKey, err
	}
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	// host, portStr, err := net.SplitHostPort(conn.RemoteAddr().String())
	// if err != nil {
	// 	fmt.Println("解析对方SplitHostPort错误", err)
	// 	return "", err
	// }
	// portInt, err := strconv.Atoi(portStr)
	// if err != nil {
	// 	fmt.Println("解析对方Port错误", err)
	// 	return "", err
	// }
	// node.Addr = host
	// node.TcpPort = uint16(portInt)

	// fmt.Println("解析到对方ip地址为", node.Addr)

	// node.Addr = strings.Split(conn.RemoteAddr().String(), ":")[0]
	// fmt.Println("RecvKey", strings.Split(conn.RemoteAddr().String(), ":")[0], conn.RemoteAddr().Network())

	//连接自己，又说自己是超级节点的，直接断开连接
	if node.GetIsSuper() {
		this.sessionEngine.DeleteConnecting(connectKey)
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		//这是一个验证是否有公网ip地址的超级节点的连接
		err = errors.New("This is a super node connection to verify whether there is a public IP address")
		return
	}
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	//如果是局域网地址，尝试局域网连接
	// if !utils.IsOnlyIp(node.Addr) && TryConn(node) {

	// 	nodeStore.AddLanNode(node)

	// 	buf = bytes.NewBuffer(nil)
	// 	binary.Write(buf, binary.LittleEndian, uint16(0))
	// 	nodeStore.AddProxyNode(node)
	// 	_, err = conn.Write(buf.Bytes())
	// 	if err != nil {
	// 		fmt.Println("连接错误 77777")
	// 		return "", err
	// 	}
	// 	select {
	// 	case isOnline <- true:
	// 	default:
	// 	}
	// 	return

	// }

	//判断对方是否是超级节点
	if this.nodeManager.NetType == config.NetType_release {
		// fmt.Println("网络类型是release")
		if !utils.IsOnlyIp(node.Addr) {
			// fmt.Println("不是公网ip地址")
			buf = bytes.NewBuffer(nil)
			binary.Write(buf, binary.LittleEndian, uint16(0))
			_, err = conn.Write(buf.Bytes())
			if err != nil {
				this.sessionEngine.DeleteConnecting(connectKey)
				// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
				return "", "", 0, nil, connectKey, err
			}
			node.Type = nodeStore.Node_type_proxy
			// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())

			return remoteName, node.MachineIDStr, node.SetGodTime, node, connectKey, nil
		}
	}
	// utils.Log.Info().Msgf("RecvKey 55555555:%s", time.Now().Sub(start))
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	// fmt.Println("判断对方是不是超级节点")
	//判断对方是否能链接上
	isSuper := false
	// if isOldVersion {
	// 	isSuper = TryConnOld(this.AreaName, node, this.nodeManager.NodeSelf)
	// } else {
	//是移动端，则不尝试连接对方了
	if !node.IsApp {
		// utils.Log.Info().Msgf("RecvKey 6666666:%s", time.Now().Sub(start))
		isSuper = TryConn(this.AreaName, node, this.nodeManager.NodeSelf)
	}
	// utils.Log.Info().Msgf("RecvKey 7777777:%s", time.Now().Sub(start))
	// }
	// node.IsSuper = isSuper
	node.SetIsSuper(isSuper)
	// fmt.Println("对方是不是超级节点", isSuper)

	buf = bytes.NewBuffer(nil)
	if isSuper {
		if err := binary.Write(buf, binary.LittleEndian, uint16(1)); err != nil {
			this.sessionEngine.DeleteConnecting(connectKey)
			// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
			return "", "", 0, nil, connectKey, err
		}
		node.Type = nodeStore.Node_type_logic
		// utils.Log.Info().Msgf("recv add super nodeid: %s", node.IdInfo.Id.B58String())
		// ok := this.nodeManager.AddNode(*node)
		// utils.Log.Info().Msgf("add super nodeid self:%s targetID:%s %t", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), node.IdInfo.Id.B58String(), ok)
		// if !ok {
		// 	//不是自己的逻辑节点，则是对方的逻辑节点
		// 	this.nodeManager.AddNodesClient(*node)
		// }
	} else {
		if err := binary.Write(buf, binary.LittleEndian, uint16(0)); err != nil {
			this.sessionEngine.DeleteConnecting(connectKey)
			// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
			return "", "", 0, nil, connectKey, err
		}
		node.Type = nodeStore.Node_type_proxy
		// this.nodeManager.AddProxyNode(*node)
	}
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		this.sessionEngine.DeleteConnecting(connectKey)
		// utils.Log.Info().Msgf("%s RecvKey 22222 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
		// this.nodeManager.DelNode(&node.IdInfo.Id)
		return "", "", 0, nil, connectKey, err
	}
	// utils.Log.Info().Msgf("打印对方nodeid:%s", node.IdInfo.Id.B58String())
	err = nil
	// utils.Log.Info().Msgf("%s RecvKey 33333 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
	//发送节点上线信号
	// select {
	// case this.isOnline <- true:
	// default:
	// }
	// utils.Log.Info().Msgf("RecvKey 88888888:%s", time.Now().Sub(start))
	return remoteName, node.MachineIDStr, node.SetGodTime, node, connectKey, nil
}

/*
发送域名称
*/
func (this *Auth) sendAreaName(conn net.Conn) error {
	//向对方发送网络id
	_, err := conn.Write(this.AreaName[:])
	return err
}

/*
接收域名称
*/
func (this *Auth) recvAreaName(conn net.Conn) error {
	return nil
}

/*
发送域名称验证结果
*/
func (this *Auth) sendAreaCheckResult(conn net.Conn) error {
	return nil
}

/*
接收域名称验证结果
*/
func (this *Auth) recvAreaCheckResult(conn net.Conn) (byte, error) {
	//接收对方的areaname验证结果
	areaNameResult := make([]byte, 1)
	_, err := io.ReadFull(conn, areaNameResult)
	if err != nil {
		utils.Log.Error().Msgf("recv remote areaNameResult error:%s", err.Error())
		return 0, err
	}
	return areaNameResult[0], nil
}

/*
发送节点信息
*/
func (this *Auth) sendNodeInfo(conn net.Conn, node *nodeStore.Node) error {
	// utils.Log.Info().Msgf("发送节点信息:%s", node.IdInfo.Id.B58String())
	bs, err := node.Proto()
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	err = binary.Write(buf, binary.LittleEndian, uint16(len(bs)))
	if err != nil {
		return err
	}
	_, err = buf.Write(bs)
	if err != nil {
		return err
	}
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

/*
接收节点信息
*/
func (this *Auth) recvNodeInfo(conn net.Conn) (*nodeStore.Node, error) {
	//接收对方的Node
	sizebs := make([]byte, 2)
	_, err := io.ReadFull(conn, sizebs)
	if err != nil {
		if err.Error() == io.EOF.Error() {
		} else if strings.Contains(err.Error(), "use of closed network connection") {
		} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
		} else {
			utils.Log.Error().Msgf("recv remote node size error:%s", err.Error())
		}
		return nil, engine.Error_different_netid
	}
	size := binary.LittleEndian.Uint16(sizebs)
	if size == 0 {
		return nil, engine.Error_different_netid
	}
	nodeBs := make([]byte, size)
	_, err = io.ReadFull(conn, nodeBs)
	if err != nil {
		// fmt.Println("接收对方node错误 44444", err)
		return nil, err
	}

	node, err := nodeStore.ParseNodeProto(nodeBs)
	// node, err = nodeStore.ParseNode(nodeBs)
	if err != nil {
		// fmt.Println("解析对方node错误 55555", err)
		return nil, err
	}
	if !nodeStore.CheckIdInfo(node.IdInfo) {
		//非法的 idinfo
		return nil, errors.New("illegal idinfo")
	}
	return node, nil
}

/*
发送节点id验证结果
*/
func (this *Auth) sendNodeCheckResult(conn net.Conn, nodeCheckResult byte) error {
	_, err := conn.Write([]byte{nodeCheckResult})
	return err
}

/*
接收节点id称验证结果
*/
func (this *Auth) recvNodeCheckResult(conn net.Conn) (byte, error) {
	//接收对方的areaname验证结果
	nodeNameResult := make([]byte, 1)
	_, err := io.ReadFull(conn, nodeNameResult)
	if err != nil {
		utils.Log.Error().Msgf("recv remote nodeNameResult error:%s", err.Error())
		return 0, err
	}
	return nodeNameResult[0], nil
}

/*
通过名称字符串获得bytes
@name   要序列化的name字符串
*/
func GetBytesForName(name string) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, int32(len(name)))
	buf.Write([]byte(name))
	return buf.Bytes()
}

/*
通过读连接中的bytes获取name字符串
*/
func GetNameForConn(conn net.Conn) (name string, err error) {
	lenghtByte := make([]byte, 4)
	io.ReadFull(conn, lenghtByte)
	nameLenght := binary.LittleEndian.Uint32(lenghtByte)
	nameByte := make([]byte, nameLenght)
	if n, e := conn.Read(nameByte); e != nil {
		err = e
		return
	} else {
		//得到对方名称
		name = string(nameByte[:n])
		return
	}
}

/*
尝试去连接一个ip地址，判断对方是否是超级节点
*/
func TryConnOld(AreaName [32]byte, srcNode *nodeStore.Node, nodeSelf *nodeStore.Node) bool {
	utils.Log.Info().Msgf("TryConn: %s:%s", srcNode.Addr, strconv.Itoa(int(srcNode.TcpPort)))
	//设置3秒钟超时
	conn, err := net.DialTimeout("tcp", srcNode.Addr+":"+strconv.Itoa(int(srcNode.TcpPort)), time.Second*3)
	if err != nil {
		return false
	}

	//向对方发送网络id
	_, err = conn.Write(AreaName[:])
	if err != nil {
		return false
	}

	// buf := bytes.NewBuffer(nil)
	// binary.Write(buf, binary.LittleEndian, uint32(engine.Netid))
	// _, err = conn.Write(buf.Bytes())
	// if err != nil {
	// 	return false
	// }

	//第一次连接，向对方发送自己的Node
	node := &nodeStore.Node{
		IdInfo:  nodeSelf.IdInfo,
		IsSuper: 1,
		Addr:    nodeSelf.Addr,
		TcpPort: nodeSelf.TcpPort,
		Version: config.Version_1,
	}
	// bs := node.Marshal()
	bs, err := node.Proto()
	if err != nil {
		return false
	}
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, uint16(len(bs)))
	_, err = buf.Write(bs)
	if err != nil {
		return false
	}
	_, err = conn.Write(buf.Bytes())

	//接收对方的Node
	sizebs := make([]byte, 2)
	_, err = io.ReadFull(conn, sizebs)
	if err != nil {
		return false
	}
	size := binary.LittleEndian.Uint16(sizebs)
	nodeBs := make([]byte, size)
	_, err = io.ReadFull(conn, nodeBs)
	if err != nil {
		return false
	}
	// node, err = nodeStore.ParseNode(nodeBs)
	node, err = nodeStore.ParseNodeProto(nodeBs)
	if err != nil {
		return false
	}
	if !nodeStore.CheckIdInfo(node.IdInfo) {
		return false
	}

	//检查这个链接是否已经存在
	//	remoteName = node.IdInfo.Id.B58String()
	//	if _, ok := engine.GetSession(remoteName); ok {
	//		err = errors.New("这个链接已经存在")
	//		return
	//	}

	utils.Log.Info().Msgf("remote ip: %s remote config ip: %s", conn.RemoteAddr().String(), node.Addr)
	//获取对方ip地址
	node.Addr = strings.Split(conn.RemoteAddr().String(), ":")[0]
	//	fmt.Println("SendKey", strings.Split(conn.RemoteAddr().String(), ":")[0], conn.RemoteAddr().Network())
	return true
}

/*
尝试去连接一个ip地址，判断对方是否是超级节点
*/
func TryConn(AreaName [32]byte, srcNode *nodeStore.Node, nodeSelf *nodeStore.Node) bool {
	// utils.Log.Info().Msgf("TryConn: %s:%s", srcNode.Addr, strconv.Itoa(int(srcNode.TcpPort)))
	//设置3秒钟超时
	conn, err := net.DialTimeout("tcp", srcNode.Addr+":"+strconv.Itoa(int(srcNode.TcpPort)), time.Second*1)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

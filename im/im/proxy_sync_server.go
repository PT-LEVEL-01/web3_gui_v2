package im

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"strconv"
	"sync"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type SyncServer struct {
	userIndex      sync.Map //保存用户数据链已经同步的index。key:string=用户地址;value:*big.Int=数据链Index;
	syncProcess    sync.Map //保存正在同步数据链的用户。key:string=用户地址;value:=;
	forwardProcess sync.Map //转发消息
}

func NewSyncServer() *SyncServer {
	return &SyncServer{
		userIndex:      sync.Map{},
		syncProcess:    sync.Map{},
		forwardProcess: sync.Map{},
	}
}

/*
开始从客户端同步数据链
*/
func (this *SyncServer) SyncDataChainForClient(addr nodeStore.AddressNet) {
	utils.Log.Info().Msgf("开始同步数据链")
	_, ok := this.syncProcess.LoadOrStore(utils.Bytes2string(addr.GetAddr()), nil)
	if ok {
		return
	}
	defer this.syncProcess.Delete(utils.Bytes2string(addr.GetAddr()))
	for {
		value, ok := this.userIndex.Load(utils.Bytes2string(addr.GetAddr()))
		if !ok {
			return
		}
		index := value.(*big.Int)
		index = new(big.Int).Add(index, big.NewInt(1))
		datachains, ERR := SyncDataChain(addr, index.Bytes())
		if !ERR.CheckSuccess() {
			return
		}
		if datachains == nil || len(datachains) == 0 {
			utils.Log.Info().Msgf("同步完成")
			break
		}
		//保存
		ERR = db.ImProxyClient_SaveDataChainMore(*Node.GetNetId(), nil, datachains...)
		if !ERR.CheckSuccess() {
			return
		}
		indexLast := datachains[len(datachains)-1].GetIndex()
		this.userIndex.Store(utils.Bytes2string(addr.GetAddr()), &indexLast)
	}
}

/*
检查并保存客户端的离线消息
*/
func (this *SyncServer) checkSaveDataChainNolink(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	//检查sendindex
	sendIndex, ERR := db.ImProxyServer_FindDatachainNoLinkSendIndex(*Node.GetNetId(), proxyItr.GetAddrFrom(), proxyItr.GetAddrTo())
	if ERR.CheckFail() {
		return ERR
	}
	if sendIndex.Cmp(big.NewInt(0)) == 0 {
		//离线消息中没有记录，则查询已经同步的sendIndex
		sendIndex, ERR = db.ImProxyServer_FindDatachainSendIndex(*Node.GetNetId(), proxyItr.GetAddrFrom(), proxyItr.GetAddrTo())
		if ERR.CheckFail() {
			return ERR
		}
	}
	proxyBase := proxyItr.GetBase()
	//
	if new(big.Int).Add(sendIndex, big.NewInt(1)).Cmp(proxyBase.SendIndex) != 0 {
		return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_sendIndex_discontinuity, "")
	}
	//保存代理消息
	ERR = db.ImProxyServer_SaveDatachainNoLink(*Node.GetNetId(), proxyItr)
	if ERR.CheckFail() {
		return ERR
	}
	//即时通知给客户端
	SendDatachainMsgToClientOrProxy(proxyItr.GetAddrTo(), proxyItr)
	return utils.NewErrorSuccess()
}

/*
检查并保存被代理节点的数据链
*/
func (this *SyncServer) CheckSaveDataChain(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	//utils.Log.Info().Msgf("检查并保存数据链:%d %+v", proxyItr.GetCmd(), proxyItr)
	proxyClientAddr := proxyItr.GetProxyClientAddr()
	userIndex := big.NewInt(0)
	//查询代理用户是否存在缓存中
	value, ok := this.userIndex.LoadOrStore(utils.Bytes2string(proxyClientAddr.GetAddr()), userIndex)
	if ok {
		userIndex = value.(*big.Int)
	} else {
		//不存在，去数据库查询
		proxyItr, ERR := db.ImProxyClient_FindDataChainLast(*Node.GetNetId(), proxyClientAddr)
		if !ERR.CheckSuccess() {
			return ERR
		}
		if proxyItr != nil {
			index := proxyItr.GetIndex()
			userIndex = &index
			this.userIndex.LoadOrStore(utils.Bytes2string(proxyClientAddr.GetAddr()), userIndex)
		}
	}
	//utils.Log.Info().Msgf("检查并保存数据链")
	//检查数据是否连续
	thisIndex := proxyItr.GetIndex()
	if new(big.Int).Add(userIndex, big.NewInt(1)).Cmp(&thisIndex) != 0 {
		//utils.Log.Info().Msgf("数据链不同步")
		//数据链不连续，开始同步。
		go this.SyncDataChainForClient(proxyClientAddr)
		return utils.NewErrorSuccess()
	}

	//utils.Log.Info().Msgf("检查并保存数据链")
	batch := new(leveldb.Batch)
	//索引连续，保存数据链记录
	ERR := db.ImProxyClient_SaveDataChainMore(*Node.GetNetId(), batch, proxyItr)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	//离线消息已上链，从未上链的离线消息中删除
	if proxyItr.GetCmd() == config.IMPROXY_Command_server_msglog_add {
		//保存已经上链的sendindex
		ERR = db.ImProxyServer_SaveDatachainSendIndex(*Node.GetNetId(), proxyItr.GetAddrFrom(), proxyItr.GetAddrTo(),
			proxyItr.GetBase().SendIndex.Bytes(), batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		//删除未上链消息
		ERR = db.ImProxyServer_RemoveDatachainNoLinkBySendIndex(*Node.GetNetId(), proxyItr.GetAddrFrom(), proxyItr.GetAddrTo(),
			proxyItr.GetBase().SendIndex.Bytes(), batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
	}
	ERR = utils.NewErrorSuccess()
	switch proxyItr.GetCmd() {
	case config.IMPROXY_Command_server_init: //初始化
	case config.IMPROXY_Command_server_forward: //转发消息
		ERR = this.forwardDataChain(proxyItr, batch)
	case config.IMPROXY_Command_server_msglog_add: //添加消息
	case config.IMPROXY_Command_server_msglog_del: //删除消息
	case config.IMPROXY_Command_server_group_create: //创建一个群
	case config.IMPROXY_Command_server_group_members: //添加或删除群成员
	case config.IMPROXY_Command_server_group_dissolve: //解散群聊
	default:
		utils.Log.Error().Msgf("未找到数据链中Proxy命令:%d", proxyItr.GetCmd())
		ERR = utils.NewErrorBus(config.ERROR_CODE_IM_datachain_cmd_exist, "未找到数据链中Proxy命令:"+strconv.Itoa(int(proxyItr.GetCmd())))
	}
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	err := db.SaveBatch(batch)
	if err != nil {
		utils.Log.Error().Msgf("提交batch错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	this.userIndex.Store(utils.Bytes2string(proxyClientAddr.GetAddr()), &thisIndex)
	return utils.NewErrorSuccess()
}

/*
转发消息
*/
func (this *SyncServer) forwardDataChain(proxyItr imdatachain.DataChainProxyItr, batch *leveldb.Batch) utils.ERROR {
	utils.Log.Info().Msgf("转发消息")
	ERR := db.ImProxyClient_SaveDataChainSendFail(*Node.GetNetId(), proxyItr, batch)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	go this.ForwardDataChain(proxyItr.GetAddrTo())
	return utils.NewErrorSuccess()
}

/*
触发转发给一个用户的所有消息
*/
func (this *SyncServer) ForwardDataChain(addr nodeStore.AddressNet) {
	utils.Log.Info().Msgf("开始转发数据链:%s", addr.B58String())
	_, ok := this.forwardProcess.LoadOrStore(utils.Bytes2string(addr.GetAddr()), nil)
	if ok {
		return
	}
	defer this.forwardProcess.Delete(utils.Bytes2string(addr.GetAddr()))
	isSendSelf := false
	//如果发送的目标地址就是自己
	if bytes.Equal(addr.GetAddr(), Node.GetNetId().GetAddr()) {
		isSendSelf = true
	}

	//查询用户的代理节点信息
	userinfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), addr)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return
	}
	addrs := make([]nodeStore.AddressNet, 0)
	if userinfo != nil {
		//放入代理节点地址
		userinfo.Proxy.Range(func(key, value interface{}) bool {
			addr := value.(nodeStore.AddressNet)
			addrs = append(addrs, addr)
			return true
		})
	}
	//如果没有代理节点地址，则发送到它自己的地址
	if len(addrs) == 0 {
		addrs = append(addrs, addr)
	}
	proxyItrs, ids, ERR := db.ImProxyClient_FindDataChainSendFailRange(*Node.GetNetId(), addr, 0)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return
	}
	if ids == nil || len(ids) == 0 {
		return
	}

	//for _, one := range proxyItrs {
	//	utils.Log.Info().Msgf("查询的离线数据链消息:%+v", one)
	//}

	for _, one := range proxyItrs {
		if isSendSelf {
			//是发送给自己的
			utils.Log.Info().Msgf("收到发送给自己的消息")
			//生成接收消息
			newItr := imdatachain.NewDataChainProxyMsgLog(one)
			//解密消息
			ERR = DecryptContent(newItr)
			if !ERR.CheckSuccess() {
				return
			}
			//并保存
			ERR = StaticProxyClientManager.ParserClient.SaveDataChain(newItr)
		} else {
			ERR = SendDatachainMsgToClientOrProxy(one.GetAddrTo(), one)
			utils.Log.Info().Msgf("发送是否成功:%+v", ERR)
		}
		if ERR.CheckFail() && ERR.Code != config.ERROR_CODE_IM_datachain_exist {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return
		}
		//发送一个，删除一个
		//ERR = db.ImProxyClient_DelDataChainSendFail(config.DBKEY_improxy_server_datachain_send_fail,
		//	config.DBKEY_improxy_server_datachain_send_fail_id, Node.GetNetId(), [][]byte{one.GetID()})
		//if !ERR.CheckSuccess() {
		//	utils.Log.Error().Msgf("错误:%s", ERR.String())
		//	return
		//}
	}
}

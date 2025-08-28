package message_center

import (
	"context"
	"sync"
	"time"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

type MessageCenter struct {
	key               keystore.KeystoreInterface            //
	nodeManager       *nodeStore.NodeManager                //
	sessionEngine     *engine.Engine                        //
	router            *Router                               //
	vm                *virtual_node.VnodeManager            //
	RatchetSession    *SessionManager                       //
	securityStore     map[string]map[string]*SearchNodeInfo // map[addr]map[machineId]*SearchNodeInfo
	securityStoreLock *sync.RWMutex                         //
	contextRoot       context.Context                       //
	// msgchannl      chan *MsgHolderOne         //
	msgchannl     *utils.ChanDynamic    //
	msgHolderLock *sync.RWMutex         //
	msgHolder     map[string]*MsgHolder // new(sync.Map)
	msgHashTask   *utils.Task           //
	msgHashMap    *sync.Map             //
	levelDB       *utilsleveldb.LevelDB //
	areaName      []byte                // 区域名称

}

func NewMessageCenter(nodeManager *nodeStore.NodeManager, sessionEngine *engine.Engine,
	vm *virtual_node.VnodeManager, key keystore.KeystoreInterface, c context.Context, areaName []byte) *MessageCenter {
	rsm := NewSessionManager()
	mc := &MessageCenter{
		key:               key,
		nodeManager:       nodeManager,
		sessionEngine:     sessionEngine,
		router:            NewRouter(),
		vm:                vm,
		RatchetSession:    &rsm,
		securityStore:     make(map[string]map[string]*SearchNodeInfo),
		securityStoreLock: new(sync.RWMutex),
		contextRoot:       c,
		// msgchannl:      make(chan *MsgHolderOne, 1000),
		msgchannl:     utils.NewChanDynamic(10000),
		msgHolderLock: new(sync.RWMutex),
		msgHolder:     make(map[string]*MsgHolder), //
		msgHashMap:    new(sync.Map),
		areaName:      areaName,
	}
	mc.msgHashTask = utils.NewTask(mc.sendhashTaskFun)
	mc.Init()

	return mc
}

/*
销毁
*/
func (this *MessageCenter) Destroy() {
	this.msgHashTask.Destroy()
}

/*
设置数据库
*/
func (this *MessageCenter) SetLevelDB(leveldb *utilsleveldb.LevelDB) {
	this.levelDB = leveldb
}

/*
检查节点是否在线
*/
func (this *MessageCenter) CheckOnline() bool {
	sessions := this.sessionEngine.GetAllSession(this.areaName)
	if len(sessions) > 0 {
		// utils.Log.Info().Msgf("连接数量:%d", len(sessions))
		// for _, one := range sessions {
		// 	this.nodeManager.GetNodesClient()
		// 	utils.Log.Info().Msgf("连接节点地址:%s RemoteHost:%s", nodeStore.AddressNet([]byte(one.GetName())).B58String(), one.GetRemoteHost())
		// }
		return true
	}
	return false
}

/*
检查节点是否在线
*/
func (this *MessageCenter) sendhashTaskFun(class string, params []byte) {
	this.msgHashMap.Delete(utils.Bytes2string(params))
}

/*
检查重复消息，检查这个消息是否发送过
@return    bool    是否重复。true=重复;false=不重复;
*/
func (this *MessageCenter) CheckRepeatHash(sendhash []byte) bool {
	_, ok := this.msgHashMap.Load(utils.Bytes2string(sendhash))
	if !ok {
		this.msgHashMap.Store(utils.Bytes2string(sendhash), nil)
		this.msgHashTask.Add(time.Now().Unix()+config.MsgCacheTimeOver, "", sendhash)
	}
	return ok
}

/*
 * 清理加密通道信息
 */
func (this *MessageCenter) CleanHEInfo(id nodeStore.AddressNet, machineId string) {
	if this.securityStore != nil {
		strKey := utils.Bytes2string(id)
		delete(this.securityStore, strKey)
	}

	this.RatchetSession.RemoveSendPipe(id, machineId)
}

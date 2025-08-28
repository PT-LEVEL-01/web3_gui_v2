package message_center

import (
	"context"
	"github.com/rs/zerolog"
	"sync"
	"time"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

type MessageCenter struct {
	key               keystore.KeystoreInterface            //
	pwd               string                                //
	nodeManager       *nodeStore.NodeManager                //
	sessionEngine     *engine.Engine                        //
	router            *Router                               //
	RatchetSession    *SessionManager                       //
	securityStoreLock *sync.RWMutex                         //
	securityStore     map[string]map[string]*SearchNodeInfo //map:string=节点地址;map:string=机器Id;value:*SearchNodeInfo=;
	contextRoot       context.Context                       //
	msgchannl         *utils.ChanDynamic                    //
	msgHolderLock     *sync.RWMutex                         //
	msgHolder         map[string]*MsgHolder                 //new(sync.Map)
	msgHashTask       *utils.Task                           //
	msgHashMap        *sync.Map                             //
	levelDB           *utilsleveldb.LevelDB                 //
	areaName          []byte                                //区域名称
	Log               *zerolog.Logger                       //日志

	msgHashCacheLock *sync.RWMutex    //
	msgHashCache     map[string]int64 //key:string=hash;value:=过期时间;
}

func NewMessageCenter(nodeManager *nodeStore.NodeManager, sessionEngine *engine.Engine, key keystore.KeystoreInterface,
	pwd string, c context.Context, areaName []byte, log *zerolog.Logger) *MessageCenter {
	rsm := NewSessionManager()
	mc := &MessageCenter{
		key:               key,
		pwd:               pwd,
		nodeManager:       nodeManager,
		sessionEngine:     sessionEngine,
		RatchetSession:    &rsm,
		securityStore:     make(map[string]map[string]*SearchNodeInfo),
		securityStoreLock: new(sync.RWMutex),
		contextRoot:       c,
		msgchannl:         utils.NewChanDynamic(10000),
		msgHolderLock:     new(sync.RWMutex),
		msgHolder:         make(map[string]*MsgHolder), //
		msgHashMap:        new(sync.Map),
		areaName:          areaName,
		Log:               log,
		msgHashCacheLock:  new(sync.RWMutex),
		msgHashCache:      make(map[string]int64),
	}
	mc.router = NewRouter(&mc.Log)
	mc.msgHashTask = utils.NewTask(mc.sendhashTaskFun)
	mc.Init()
	go mc.loopCleanMsgHashCache()
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
	sessions := this.sessionEngine.GetSessionAll()
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
	_, ok := this.msgHashMap.LoadOrStore(utils.Bytes2string(sendhash), nil)
	if !ok {
		this.msgHashTask.Add(time.Now().Unix()+config.MsgCacheTimeOver, "", sendhash)
	}
	return ok
}

/*
 * 清理加密通道信息
 */
func (this *MessageCenter) CleanHEInfo(id *nodeStore.AddressNet, machineId []byte) {
	if this.securityStore != nil {
		strKey := utils.Bytes2string(id.Data())
		delete(this.securityStore, strKey)
	}
	this.RatchetSession.RemoveSendPipe(id, machineId)
}

/*
循环清理消息hash
*/
func (this *MessageCenter) loopCleanMsgHashCache() {
	for range time.NewTicker(time.Second * 10).C {
		now := time.Now().Unix()
		this.msgHashCacheLock.Lock()
		for k, one := range this.msgHashCache {
			if now > one {
				delete(this.msgHashCache, k)
			}
		}
		this.msgHashCacheLock.Unlock()
	}
}

func (this *MessageCenter) CheckHaveMsgHash(shareKey *ShareKey) bool {
	this.msgHashCacheLock.Lock()
	defer this.msgHashCacheLock.Unlock()
	key := append(shareKey.B_DH_PUK[:], shareKey.A_DH_PUK[:]...)
	_, ok := this.msgHashCache[utils.Bytes2string(key)]
	if ok {
		return true
	}
	this.msgHashCache[utils.Bytes2string(key)] = time.Now().Unix() + 10
	return false
}

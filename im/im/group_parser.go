package im

import (
	"bytes"
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/oklog/ulid/v2"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"strconv"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/im/subscription"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type GroupParserManager struct {
	lock               *sync.RWMutex                      //
	groups             *sync.Map                          //保存自己加入的群
	groupOperationChan chan imdatachain.DataChainProxyItr //群操作管道
}

func NewGroupParserManager(groupOperationChan chan imdatachain.DataChainProxyItr) (*GroupParserManager, utils.ERROR) {
	gkm := GroupParserManager{
		lock:               new(sync.RWMutex),
		groups:             new(sync.Map),
		groupOperationChan: groupOperationChan,
	}
	ERR := gkm.LoadDB()
	return &gkm, ERR
}

func (this *GroupParserManager) LoadDB() utils.ERROR {
	//查询自己创建的群列表
	createGroups, ERR := db.ImProxyClient_FindGroupList(*config.DBKEY_improxy_group_list_create, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		return ERR
	}
	joinGroup, ERR := db.ImProxyClient_FindGroupList(*config.DBKEY_improxy_group_list_join, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		return ERR
	}
	groups := append(*createGroups, *joinGroup...)

	//加载快照
	for _, group := range groups {
		//utils.Log.Info().Str("加载群", group.Nickname).Send()
		bs, ERR := db.ImProxyClient_LoadGroupShot(*config.DBKEY_improxy_group_shot_index_parser, *Node.GetNetId(), group.GroupID)
		if !ERR.CheckSuccess() {
			return ERR
		}
		gp, err := ParseGroupParser(bs)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		gp.groupOperationChan = this.groupOperationChan
		//utils.Log.Info().Str("这里启动了一次", "").Send()
		gp.run()
		this.groups.Store(utils.Bytes2string(group.GroupID), gp)
	}
	return utils.NewErrorSuccess()
}

/*
获取自己发送给群的sendIndex
*/
func (this *GroupParserManager) GetGroupParse(groupId []byte) (*GroupParser, utils.ERROR) {
	pgItr, ok := this.groups.Load(utils.Bytes2string(groupId))
	if ok {
		pg := pgItr.(*GroupParser)
		return pg, utils.NewErrorSuccess()
		//return sendIndex, utils.NewErrorSuccess()
	}
	utils.Log.Info().Msgf("群不存在")
	return nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_not_exist, "")
}

/*
获取自己发送给群的sendIndex
*/
func (this *GroupParserManager) GetSendIndex(groupId []byte) (*big.Int, utils.ERROR) {
	pgItr, ok := this.groups.Load(utils.Bytes2string(groupId))
	if ok {
		pg := pgItr.(*GroupParser)
		sendIndex := pg.GetSendIndex()
		utils.Log.Info().Msgf("查询群sendIndex:%+v %+v", sendIndex.Bytes(), groupId)
		return sendIndex, utils.NewErrorSuccess()
	}
	utils.Log.Info().Msgf("群不存在")
	return nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_not_exist, "")
}

/*
获取自己发送给群的sendIndex
*/
func (this *GroupParserManager) GetShareKey(groupId []byte) ([]byte, utils.ERROR) {
	pgItr, ok := this.groups.Load(utils.Bytes2string(groupId))
	if ok {
		pg := pgItr.(*GroupParser)
		shareKey := pg.GetShareKey()
		//pg.lock.Lock()
		//shareKey := pg.ShareKey
		//pg.lock.Unlock()
		return shareKey, utils.NewErrorSuccess()
	}
	utils.Log.Info().Msgf("群不存在")
	return nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_not_exist, "")
}

/*
保存群数据链记录
*/
func (this *GroupParserManager) SaveGroupDataChain(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	utils.Log.Info().Msgf("parse group SaveGroupDataChain解析群:%d", proxyItr.GetCmd())
	if proxyItr.GetCmd() == config.IMPROXY_Command_server_group_create {
		//创建一个群
		createGroup := proxyItr.(*imdatachain.DataChainCreateGroup)
		pg := NewGroupParser(createGroup, this.groupOperationChan)
		ERR := pg.ParseDataChain(proxyItr)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("解析群 错误:%s", ERR.String())
			return ERR
		}
		_, ok := this.groups.LoadOrStore(utils.Bytes2string(createGroup.GroupID), pg)
		if !ok {
			pg.run()
		}
		return utils.NewErrorSuccess()
	} else if proxyItr.GetCmd() == config.IMPROXY_Command_server_group_members {
		//加入群聊
		groupMember := proxyItr.(*imdatachain.DataChainGroupMember)
		have := false
		for _, one := range groupMember.MembersAddr {
			if bytes.Equal(one.GetAddr(), Node.GetNetId().GetAddr()) {
				have = true
				break
			}
		}
		this.lock.Lock()
		defer this.lock.Unlock()
		pgItr, ok := this.groups.Load(utils.Bytes2string(groupMember.GroupID))
		if !ok && have {
			//半路加入群聊
			pg := NewGroupParserByGroupMember(groupMember, this.groupOperationChan)
			//utils.Log.Info().Str("这里启动了一次", "").Send()
			pg.run()
			ERR := pg.ParseDataChain(proxyItr)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("解析群 错误:%s", ERR.String())
				return ERR
			}
			this.groups.LoadOrStore(utils.Bytes2string(groupMember.GroupID), pg)
			return utils.NewErrorSuccess()
		}
		//退出了群聊
		if ok && !have {
			this.groups.Delete(utils.Bytes2string(groupMember.GroupID))
			return utils.NewErrorSuccess()
		}
		//其他成员变动，和自己无关
		if ok && have {
			pg := pgItr.(*GroupParser)
			ERR := pg.ParseDataChain(proxyItr)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("解析群 错误:%s", ERR.String())
				return ERR
			}
			return utils.NewErrorSuccess()
		}
		return utils.NewErrorSuccess()
	} else {
		//非创建群指令
		groupId := proxyItr.GetBase().GroupID
		pgItr, ok := this.groups.Load(utils.Bytes2string(groupId))
		if ok {
			pg := pgItr.(*GroupParser)
			ERR := pg.ParseDataChain(proxyItr)
			if !ERR.CheckSuccess() {
				if ERR.Code == config.ERROR_CODE_IM_index_discontinuity {
					//index不连续，则触发同步
					select {
					case pg.syncDownloadSignal <- nil:
					default:
					}
				}
				utils.Log.Info().Msgf("解析群 错误:%s", ERR.String())
				return ERR
			}
			return utils.NewErrorSuccess()
		} else {
			utils.Log.Info().Msgf("未找到群")
			return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_exist, "")
		}
	}
}

/*
群成员丢失了本地群记录，需要重新同步记录到本地
*/
func (this *GroupParserManager) DownloadBuildGroupDataChain(groupId []byte, addrProxyServer nodeStore.AddressNet) {
	startDataChain, ERR := GetGroupDataChainStartIndex(addrProxyServer, groupId)
	if !ERR.CheckSuccess() {
		return
	}
	ERR = this.SaveGroupDataChain(startDataChain)
	if !ERR.CheckSuccess() {
		return
	}
	//this.downloadGroupDataChain(groupId)
}

/*
群成员丢失了本地群记录，需要重新同步记录到本地
*/
//func (this *GroupParserManager) downloadGroupDataChain(groupId []byte) {
//	pgItr, ok := this.groups.Load(utils.Bytes2string(groupId))
//	if !ok {
//		return
//	}
//	pg := pgItr.(*GroupParser)
//	pg.downloadGroupDataChain()
//	return
//}

type GroupParser struct {
	GroupID            []byte                             //群id
	GroupKnit          nodeStore.AddressNet               //群构建者
	lock               *sync.RWMutex                      //锁
	IndexParse         *big.Int                           //本地解析到的数据链高度
	SendIndex          *big.Int                           //自己给群发送消息的index，必须连续的自增长ID
	PreHash            []byte                             //
	ShareKey           []byte                             //群共享密钥，用于加解密群内消息
	members            map[string]nodeStore.AddressNet    //群成员
	uploadChan         chan imdatachain.DataChainProxyItr //
	sendChan           chan imdatachain.DataChainProxyItr //发送队列
	syncDownloadSignal chan imdatachain.DataChainProxyItr //异步下载新消息信号
	syncSendTextSignal chan *SendTextTask                 //异步发送文本消息信号
	syncSendFileSignal chan bool                          //异步发送文件信号
	addrAdmin          nodeStore.AddressNet               //管理员地址
	addrKnitProxys     nodeStore.AddressNet               //群构建节点地址
	shoutUp            bool                               //是否禁言
	groupOperationChan chan imdatachain.DataChainProxyItr //群操作管道
	cancel             context.CancelFunc                 //
	ctx                context.Context                    //
	clone              *GroupParser                       //
	parseLock          *sync.Mutex                        //解析锁，一次只运行一个数据链解析方法
}

func NewGroupParser(createGroup *imdatachain.DataChainCreateGroup, groupOperationChan chan imdatachain.DataChainProxyItr) *GroupParser {
	groupParse := NewGroupParserByGroupMember(&createGroup.DataChainGroupMember, groupOperationChan)
	utils.Log.Info().Str("这里启动了一次", "").Send()
	//groupParse.run()
	return groupParse
}

func NewGroupParserByGroupMember(createGroup *imdatachain.DataChainGroupMember, groupOperationChan chan imdatachain.DataChainProxyItr) *GroupParser {
	ctx, cancel := context.WithCancel(context.Background())

	pg := GroupParser{
		//GroupID:            createGroup.GroupID,
		//GroupKnit:          createGroup.ProxyMajor,
		lock:               new(sync.RWMutex),
		SendIndex:          big.NewInt(0),
		PreHash:            nil,
		uploadChan:         nil,
		sendChan:           make(chan imdatachain.DataChainProxyItr, 10000), //
		IndexParse:         big.NewInt(0),
		syncDownloadSignal: make(chan imdatachain.DataChainProxyItr, 100),
		syncSendTextSignal: make(chan *SendTextTask, 100),
		syncSendFileSignal: make(chan bool, 1),
		//addrAdmin:          createGroup.AddrFrom,
		//addrKnitProxys:     createGroup.ProxyMajor,
		members: make(map[string]nodeStore.AddressNet),
		//shoutUp:            createGroup.ShoutUp,
		groupOperationChan: groupOperationChan,
		cancel:             cancel,
		ctx:                ctx,
		parseLock:          new(sync.Mutex), //
	}
	if createGroup != nil {
		pg.GroupID = createGroup.GroupID
		pg.GroupKnit = createGroup.ProxyMajor
		pg.addrAdmin = createGroup.AddrFrom
		pg.addrKnitProxys = createGroup.ProxyMajor
		pg.shoutUp = createGroup.ShoutUp
	}
	//pg.run()
	return &pg
}

/*
启动协程方法
*/
func (this *GroupParser) run() {
	go this.downloadGroupDataChain()
	go this.loopSendFile()
}

/*
获取解析高度
*/
func (this *GroupParser) GetIndexParse() *big.Int {
	this.lock.RLock()
	index := new(big.Int).SetBytes(this.IndexParse.Bytes())
	this.lock.RUnlock()
	return index
}

/*
设置解析高度
*/
func (this *GroupParser) SetIndexParse(newIndex *big.Int) {
	this.lock.Lock()
	this.IndexParse = newIndex
	this.lock.Unlock()
}

/*
获取发送高度
*/
func (this *GroupParser) GetSendIndex() *big.Int {
	this.lock.RLock()
	index := new(big.Int).SetBytes(this.SendIndex.Bytes())
	this.lock.RUnlock()
	return index
}

/*
设置发送高度
*/
func (this *GroupParser) SetSendIndex(newIndex *big.Int) {
	this.lock.Lock()
	this.SendIndex = newIndex
	this.lock.Unlock()
}

/*
获取发送高度
*/
func (this *GroupParser) GetShareKey() []byte {
	this.lock.RLock()
	shareKey := this.ShareKey
	this.lock.RUnlock()
	return shareKey
}

/*
设置发送高度
*/
func (this *GroupParser) SetShareKey(shareKey []byte) {
	this.lock.Lock()
	this.ShareKey = shareKey
	this.lock.Unlock()
}

/*
克隆，方便回滚
*/
func (this *GroupParser) Clone() {
	this.lock.Lock()
	defer this.lock.Unlock()
	gk := GroupParser{
		GroupKnit:      this.GroupKnit,
		SendIndex:      new(big.Int).SetBytes(this.SendIndex.Bytes()),
		PreHash:        this.PreHash,
		IndexParse:     new(big.Int).SetBytes(this.IndexParse.Bytes()),
		addrKnitProxys: this.addrKnitProxys,
		members:        make(map[string]nodeStore.AddressNet),
		shoutUp:        this.shoutUp,
		ShareKey:       this.ShareKey,
	}
	//gk.SendIndex = this.GetSendIndex()
	//gk.IndexParse = this.GetIndexParse()
	for k, v := range this.members {
		gk.members[k] = v
	}
	this.clone = &gk
	return
}

/*
回滚
*/
func (this *GroupParser) RollBACK(ERR utils.ERROR) {
	if ERR.CheckSuccess() {
		return
	}
	utils.Log.Error().Msgf("group parse 回滚 RollBACK")
	this.lock.Lock()
	defer this.lock.Unlock()
	this.GroupKnit = this.clone.GroupKnit
	this.SendIndex = this.clone.SendIndex
	this.PreHash = this.clone.PreHash
	this.IndexParse = this.clone.IndexParse
	this.addrKnitProxys = this.clone.addrKnitProxys
	this.members = this.clone.members
	this.shoutUp = this.clone.shoutUp
	this.ShareKey = this.ShareKey
}

/*
循环定时下载群数据链
*/
func (this *GroupParser) downloadGroupDataChain() {
	utils.Log.Info().Str("循环下载群数据链11", "").Send()
	var proxyItrs []imdatachain.DataChainProxyItr
	var ERR utils.ERROR
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-this.ctx.Done():
			return
		default:
		}
		select {
		case <-ticker.C:
		case <-this.syncDownloadSignal:
			utils.Log.Info().Str("循环下载群数据链", "").Send()
		case <-this.ctx.Done():
			return
		}
		indexParse := this.GetIndexParse()
		utils.Log.Info().Str("循环下载群数据链 indexParse", indexParse.String()).Send()
		proxyItrs, ERR = DownloadGroupDataChain(this.addrKnitProxys, this.GroupID, indexParse.Bytes())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("下载群数据链 错误:%s", ERR.String())
			continue
		}
		utils.Log.Info().Int("循环下载群数据链", len(proxyItrs)).Send()
		if len(proxyItrs) == 0 {
			continue
		}
		for _, one := range proxyItrs {
			ERR = this.ParseDataChain(one)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("解析群数据链 错误:%s", ERR.String())
				continue
			}
		}
		if ERR.CheckFail() {
			continue
		}
		//还有消息，并且处理正确，则继续下载
		select {
		case this.syncDownloadSignal <- nil:
		default:
		}
	}
}

/*
发送文本消息
*/
func (this *GroupParser) SendText(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	utils.Log.Info().Str("开始发送文本消息", "").Send()
	task := NewSendTextTask(proxyItr)
	utils.Log.Info().Int("开始发送文本消息", len(this.syncSendTextSignal)).Send()
	this.syncSendTextSignal <- task
	utils.Log.Info().Str("开始发送文本消息", "").Send()
	ERR := <-task.errChan
	if ERR.CheckFail() {
		if ERR.Code == config.ERROR_CODE_IM_datachain_sendIndex_discontinuity {
			//index不连续，则触发同步
			select {
			case this.syncDownloadSignal <- nil:
			default:
			}
		}
	}
	utils.Log.Info().Str("开始发送文本消息", "").Send()
	return ERR
}

/*
发送文件
*/
func (this *GroupParser) SendFile() {
	this.syncSendFileSignal <- false
}

/*
循环发送文件
*/
func (this *GroupParser) loopSendFile() {
	utils.Log.Info().Str("循环发送群消息", "").Send()
	var ERR utils.ERROR
	for {
		if ERR.CheckFail() {
			utils.Log.Info().Str("循环发送群消息 等待50秒", ERR.String()).Send()
			time.Sleep(time.Second * 50)
		}
		//获取自己的发送index
		sendIndex := this.GetSendIndex()
		utils.Log.Info().Str("循环发送群消息", "").Send()

		var sendFileInfo *imdatachain.SendFileInfo
		var sendTextTask *SendTextTask
		for {
			utils.Log.Info().Str("循环发送群消息", "").Send()
			sendTextTask = nil
			//优先检查是否已经退出
			select {
			case <-this.ctx.Done():
				return
			default:
			}
			utils.Log.Info().Str("循环发送群消息", "").Send()
			//优先发送文本消息
			select {
			case sendTextTask = <-this.syncSendTextSignal:
				utils.Log.Info().Str("循环发送群消息", "").Send()
			default:
			}
			if sendTextTask == nil && sendFileInfo != nil {
				utils.Log.Info().Str("循环发送群文件", "").Send()
				//开始发送文件
				var datachainFile *imdatachain.DatachainFile
				datachainFile, ERR = sendFileInfo.BuildDatachainFile(*Node.GetNetId())
				if ERR.CheckFail() {
					utils.Log.Error().Str("ERR", ERR.String()).Send()
					break
				}
				//设置sendIndex
				datachainFile.GetProxyItr().GetBase().SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
				//加密
				ERR = datachainFile.GetProxyItr().EncryptContent(this.GetShareKey())
				if !ERR.CheckSuccess() {
					utils.Log.Error().Str("ERR", ERR.String()).Send()
					break
				}
				//发送给群
				ERR = SendGroupDataChain(this.addrKnitProxys, datachainFile.GetProxyItr())
				if !ERR.CheckSuccess() {
					utils.Log.Error().Str("ERR", ERR.String()).Send()
					if ERR.Code == config.ERROR_CODE_IM_datachain_sendIndex_discontinuity {
						//更新sendIndex，再发送一次
						break
					}
					if ERR.Code == config.ERROR_CODE_IM_index_discontinuity {
						//index不连续，则触发同步
						select {
						case this.syncDownloadSignal <- nil:
						default:
						}
						break
					}
					break
				} else {
					sendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
					sendFileInfo.BlockIndex++
					if sendFileInfo.BlockTotal == sendFileInfo.BlockIndex {
						//传完后删除
						ERR = db.ImProxyClient_DelSendFileList(*Node.GetNetId(), sendFileInfo)
						if ERR.CheckFail() {
							utils.Log.Error().Str("ERR", ERR.String()).Send()
							break
						}
						sendFileInfo = nil
					} else {
						//跟新上传进度
						ERR = db.ImProxyClient_SaveSendFileList(*Node.GetNetId(), sendFileInfo)
						if ERR.CheckFail() {
							utils.Log.Error().Str("ERR", ERR.String()).Send()
							break
						}
					}
				}
				continue
			}

			if sendTextTask == nil && sendFileInfo == nil {
				utils.Log.Info().Hex("循环发送群消息", this.GroupID).Send()
				select {
				case <-this.syncSendFileSignal:
					utils.Log.Info().Str("循环发送群消息", "").Send()
					//查询待发送的离线文件
					sendFileInfo, ERR = db.ImProxyClient_FindSendFileList(*Node.GetNetId(), *nodeStore.NewAddressNet(this.GroupID))
					if ERR.CheckFail() {
						utils.Log.Error().Str("ERR", ERR.String()).Send()
						break
					}
				case sendTextTask = <-this.syncSendTextSignal:
					utils.Log.Info().Str("循环发送群消息", "").Send()
				case <-this.ctx.Done():
					utils.Log.Info().Str("循环发送群消息", "").Send()
					return
				}
				utils.Log.Info().Hex("循环发送群消息", this.GroupID).Send()
			}

			if sendTextTask != nil {
				utils.Log.Info().Str("循环发送群消息", sendIndex.String()).Send()
				//设置sendIndex
				sendTextTask.proxyItr.GetBase().SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
				//发送给群
				ERR = SendGroupDataChain(this.addrKnitProxys, sendTextTask.proxyItr)
				if ERR.CheckFail() {
					utils.Log.Error().Msgf("发送群消息 错误:%s", ERR.String())
					if ERR.Code == config.ERROR_CODE_IM_datachain_sendIndex_discontinuity {
						//更新sendIndex，再发送一次
						sendIndex = this.GetSendIndex()
						//设置sendIndex
						sendTextTask.proxyItr.GetBase().SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
						//发送给群
						ERR = SendGroupDataChain(this.addrKnitProxys, sendTextTask.proxyItr)
						if ERR.CheckFail() {
							utils.Log.Error().Str("ERR", ERR.String()).Send()
							break
						}
					}
					if ERR.Code == config.ERROR_CODE_IM_index_discontinuity {
						//index不连续，则触发同步
						select {
						case this.syncDownloadSignal <- nil:
						default:
						}
						break
					}
				}
				if ERR.CheckSuccess() {
					sendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
				}
				utils.Log.Info().Str("循环发送群消息", "").Send()
				sendTextTask.errChan <- ERR
				continue
			}
		}
	}
}

/*
按顺序解析数据链
*/
func (this *GroupParser) ParseDataChain(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	this.parseLock.Lock()
	defer this.parseLock.Unlock()
	indexParse := this.GetIndexParse()
	index := proxyItr.GetIndex()
	utils.Log.Info().Msgf("parse 解析群信息:%s cmd:%d id:%+v", index.String(), proxyItr.GetCmd(), proxyItr.GetID())
	//只有刚刚创建的解析器不用验证
	if !(proxyItr.GetCmd() == config.IMPROXY_Command_server_group_create ||
		(proxyItr.GetCmd() == config.IMPROXY_Command_server_group_members && indexParse.Cmp(big.NewInt(0)) == 0)) {
		//utils.Log.Info().Msgf("groupParse 导入群数据链:%d", proxyItr.GetCmd())
		//判断群数据链index是否连续
		indexParse := new(big.Int).Add(indexParse, big.NewInt(1))
		//index := proxyItr.GetIndex()
		if !bytes.Equal(index.Bytes(), indexParse.Bytes()) {
			return utils.NewErrorBus(config.ERROR_CODE_IM_index_discontinuity, "")
		}
		//判断成员sendIndex
		sendIndex, ERR := db.ImProxyClient_FindGroupSendIndex(*config.DBKEY_improxy_group_datachain_sendIndex_parse,
			proxyItr.GetAddrFrom(), proxyItr.GetBase().GroupID)
		if !ERR.CheckSuccess() {
			return ERR
		}
		//utils.Log.Info().Msgf("groupParse 导入群数据链:%d", proxyItr.GetCmd())
		sendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
		if !bytes.Equal(proxyItr.GetBase().SendIndex.Bytes(), sendIndex.Bytes()) {
			return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_sendIndex_discontinuity, "")
		}
		//是不是群里的人发的消息
		_, ok := this.members[utils.Bytes2string(proxyItr.GetAddrFrom().GetAddr())]
		if !ok {
			return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
		}
		//禁言后只有管理员可以发言
		if this.shoutUp && !bytes.Equal(proxyItr.GetAddrFrom().GetAddr(), this.addrAdmin.GetAddr()) {
			return utils.NewErrorBus(config.ERROR_CODE_IM_group_shoutup, "")
		}
	}
	//utils.Log.Info().Msgf("groupParse 导入群数据链:%d", proxyItr.GetCmd())
	if proxyItr.GetBase().Content != nil && len(proxyItr.GetBase().Content) > 0 {
		shareKey := this.GetShareKey()
		//解密内容
		ERR := proxyItr.DecryptContent(shareKey)
		if !ERR.CheckSuccess() {
			return ERR
		}
		clientItrOne := proxyItr.GetClientItr()
		//for _, clientItrOne := range clientItrs {
		clientItrOne.SetProxyItr(proxyItr)
		ok := clientItrOne.CheckCmd()
		if !ok {
			return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_cmd_fail, "")
		}
	}

	//把数据链保存到数据库
	ERR := db.ImProxyClient_SaveGroupDataChain(*Node.GetNetId(), proxyItr, nil)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("错误:%s", ERR.String())
		return ERR
	}

	//}
	//utils.Log.Info().Msgf("groupParse 导入群数据链:%d", proxyItr.GetCmd())
	//
	//var ERR utils.ERROR
	this.Clone()
	defer func() {
		this.RollBACK(ERR)
	}()
	var mcVO *model.MessageContentVO
	batch := new(leveldb.Batch)
	switch proxyItr.GetCmd() {
	case config.IMPROXY_Command_server_group_create: //创建一个群
		createGroup := proxyItr.(*imdatachain.DataChainCreateGroup)
		mcVO, ERR = this.parseGroupCreate(createGroup, batch)
	case config.IMPROXY_Command_server_group_update: //修改群信息
		updateGroup := proxyItr.(*imdatachain.DataChainUpdateGroup)
		mcVO, ERR = this.parserGroupUpdate(updateGroup, batch)
	case config.IMPROXY_Command_server_group_members: //改变群成员群
		membersGroup := proxyItr.(*imdatachain.DataChainGroupMember)
		mcVO, ERR = this.parserGroupMembers(membersGroup, batch)
	case config.IMPROXY_Command_server_group_quit: //成员退出群聊
		quitGroup := proxyItr.(*imdatachain.ProxyGroupMemberQuit)
		mcVO, ERR = this.parserGroupQuit(quitGroup, batch)
	case config.IMPROXY_Command_server_group_dissolve: //解散群聊
		dissolveGroup := proxyItr.(*imdatachain.ProxyGroupDissolve)
		mcVO, ERR = this.parserGroupDissolve(dissolveGroup, batch)
	case config.IMPROXY_Command_server_forward: //转发群消息
		clientItr := proxyItr.GetClientItr()
		switch clientItr.GetClientCmd() {
		case config.IMPROXY_Command_client_group_sendText: //发送文本消息
			sendText := clientItr.(*imdatachain.DataChainGroupSendText)
			mcVO, ERR = this.parserGroupSendText(sendText, batch)
		case config.IMPROXY_Command_client_file: //发送文件
			//utils.Log.Info().Msgf("解析一条数据链消息:%d", itr.GetCmd())
			mcVO, ERR = this.sendGroupFile(clientItr, batch)
		default:
			utils.Log.Error().Msgf("未找到数据链中Client命令:%d", clientItr.GetClientCmd())
			ERR = utils.NewErrorBus(config.ERROR_CODE_IM_datachain_cmd_exist, "未找到数据链中Client命令:"+strconv.Itoa(clientItr.GetClientCmd()))
		}
	case config.IMPROXY_Command_server_group_msglog_add: //添加群消息
	case config.IMPROXY_Command_server_group_msglog_del: //删除群消息
	}
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("错误:%s", ERR.String())
		return ERR
	}

	//保存解析高度
	//proxyClientAddr := proxyItr.GetProxyClientAddr()
	groupId := proxyItr.GetBase().GroupID
	//index := proxyItr.GetIndex()
	//this.IndexParse = &index
	this.SetIndexParse(&index)

	//utils.Log.Info().Msgf("群发送者sendIndex:%+v %+v", sendIndex.Bytes(), proxyBase.SendIndex.Bytes())
	//保存发送者index
	ERR = db.ImProxyClient_SaveGroupSendIndex(*config.DBKEY_improxy_group_datachain_sendIndex_parse, proxyItr.GetAddrFrom(),
		groupId, proxyItr.GetBase().SendIndex.Bytes(), batch)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("错误:%s", ERR.String())
		return ERR
	}

	//更新自己发送到群消息的index
	if bytes.Equal(proxyItr.GetAddrFrom().GetAddr(), Node.GetNetId().GetAddr()) {
		this.SetSendIndex(proxyItr.GetBase().SendIndex)
	}

	//序列化快照
	bs, err := this.Proto()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		return ERR
	}
	//保存快照
	ERR = db.ImProxyClient_SaveGroupShot(*config.DBKEY_improxy_group_shot_index_parser, *Node.GetNetId(), groupId, *bs, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//utils.Log.Info().Msgf("parse保存群发送者的sendIndex:%+v", proxyItr.GetBase().SendIndex.Bytes())
	//一次性提交
	err = db.SaveBatch(batch)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		return ERR
	}

	if mcVO != nil {
		subscription.AddSubscriptionMsg(mcVO)
	}
	ERR = utils.NewErrorSuccess()
	return ERR
}

/*
创建群
*/
func (this *GroupParser) parseGroupCreate(createGroup *imdatachain.DataChainCreateGroup, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	//加入群列表
	ERR := db.ImProxyClient_SaveGroupList(*config.DBKEY_improxy_group_list_create, *Node.GetNetId(), createGroup, batch)
	//
	userInfo := make([]model.UserInfo, 0, 1)
	user := model.UserInfo{
		Addr:            createGroup.AddrFrom,
		Nickname:        "",
		RemarksName:     "",
		HeadNum:         0,
		Status:          0,
		Time:            0,
		CircleClass:     nil,
		Tray:            false,
		Proxy:           new(sync.Map),
		IsGroup:         true,                           //是否是群
		GroupId:         createGroup.GroupID,            //群ID
		AddrAdmin:       createGroup.AddrFrom,           //群管理员地址
		GroupAcceptTime: createGroup.MembersTime[0],     //同意入群时间
		GroupSign:       createGroup.MembersSign[0],     //群签名
		GroupSignPuk:    createGroup.MembersSignPuk[0],  //群签名用的公钥
		GroupShareKey:   createGroup.MembersShareKey[0], //群协商密码，用于群消息加解密
		GroupDHPuk:      createGroup.MembersDHPuk[0],    //协商密钥用公钥
		Admin:           true,
	}
	userInfo = append(userInfo, user)
	//把管理员保存到成员列表
	ERR = db.ImProxyClient_SaveGroupMembers(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), createGroup.GroupID, userInfo, batch)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	this.members[utils.Bytes2string(user.Addr.GetAddr())] = user.Addr

	//计算共享密钥，并解密群聊天密码
	Node.Keystore.GetDhAddrKeyPair(config.AddrPre)
	dhKeyPair, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dhPrk := dhKeyPair.PrivateKey
	var memberPuk [32]byte
	copy(memberPuk[:], createGroup.MembersDHPuk[0][:])
	//utils.Log.Info().Msgf("打印公钥:%+v 私钥:%+v", one.GroupDHPuk, prk)
	//生成共享密钥sharekey
	sharekey, err := keystore.KeyExchange(keystore.NewDHPair(dhPrk, memberPuk))
	if err != nil {
		utils.Log.Info().Msgf("生成共享密钥 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("加密共享密钥参数:%+v %+v", sharekey, randKey)
	plainText, err := utils.AesCTR_Decrypt(sharekey[:], nil, createGroup.MembersShareKey[0])
	if err != nil {
		utils.Log.Info().Msgf("用协商密钥加密 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}
	this.SetShareKey(plainText)
	//this.lock.Lock()
	//this.ShareKey = plainText
	//this.lock.Unlock()
	//给前端发送一个通知
	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
	//AddSubscriptionMsg(&msgInfo)
	return &msgInfo, utils.NewErrorSuccess()
}

/*
修改群信息
*/
func (this *GroupParser) parserGroupUpdate(updateGroup *imdatachain.DataChainUpdateGroup, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("parse 修改群信息:%+v", updateGroup)
	//改变了构建者
	if !bytes.Equal(this.GroupKnit.GetAddr(), updateGroup.ProxyMajor.GetAddr()) {
		this.GroupKnit = updateGroup.ProxyMajor
		this.groupOperationChan <- updateGroup
	}

	createGroup, ERR := db.FindGroupInfoAllList(*Node.GetNetId(), updateGroup.GroupID)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	createGroup.Nickname = updateGroup.Nickname
	createGroup.ShoutUp = updateGroup.ShoutUp
	createGroup.ProxyMajor = updateGroup.ProxyMajor
	if bytes.Equal(this.addrAdmin.GetAddr(), Node.GetNetId().GetAddr()) {
		ERR = db.ImProxyClient_SaveGroupList(*config.DBKEY_improxy_group_list_create, *Node.GetNetId(), createGroup, batch)
	} else {
		ERR = db.ImProxyClient_SaveGroupList(*config.DBKEY_improxy_group_list_join, *Node.GetNetId(), createGroup, batch)
	}
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//给前端发送一个通知
	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_group_members}
	//AddSubscriptionMsg(&msgInfo)
	return &msgInfo, utils.NewErrorSuccess()
}

/*
修改群成员
*/
func (this *GroupParser) parserGroupMembers(memberGroup *imdatachain.DataChainGroupMember, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {

	var subscription uint64 = config.SUBSCRIPTION_type_group_members
	//先在自己创建的群列表中查询
	ok, ERR := db.ImProxyClient_FindGroupListExist(*config.DBKEY_improxy_group_list_create, *Node.GetNetId(), memberGroup.GroupID)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	if !ok {
		//不在创建群列表，则查询加入的群列表
		ok, ERR = db.ImProxyClient_FindGroupListExist(*config.DBKEY_improxy_group_list_join, *Node.GetNetId(), memberGroup.GroupID)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		if !ok {
			//不在列表中，则保存到加入列表
			createGroup := &imdatachain.DataChainCreateGroup{DataChainGroupMember: *memberGroup}
			ERR = db.ImProxyClient_SaveGroupList(*config.DBKEY_improxy_group_list_join, *Node.GetNetId(), createGroup, batch)
			if !ERR.CheckSuccess() {
				return nil, ERR
			}
			subscription = config.SUBSCRIPTION_type_agreeFriend
		}
	}

	//查询旧有成员
	users, ERR := db.ImProxyClient_FindGroupMembers(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), this.GroupID)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//把旧有成员装进map
	oldUsersMap := make(map[string]*model.UserInfo)
	for _, one := range *users {
		oldUsersMap[utils.Bytes2string(one.Addr.GetAddr())] = &one
	}

	//过滤为现有成员
	newUsersMap := make(map[string]*model.UserInfo)
	for i, one := range memberGroup.MembersAddr {
		k := utils.Bytes2string(one.GetAddr())
		userinfo, ok := oldUsersMap[k]
		if !ok {
			//新添加的成员
			userinfo, ERR = db.ImProxyClient_FindUserinfo(*Node.GetNetId(), one)
			if !ERR.CheckSuccess() {
				return nil, ERR
			}
			if userinfo == nil {
				userinfo = model.NewUserInfo(one)
			}
		}
		//赋值新的共享密钥
		userinfo.GroupAcceptTime = memberGroup.MembersTime[i]
		userinfo.GroupSign = memberGroup.MembersSign[i]
		userinfo.GroupSignPuk = memberGroup.MembersDHPuk[i]
		userinfo.GroupDHPuk = memberGroup.MembersDHPuk[i]
		userinfo.GroupShareKey = memberGroup.MembersShareKey[i]
		newUsersMap[k] = userinfo
	}
	this.members = make(map[string]nodeStore.AddressNet)
	newMembers := make([]model.UserInfo, 0, len(newUsersMap))
	//整理成员以及成员的代理节点
	for k, one := range newUsersMap {
		newMembers = append(newMembers, *one)
		this.members[k] = one.Addr
	}
	//保存群成员
	ERR = db.ImProxyClient_SaveGroupMembers(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), memberGroup.GroupID, newMembers, batch)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}

	//需要重置所有成员的sendindex
	for _, one := range newMembers {
		ERR = db.ImProxyClient_SaveGroupSendIndex(*config.DBKEY_improxy_group_datachain_sendIndex_parse, one.Addr,
			memberGroup.GroupID, nil, batch)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
	}

	//计算共享密钥，并解密群聊天密码
	dhKeyPair, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	dhPrk := dhKeyPair.PrivateKey

	//utils.Log.Info().Msgf("打印自己的公钥:%s 私钥:%s", hex.EncodeToString(dhKeyPair.KeyPair.PublicKey[:]), hex.EncodeToString(dhPrk[:]))

	//查找自己的信息
	userSelf, ok := newUsersMap[utils.Bytes2string(Node.GetNetId().GetAddr())]
	if !ok {
		//自己被踢出群
		this.cancel()
		return nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
	}

	//查找管理员
	adminUser, ok := newUsersMap[utils.Bytes2string(memberGroup.AddrFrom.GetAddr())]
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
	}
	//取管理员的公钥
	var memberPuk [32]byte
	copy(memberPuk[:], adminUser.GroupDHPuk[:])

	//utils.Log.Info().Msgf("打印协商用公钥:%s 私钥:%s", hex.EncodeToString(memberPuk[:]), hex.EncodeToString(dhPrk[:]))

	//管理员的公钥，和自己的私钥，生成共享密钥sharekey
	sharekey, err := keystore.KeyExchange(keystore.NewDHPair(dhPrk, memberPuk))
	if err != nil {
		utils.Log.Info().Msgf("生成共享密钥 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}

	//utils.Log.Info().Msgf("本节点与管理员协商密钥是:%s 解密前的群密码是:%s", hex.EncodeToString(sharekey[:]), hex.EncodeToString(userSelf.GroupShareKey))

	plainText, err := utils.AesCTR_Decrypt(sharekey[:], nil, userSelf.GroupShareKey)
	if err != nil {
		//utils.Log.Info().Msgf("用协商密钥加密 错误:%s", err.Error())
		return nil, utils.NewErrorSysSelf(err)
	}

	//utils.Log.Info().Msgf("解密后的群密码是:%s", hex.EncodeToString(plainText))
	this.SetShareKey(plainText)
	//this.lock.Lock()
	//this.ShareKey = plainText
	//this.lock.Unlock()

	//给前端发送一个通知
	msgInfo := model.MessageContentVO{Subscription: subscription}
	return &msgInfo, utils.NewErrorSuccess()
}

/*
群成员退出群聊
*/
func (this *GroupParser) parserGroupQuit(quitGroup *imdatachain.ProxyGroupMemberQuit, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("退出群聊")
	var msgInfo *model.MessageContentVO
	var ERR utils.ERROR
	if bytes.Equal(Node.GetNetId().GetAddr(), quitGroup.AddrFrom.GetAddr()) {
		//自己退出，则删除群列表
		ERR = db.ImProxyClient_RemoveGroupList(*config.DBKEY_improxy_group_list_join, *Node.GetNetId(), quitGroup.GroupID, batch)
		//给前端发送一个通知
		msgInfo = &model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
	} else {
		//别人退出，则删除群成员
		ERR = db.ImProxyClient_RemoveGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), quitGroup.GroupID, quitGroup.AddrFrom, batch)
		//给前端发送一个通知
		msgInfo = &model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_group_members}
	}
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return msgInfo, utils.NewErrorSuccess()
}

/*
解散群聊
*/
func (this *GroupParser) parserGroupDissolve(dissolveGroup *imdatachain.ProxyGroupDissolve, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("解散群聊")
	//群成员收到解散，从加入的群列表中删除
	ERR := db.ImProxyClient_RemoveGroupList(*config.DBKEY_improxy_group_list_join, *Node.GetNetId(), dissolveGroup.GroupID, batch)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}

	//管理员收到解散，从创建的群列表中删除
	ERR = db.ImProxyClient_RemoveGroupList(*config.DBKEY_improxy_group_list_create, *Node.GetNetId(), dissolveGroup.GroupID, batch)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	//给前端发送一个通知
	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
	return &msgInfo, utils.NewErrorSuccess()
}

/*
群文本消息
*/
func (this *GroupParser) parserGroupSendText(sendText *imdatachain.DataChainGroupSendText, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("解析并保存文本消息记录:%d", sendText.GetClientCmd())
	//clientBase := clientItr.(*imdatachain.DataChainSendText)
	if bytes.Equal(Node.GetNetId().GetAddr(), sendText.AddrFrom.GetAddr()) {
		utils.Log.Info().Msgf("解析并保存文本消息记录:%d", sendText.GetClientCmd())
		//id := ulid.Make()
		//utils.Log.Info().Msgf("给好友发送消息：%s", content)
		messageOne := &model.MessageContent{
			Type:       config.MSG_type_text,                                               //消息类型
			FromIsSelf: true,                                                               //是否自己发出的
			From:       sendText.AddrFrom,                                                  //发送者
			To:         *nodeStore.NewAddressNet(sendText.GetProxyItr().GetBase().GroupID), //接收者
			Content:    sendText.Content,                                                   //消息内容
			Time:       time.Now().Unix(),                                                  //时间
			SendID:     sendText.GetProxyItr().GetID(),                                     //
			QuoteID:    sendText.QuoteID,                                                   //
			State:      config.MSG_GUI_state_success,                                       //
			//PullAndPushID: pullAndPushID,                 //
			IsGroup: true, //
		}
		//utils.Log.Info().Msgf("发送消息：%+v", messageOne)
		//保存未发送状态的消息
		messageOne, ERR := db.AddMessageHistoryV2(*Node.GetNetId(), messageOne, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
			return nil, ERR
		}
		utils.Log.Info().Msgf("解析并保存文本消息记录:%d", sendText.GetClientCmd())
		msgVO := messageOne.ConverVO()
		msgVO.Subscription = config.SUBSCRIPTION_type_msg
		msgVO.State = config.MSG_GUI_state_success
		//AddSubscriptionMsg(msgVO)
		return msgVO, utils.NewErrorSuccess()
	} else {
		utils.Log.Info().Msgf("解析并保存文本消息记录:%d", sendText.GetClientCmd())
		//自己是接收者
		//保存聊天记录
		messageOne := &model.MessageContent{
			Type:       config.MSG_type_text,                                               //消息类型
			FromIsSelf: false,                                                              //是否自己发出的
			From:       sendText.AddrFrom,                                                  //发送者
			To:         *nodeStore.NewAddressNet(sendText.GetProxyItr().GetBase().GroupID), //接收者
			Content:    sendText.Content,                                                   //消息内容
			Time:       time.Now().Unix(),                                                  //时间
			SendID:     sendText.GetProxyItr().GetID(),                                     //消息唯一ID
			RecvID:     ulid.Make().Bytes(),                                                //
			QuoteID:    sendText.QuoteID,                                                   //
			State:      config.MSG_GUI_state_success,                                       //
			IsGroup:    true,                                                               //
		}
		messageOne, ERR := db.AddMessageHistoryV2(*Node.GetNetId(), messageOne, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("添加消息记录 错误:%s", ERR.String())
			return nil, ERR
		}
		//给前端发送一个通知
		msgInfo := messageOne.ConverVO()
		msgInfo.Subscription = config.SUBSCRIPTION_type_msg
		//AddSubscriptionMsg(msgInfo)
		utils.Log.Info().Msgf("解析并保存文本消息记录:%d", sendText.GetClientCmd())
		return msgInfo, utils.NewErrorSuccess()
	}
}

/*
发送文件
*/
func (this *GroupParser) sendGroupFile(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("解析并保存文件")
	dataChainFile := clientItr.(*imdatachain.DatachainFile)
	//utils.Log.Info().Msgf("解析并保存文件:%+v %+v", dataChainFile.Hash, dataChainFile.SendTime)
	//index := clientItr.GetProxyItr().GetIndex()
	//addrRemote := clientItr.GetAddrFrom()
	fromIsSelf := false
	if bytes.Equal(Node.GetNetId().GetAddr(), dataChainFile.GetAddrFrom().GetAddr()) {
		fromIsSelf = true
		//addrRemote = clientItr.GetAddrTo()
	}

	//utils.Log.Info().Msgf("创建群或者修改群信息")

	//是一个新文件，则保存新记录
	if dataChainFile.BlockIndex == 0 {
		utils.Log.Info().Msgf("解析并保存文件")
		var msgContent *model.MessageContent
		//是base64编码图片
		//if dataChainFile.Type == config.FILE_type_image_base64 {
		if dataChainFile.MimeType[:5] == "image" && dataChainFile.Size <= uint64(config.FILE_image_size_max) {
			msgContent = model.NewMsgContentImgBase64(fromIsSelf, dataChainFile.GetAddrFrom(), dataChainFile.GetAddrTo(),
				time.Now().Unix(), dataChainFile.SendTime, dataChainFile.GetProxyItr().GetID(), dataChainFile.Name,
				dataChainFile.MimeType, dataChainFile.Size, dataChainFile.Hash, dataChainFile.BlockTotal,
				dataChainFile.BlockIndex, [][]byte{dataChainFile.GetProxyItr().GetID()})
		} else {
			msgContent = model.NewMsgContentFile(fromIsSelf, dataChainFile.GetAddrFrom(), dataChainFile.GetAddrTo(),
				time.Now().Unix(), dataChainFile.SendTime, dataChainFile.GetProxyItr().GetID(), dataChainFile.Name,
				dataChainFile.MimeType, dataChainFile.Size, dataChainFile.Hash, dataChainFile.BlockTotal,
				dataChainFile.BlockIndex, [][]byte{dataChainFile.GetProxyItr().GetID()})
		}
		msgContent.IsGroup = true
		msgContent.State = config.MSG_GUI_state_not_send
		//utils.Log.Info().Msgf("发送消息：%+v", messageOne)
		//保存未发送状态的消息
		messageOne, ERR := db.AddMessageHistoryV2(*Node.GetNetId(), msgContent, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("解析并保存文件id:%+v %+v %+v", dataChainFile.GetProxyItr().GetID(), messageOne.FileContent, dataChainFile.Hash)
		utils.Log.Info().Msgf("解析并保存文件")
		//判断图片文件是否已经传完
		//图片传完的，拼接起来
		ERR = JoinImageBase64(messageOne)
		if ERR.CheckFail() {
			return nil, ERR
		}

		utils.Log.Info().Msgf("解析并保存文件")
		msgVO := messageOne.ConverVO()
		msgVO.Subscription = config.SUBSCRIPTION_type_msg
		msgVO.State = config.MSG_GUI_state_not_send
		return msgVO, utils.NewErrorSuccess()
	}
	utils.Log.Info().Msgf("解析并保存文件")
	//是文件续传
	key := config.DBKEY_improxy_user_message_history_fileHash_recv
	if fromIsSelf {
		key = config.DBKEY_improxy_user_message_history_fileHash_send
	}
	//utils.Log.Info().Msgf("保存记录:%d %d %+v", dataChainFile.BlockIndex, dataChainFile.BlockTotal, dataChainFile.Hash)
	msgContentOld, ERR := db.FindMessageHistoryByFileHash(*key, *Node.GetNetId(),
		*nodeStore.NewAddressNet(clientItr.GetProxyItr().GetBase().GroupID), dataChainFile.Hash, dataChainFile.SendTime)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
		return nil, ERR
	}
	msgContentOld.FileContent = append(msgContentOld.FileContent, dataChainFile.GetProxyItr().GetID())
	msgContentOld.FileBlockIndex = dataChainFile.BlockIndex
	//保存传送进度到数据库
	msgContentOld.TransProgress = int(float64(msgContentOld.FileBlockIndex+1) / float64(msgContentOld.FileBlockTotal) * 100)

	ERR = db.UpdateMessageHistoryByFileHash(*Node.GetNetId(), msgContentOld, batch)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
		return nil, ERR
	}
	//判断图片文件是否已经传完
	//if msgContentOld.FileType == config.FILE_type_image_base64 && msgContentOld.FileBlockTotal == msgContentOld.FileBlockIndex+1 {
	//图片传完的，拼接起来
	ERR = JoinImageBase64(msgContentOld)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//}
	utils.Log.Info().Msgf("解析并保存文件")
	msgInfo := msgContentOld.ConverVO()
	msgInfo.Subscription = config.SUBSCRIPTION_type_msg
	//给前端发送一个通知
	//msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_msg}
	//AddSubscriptionMsg(&msgInfo)
	return msgInfo, utils.NewErrorSuccess()
}

/*
回滚
*/
func (this *GroupParser) Proto() (*[]byte, error) {
	base := go_protos.GroupShot{
		GroupID:        this.GroupID,
		GroupKnit:      this.GroupKnit.GetAddr(),
		SendIndex:      nil,
		PreHash:        this.PreHash,
		IndexParse:     nil,
		AddrAdmin:      this.addrAdmin.GetAddr(),
		AddrKnitProxys: this.addrKnitProxys.GetAddr(),
		Members:        make([][]byte, 0, len(this.members)),
		ShoutUp:        this.shoutUp,
		ShareKey:       this.ShareKey,
	}
	base.SendIndex = this.GetSendIndex().Bytes()
	base.IndexParse = this.GetIndexParse().Bytes()
	for _, member := range this.members {
		base.Members = append(base.Members, member.GetAddr())
	}
	bs, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func ParseGroupParser(bs []byte) (*GroupParser, error) {
	groupShot := go_protos.GroupShot{}
	err := proto.Unmarshal(bs, &groupShot)
	if err != nil {
		return nil, err
	}

	group := NewGroupParserByGroupMember(nil, nil)
	group.GroupID = groupShot.GroupID
	group.GroupKnit = *nodeStore.NewAddressNet(groupShot.GroupKnit)
	group.SendIndex = new(big.Int).SetBytes(groupShot.SendIndex)
	group.PreHash = groupShot.PreHash
	group.IndexParse = new(big.Int).SetBytes(groupShot.IndexParse)
	group.addrAdmin = *nodeStore.NewAddressNet(groupShot.AddrAdmin)
	group.addrKnitProxys = *nodeStore.NewAddressNet(groupShot.AddrKnitProxys)
	group.shoutUp = groupShot.ShoutUp
	group.ShareKey = groupShot.ShareKey
	for _, one := range groupShot.Members {
		group.members[utils.Bytes2string(one)] = *nodeStore.NewAddressNet(one)
	}
	return group, nil
}

type SendTextTask struct {
	errChan  chan utils.ERROR
	proxyItr imdatachain.DataChainProxyItr
}

func NewSendTextTask(proxyItr imdatachain.DataChainProxyItr) *SendTextTask {
	sft := SendTextTask{
		errChan:  make(chan utils.ERROR, 1),
		proxyItr: proxyItr,
	}
	return &sft
}

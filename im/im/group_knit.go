package im

import (
	"bytes"
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type GroupKnitManager struct {
	lock   *sync.Mutex //
	groups *sync.Map   //保存自己管理的群
}

func NewGroupKnitManager() (*GroupKnitManager, utils.ERROR) {
	gkm := GroupKnitManager{
		lock:   new(sync.Mutex),
		groups: new(sync.Map),
	}
	ERR := gkm.LoadDB()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	go gkm.LoopCleanGroup()
	return &gkm, utils.NewErrorSuccess()
}

func (this *GroupKnitManager) LoadDB() utils.ERROR {
	//查询自己创建的群列表
	knitGroups, ERR := db.ImProxyClient_FindGroupList(*config.DBKEY_improxy_group_list_knit, *Node.GetNetId())
	if !ERR.CheckSuccess() {
		return ERR
	}

	//加载快照
	for _, group := range *knitGroups {
		bs, ERR := db.ImProxyClient_LoadGroupShot(*config.DBKEY_improxy_group_shot_index_knit, *Node.GetNetId(), group.GroupID)
		if !ERR.CheckSuccess() {
			return ERR
		}
		gk, err := ParseGroupKnit(bs)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		this.groups.Store(utils.Bytes2string(group.GroupID), gk)
	}
	return utils.NewErrorSuccess()
}

/*
循环清理被解散的群
*/
func (this *GroupKnitManager) LoopCleanGroup() {
	for range time.NewTicker(time.Hour).C {
		items, ERR := db.ImProxyClient_FindDissolveGroupList(*Node.GetNetId())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询解散群列表 错误:%s", ERR.String())
			continue
		}
		batch := new(leveldb.Batch)
		for _, one := range items {
			//检查解散的群是否超期
			timeUnix := utils.BytesToUint64ByBigEndian(one.Value)
			dissolveTime := time.Unix(int64(timeUnix), 0)
			if dissolveTime.Add(config.DissolveGroupOverTime).After(time.Now()) {
				//未超时
				continue
			}
			groupId, ERR := one.Key.BaseKey()
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("解析群ID 错误:%s", ERR.String())
				continue
			}
			//超时了，清理数据
			ERR = db.ImProxyClient_RemoveDissolveGroupList(*Node.GetNetId(), groupId, batch)
		}
		err := db.SaveBatch(batch)
		if err != nil {
			utils.Log.Error().Msgf("清理群批量保存 错误:%s", err.Error())
		}
	}
}

func (this *GroupKnitManager) KnitDataChain(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	proxyItr.GetBase().AddrProxyServer = *Node.GetNetId()
	utils.Log.Info().Msgf("knit 构建一个群消息:%d", proxyItr.GetCmd())
	//创建一个群
	if proxyItr.GetCmd() == config.IMPROXY_Command_server_group_create {
		createGroup := proxyItr.(*imdatachain.DataChainCreateGroup)
		//指定构建节点地址不是自己
		if !bytes.Equal(Node.GetNetId().GetAddr(), createGroup.ProxyMajor.GetAddr()) {
			return utils.NewErrorSuccess()
		}
		this.lock.Lock()
		defer this.lock.Unlock()
		_, ok := this.groups.Load(utils.Bytes2string(createGroup.GroupID))
		if ok {
			return utils.NewErrorBus(config.ERROR_CODE_IM_group_exist, "")
		}
		pg := NewGroupKnit(createGroup.GroupID, big.NewInt(0), createGroup.Hash, createGroup.AddrFrom, nil, createGroup.ShoutUp)
		ERR := pg.KnitDataChain(proxyItr)
		if ERR.CheckSuccess() {
			utils.Log.Info().Msgf("创建群成功")
			this.groups.Store(utils.Bytes2string(createGroup.GroupID), pg)
			return utils.NewErrorSuccess()
		}
		utils.Log.Error().Msgf("创建群 错误:%s", ERR.String())
		return ERR
	} else if proxyItr.GetCmd() == config.IMPROXY_Command_server_group_update {
		//utils.Log.Info().Msgf("修改群:%+v", proxyItr)
		updateGroup := proxyItr.(*imdatachain.DataChainUpdateGroup)
		//指定构建节点地址不是自己
		if !bytes.Equal(Node.GetNetId().GetAddr(), updateGroup.ProxyMajor.GetAddr()) {
			//utils.Log.Info().Msgf("已经存在，则删除")
			//如果存在，则删除
			this.groups.Delete(utils.Bytes2string(updateGroup.GroupID))
			return utils.NewErrorSuccess()
		}
		this.lock.Lock()
		defer this.lock.Unlock()
		//查询是否自己已经在构建这个群的数据链
		var pg *GroupKnit
		pgItr, ok := this.groups.Load(utils.Bytes2string(updateGroup.GroupID))
		if ok {
			//utils.Log.Info().Msgf("已经存在")
			pg = pgItr.(*GroupKnit)
		} else {
			//utils.Log.Info().Msgf("创建新的")
			pg = NewGroupKnit(updateGroup.GroupID, big.NewInt(0), updateGroup.Hash, updateGroup.AddrFrom, nil, updateGroup.ShoutUp)
		}
		//utils.Log.Info().Msgf("开始解析")
		ERR := pg.KnitDataChain(proxyItr)
		if ERR.CheckSuccess() {
			if !ok {
				this.groups.Store(utils.Bytes2string(updateGroup.GroupID), pg)
			}
		}
		return ERR
	} else {
		//_, _, ERR := GroupAuthMember(config.DBKEY_improxy_group_list_knit, proxyItr.GetBase().GroupID, proxyItr.GetAddrFrom())
		//if !ERR.CheckSuccess() {
		//	utils.Log.Info().Msgf("群权限验证 失败:%s", ERR.String())
		//	return ERR
		//}
		utils.Log.Info().Msgf("knit 构建一个群消息:%d", proxyItr.GetCmd())
		//
		groupId := proxyItr.GetBase().GroupID
		pgItr, ok := this.groups.Load(utils.Bytes2string(groupId))
		if !ok {
			utils.Log.Info().Msgf("knit 不是自己负责构建的群:%d", proxyItr.GetCmd())
			//不是自己负责构建链的群
			return utils.NewErrorSuccess()
		}
		pg := pgItr.(*GroupKnit)
		ERR := pg.KnitDataChain(proxyItr)
		return ERR
	}
}

type GroupKnit struct {
	GroupID        []byte                             //群id
	lock           *sync.RWMutex                      //锁
	Index          *big.Int                           //必须连续的自增长ID
	PreHash        []byte                             //
	uploadChan     chan imdatachain.DataChainProxyItr //
	sendChan       chan imdatachain.DataChainProxyItr //发送队列
	indexParseLock *sync.RWMutex                      //锁
	//IndexParse         *big.Int                           //本地解析到的数据链高度
	parseSignal        chan imdatachain.DataChainProxyItr //新消息需要解析的信号
	addrAdmin          nodeStore.AddressNet               //管理员地址
	members            map[string]nodeStore.AddressNet    //群成员，包括管理员
	membersProxys      map[string]nodeStore.AddressNet    //群成员的地址，或者是群成员的代理节点的地址，包括群主的代理节点地址
	shoutUp            bool                               //是否禁言
	multicastCacheLock *sync.Mutex                        //
	multicastCache     []GroupMulticastMembers            //群内累积的消息广播
	multicastChan      chan bool                          //新的群消息信号
	cancel             context.CancelFunc                 //
	ctx                context.Context                    //
	clone              *GroupKnit                         //克隆，方便回滚
}

func NewGroupKnit(groupId []byte, index *big.Int, preHash []byte, addrAdmin nodeStore.AddressNet,
	members map[string]nodeStore.AddressNet, shoutUp bool) *GroupKnit {
	ctx, cancel := context.WithCancel(context.Background())
	pg := GroupKnit{
		GroupID:        groupId,
		lock:           new(sync.RWMutex),
		Index:          index,
		PreHash:        preHash,
		uploadChan:     nil,
		sendChan:       make(chan imdatachain.DataChainProxyItr, 10000), //
		indexParseLock: nil,
		//IndexParse:         nil,
		parseSignal:        nil,
		addrAdmin:          addrAdmin,
		members:            members,
		membersProxys:      make(map[string]nodeStore.AddressNet),
		shoutUp:            shoutUp,
		multicastCacheLock: new(sync.Mutex),
		multicastCache:     make([]GroupMulticastMembers, 0),
		multicastChan:      make(chan bool, 1),
		cancel:             cancel,
		ctx:                ctx,
	}
	go pg.loopMulticastDataChain()
	return &pg
}

/*
克隆，方便回滚
*/
func (this *GroupKnit) Clone() {
	gk := GroupKnit{
		//Index:              new(big.Int).SetBytes(this.Index.Bytes()),
		PreHash: this.PreHash,
		//IndexParse:         new(big.Int).SetBytes(this.IndexParse.Bytes()),
		members:       make(map[string]nodeStore.AddressNet),
		membersProxys: make(map[string]nodeStore.AddressNet),
		shoutUp:       this.shoutUp,
	}
	if this.Index != nil {
		gk.Index = new(big.Int).SetBytes(this.Index.Bytes())
	}
	//if this.IndexParse != nil {
	//	gk.IndexParse = new(big.Int).SetBytes(this.IndexParse.Bytes())
	//}
	for k, v := range this.members {
		gk.members[k] = v
	}
	for k, v := range this.membersProxys {
		gk.membersProxys[k] = v
	}
	this.clone = &gk
	return
}

/*
回滚
*/
func (this *GroupKnit) RollBACK(ERR utils.ERROR) {
	if ERR.CheckSuccess() {
		return
	}
	utils.Log.Error().Msgf("group knit 回滚 RollBACK")
	this.Index = this.clone.Index
	this.PreHash = this.clone.PreHash
	//this.IndexParse = this.clone.IndexParse
	this.members = this.clone.members
	this.membersProxys = this.clone.membersProxys
	this.shoutUp = this.clone.shoutUp
}

/*
按顺序构建数据链
*/
func (this *GroupKnit) KnitDataChain(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {

	utils.Log.Info().Msgf("knit 构建群数据链:%d", proxyItr.GetCmd())
	this.lock.Lock()
	defer this.lock.Unlock()

	proxyBase := proxyItr.GetBase()
	addrFrom := proxyItr.GetAddrFrom()
	groupId := proxyBase.GroupID
	if groupId == nil || len(groupId) == 0 {
		utils.Log.Info().Msgf("参数错误 GroupID")
		return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_params_fail, "GroupID")
	}
	//utils.Log.Info().Msgf("knit 构建群数据链:%d", proxyItr.GetCmd())
	//查询这条消息是否重复
	if proxyItr.GetSendID() != nil && len(proxyItr.GetSendID()) != 0 {
		//utils.Log.Info().Msgf("保存数据链")
		ok, ERR := db.ImProxyClient_FindGroupDataChainBySendID(*Node.GetNetId(), groupId, proxyItr.GetSendID())
		if !ERR.CheckSuccess() {
			return ERR
		}
		//utils.Log.Info().Msgf("保存数据链")
		//已经存在
		if ok {
			return utils.NewErrorSuccess()
		}
	}
	//utils.Log.Info().Msgf("knit 构建群数据链:%d", proxyItr.GetCmd())
	//只有创建命令，不验证群成员
	if proxyItr.GetCmd() == config.IMPROXY_Command_server_group_create {
	} else if proxyItr.GetCmd() == config.IMPROXY_Command_server_group_update {
		//是不是管理员
		if !bytes.Equal(this.addrAdmin.GetAddr(), proxyItr.GetAddrFrom().GetAddr()) {
			return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_admin, "")
		}
	} else {
		//是不是群成员
		_, ok := this.members[utils.Bytes2string(addrFrom.GetAddr())]
		if !ok {
			return utils.NewErrorBus(config.ERROR_CODE_IM_group_not_member, "")
		}
	}
	//utils.Log.Info().Msgf("knit 构建群数据链:%d", proxyItr.GetCmd())
	//验证发送者的sendIndex是否连续
	sendIndex, ERR := db.ImProxyClient_FindGroupSendIndex(*config.DBKEY_improxy_group_datachain_sendIndex_knit, addrFrom, groupId)
	if !ERR.CheckSuccess() {
		return ERR
	}
	utils.Log.Info().Str("addr", addrFrom.B58String()).Hex("gid", groupId).Msgf("群发送者sendIndex:%+v %+v",
		sendIndex.Bytes(), proxyBase.SendIndex.Bytes())
	sendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
	if !bytes.Equal(proxyBase.SendIndex.Bytes(), sendIndex.Bytes()) {
		return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_sendIndex_discontinuity, "")
	}
	//禁言后，只有管理员能发消息
	if this.shoutUp && !bytes.Equal(addrFrom.GetAddr(), this.addrAdmin.GetAddr()) {
		return utils.NewErrorBus(config.ERROR_CODE_IM_group_shoutup, "")
	}
	//utils.Log.Info().Msgf("knit 构建群数据链:%d", proxyItr.GetCmd())
	newIndex := new(big.Int).Add(this.Index, big.NewInt(1))
	proxyItr.SetIndex(*newIndex)
	proxyItr.SetPreHash(this.PreHash)
	proxyItr.BuildHash()

	gmm := NewGroupMulticastMembers(proxyItr, this.membersProxys)
	this.Clone()
	defer func() {
		this.RollBACK(ERR)
	}()
	batch := new(leveldb.Batch)
	this.Index = newIndex
	this.PreHash = proxyItr.GetHash()

	ERR = this.parseProxyDataChain(proxyItr, batch)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("解析 错误:%s", ERR.String())
		return ERR
	}
	//把数据链保存到数据库
	ERR = db.ImProxyClient_SaveGroupDataChain(*Node.GetNetId(), proxyItr, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存解析高度
	//proxyClientAddr := proxyItr.GetProxyClientAddr()
	//index := proxyItr.GetBase().Index

	//序列化快照
	bs, err := this.Proto()
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		return ERR
	}
	//保存快照
	ERR = db.ImProxyClient_SaveGroupShot(*config.DBKEY_improxy_group_shot_index_knit, *Node.GetNetId(), groupId, *bs, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//保存发送者index
	ERR = db.ImProxyClient_SaveGroupSendIndex(*config.DBKEY_improxy_group_datachain_sendIndex_knit, addrFrom,
		groupId, proxyBase.SendIndex.Bytes(), batch)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//utils.Log.Info().Msgf("knit 构建群数据链")
	err = db.SaveBatch(batch)
	if err != nil {
		utils.Log.Info().Msgf("保存解析结果 错误:%s", err.Error())
		ERR = utils.NewErrorSysSelf(err)
		return ERR
	}
	index := proxyItr.GetIndex()
	utils.Log.Info().Msgf("knit 构建群数据链:%s %d", index.String(), proxyItr.GetCmd())

	//utils.Log.Info().Msgf("保存解析结果")
	gmm.MergeMembers(this.membersProxys)
	this.multicastCacheLock.Lock()
	this.multicastCache = append(this.multicastCache, *gmm)
	select {
	case this.multicastChan <- false:
	default:
	}
	this.multicastCacheLock.Unlock()
	ERR = utils.NewErrorSuccess()
	return ERR
}

/*
直接发送给成员或者成员的代理节点
*/
func (this *GroupKnit) loopMulticastDataChain() {
	for {
		select {
		case <-this.multicastChan:
		case <-this.ctx.Done():
			//销毁了，也要检查消息是否发送出去了
			this.multicastCacheLock.Lock()
			list := this.multicastCache
			this.multicastCacheLock.Unlock()
			if len(list) == 0 {
				//发送完了，就可以退出，自动销毁协程
				return
			}
		}
		this.multicastCacheLock.Lock()
		//取出带发送消息缓存，然后清空缓存
		list := this.multicastCache
		this.multicastCache = make([]GroupMulticastMembers, 0)
		select {
		case <-this.multicastChan:
		default:
		}
		this.multicastCacheLock.Unlock()
		for _, m := range list {
			utils.Log.Info().Msgf("本次群聊，要发送给的成员")
			for _, addrOne := range m.members {
				MulticastGroupDataChain(addrOne, []imdatachain.DataChainProxyItr{m.msg})
			}
		}
	}
}

func (this *GroupKnit) parseProxyDataChain(proxyItr imdatachain.DataChainProxyItr, batch *leveldb.Batch) utils.ERROR {
	index := proxyItr.GetIndex()
	utils.Log.Info().Msgf("knit 解析群信息:%s %d", index.String(), proxyItr.GetCmd())
	ERR := utils.NewErrorSuccess()
	switch proxyItr.GetCmd() {
	case config.IMPROXY_Command_server_group_create: //创建一个群
		createGroup := proxyItr.(*imdatachain.DataChainCreateGroup)
		ERR = this.knitGroupCreate(createGroup, batch)
	case config.IMPROXY_Command_server_group_update: //修改群信息
		memberGroup := proxyItr.(*imdatachain.DataChainUpdateGroup)
		ERR = this.knitGroupUpdate(memberGroup, batch)
	case config.IMPROXY_Command_server_group_members: //改变群成员群
		memberGroup := proxyItr.(*imdatachain.DataChainGroupMember)
		ERR = this.knitGroupMembers(memberGroup, batch)
	case config.IMPROXY_Command_server_group_quit: //成员退出群聊
		quitGroup := proxyItr.(*imdatachain.ProxyGroupMemberQuit)
		ERR = this.knitGroupQuit(quitGroup, batch)
	case config.IMPROXY_Command_server_group_dissolve: //解散群聊
		dissolveGroup := proxyItr.(*imdatachain.ProxyGroupDissolve)
		ERR = db.ImProxyClient_RemoveGroupList(*config.DBKEY_improxy_group_list_knit, *Node.GetNetId(), dissolveGroup.GroupID, batch)
		if !ERR.CheckSuccess() {
			return ERR
		}
		ERR = db.ImProxyClient_SaveDissolveGroupList(*Node.GetNetId(), dissolveGroup.GroupID, batch)
	case config.IMPROXY_Command_server_forward: //转发群消息
	case config.IMPROXY_Command_server_group_msglog_add: //添加群消息
	case config.IMPROXY_Command_server_group_msglog_del: //删除群消息
	}
	return ERR
}

/*
创建群
*/
func (this *GroupKnit) knitGroupCreate(createGroup *imdatachain.DataChainCreateGroup, batch *leveldb.Batch) utils.ERROR {
	//加入群列表
	ERR := db.ImProxyClient_SaveGroupList(*config.DBKEY_improxy_group_list_knit, *Node.GetNetId(), createGroup, batch)
	//
	userInfo := make([]model.UserInfo, 0, 1)
	user := model.NewUserInfo(createGroup.AddrFrom)
	//查询用户信息
	if bytes.Equal(Node.GetNetId().GetAddr(), createGroup.AddrFrom.GetAddr()) {
		userinfo, ERR := db.GetSelfInfo(*Node.GetNetId())
		if !ERR.CheckSuccess() {
			return ERR
		}
		user.Proxy = userinfo.Proxy
	} else {
		userinfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), createGroup.AddrFrom)
		if !ERR.CheckSuccess() {
			return ERR
		}
		if userinfo != nil {
			//设置节点的代理节点信息
			user.Proxy = userinfo.Proxy
		}
	}
	proxys := make(map[string]nodeStore.AddressNet)
	user.Proxy.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		addr := value.(nodeStore.AddressNet)
		proxys[keyStr] = addr
		return true
	})
	//如果没有代理节点，则要通知它本人
	if len(proxys) == 0 {
		proxys[utils.Bytes2string(createGroup.AddrFrom.GetAddr())] = createGroup.AddrFrom
	}
	this.membersProxys = proxys
	userInfo = append(userInfo, *user)
	//把管理员保存到成员列表
	ERR = db.ImProxyClient_SaveGroupMembers(*config.DBKEY_improxy_group_members_knit, *Node.GetNetId(), createGroup.GroupID, userInfo, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	this.addrAdmin = createGroup.AddrFrom
	this.members = map[string]nodeStore.AddressNet{utils.Bytes2string(createGroup.AddrFrom.GetAddr()): createGroup.AddrFrom}
	return utils.NewErrorSuccess()
}

/*
修改群信息
*/
func (this *GroupKnit) knitGroupUpdate(updateGroup *imdatachain.DataChainUpdateGroup, batch *leveldb.Batch) utils.ERROR {
	utils.Log.Info().Msgf("修改群信息")
	createGroup, ERR := db.ImProxyClient_FindGroupInfo(*config.DBKEY_improxy_group_list_knit, *Node.GetNetId(), updateGroup.GroupID)
	if !ERR.CheckSuccess() {
		return ERR
	}
	createGroup.Nickname = updateGroup.Nickname
	createGroup.ShoutUp = updateGroup.ShoutUp
	createGroup.ProxyMajor = updateGroup.ProxyMajor
	ERR = db.ImProxyClient_SaveGroupList(*config.DBKEY_improxy_group_list_knit, *Node.GetNetId(), createGroup, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if this.members == nil || len(this.members) == 0 {
		//新创建的，查询并构建成员地址
		users, ERR := db.ImProxyClient_FindGroupMembers(*config.DBKEY_improxy_group_members_knit, *Node.GetNetId(), this.GroupID)
		if !ERR.CheckSuccess() {
			return ERR
		}
		for _, one := range *users {
			proxys := make(map[string]nodeStore.AddressNet)
			one.Proxy.Range(func(key, value interface{}) bool {
				keyStr := key.(string)
				addr := value.(nodeStore.AddressNet)
				proxys[keyStr] = addr
				return true
			})
			if len(proxys) == 0 {
				//没有代理就保存它本人
				this.membersProxys[utils.Bytes2string(one.Addr.GetAddr())] = one.Addr
			} else {
				//有代理，则保存代理
				for key, value := range proxys {
					this.membersProxys[key] = value
				}
			}
		}
	} else {
		this.shoutUp = updateGroup.ShoutUp
	}
	return utils.NewErrorSuccess()
}

/*
修改群成员
*/
func (this *GroupKnit) knitGroupMembers(memberGroup *imdatachain.DataChainGroupMember, batch *leveldb.Batch) utils.ERROR {
	index := memberGroup.GetIndex()
	//查询旧有成员
	users, ERR := db.ImProxyClient_FindGroupMembers(*config.DBKEY_improxy_group_members_knit, *Node.GetNetId(), this.GroupID)
	if !ERR.CheckSuccess() {
		return ERR
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
		v, ok := oldUsersMap[k]
		if ok {
			newUsersMap[k] = v
		} else {
			//新添加的成员
			//保存新成员的起始index
			ERR := db.ImProxyClient_SaveGroupMemberStartIndex(*Node.GetNetId(), memberGroup.GroupID, one, index.Bytes(), batch)
			if !ERR.CheckSuccess() {
				return ERR
			}
			//构建新成员信息
			userinfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), one)
			if !ERR.CheckSuccess() {
				return ERR
			}
			if userinfo != nil {
				userinfo.GroupSign = memberGroup.MembersSign[i]
				userinfo.GroupSignPuk = memberGroup.MembersDHPuk[i]
				userinfo.GroupDHPuk = memberGroup.MembersDHPuk[i]
			} else {
				userinfo = &model.UserInfo{
					Addr:         one,
					Nickname:     "",
					RemarksName:  "",
					HeadNum:      0,
					Status:       0,
					Time:         0,
					CircleClass:  nil,
					Tray:         false,
					Proxy:        new(sync.Map),
					IsGroup:      false,
					GroupSign:    memberGroup.MembersSign[i],
					GroupSignPuk: memberGroup.MembersDHPuk[i],
					//GroupKey:     memberGroup.MembersKey[i],
					GroupDHPuk: memberGroup.MembersDHPuk[i],
					Admin:      false,
				}
			}
			newUsersMap[k] = userinfo
		}
	}
	this.members = make(map[string]nodeStore.AddressNet)
	this.membersProxys = make(map[string]nodeStore.AddressNet)
	newMembers := make([]model.UserInfo, 0, len(newUsersMap))
	//整理成员已经成员的代理节点
	for k, one := range newUsersMap {
		newMembers = append(newMembers, *one)
		this.members[k] = one.Addr
		proxys := make(map[string]nodeStore.AddressNet)
		one.Proxy.Range(func(key, value interface{}) bool {
			keyStr := key.(string)
			addr := value.(nodeStore.AddressNet)
			proxys[keyStr] = addr
			return true
		})
		if len(proxys) == 0 {
			//没有代理就保存它本人
			this.membersProxys[utils.Bytes2string(one.Addr.GetAddr())] = one.Addr
		} else {
			//有代理，则保存代理
			for key, value := range proxys {
				this.membersProxys[key] = value
			}
		}
	}
	//保存群成员
	ERR = db.ImProxyClient_SaveGroupMembers(*config.DBKEY_improxy_group_members_knit, *Node.GetNetId(), memberGroup.GroupID, newMembers, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}

	//需要重置所有成员的sendindex
	for _, one := range newMembers {
		ERR = db.ImProxyClient_SaveGroupSendIndex(*config.DBKEY_improxy_group_datachain_sendIndex_knit, one.Addr,
			memberGroup.GroupID, nil, batch)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}

	return utils.NewErrorSuccess()
}

/*
群成员退出群聊
*/
func (this *GroupKnit) knitGroupQuit(quitGroup *imdatachain.ProxyGroupMemberQuit, batch *leveldb.Batch) utils.ERROR {
	ERR := db.ImProxyClient_RemoveGroupMember(*config.DBKEY_improxy_group_members_knit, *Node.GetNetId(), quitGroup.GroupID, quitGroup.AddrFrom, batch)
	if !ERR.CheckSuccess() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
回滚
*/
func (this *GroupKnit) Proto() (*[]byte, error) {
	base := go_protos.GroupShot{
		GroupID: this.GroupID,
		//GroupKnit:      this.GroupKnit,
		SendIndex:  nil,
		PreHash:    this.PreHash,
		IndexParse: nil,
		AddrAdmin:  this.addrAdmin.GetAddr(),
		//AddrKnitProxys: this.addrKnitProxys,
		Members:       make([][]byte, 0, len(this.members)),
		MembersProxys: make([][]byte, 0, len(this.membersProxys)),
		ShoutUp:       this.shoutUp,
		//ShareKey:       this.ShareKey,
	}
	if this.Index != nil {
		base.IndexParse = this.Index.Bytes()
	}
	for _, member := range this.members {
		base.Members = append(base.Members, member.GetAddr())
	}
	for _, member := range this.membersProxys {
		base.MembersProxys = append(base.MembersProxys, member.GetAddr())
	}
	bs, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func ParseGroupKnit(bs []byte) (*GroupKnit, error) {
	ctx, cancel := context.WithCancel(context.Background())
	groupShot := go_protos.GroupShot{}
	err := proto.Unmarshal(bs, &groupShot)
	if err != nil {
		return nil, err
	}
	pg := GroupKnit{
		GroupID:        groupShot.GroupID,
		lock:           new(sync.RWMutex),
		Index:          new(big.Int).SetBytes(groupShot.IndexParse),
		PreHash:        groupShot.PreHash,
		uploadChan:     nil,
		sendChan:       make(chan imdatachain.DataChainProxyItr, 10000), //
		indexParseLock: nil,
		//IndexParse:         nil,
		parseSignal:        nil,
		addrAdmin:          *nodeStore.NewAddressNet(groupShot.AddrAdmin),
		members:            make(map[string]nodeStore.AddressNet),
		membersProxys:      make(map[string]nodeStore.AddressNet),
		shoutUp:            groupShot.ShoutUp,
		multicastCacheLock: new(sync.Mutex),
		multicastCache:     make([]GroupMulticastMembers, 0),
		multicastChan:      make(chan bool, 1),
		cancel:             cancel,
		ctx:                ctx,
	}
	for _, one := range groupShot.Members {
		pg.members[utils.Bytes2string(one)] = *nodeStore.NewAddressNet(one)
	}
	for _, one := range groupShot.MembersProxys {
		pg.membersProxys[utils.Bytes2string(one)] = *nodeStore.NewAddressNet(one)
	}
	go pg.loopMulticastDataChain()
	return &pg, nil
}

/*
保存群内广播的群成员
当有群成员变动的时候，消息不能发给旧有成员的问题
*/
type GroupMulticastMembers struct {
	members map[string]nodeStore.AddressNet
	msg     imdatachain.DataChainProxyItr
}

/*
合并成员，当群成员变动的时候，消息需要发送给旧有成员和新增成员，所以需要合并，取并集。
*/
func (this *GroupMulticastMembers) MergeMembers(members map[string]nodeStore.AddressNet) {
	for k, v := range members {
		this.members[k] = v
	}
}

func NewGroupMulticastMembers(proxyItr imdatachain.DataChainProxyItr, members map[string]nodeStore.AddressNet) *GroupMulticastMembers {
	gmm := &GroupMulticastMembers{
		members: members,
		msg:     proxyItr,
	}
	return gmm
}

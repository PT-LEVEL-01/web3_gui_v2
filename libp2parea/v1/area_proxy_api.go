package libp2parea

/*
 * p2p内部维护代理, 发消息时, 如果是给客户端发送消息时, 自动根据客户端地址获取其代理节点地址, 然后把自己和对方的代理节点信息传入进去
 */

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"sync"
	"time"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	proxyrec "web3_gui/libp2parea/v1/proxy_rec"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

// 消息返回值, 带有消息和错误信息
type ResSimpleMsgErr struct {
	MachineID string                  // 机器id
	Msg       *message_center.Message // 发送的消息
	Err       error                   // 错误信息
}

// 消息返回值, 带有消息、是否发送成功、是否发送给自己和错误信息
type ResMsgErr struct {
	MachineID   string                  // 机器id
	Msg         *message_center.Message // 发送的消息
	SendSuccess bool                    // 是否发送成功
	SendSelf    bool                    // 消息是发给自己
	Err         error                   // 错误信息
}

// 消息返回值, 带有字节数组和错误信息
type ResSimpleBsErr struct {
	MachineID string  // 机器id
	Bs        *[]byte // 返回的字节数组
	Err       error   // 错误信息
}

// 消息返回值, 带有字节数组、是否发送成功、是否发送给自己和错误信息
type ResBsErr struct {
	MachineID   string  // 机器id
	Bs          *[]byte // 返回的字节数组
	SendSuccess bool    // 是否发送成功
	SendSelf    bool    // 消息是发给自己
	Err         error   // 错误信息
}

/*
 * 获取指定地址和机器id的代理信息
 *
 * @param	recvid			addressnet	接收者节点id
 * @param	recvMachineId	string		接收方机器Id, 如果传入空串, 将获取地址对应所有machineid的客户端地址
 * @return	proxyInfoes		[]ProxyInfo 代理信息数组
 * @return  err		 		error       错误信息
 */
func (this *Area) GetAddrProxy(recvid *nodeStore.AddressNet, recvMachineId string) (res []*proxyrec.ProxyInfo, err error) {
	// 1. 查询缓存中的内容
	var exist bool
	exist, res = this.ProxyCache.GetNodeIdProxy(recvid, recvMachineId)
	// 1.1 查到数据
	if exist && len(res) > 0 {
		// utils.Log.Error().Msgf("从缓存中获取到数据!!!!!!")
		return
	}
	// 1.2 从本地中查询
	{
		_, proxyInfoes := this.ProxyData.GetNodeIdProxy2(recvid, recvMachineId)
		if len(proxyInfoes) > 0 {
			return proxyInfoes, nil
		}
	}

	// 2. 根据地址查询存储节点地址
	var reqData go_protobuf.ProxyInfo
	reqData.Id = *recvid
	reqData.MachineID = recvMachineId
	bs, err := reqData.Marshal()
	if err != nil {
		// 解析失败
		return nil, err
	}

	// 3. 查询地址对应的代理信息
	saveAddrs, err := this.MessageCenter.GetSearchNetAddr(recvid, nil, nil, true, proxyrec.MAX_REC_PROXY_CNT)
	if err != nil {
		return nil, err
	} else if len(saveAddrs) == 0 {
		return nil, errors.New("存储地址为空")
	}

	// 4. 依次向存储节点发送查询消息, 根据结果获取版本号最大的值
	machineIdProxy := make(map[string]*proxyrec.ProxyInfo)
	var lock sync.Mutex
	var wg sync.WaitGroup

	for i := range saveAddrs {
		wg.Add(1)

		// 启动协程依次请求获取结果
		go func(nIndex int) {
			defer wg.Done()
			bs2, _, _, err := this.MessageCenter.SendP2pMsgProxyWaitRequest(config.MSGID_search_addr_proxy, &saveAddrs[nIndex], nil, this.GodID, "", &bs, time.Second*5)
			// 查询出错或者没有找到, 则直接返回
			if err != nil || bs2 == nil || len(*bs2) == 0 {
				return
			}

			// 查询出错或者没有找到, 则直接返回
			if err != nil || bs2 == nil || len(*bs2) == 0 {
				return
			}

			// 5. 解析返回的代理列表信息
			proxyInfoes, err := proxyrec.ParseProxyesProto(bs2)
			if err != nil {
				return
			}

			lock.Lock()
			defer lock.Unlock()

			// 6. 遍历所有的代理信息，根据版本号对比，记录信息
			for ii := range proxyInfoes {
				// utils.Log.Error().Msgf("[%s] proxyInfoes addr:%s, mid:%s, proxyId:%s, version:%d", saveAddrs[nIndex].B58String(), proxyInfoes[ii].NodeId.B58String(), proxyInfoes[ii].MachineId, proxyInfoes[ii].ProxyId.B58String(), proxyInfoes[ii].Version)
				if v, exist := machineIdProxy[proxyInfoes[ii].MachineId]; !exist {
					var proxyInfo proxyrec.ProxyInfo
					proxyInfo.NodeId = recvid
					proxyInfo.ProxyId = proxyInfoes[ii].ProxyId
					proxyInfo.MachineId = proxyInfoes[ii].MachineId
					proxyInfo.Version = proxyInfoes[ii].Version

					machineIdProxy[proxyInfoes[ii].MachineId] = &proxyInfo
				} else if v.Version < proxyInfoes[ii].Version {
					v.ProxyId = proxyInfoes[ii].ProxyId
					v.Version = proxyInfoes[ii].Version
				}
			}
		}(i)
	}

	wg.Wait()

	// 7. 组装结果
	for k := range machineIdProxy {
		res = append(res, machineIdProxy[k])
	}

	// 8. 添加到缓存记录中
	if len(res) > 0 && !this.NodeManager.NodeSelf.GetIsSuper() {
		this.ProxyCache.AddOrUpdateProxyRec(res)
	}

	return res, nil
}

/*
 * 发送一个新的查找超级节点消息, 自动处理代理信息
 * @param		msgid		uint64		消息号
 * @param		recvid		addressnet	接收者节点id
 * @param		content		[]byte		发送内容
 * @return		msg			*Message	返回的消息
 * @return		err			error		错误信息
 */
func (this *Area) SendSearchSuperMsgAuto(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendSearchSuperMsgProxy(msgid, recvid, nil, this.GodID, content)
}

/*
 * 发送一个新的查找超级节点消息, 自动处理代理信息
 *
 * @param			msgid			uint64		消息号
 * @param			recvid			addressnet	接收者节点id
 * @param			content			[]byte		发送内容
 * @param			timeout			Duration	超时时间
 * @return			res				*[]byte		返回结果
 * @return			err				error		错误信息
 */
func (this *Area) SendSearchSuperMsgAutoWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendSearchSuperMsgProxyWaitRequest(msgid, recvid, nil, this.GodID, content, timeout, false)
}

/*
 * 发送一个P2p消息, 自动处理代理信息
 * @param		msgid			uint64			消息号
 * @param		recvid			addressnet		接收者节点id
 * @param		machineID		string			接收方的机器id, 如果传入空串, 将发送给地址对应所有machineid的节点
 * @param		content			[]byte			发送内容
 * @param		isSuperRecv		bool			接收者是不是超级节点, true: 不会搜索它的代理节点, false: 自动搜索接收方的代理节点
 * @return    	res				[]*ResMsgErr    发送的结果列表
 */
func (this *Area) SendP2pMsgAuto(msgid uint64, recvid *nodeStore.AddressNet, machineID string, content *[]byte, isSuperRecv bool) (res []*ResMsgErr) {
	// 代理列表信息 map[proxyId]*proxyrec.ProxyInfo
	// 使用map可以加快访问速度, 如果多个设备连接的是同一个代理, 只需要发送一次就可以了
	var recvProxyIdMap map[string]*proxyrec.ProxyInfo

	// 接收方不是超级节点, 则需要查询其代理节点
	if !isSuperRecv {
		recvProxyIds, err := this.GetAddrProxy(recvid, machineID)
		if err != nil {
			res = append(res, &ResMsgErr{machineID, nil, false, false, err})
			return
		}

		recvProxyIdMap = make(map[string]*proxyrec.ProxyInfo)

		// 通过map处理, 加快访问速度
		if len(recvProxyIds) > 0 {
			for i := range recvProxyIds {
				if recvProxyIds[i] == nil {
					continue
				}

				strProxyId := utils.Bytes2string(*recvProxyIds[i].ProxyId)
				if _, exist := recvProxyIdMap[strProxyId]; !exist {
					recvProxyIdMap[strProxyId] = recvProxyIds[i]
				}
			}
		}
	}

	if len(recvProxyIdMap) > 1 {
		// 地址有多个代理信息, 依次进行发送
		var wg sync.WaitGroup
		chRes := make(chan *ResMsgErr, len(recvProxyIdMap))
		for k := range recvProxyIdMap {
			wg.Add(1)
			// 开协程加快访问速度
			go func(key string) {
				var recvProxyId *nodeStore.AddressNet
				var recvMid string
				if recvProxyIdMap[key] != nil {
					recvProxyId = recvProxyIdMap[key].ProxyId
					recvMid = recvProxyIdMap[key].MachineId
				}

				msg, suc, toSelf, err := this.MessageCenter.SendP2pMsgProxy(msgid, recvid, recvProxyId, this.GodID, machineID, content)
				chRes <- &ResMsgErr{recvMid, msg, suc, toSelf, err}
			}(k)
		}

		// 依次等待结果返回
		for i := 0; i < len(recvProxyIdMap); i++ {
			// res = append(res, <-chRes)
			// 最多等待10秒, 防止无限等待的问题
			select {
			case tempRes := <-chRes:
				res = append(res, tempRes)
			case <-time.NewTimer(time.Second * 10).C:
			}
		}
	} else if len(recvProxyIdMap) == 1 {
		// 地址只有一个代理信息, 直接发送即可
		var recvProxyId *nodeStore.AddressNet
		var recvMid string
		for _, v := range recvProxyIdMap {
			if v == nil {
				continue
			}

			recvProxyId = v.ProxyId
			recvMid = v.MachineId
			break
		}

		msg, suc, toSelf, err := this.MessageCenter.SendP2pMsgProxy(msgid, recvid, recvProxyId, this.GodID, machineID, content)
		res = append(res, &ResMsgErr{recvMid, msg, suc, toSelf, err})
	} else {
		// 没有设置代理信息, 或者接收方是超级节点
		msg, suc, toSelf, err := this.MessageCenter.SendP2pMsgProxy(msgid, recvid, nil, this.GodID, machineID, content)
		res = append(res, &ResMsgErr{machineID, msg, suc, toSelf, err})
	}

	// 返回
	return
}

/*
 * 给指定节点发送一个消息, 自动处理代理信息
 *
 * @param		msgid			uint64			消息号
 * @param		recvid			addressnet		接收者节点id
 * @param		machineID	string				接收方的机器id, 如果传入空串, 将发送给地址对应所有machineid的节点
 * @param		content			[]byte			发送内容
 * @param		timeout			Duration		超时时间
 * @param		isSuperRecv		bool			接收者是不是超级节点, true: 不会搜索它的代理节点, false: 自动搜索接收方的代理节点
 * @return    	res				[]*ResBsErr    	发送的结果列表
 */
func (this *Area) SendP2pMsgAutoWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, machineID string, content *[]byte, timeout time.Duration, isSuperRecv bool) (res []*ResBsErr) {
	// 代理列表信息 map[proxyId]*proxyrec.ProxyInfo
	// 使用map可以加快访问速度, 如果多个设备连接的是同一个代理, 只需要发送一次就可以了
	var recvProxyIdMap map[string]*proxyrec.ProxyInfo

	// 接收方不是超级节点, 则需要查询其代理节点
	if !isSuperRecv {
		recvProxyIds, err := this.GetAddrProxy(recvid, machineID)
		if err != nil {
			res = append(res, &ResBsErr{machineID, nil, false, false, err})
			return
		}

		recvProxyIdMap = make(map[string]*proxyrec.ProxyInfo)

		// 通过map处理, 加快访问速度
		if len(recvProxyIds) > 0 {
			for i := range recvProxyIds {
				if recvProxyIds[i] == nil {
					continue
				}

				strProxyId := utils.Bytes2string(*recvProxyIds[i].ProxyId)
				if _, exist := recvProxyIdMap[strProxyId]; !exist {
					recvProxyIdMap[strProxyId] = recvProxyIds[i]
				}
			}
		}
	}

	if len(recvProxyIdMap) > 1 {
		// 地址有多个代理信息, 依次进行发送
		var wg sync.WaitGroup
		chRes := make(chan *ResBsErr, len(recvProxyIdMap))
		for k := range recvProxyIdMap {
			wg.Add(1)
			// 开协程加快访问速度
			go func(key string) {
				var recvProxyId *nodeStore.AddressNet
				var recvMid string
				if recvProxyIdMap[key] != nil {
					recvProxyId = recvProxyIdMap[key].ProxyId
					recvMid = recvProxyIdMap[key].MachineId
				}

				bs, suc, toSelf, err := this.MessageCenter.SendP2pMsgProxyWaitRequest(msgid, recvid, recvProxyId, this.GodID, machineID, content, timeout)
				chRes <- &ResBsErr{recvMid, bs, suc, toSelf, err}
			}(k)
		}

		// 依次等待结果返回
		for i := 0; i < len(recvProxyIdMap); i++ {
			// res = append(res, <-chRes)
			// 最多等待10秒, 防止无限等待的问题
			select {
			case tempRes := <-chRes:
				res = append(res, tempRes)
			case <-time.NewTimer(time.Second * 10).C:
			}
		}
	} else if len(recvProxyIdMap) == 1 {
		// 地址只有一个代理信息, 直接发送即可
		var recvProxyId *nodeStore.AddressNet
		var recvMid string
		for _, v := range recvProxyIdMap {
			if v == nil {
				continue
			}

			recvProxyId = v.ProxyId
			recvMid = v.MachineId
			break
		}

		bs, suc, toSelf, err := this.MessageCenter.SendP2pMsgProxyWaitRequest(msgid, recvid, recvProxyId, this.GodID, machineID, content, timeout)
		res = append(res, &ResBsErr{recvMid, bs, suc, toSelf, err})
	} else {
		// 没有设置代理信息, 或者接收方是超级节点
		bs, suc, toSelf, err := this.MessageCenter.SendP2pMsgProxyWaitRequest(msgid, recvid, nil, this.GodID, machineID, content, timeout)
		res = append(res, &ResBsErr{machineID, bs, suc, toSelf, err})
	}

	// 返回
	return
}

/*
 * 发送一个加密消息，包括消息头也加密, 自动处理代理信息
 *
 * @param		msgid			uint64			消息号
 * @param		recvid			addressnet		接收者节点id
 * @param		machineID		string			接收方的机器id, 如果传入空串, 将发送给地址对应所有machineid的节点
 * @param		content			[]byte			发送内容
 * @param		isSuperRecv		bool			接收者是不是超级节点, true: 不会搜索它的代理节点, false: 自动搜索接收方的代理节点
 * @return    	res				[]*ResMsgErr    发送的结果列表
 */
func (this *Area) SendP2pMsgHEAuto(msgid uint64, recvid *nodeStore.AddressNet, machineID string, content *[]byte, isSuperRecv bool) (res []*ResMsgErr) {
	// p2pHE消息必须根据machineid分开发送, 否则会因为machineId不匹配, 密钥错误
	// 代理列表信息
	var recvProxyIds []*proxyrec.ProxyInfo

	// 接收方不是超级节点, 则需要查询其代理节点
	if !isSuperRecv {
		var err error
		recvProxyIds, err = this.GetAddrProxy(recvid, machineID)
		if err != nil {
			res = append(res, &ResMsgErr{machineID, nil, false, false, err})
			return
		}
	}

	if len(recvProxyIds) > 1 {
		// 地址有多个代理信息, 依次进行发送
		var wg sync.WaitGroup
		chRes := make(chan *ResMsgErr, len(recvProxyIds))
		for i := range recvProxyIds {
			if recvProxyIds[i] == nil || recvProxyIds[i].ProxyId == nil || len(*recvProxyIds[i].ProxyId) == 0 {
				continue
			}

			wg.Add(1)
			// 开协程加快访问速度
			go func(nIndex int) {
				recvProxyId := recvProxyIds[nIndex].ProxyId
				recvMid := recvProxyIds[nIndex].MachineId

				// utils.Log.Error().Msgf("recvId:%s recvProxy:%s recvMid:%s", recvid.B58String(), recvProxyId.B58String(), recvMid)
				msg, suc, toSelf, err := this.MessageCenter.SendP2pMsgHEProxy(msgid, recvid, recvProxyId, this.GodID, recvMid, content)
				chRes <- &ResMsgErr{recvMid, msg, suc, toSelf, err}
			}(i)
		}

		// 依次等待结果返回
		for i := 0; i < len(recvProxyIds); i++ {
			// res = append(res, <-chRes)
			// 最多等待10秒, 防止无限等待的问题
			select {
			case tempRes := <-chRes:
				res = append(res, tempRes)
			case <-time.NewTimer(time.Second * 10).C:
			}
		}
	} else if len(recvProxyIds) == 1 {
		// 地址只有一个代理信息, 直接发送即可
		var recvProxyId *nodeStore.AddressNet
		var recvMid string
		for i := range recvProxyIds {
			if recvProxyIds[i] == nil || recvProxyIds[i].ProxyId == nil {
				continue
			}

			recvProxyId = recvProxyIds[i].ProxyId
			recvMid = recvProxyIds[i].MachineId
			break
		}

		msg, suc, toSelf, err := this.MessageCenter.SendP2pMsgHEProxy(msgid, recvid, recvProxyId, this.GodID, recvMid, content)
		res = append(res, &ResMsgErr{recvMid, msg, suc, toSelf, err})
	} else {
		// 没有设置代理信息, 或者接收方是超级节点
		msg, suc, toSelf, err := this.MessageCenter.SendP2pMsgHEProxy(msgid, recvid, nil, this.GodID, machineID, content)
		res = append(res, &ResMsgErr{machineID, msg, suc, toSelf, err})
	}

	// 返回
	return
}

/*
 * 发送一个加密消息，包括消息头也加密, 自动处理代理信息
 *
 * @param		msgid			uint64			消息号
 * @param		recvid			addressnet		接收者节点id
 * @param		machineID	string				接收方的机器id, 如果传入空串, 将发送给地址对应所有machineid的节点
 * @param		content			[]byte			发送内容
 * @param		timeout			Duration		超时时间
 * @param		isSuperRecv		bool			接收者是不是超级节点, true: 不会搜索它的代理节点, false: 自动搜索接收方的代理节点
 * @return    	res				[]*ResBsErr    	发送的结果列表
 */
func (this *Area) SendP2pMsgHEAutoWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, machineID string, content *[]byte, timeout time.Duration, isSuperRecv bool) (res []*ResBsErr) {
	// p2pHE消息必须根据machineid分开发送, 否则会因为machineId不匹配, 密钥错误
	// 代理列表信息
	var recvProxyIds []*proxyrec.ProxyInfo

	// 接收方不是超级节点, 则需要查询其代理节点
	if !isSuperRecv {
		var err error
		recvProxyIds, err = this.GetAddrProxy(recvid, machineID)
		if err != nil {
			res = append(res, &ResBsErr{machineID, nil, false, false, err})
			return
		}
	}

	if len(recvProxyIds) > 1 {
		// 地址有多个代理信息, 依次进行发送
		var wg sync.WaitGroup
		chRes := make(chan *ResBsErr, len(recvProxyIds))
		for i := range recvProxyIds {
			if recvProxyIds[i] == nil || recvProxyIds[i].ProxyId == nil || len(*recvProxyIds[i].ProxyId) == 0 {
				continue
			}

			wg.Add(1)
			// 开协程加快访问速度
			go func(nIndex int) {
				recvProxyId := recvProxyIds[nIndex].ProxyId
				recvMid := recvProxyIds[nIndex].MachineId

				msg, suc, toSelf, err := this.MessageCenter.SendP2pMsgHEProxyWaitRequest(msgid, recvid, recvProxyId, this.GodID, recvMid, content, timeout)
				chRes <- &ResBsErr{recvMid, msg, suc, toSelf, err}
			}(i)
		}

		// 依次等待结果返回
		for i := 0; i < len(recvProxyIds); i++ {
			// res = append(res, <-chRes)
			// 最多等待10秒, 防止无限等待的问题
			select {
			case tempRes := <-chRes:
				res = append(res, tempRes)
			case <-time.NewTimer(time.Second * 10).C:
			}
		}
	} else if len(recvProxyIds) == 1 {
		// 地址只有一个代理信息, 直接发送即可
		var recvProxyId *nodeStore.AddressNet
		var recvMid string
		for i := range recvProxyIds {
			if recvProxyIds[i] == nil || recvProxyIds[i].ProxyId == nil {
				continue
			}

			recvProxyId = recvProxyIds[i].ProxyId
			recvMid = recvProxyIds[i].MachineId
			break
		}

		bs, suc, toSelf, err := this.MessageCenter.SendP2pMsgHEProxyWaitRequest(msgid, recvid, recvProxyId, this.GodID, recvMid, content, timeout)
		res = append(res, &ResBsErr{recvMid, bs, suc, toSelf, err})
	} else {
		// 没有设置代理信息, 或者接收方是超级节点
		bs, suc, toSelf, err := this.MessageCenter.SendP2pMsgHEProxyWaitRequest(msgid, recvid, nil, this.GodID, machineID, content, timeout)
		res = append(res, &ResBsErr{machineID, bs, suc, toSelf, err})
	}

	// 返回
	return
}

/*
 * 发送虚拟节点搜索节点消息, 自动处理代理信息
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvNId			*Addressnet			接收方真实节点id
 * @param	machineID		string				接收方的机器id, 如果传入空串, 将发送给地址对应所有machineid的节点
 * @param	content			*[]byte				内容
 * @param	isSuperRecv		bool				接收者是不是超级节点, true: 不会搜索它的代理节点, false: 自动搜索接收方的代理节点
 * @return	msssage			*Message			消息
 * @return	err				error				错误信息
 */
func (this *Area) SendVnodeSearchMsgAuto(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, content *[]byte) (*message_center.Message, error) {
	return this.MessageCenter.SendVnodeSearchMsg(msgid, sendVnodeid, recvVnodeid, nil, this.GodID, content)
}

/*
 * 发送虚拟节点搜索节点消息, 自动处理代理信息
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	content			*[]byte				内容
 * @return	msssage			*Message			消息
 * @return	err				error				错误信息
 */
func (this *Area) SendVnodeSearchMsgAutoWaitRequest(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, content *[]byte, timeout time.Duration) (*[]byte, error) {
	return this.MessageCenter.SendVnodeSearchMsgWaitRequest(msgid, sendVnodeid, recvVnodeid, nil, this.GodID, content, timeout, false)
}

/*
 * 发送虚拟节点之间点对点消息, 自动处理代理信息
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvNodeId		*AddressNet			接收方真实地址, 必须传入
 * @param	machineID		string				接收方的机器id, 如果传入空串, 将发送给地址对应所有machineid的节点
 * @param	content			*[]byte				内容
 * @param	isSuperRecv		bool				接收者是不是超级节点, true: 不会搜索它的代理节点, false: 自动搜索接收方的代理节点
 * @return  res				[]*ResSimpleMsgErr  发送的结果列表
 */
func (this *Area) SendVnodeP2pMsgHEAuto(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, recvNodeId *nodeStore.AddressNet, machineID string, content *[]byte, isSuperRecv bool) (res []*ResSimpleMsgErr) {
	// 代理列表信息 map[proxyId]*proxyrec.ProxyInfo
	// 使用map可以加快访问速度, 如果多个设备连接的是同一个代理, 只需要发送一次就可以了
	var recvProxyIdMap map[string]*proxyrec.ProxyInfo

	// 接收方不是超级节点, 则需要查询其代理节点
	if !isSuperRecv {
		recvProxyIds, err := this.GetAddrProxy(recvNodeId, machineID)
		if err != nil {
			res = append(res, &ResSimpleMsgErr{machineID, nil, err})
			return
		}

		recvProxyIdMap = make(map[string]*proxyrec.ProxyInfo)

		// 通过map处理, 加快访问速度
		if len(recvProxyIds) > 0 {
			for i := range recvProxyIds {
				if recvProxyIds[i] == nil {
					continue
				}

				strProxyId := utils.Bytes2string(*recvProxyIds[i].ProxyId)
				if _, exist := recvProxyIdMap[strProxyId]; !exist {
					recvProxyIdMap[strProxyId] = recvProxyIds[i]
				}
			}
		}
	}

	if len(recvProxyIdMap) > 1 {
		// 地址有多个代理信息, 依次进行发送
		var wg sync.WaitGroup
		chRes := make(chan *ResSimpleMsgErr, len(recvProxyIdMap))
		for k := range recvProxyIdMap {
			wg.Add(1)
			// 开协程加快访问速度
			go func(key string) {
				var recvProxyId *nodeStore.AddressNet
				var recvMid string
				if recvProxyIdMap[key] != nil {
					recvProxyId = recvProxyIdMap[key].ProxyId
					recvMid = recvProxyIdMap[key].MachineId
				}

				msg, err := this.MessageCenter.SendVnodeP2pMsgHE(msgid, sendVnodeid, recvVnodeid, recvNodeId, recvProxyId, this.GodID, machineID, content)
				chRes <- &ResSimpleMsgErr{recvMid, msg, err}
			}(k)
		}

		// 依次等待结果返回
		for i := 0; i < len(recvProxyIdMap); i++ {
			// res = append(res, <-chRes)
			// 最多等待10秒, 防止无限等待的问题
			select {
			case tempRes := <-chRes:
				res = append(res, tempRes)
			case <-time.NewTimer(time.Second * 10).C:
			}
		}
	} else if len(recvProxyIdMap) == 1 {
		// 地址只有一个代理信息, 直接发送即可
		var recvProxyId *nodeStore.AddressNet
		var recvMid string
		for _, v := range recvProxyIdMap {
			if v == nil {
				continue
			}

			recvProxyId = v.ProxyId
			recvMid = v.MachineId
			break
		}

		msg, err := this.MessageCenter.SendVnodeP2pMsgHE(msgid, sendVnodeid, recvVnodeid, recvNodeId, recvProxyId, this.GodID, machineID, content)
		res = append(res, &ResSimpleMsgErr{recvMid, msg, err})
	} else {
		// 没有设置代理信息, 或者接收方是超级节点
		msg, err := this.MessageCenter.SendVnodeP2pMsgHE(msgid, sendVnodeid, recvVnodeid, recvNodeId, nil, this.GodID, machineID, content)
		res = append(res, &ResSimpleMsgErr{machineID, msg, err})
	}

	// 返回
	return
}

/*
 * 发送虚拟节点之间点对点消息, 等待消息返回, 自动处理代理信息
 *
 * @param	msgid			uint64				消息号
 * @param	sendVnodeId		*AddressNetExtend	发送方虚拟节点地址
 * @param	recvVnodeid		*AddressNetExtend	接收方虚拟地址
 * @param	recvNId			*AddressNet			接收方真实地址, 必须传入
 * @param	machineID		string				接收方的机器id, 如果传入空串, 将发送给地址对应所有machineid的节点
 * @param	machineID		string				接收方机器Id
 * @param	content			*[]byte				内容
 * @param	timeout			time.Duration		超时时间
 * @param	isSuperRecv		bool				接收者是不是超级节点, true: 不会搜索它的代理节点, false: 自动搜索接收方的代理节点
 * @return  res				[]*ResSimpleBsErr  	发送的结果列表
 */
func (this *Area) SendVnodeP2pMsgHEAutoWaitRequest(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, recvNId *nodeStore.AddressNet, machineID string, content *[]byte, timeout time.Duration, isSuperRecv bool) (res []*ResSimpleBsErr) {
	// 代理列表信息 map[proxyId]*proxyrec.ProxyInfo
	// 使用map可以加快访问速度, 如果多个设备连接的是同一个代理, 只需要发送一次就可以了
	var recvProxyIdMap map[string]*proxyrec.ProxyInfo

	// 接收方不是超级节点, 则需要查询其代理节点
	if !isSuperRecv {
		recvProxyIds, err := this.GetAddrProxy(recvNId, machineID)
		if err != nil {
			res = append(res, &ResSimpleBsErr{machineID, nil, err})
			return
		}

		recvProxyIdMap = make(map[string]*proxyrec.ProxyInfo)

		// 通过map处理, 加快访问速度
		if len(recvProxyIds) > 0 {
			for i := range recvProxyIds {
				if recvProxyIds[i] == nil {
					continue
				}

				strProxyId := utils.Bytes2string(*recvProxyIds[i].ProxyId)
				if _, exist := recvProxyIdMap[strProxyId]; !exist {
					recvProxyIdMap[strProxyId] = recvProxyIds[i]
				}
			}
		}
	}

	if len(recvProxyIdMap) > 1 {
		// 地址有多个代理信息, 依次进行发送
		var wg sync.WaitGroup
		chRes := make(chan *ResSimpleBsErr, len(recvProxyIdMap))
		for k := range recvProxyIdMap {
			wg.Add(1)
			// 开协程加快访问速度
			go func(key string) {
				var recvProxyId *nodeStore.AddressNet
				var recvMid string
				if recvProxyIdMap[key] != nil {
					recvProxyId = recvProxyIdMap[key].ProxyId
					recvMid = recvProxyIdMap[key].MachineId
				}

				bs, err := this.MessageCenter.SendVnodeP2pMsgHEWaitRequest(msgid, sendVnodeid, recvVnodeid, recvNId, recvProxyId, this.GodID, machineID, content, timeout)
				chRes <- &ResSimpleBsErr{recvMid, bs, err}
			}(k)
		}

		// 依次等待结果返回
		for i := 0; i < len(recvProxyIdMap); i++ {
			// res = append(res, <-chRes)
			// 最多等待10秒, 防止无限等待的问题
			select {
			case tempRes := <-chRes:
				res = append(res, tempRes)
			case <-time.NewTimer(time.Second * 10).C:
			}
		}
	} else if len(recvProxyIdMap) == 1 {
		// 地址只有一个代理信息, 直接发送即可
		var recvProxyId *nodeStore.AddressNet
		var recvMid string
		for _, v := range recvProxyIdMap {
			if v == nil {
				continue
			}

			recvProxyId = v.ProxyId
			recvMid = v.MachineId
			break
		}

		bs, err := this.MessageCenter.SendVnodeP2pMsgHEWaitRequest(msgid, sendVnodeid, recvVnodeid, recvNId, recvProxyId, this.GodID, machineID, content, timeout)
		res = append(res, &ResSimpleBsErr{recvMid, bs, err})
	} else {
		// 没有设置代理信息, 或者接收方是超级节点
		bs, err := this.MessageCenter.SendVnodeP2pMsgHEWaitRequest(msgid, sendVnodeid, recvVnodeid, recvNId, nil, this.GodID, machineID, content, timeout)
		res = append(res, &ResSimpleBsErr{machineID, bs, err})
	}

	// 返回
	return
}

/*
 * 根据磁力地址查询匹配的一个虚拟地址, 自动处理代理信息
 *
 * @param	vnodeId			*AddressNetExtend	虚拟磁力地址
 * @return  vs				*AddressNetExtend  	获取到的虚拟地址
 * @return  err				error	         	错误信息
 */
func (this *Area) SearchVnodeIdAuto(vnodeid *virtual_node.AddressNetExtend) (*virtual_node.AddressNetExtend, error) {
	nodeBs, err := this.Vc.SearchVnodeId(vnodeid, nil, this.GodID, false, 1)
	if err != nil {
		return nil, err
	} else if nodeBs == nil || len(*nodeBs) == 0 {
		return nil, err
	}

	virtualNode := virtual_node.AddressNetExtend(*nodeBs)
	return &virtualNode, err
}

/*
 * 根据磁力地址查询匹配的虚拟地址列表, 自动处理代理信息
 *
 * @param	vnodeId			*AddressNetExtend	虚拟磁力地址
 * @param	num				uint16				需要返回的最大数量
 * @return  vs				[]VnodeInfo     	获取到的虚拟地址信息数组
 * @return  err				error	        	错误信息
 */
func (this *Area) SearchVnodeIdOnebyoneAuto(vnodeid *virtual_node.AddressNetExtend, num uint16) ([]virtual_node.Vnodeinfo, error) {
	var rbs []virtual_node.Vnodeinfo

	bs, err := this.Vc.SearchVnodeId(vnodeid, nil, this.GodID, true, num)
	if err != nil || bs == nil {
		return nil, err
	}

	vrp := new(go_protobuf.VnodeinfoRepeated)
	err = proto.Unmarshal(*bs, vrp)
	if err != nil {
		utils.Log.Warn().Msgf("SearchVnodeIdOnebyoneAuto 不能解析proto:%s", err.Error())
		return nil, err
	}

	for _, one := range vrp.Vnodes {
		rbs = append(rbs, virtual_node.Vnodeinfo{
			Nid:   one.Nid,
			Index: one.Index,
			Vid:   one.Vid,
		})
	}
	return rbs, nil
}

/*
 * 搜索磁力节点网络地址, 自动处理代理信息
 *
 * @param	nodeId		*AddressNet		磁力地址
 * @return	nodeId		*AddressNet		搜索到的磁力地址
 * @return	err			error			错误信息
 */
func (this *Area) SearchNetAddrAuto(nodeId *nodeStore.AddressNet) (*nodeStore.AddressNet, error) {
	return this.MessageCenter.GetSearchNetAddrProxy(nodeId, nil, this.GodID)
}

/*
 * 搜索磁力节点网络地址, 指定返回的数量地址, 自动处理代理信息
 *
 * @param	nodeId		*AddressNet		磁力地址
 * @param	num			uint16			返回的数量
 * @return	nodeIds		[]AddressNet	搜索到的磁力地址列表
 * @return	err			error			错误信息
 */
func (this *Area) SearchNetAddrWithNumAuto(nodeId *nodeStore.AddressNet, num uint16) ([]nodeStore.AddressNet, error) {
	res, err := this.MessageCenter.GetSearchNetAddr(nodeId, nil, this.GodID, true, num)
	return res, err
}

/*
 * 搜索磁力节点网络地址, 自动处理代理信息
 *
 * @param	nodeId		*AddressNetExtend		磁力地址
 * @param	num			uint16					返回的最大数量
 * @return	nodeIds		[]AddressNet			搜索到的磁力地址列表
 * @return	err			error					错误信息
 */
func (this *Area) SearchNetAddrOneByOneAuto(nodeId *nodeStore.AddressNet, num uint16) ([]nodeStore.AddressNet, error) {
	return this.MessageCenter.GetSearchNetAddr(nodeId, nil, this.GodID, true, num)
}

/*
 * 搜索磁力虚拟节点网络地址, 自动处理代理信息
 *
 * @param	nodeId	*AddressNetExtend	磁力地址
 * @return	vnodeId	*AddressNetExtend	搜索到的磁力地址
 * @return	err		error				错误信息
 */
func (this *Area) SearchNetAddrVnodeAuto(nodeId *virtual_node.AddressNetExtend) (*virtual_node.AddressNetExtend, error) {
	netIdBytes, err := this.Vc.SearchVnodeId(nodeId, nil, this.GodID, false, 1)
	if err != nil {
		return nil, err
	} else if netIdBytes == nil || len(*netIdBytes) == 0 {
		return nil, nil
	}

	vnodeId := virtual_node.AddressNetExtend(*netIdBytes)
	return &vnodeId, nil
}

/*
 * 检查被代理的客户端是否在线
 *
 * @param		recvid			addressnet		接收者节点id
 * @param		machineID		string			接收方的机器id, 如果传入空串, 将发送给地址对应所有machineid的节点
 * @return		isOnline		bool			是否在线
 */
func (this *Area) CheckProxyClientIsOnline(nodeId *nodeStore.AddressNet, machineID string) bool {
	// 1. 检查参数的合法性
	if nodeId == nil || len(*nodeId) == 0 {
		return false
	}

	// 2. 从本地查看是否存在和目标的连接信息
	if session, ok := this.SessionEngine.GetSession(this.AreaName[:], utils.Bytes2string(*nodeId)); ok {
		// 存在和目标的连接信息, 则目标在线
		if machineID != "" && session != nil {
			// 如果指定了machineid, 则进一步检测machineid是否相同
			if session.GetMachineID() == machineID {
				return true
			}
		} else {
			return true
		}
	}

	// 3. 获取目标对应的代理信息
	recvProxyIds, err := this.GetAddrProxy(nodeId, machineID)
	if err != nil {
		return false
	}

	// 4. 不存在代理信息, 则客户端不在线
	if len(recvProxyIds) == 0 {
		return false
	}

	return true
}

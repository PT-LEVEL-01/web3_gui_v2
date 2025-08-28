package proxyrec

import (
	"context"
	"math/big"
	"sort"
	"sync"
	"time"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/utils"
)

/*
 * 代理节点记录信息
 */
type ProxyNodeRecInfo struct {
	NodeId        *nodeStore.AddressNet  // 客户端地址
	ProxyId       *nodeStore.AddressNet  // 代理地址
	SaveNodes     map[string]interface{} // 保存数据的节点地址 key:string=节点地址;value:nil=nil;
	MachineInfoes map[string]int64       // 机器id信息map key:string=机器ID;value:int64=版本号;
}

/*
 * 代理节点详情信息
 */
type ProxyDetailInfo struct {
	proxyMapInfo map[string]*ProxyNodeRecInfo  // 代理信息记录map map[nodeId]map[machineId]ProxyInfo
	optLock      sync.RWMutex                  // 操作锁
	ctx          context.Context               // 上下文
	msgCenter    *message_center.MessageCenter // 消息中心
	allSaveAddrs map[string]uint32             // 保存数据的服务器节点map, map[serverAddr]saveCnt
	whiteNodes   map[string]interface{}        // 保留所有数据的节点
}

/*
 * NewProxyData 创建代理详情对象
 */
func NewProxyDetail(ctx context.Context, msgCenter *message_center.MessageCenter) *ProxyDetailInfo {
	if msgCenter == nil {
		utils.Log.Error().Msgf("请传入有效的消息管理器!!!!")
		return nil
	}

	pd := ProxyDetailInfo{}
	pd.proxyMapInfo = make(map[string]*ProxyNodeRecInfo)
	pd.ctx = ctx
	pd.msgCenter = msgCenter
	pd.allSaveAddrs = make(map[string]uint32)
	pd.whiteNodes = make(map[string]interface{})

	// 开启协程, 定期广播自己拥有的代理信息
	go pd.loopSendProxyInfo()

	return &pd
}

/*
 * loopSendProxyInfo 定时广播自己拥有的代理信息
 */
func (pd *ProxyDetailInfo) loopSendProxyInfo() {
	ticker := time.NewTicker(ProxySyncTime)
	defer ticker.Stop()

	// 定时器定时触发同步操作
	for range ticker.C {
		// 1. 如果area已经退出，则退出定时器
		select {
		case <-pd.ctx.Done():
			return
		default:
		}

		// 2. 如果没有代理信息, 则等待下一次的定时器
		if len(pd.proxyMapInfo) == 0 {
			continue
		}

		// 3. 加读锁, 注意: 一定要解锁
		pd.optLock.RLock()

		// 4. 整理已有数据
		syncData := pd.getSyncProxyInfo()
		if len(syncData) == 0 {
			pd.optLock.RUnlock()
			continue
		}

		// 5. 给客户端发送对应数据
		for k, v := range syncData {
			recvId := nodeStore.AddressNet([]byte(k))
			content, err := v.Marshal()
			if err != nil {
				continue
			}
			pd.msgCenter.SendP2pMsg(config.MSGID_sync_proxy, &recvId, &content)
		}

		// 6. 解锁
		pd.optLock.RUnlock()
	}
}

/*
 * AddProxyDetail 增加代理详情
 *
 * @param	clientAddr	AddressNet		被代理的节点地址
 * @param	proxyId		AddressNet		代理节点地址
 * @param	machineId	string			被代理的节点机器id
 * @param	version		int64			设置代理的版本号
 * @return	success		bool			是否添加成功
 */
func (pd *ProxyDetailInfo) AddProxyDetail(clientAddr, proxyId nodeStore.AddressNet, machineId string, version int64) bool {
	// 如果还没有初始化, 则直接返回
	if pd == nil {
		return false
	}

	// 检测参数的合法性
	if len(clientAddr) == 0 || len(proxyId) == 0 {
		return false
	}

	// utils.Log.Error().Msgf("[代理:%s] 增加代理详情 客户端:%s mId:%s version:%d", proxyId.B58String(), clientAddr.B58String(), machineId, version)

	// 加写锁
	pd.optLock.Lock()
	defer pd.optLock.Unlock()

	// 立刻发送同步信息标识
	var bNeedSyncData = true
	// 客户端id的key值
	strAddr := utils.Bytes2string(clientAddr)
	// 根据客户端key获取详情
	detailInfo, exist := pd.proxyMapInfo[strAddr]
	// 判断是否存在, 存在则更新, 不存在则添加
	if exist && detailInfo != nil {
		if detailInfo.MachineInfoes != nil {
			machineVersion, exist := detailInfo.MachineInfoes[machineId]
			if exist {
				if machineVersion < version {
					detailInfo.MachineInfoes[machineId] = version
				} else {
					// 版本号较小时, 不做同步处理
					bNeedSyncData = false
				}
			} else {
				detailInfo.MachineInfoes[machineId] = version
			}
		} else {
			detailInfo.MachineInfoes = make(map[string]int64)
			detailInfo.MachineInfoes[machineId] = version
		}
	} else {
		// 查询地址对应存储节点
		saveAddrs, err := pd.msgCenter.GetSearchNetAddr(&clientAddr, nil, nil, true, MAX_REC_PROXY_CNT)
		if err != nil {
			utils.Log.Error().Msgf("获取存储节点出错!!!! err:%s", err)
			return false
		} else if len(saveAddrs) == 0 {
			utils.Log.Error().Msgf("没有找到任何存储代理数据的节点!!!!")
			return false
		}

		// 构建代理详情数据
		var newDetailInfo ProxyNodeRecInfo
		newDetailInfo.NodeId = &clientAddr
		newDetailInfo.ProxyId = &proxyId
		newDetailInfo.SaveNodes = make(map[string]interface{})
		for i := range saveAddrs {
			// utils.Log.Error().Msgf("[代理:%s] 增加代理详情 客户端:%s mId:%s version:%d 存储节点:%s", proxyId.B58String(), clientAddr.B58String(), machineId, version, saveAddrs[i].B58String())
			newDetailInfo.SaveNodes[utils.Bytes2string(saveAddrs[i])] = struct{}{}
		}
		newDetailInfo.MachineInfoes = make(map[string]int64)
		newDetailInfo.MachineInfoes[machineId] = version

		// 记录客户端地址对应的代理信息
		pd.proxyMapInfo[strAddr] = &newDetailInfo

		// 保存
		for i := range saveAddrs {
			pd.allSaveAddrs[utils.Bytes2string(saveAddrs[i])]++
		}
	}

	if bNeedSyncData {
		// 立刻同步一次代理信息
		// 根据客户端地址整理已有数据
		syncData := pd.getSyncProxyInfoAppClientAddr(strAddr, machineId)
		// utils.Log.Warn().Msgf("需要同步的长度为: %d", len(syncData))
		if len(syncData) > 0 {
			// 给客户端发送对应数据
			for k, v := range syncData {
				recvId := nodeStore.AddressNet([]byte(k))
				content, err := v.Marshal()
				if err != nil {
					continue
				}
				pd.msgCenter.SendP2pMsg(config.MSGID_sync_proxy, &recvId, &content)
			}
		}
	}

	return true
}

/*
 * RemoveProxyDetail 删除代理详情
 *
 * @param	addr		AddressNet		被代理的节点地址
 * @param	machineId	string			被代理的节点机器id
 * @param	version		int64			设置代理的版本号
 * @return 	exist		bool			之前是否存在代理信息
 */
func (pd *ProxyDetailInfo) RemoveProxyDetail(addr *nodeStore.AddressNet, machineId string, version int64) bool {
	// utils.Log.Info().Msgf("客户端下线, 删除代理信息 cid:%s mid:%s version:%d", addr.B58String(), machineId, version)
	// 如果还没有初始化, 则直接返回
	if pd == nil {
		utils.Log.Error().Msgf("还没有初始化!!!!!")
		return false
	}

	// 加写锁
	pd.optLock.Lock()
	defer pd.optLock.Unlock()

	// 获取客户端地址key
	strAddr := utils.Bytes2string(*addr)

	// 删除客户端代理信息
	// 根据节点地址, 获取代理详情信息
	proxyDetailInfo, exist := pd.proxyMapInfo[strAddr]
	if !exist || proxyDetailInfo == nil {
		utils.Log.Error().Msgf("没有找到代理信息!!!!!")
		return false
	}

	// 根据机器id, 检查代理信息是否存在
	versionRec, exist := proxyDetailInfo.MachineInfoes[machineId]
	if !exist {
		utils.Log.Error().Msgf("根据机器id, 没有找到代理信息!!!!!")
		return false
	}

	// 对比版本号, 如果是之前的版本号, 则直接返回
	if versionRec > version {
		// 节点的版本号, 大于需要删除的版本号, 忽略本次操作
		utils.Log.Error().Msgf("节点的版本号, 大于需要删除的版本号, 忽略本次操作!!!!!")
		return false
	}

	// 删除对应的记录信息
	delete(proxyDetailInfo.MachineInfoes, machineId)

	// 如果没有任何的记录信息, 则清除该节点地址对应的所有记录信息
	if len(proxyDetailInfo.MachineInfoes) == 0 {
		delete(pd.proxyMapInfo, strAddr)
	}

	return true
}

/*
 * NodeOfflineDeal 存储节点下线处理
 *
 * @param	offlineAddr		AddressNet		下线节点地址
 */
func (pd *ProxyDetailInfo) NodeOfflineDeal(offlineAddr *nodeStore.AddressNet) {
	if pd == nil {
		utils.Log.Error().Msgf("还没有初始化!!!!!")
		return
	}
	// 检查参数的合法性
	if offlineAddr == nil || len(*offlineAddr) == 0 {
		utils.Log.Error().Msgf("下线地址信息为空!!!!!")
		return
	}

	// utils.Log.Info().Msgf("节点下线处理 sid:%s", offlineAddr.B58String())

	// 加锁
	pd.optLock.Lock()
	defer pd.optLock.Unlock()

	// 节点地址key
	strAddr := utils.Bytes2string(*offlineAddr)

	// 判断节点是否需要处理
	if _, exist := pd.allSaveAddrs[strAddr]; !exist {
		// utils.Log.Error().Msgf("节点信息不存在!!!!!")
		return
	}

	// 超级节点，需要判断是不是需要更新客户端代理节点信息
	for _, v := range pd.proxyMapInfo {
		if _, exist := v.SaveNodes[strAddr]; !exist {
			continue
		}

		// 删除存储节点信息
		delete(v.SaveNodes, strAddr)
		if cnt := pd.allSaveAddrs[strAddr]; cnt > 1 {
			pd.allSaveAddrs[strAddr]--
		} else {
			delete(pd.allSaveAddrs, strAddr)
		}
		// utils.Log.Error().Msgf("删除存储节点信息 sid:%s cnt:%d", offlineAddr.B58String(), pd.allSaveAddrs[strAddr])

		// 更新存储节点地址
		saveAddrs, err := pd.msgCenter.GetSearchNetAddr(v.ProxyId, nil, nil, true, MAX_REC_PROXY_CNT)
		if err == nil && len(saveAddrs) > 0 {
			for ii := range saveAddrs {
				strTempKey := utils.Bytes2string(saveAddrs[ii])
				if _, exist := v.SaveNodes[strTempKey]; exist {
					continue
				}

				v.SaveNodes[strTempKey] = struct{}{}
				pd.allSaveAddrs[strTempKey]++
			}
		}
	}
}

/*
 * NodeOnlineDeal 超级节点上线处理
 *
 * @param	node		AddressNet		超级节点信息
 */
func (pd *ProxyDetailInfo) NodeOnlineDeal(node *nodeStore.Node) {
	// utils.Log.Info().Msgf("节点上线处理 sid:%s", node.IdInfo.Id.B58String())
	// 如果没有初始化, 或者节点不是超级节点, 则直接返回
	if pd == nil || !node.GetIsSuper() {
		return
	}

	// 如果没有任何保存的数据, 则不做处理
	if len(pd.proxyMapInfo) == 0 {
		return
	}

	// 加锁
	pd.optLock.Lock()
	defer pd.optLock.Unlock()

	// 获取上线节点地址的key值
	strKey := utils.Bytes2string(node.IdInfo.Id)

	// 遍历代理信息, 依次进行处理
	for _, v := range pd.proxyMapInfo {
		if v.SaveNodes == nil || len(v.SaveNodes) == 0 {
			v.SaveNodes = make(map[string]interface{})
			v.SaveNodes[strKey] = struct{}{}
			pd.allSaveAddrs[strKey]++
		}

		// 存储节点中已经包含了该节点, 则不再处理
		if _, exist := v.SaveNodes[strKey]; exist {
			continue
		}

		// 如果存储节点没有达到设定的上限, 则直接设置存储节点
		if len(v.SaveNodes) < MAX_REC_PROXY_CNT {
			v.SaveNodes[strKey] = struct{}{}
			pd.allSaveAddrs[strKey]++
			continue
		}

		// 上线的节点地址, 在保存的地址最大和最小之间, 需要重新查询地址
		if pd.checkNeedServerNode(&node.IdInfo.Id, v.NodeId, v.SaveNodes) {
			// utils.Log.Error().Msgf("需要更新存储节点!!!!!!")
			saveAddrs, err := pd.msgCenter.GetSearchNetAddr(v.NodeId, nil, nil, true, MAX_REC_PROXY_CNT)
			if err == nil && len(saveAddrs) > 0 {
				// 先清空之前的记录信息
				for k := range v.SaveNodes {
					pd.allSaveAddrs[k]--
					// 如果存储节点引用次数为0, 则删除map中的记录
					if pd.allSaveAddrs[k] <= 0 {
						delete(pd.allSaveAddrs, k)
					}
				}
				v.SaveNodes = make(map[string]interface{})
				// utils.Log.Warn().Msgf("清空了所有地址信息!!!!!! saveAddr len:%d", len(saveAddrs))

				// 依次保留最新的保存地址信息
				for ii := range saveAddrs {
					strTempKey := utils.Bytes2string(saveAddrs[ii])
					if _, exist := v.SaveNodes[strTempKey]; exist {
						continue
					}

					v.SaveNodes[strTempKey] = struct{}{}
					pd.allSaveAddrs[strTempKey]++
					// utils.Log.Warn().Msgf("添加的地址: %s", saveAddrs[ii].B58String())
				}
			}
		}
	}
}

/*
 * CheckIsSaveAddr 检查地址是不是存储节点地址
 *
 * @param	offlineAddr		AddressNet		下线节点地址
 * @return	exist			bool			是不是存储节点标识
 */
func (pd *ProxyDetailInfo) CheckIsSaveAddr(addr *nodeStore.AddressNet) bool {
	// 检查是否初始化
	if pd == nil {
		return false
	}
	// 检查参数的合法性
	if addr == nil || len(*addr) == 0 {
		return false
	}

	// 加锁
	pd.optLock.Lock()
	defer pd.optLock.Unlock()

	// 节点地址key
	strAddr := utils.Bytes2string(*addr)

	// 判断节点是否需要处理
	cnt, exist := pd.allSaveAddrs[strAddr]
	if !exist || cnt <= 0 {
		return false
	}

	// 返回
	return true
}

/*
 * 获取需要同步的代理信息
 *
 * @return res		map[string][]ProxyInfo	需要同步的数据, key为代理节点的地址
 */
func (pd *ProxyDetailInfo) getSyncProxyInfo() (res map[string]go_protobuf.ProxyRepeated) {
	// 如果没有初始化, 或者数据为空, 则直接返回
	if pd == nil || len(pd.proxyMapInfo) == 0 {
		return
	}

	// 依次组装相关数据
	res = make(map[string]go_protobuf.ProxyRepeated)
	allProxyRepeat := go_protobuf.ProxyRepeated{
		Proxys: make([]*go_protobuf.ProxyInfo, 0),
	}
	for _, v := range pd.proxyMapInfo {
		// 依次遍历代理节点, 根据其存储节点, 进行记录
		for k := range v.SaveNodes {
			// 如果存储节点是保存所有
			if _, exist := pd.whiteNodes[k]; exist {
				continue
			}

			// 先判断map是否存在, 不存在, 则创建
			if _, exist := res[k]; !exist {
				proxyRepeat := go_protobuf.ProxyRepeated{
					Proxys: make([]*go_protobuf.ProxyInfo, 0),
				}
				res[k] = proxyRepeat
			}

			// 多端会记录多条数据
			for kk, vv := range v.MachineInfoes {
				// 构建代理信息
				var proxyInfo go_protobuf.ProxyInfo
				proxyInfo.Id = *v.NodeId
				proxyInfo.ProxyId = *v.ProxyId
				proxyInfo.MachineID = kk
				proxyInfo.Version = vv

				proxyRepeat := res[k]
				proxyRepeat.Proxys = append(proxyRepeat.Proxys, &proxyInfo)
				res[k] = proxyRepeat
			}
		}

		// 多端会记录多条数据
		for kk, vv := range v.MachineInfoes {
			// 构建代理信息
			var proxyInfo go_protobuf.ProxyInfo
			proxyInfo.Id = *v.NodeId
			proxyInfo.ProxyId = *v.ProxyId
			proxyInfo.MachineID = kk
			proxyInfo.Version = vv

			// 追加到所有的代理记录信息中
			allProxyRepeat.Proxys = append(allProxyRepeat.Proxys, &proxyInfo)
		}
	}

	// 组装所有的存储节点
	for k := range pd.whiteNodes {
		// 先判断map是否存在, 不存在, 则创建
		if _, exist := res[k]; !exist {
			proxyRepeat := go_protobuf.ProxyRepeated{
				Proxys: make([]*go_protobuf.ProxyInfo, 0),
			}
			res[k] = proxyRepeat
		}

		res[k] = allProxyRepeat
	}

	// 返回
	return
}

/*
 * 获取指定节点的同步代理信息
 *
 * @return res		map[string][]ProxyInfo	需要同步的数据, key为代理节点的地址
 */
func (pd *ProxyDetailInfo) getSyncProxyInfoAppClientAddr(nodeAddr, machineID string) (res map[string]go_protobuf.ProxyRepeated) {
	// 如果没有初始化, 或者数据为空, 则直接返回
	if pd == nil || len(pd.proxyMapInfo) == 0 || nodeAddr == "" {
		utils.Log.Error().Msgf("参数错误!!!!!!")
		return
	}

	// 依次组装相关数据
	res = make(map[string]go_protobuf.ProxyRepeated)
	// 根据客户端节点获取代理信息
	proxyDetailInfo, exist := pd.proxyMapInfo[nodeAddr]
	if !exist {
		utils.Log.Error().Msgf("不存在代理信息!!!!")
		return
	}
	// 构建返回代理信息
	proxyRepeat := go_protobuf.ProxyRepeated{
		Proxys: make([]*go_protobuf.ProxyInfo, 0),
	}
	// 依次遍历存储节点, 获取节点对应的代理信息
	for k, v := range proxyDetailInfo.MachineInfoes {
		if machineID != k {
			continue
		}

		// 构建代理信息
		var proxyInfo go_protobuf.ProxyInfo
		proxyInfo.Id = *proxyDetailInfo.NodeId
		proxyInfo.ProxyId = *proxyDetailInfo.ProxyId
		proxyInfo.MachineID = k
		proxyInfo.Version = v

		// 追加代理信息
		proxyRepeat.Proxys = append(proxyRepeat.Proxys, &proxyInfo)
	}
	// 赋值到存储节点上
	for k := range proxyDetailInfo.SaveNodes {
		res[k] = proxyRepeat
	}

	// 组装所有的存储节点
	for k := range pd.whiteNodes {
		if _, exist := res[k]; exist {
			continue
		}

		res[k] = proxyRepeat
	}

	// 返回
	return
}

/*
 * 检查是否需要指定的超级节点地址
 */
func (pd *ProxyDetailInfo) checkNeedServerNode(nodes, clientId *nodeStore.AddressNet, saveNodes map[string]interface{}) bool {
	// 1. 判断参数是否合法
	if nodes == nil || len(*nodes) == 0 || clientId == nil || len(*clientId) == 0 || saveNodes == nil {
		return false
	}

	// 2. 获取key
	strKey := utils.Bytes2string(*nodes)

	// 3. 判断key是否存在
	if _, exist := saveNodes[strKey]; exist {
		return false
	}

	// 4. 判断保存的节点是否达到要求
	if len(saveNodes) < MAX_REC_PROXY_CNT {
		return true
	}

	// 5. 组装排序数组
	onebyone := new(nodeStore.IdDESC)
	for k := range saveNodes {
		*onebyone = append(*onebyone, new(big.Int).SetBytes([]byte(k)))
	}
	newNodeBigId := new(big.Int).SetBytes([]byte(*nodes))
	*onebyone = append(*onebyone, newNodeBigId)
	sort.Sort(onebyone)

	// 6. 客户端地址big的int值
	clientBigInt := new(big.Int).SetBytes(*clientId)

	// 7. 对比，获取大于客户端节点值的数量
	var nIndex int
	for i := len(*onebyone) - 1; i >= 0; i-- {
		one := (*onebyone)[i]
		if one.Cmp(clientBigInt) == 1 {
			nIndex = i
			break
		}
	}

	// 8. 判断目标是否在
	if (*onebyone)[nIndex].Cmp(newNodeBigId) == 0 {
		return true
	}
	var nCheckCnt int = 1
	var checkValid int
	for i := 1; ; i++ {
		checkValid = 0
		// 找upsession，由近及远
		if u := nIndex - i; u >= 0 {
			checkValid++
			nCheckCnt++
			if (*onebyone)[u].Cmp(newNodeBigId) == 0 {
				return true
			}
		}

		if nCheckCnt >= MAX_REC_PROXY_CNT {
			return false
		}

		// 找downsession，由近及远
		if d := nIndex + i; d < len(*onebyone) {
			checkValid++
			nCheckCnt++
			if (*onebyone)[d].Cmp(newNodeBigId) == 0 {
				return true
			}
		}

		// 如果检查次数大于最大记录数, 则直接返回
		if nCheckCnt >= MAX_REC_PROXY_CNT {
			return false
		}

		if checkValid == 0 {
			break
		}
	}

	return false
}

/*
 * 添加保存节点白名单, 即会保存所有节点代理信息
 *
 * @param	nodeId	*AddressNet		添加的节点地址
 */
func (pd *ProxyDetailInfo) AddWhiteNode(nodeId *nodeStore.AddressNet) bool {
	// 1. 如果还没有初始化, 则直接返回
	if pd == nil {
		return false
	}

	// 2. 检查参数是否有效
	if nodeId == nil || len(*nodeId) == 0 {
		return false
	}

	// 3. 加锁
	pd.optLock.Lock()
	defer pd.optLock.Unlock()

	// 4. 添加白名单
	pd.whiteNodes[utils.Bytes2string(*nodeId)] = struct{}{}

	// 5. 返回
	return true
}

/*
 * 添加保存节点白名单列表, 即会保存所有节点代理信息
 *
 * @param	nodeIds	[]*AddressNet	添加的节点地址
 */
func (pd *ProxyDetailInfo) AddWhiteNodes(nodeIds []nodeStore.AddressNet) bool {
	// 1. 如果还没有初始化, 则直接返回
	if pd == nil {
		return false
	}

	// 2. 检查参数是否有效
	if len(nodeIds) == 0 {
		return false
	}

	// 3. 加锁
	pd.optLock.Lock()
	defer pd.optLock.Unlock()

	// 4. 添加白名单
	for i := range nodeIds {
		if nodeIds[i] == nil || len(nodeIds[i]) == 0 {
			continue
		}

		pd.whiteNodes[utils.Bytes2string(nodeIds[i])] = struct{}{}
	}

	// 5. 返回
	return true
}

/*
 * 移除保存节点白名单列表
 *
 * @param	nodeId	*AddressNet		删除的白名单节点地址
 */
func (pd *ProxyDetailInfo) RemoveWhiteNode(nodeId *nodeStore.AddressNet) {
	// 1. 如果还没有初始化, 则直接返回
	if pd == nil {
		return
	}

	// 2. 检查参数是否有效
	if nodeId == nil || len(*nodeId) == 0 {
		return
	}

	// 3. 加锁
	pd.optLock.Lock()
	defer pd.optLock.Unlock()

	// 4. 删除白名单
	delete(pd.whiteNodes, utils.Bytes2string(*nodeId))
}

/*
 * 获取所有的保存节点白名单列表
 *
 * @return	res		[]*AddressNet	白名单节点地址列表
 */
func (pd *ProxyDetailInfo) GetWhiteNode() []*nodeStore.AddressNet {
	// 1. 如果还没有初始化, 则直接返回
	if pd == nil {
		return nil
	}

	// 2. 加锁
	pd.optLock.RLock()
	defer pd.optLock.RUnlock()

	// 3. 获取所有的白名单
	res := make([]*nodeStore.AddressNet, 0)
	for k := range pd.whiteNodes {
		nodeId := nodeStore.AddressNet([]byte(k))
		res = append(res, &nodeId)
	}

	// 4. 返回
	return res
}

/*
 * 检查节点是不是白名单节点
 *
 * @param	nodeId	*AddressNet		检查的节点地址
 * @return	exist	bool			是不是白名单标识
 */
func (pd *ProxyDetailInfo) CheckWhiteNode(nodeId *nodeStore.AddressNet) bool {
	// 1. 如果还没有初始化, 则直接返回
	if pd == nil {
		return false
	}

	// 2. 检查参数是否有效
	if nodeId == nil || len(*nodeId) == 0 {
		return false
	}

	// 3. 加锁
	pd.optLock.RLock()
	defer pd.optLock.RUnlock()

	// 4. 删除白名单
	_, exist := pd.whiteNodes[utils.Bytes2string(*nodeId)]

	// 5. 返回
	return exist
}

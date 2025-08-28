package proxyrec

import (
	"context"
	"sync"
	"time"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

// 代理记录信息
type ProxyData struct {
	proxyMapInfo map[string]map[string]*ProxyInfo // 代理信息记录map key:string=客户端id;value:map={key:string=机器ID;value:*ProxyInfo=代理信息;};
	optLock      sync.RWMutex                     // 操作锁
	ctx          context.Context                  // 上下文
}

/*
 * NewProxyData 创建代理数据对象
 */
func NewProxyData(ctx context.Context) *ProxyData {
	pr := ProxyData{}
	pr.proxyMapInfo = make(map[string]map[string]*ProxyInfo)
	pr.ctx = ctx

	// 开启协程，定期检查数据是否过期，过期则删除
	go pr.loopCleanData()

	return &pr
}

/*
 * loopCleanData 循环定期检测并清理过期的代理数据
 */
func (pd *ProxyData) loopCleanData() {
	ticker := time.NewTicker(ProxySyncTime)
	defer ticker.Stop()

	for range ticker.C {
		// 1. 如果area已经退出，则退出定时器
		select {
		case <-pd.ctx.Done():
			return
		default:
		}

		// 获取当前毫秒数
		curTime := time.Now()

		// 2. 遍历所有的数据，依次检查合法性
		pd.optLock.Lock()
		for k, v := range pd.proxyMapInfo {
			// 如果k对应的值为空，则删除map中的记录信息
			if len(v) == 0 {
				delete(pd.proxyMapInfo, k)
				continue
			}
			for kk, vv := range v {
				// 如果值信息无效，则删除map中的记录信息
				if vv.NodeId == nil || vv.ProxyId == nil || vv.Version == 0 {
					delete(pd.proxyMapInfo[k], kk)
					continue
				}

				// 如果长时间没有同步到数据，则删除对应的记录信息
				if curTime.Sub(vv.syncTime) >= ProxyDataInvalieTime*ProxySyncTime {
					// utils.Log.Error().Msgf("删除过期的记录信息 addr:%s mid:%s curTime:%d syncTime:%d", vv.NodeId.B58String(), vv.MachineId, curTime.UnixMilli(), vv.syncTime.UnixMilli())
					delete(pd.proxyMapInfo[k], kk)
				}
			}

			// 如果k对应的值为空，则删除map中的记录信息
			if len(pd.proxyMapInfo[k]) == 0 {
				delete(pd.proxyMapInfo, k)
			}
		}
		pd.optLock.Unlock()
	}
}

/*
 * AddOrUpdateProxyRec 添加或更新代理记录信息
 *
 * @param	nodes		[]ProxyInfo			代理信息列表
 * @param	proxyId		Address				代理节点地址
 * @param	recvId		Address				存储节点地址
 * @return	success		bool				是否添加成功
 */
func (pr *ProxyData) AddOrUpdateProxyRec(nodes []ProxyInfo, proxyId, recvId *nodeStore.AddressNet) bool {
	// utils.Log.Warn().Msgf("[存储节点:%s]添加或更新代理记录信息, 代理地址:%s", recvId.B58String(), proxyId.B58String())
	// 1. 判断是否初始化
	if pr == nil {
		utils.Log.Error().Msgf("还未初始化!!!!!!")
		return false
	}

	// 2. 检测area是否退出
	select {
	case <-pr.ctx.Done():
		utils.Log.Error().Msgf("应用结束!!!!!!")
		return false
	default:
	}

	// 3. 加写锁
	pr.optLock.Lock()
	defer pr.optLock.Unlock()

	// 4. 获取当前时间, 用以记录同步时间, 进而方便维护代理的处理
	curTime := time.Now()

	// 5. 遍历代理列表, 依次进行处理
	for i := range nodes {
		// 检查参数是否合法
		if nodes[i].NodeId == nil || len(*nodes[i].NodeId) == 0 || nodes[i].ProxyId == nil || len(*nodes[i].ProxyId) == 0 {
			continue
		}

		// 5.1 获取节点地址的base58字符串
		strNodeId := utils.Bytes2string(*nodes[i].NodeId)

		// utils.Log.Warn().Msgf("[存储节点:%s] 代理地址:%s 客户端:%s mId:%s version:%d", recvId.B58String(), proxyId.B58String(), nodes[i].NodeId.B58String(), nodes[i].MachineId, nodes[i].Version)

		// 5.2 获取机器id对应的map信息
		machineIdInfo, exist := pr.proxyMapInfo[strNodeId]
		if exist && machineIdInfo != nil {
			// 存在, 则更新相关信息
			// 查询机器id map对应的数据是否存在
			proxyInfo, exist2 := machineIdInfo[nodes[i].MachineId]
			if exist2 && proxyInfo != nil {
				// 只有版本号不大于更新信息的版本号时，才允许更新
				if proxyInfo.Version <= nodes[i].Version {
					// 存在, 则更新代理、同步时间和版本号
					proxyInfo.ProxyId = nodes[i].ProxyId
					proxyInfo.syncTime = curTime
					proxyInfo.Version = nodes[i].Version
				}
			} else {
				// 不存在
				var newProxyInfo ProxyInfo
				newProxyInfo.NodeId = nodes[i].NodeId
				newProxyInfo.ProxyId = nodes[i].ProxyId
				newProxyInfo.MachineId = nodes[i].MachineId
				newProxyInfo.Version = nodes[i].Version
				newProxyInfo.syncTime = curTime

				// 保存
				machineIdInfo[nodes[i].MachineId] = &newProxyInfo
			}
		} else {
			// 不存在, 则添加机器id对应的map信息
			// 构建代理信息
			var newProxyInfo ProxyInfo
			newProxyInfo.NodeId = nodes[i].NodeId
			newProxyInfo.ProxyId = nodes[i].ProxyId
			newProxyInfo.MachineId = nodes[i].MachineId
			newProxyInfo.Version = nodes[i].Version
			newProxyInfo.syncTime = curTime

			// 构建机器id的map信息
			mapMachineValue := make(map[string]*ProxyInfo)
			// 保存机器id map对应的数据
			mapMachineValue[nodes[i].MachineId] = &newProxyInfo

			// 保存代理对应的机器id信息
			pr.proxyMapInfo[strNodeId] = mapMachineValue
		}
	}

	return true
}

/*
 * GetNodeIdProxy 根据节点地址和机器id获取对应的代理信息，如果机器id传入空串，将会获取该地址对应的所有端的代理信息
 *
 * @param	clientNode	*ProxyInfo			代理信息
 * @return	exist		bool				是否存在该节点对应的代理信息
 * @return	res			[]*ProxyInfo		代理信息列表
 */
func (pr *ProxyData) GetNodeIdProxy(clientNode *ProxyInfo) (exist bool, res []*ProxyInfo) {
	// 1. 判断是否初始化过，如果没有初始化，则返回false
	if pr == nil {
		return false, nil
	}
	if clientNode.NodeId == nil || len(*clientNode.NodeId) == 0 {
		return false, nil
	}

	// 2. 加读锁
	pr.optLock.RLock()
	defer pr.optLock.RUnlock()

	// 3. 获取目标节点地址的base58字符串
	strNodeId := utils.Bytes2string(*clientNode.NodeId)

	// 4. 根据节点地址, 获取机器id对应的设备信息
	machineIdInfo, exist := pr.proxyMapInfo[strNodeId]
	if !exist || machineIdInfo == nil {
		return false, nil
	}

	// 构建返回的参数
	res = make([]*ProxyInfo, 0)

	// 5. 根据设备信息, 获取代理信息
	if clientNode.MachineId != "" {
		proxyInfo, exist := machineIdInfo[clientNode.MachineId]
		if proxyInfo == nil || !exist {
			return false, nil
		}

		// 判断代理信息的合法性
		if proxyInfo.ProxyId == nil || len(*proxyInfo.ProxyId) == 0 {
			return false, nil
		}

		res = append(res, proxyInfo)
	} else {
		// 获取节点地址所有的代理信息
		for k := range machineIdInfo {
			proxyInfo := machineIdInfo[k]
			if proxyInfo == nil || proxyInfo.ProxyId == nil || len(*proxyInfo.ProxyId) == 0 {
				continue
			}

			res = append(res, proxyInfo)
		}
	}

	// 7. 返回
	return true, res
}

/*
 * GetNodeIdProxy2 根据节点地址和机器id获取对应的代理信息，如果机器id传入空串，将会获取该地址对应的所有端的代理信息
 *
 * @param	nodeId		*AddressNet			节点地址
 * @param	machineId	string				机器id，如果机器id传入空串，将会获取该地址对应的所有端的代理信息
 * @return	exist		bool				是否存在该节点对应的代理信息
 * @return	res			[]*ProxyInfo		代理信息列表
 */
func (pr *ProxyData) GetNodeIdProxy2(nodeId *nodeStore.AddressNet, machineId string) (exist bool, res []*ProxyInfo) {
	// 1. 判断是否初始化过，如果没有初始化，则返回false
	if pr == nil {
		return false, nil
	}
	if nodeId == nil || len(*nodeId) == 0 {
		return false, nil
	}

	// 2. 加读锁
	pr.optLock.RLock()
	defer pr.optLock.RUnlock()

	// 3. 获取目标节点地址的base58字符串
	strNodeId := utils.Bytes2string(*nodeId)

	// 4. 根据节点地址, 获取机器id对应的设备信息
	machineIdInfo, exist := pr.proxyMapInfo[strNodeId]
	if !exist || machineIdInfo == nil {
		return false, nil
	}

	// 构建返回的参数
	res = make([]*ProxyInfo, 0)

	// 5. 根据设备信息, 获取代理信息
	if machineId != "" {
		proxyInfo, exist := machineIdInfo[machineId]
		if proxyInfo == nil || !exist {
			return false, nil
		}

		// 判断代理信息的合法性
		if proxyInfo.ProxyId == nil || len(*proxyInfo.ProxyId) == 0 {
			return false, nil
		}

		res = append(res, proxyInfo)
	} else {
		// 获取节点地址所有的代理信息
		for k := range machineIdInfo {
			proxyInfo := machineIdInfo[k]
			if proxyInfo == nil || proxyInfo.ProxyId == nil || len(*proxyInfo.ProxyId) == 0 {
				continue
			}

			res = append(res, proxyInfo)
		}
	}

	// 6. 返回
	return true, res
}

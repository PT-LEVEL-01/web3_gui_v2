package proxyrec

import (
	"context"
	"sync"
	"time"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

// 代理记录信息
type ProxyCache struct {
	proxyMapInfo map[string]map[string]ProxyInfo // 代理信息记录map key:string=客户端id;value:map={key:string=机器ID;value:int64=版本号;};
	optLock      sync.Mutex                      // 操作锁
	ctx          context.Context                 // 上下文
}

/*
 * NewProxyData 创建代理数据对象
 */
func NewProxyCache(ctx context.Context) *ProxyCache {
	pc := ProxyCache{}
	pc.proxyMapInfo = make(map[string]map[string]ProxyInfo)
	pc.ctx = ctx

	return &pc
}

/*
 * AddOrUpdateProxyRec 添加或更新代理记录信息
 *
 * @param	nodes		[]ProxyInfo			代理信息列表
 * @param	proxyId		Address				代理节点地址
 * @param	recvId		Address				存储节点地址
 * @return	success		bool				是否添加成功
 */
func (pc *ProxyCache) AddOrUpdateProxyRec(nodes []*ProxyInfo) bool {
	// utils.Log.Warn().Msgf("[存储节点:%s]添加或更新代理记录信息, 代理地址:%s", recvId.B58String(), proxyId.B58String())
	// 1. 判断是否初始化
	if pc == nil {
		utils.Log.Error().Msgf("还未初始化!!!!!!")
		return false
	}

	// 2. 判断参数是否有效
	if len(nodes) == 0 {
		utils.Log.Error().Msgf("代理信息为空!!!!!!")
		return false
	}

	// 3. 检测是否退出
	select {
	case <-pc.ctx.Done():
		utils.Log.Error().Msgf("应用结束!!!!!!")
		return false
	default:
	}

	// 4. 加写锁
	pc.optLock.Lock()
	defer pc.optLock.Unlock()

	// 5. 获取当前时间, 用以记录同步时间, 进而方便维护代理的处理
	curTime := time.Now()

	// 6. 遍历代理列表, 依次进行处理
	for i := range nodes {
		if nodes[i] == nil {
			continue
		}
		// 6.1 获取节点地址的base58字符串
		strNodeId := utils.Bytes2string(*nodes[i].NodeId)

		// utils.Log.Warn().Msgf("[存储节点:%s] 代理地址:%s 客户端:%s mId:%s version:%d", recvId.B58String(), proxyId.B58String(), nodes[i].NodeId.B58String(), nodes[i].MachineId, nodes[i].Version)

		// 6.2 获取机器id对应的map信息
		machineIdInfo, exist := pc.proxyMapInfo[strNodeId]
		if exist && machineIdInfo != nil {
			// 存在, 则更新相关信息
			// 查询机器id map对应的数据是否存在
			proxyInfo, exist2 := machineIdInfo[nodes[i].MachineId]
			if exist2 {
				// 只有版本号不大于更新信息的版本号时，才允许更新
				if proxyInfo.Version <= nodes[i].Version {
					// 存在, 则更新代理、同步时间和版本号
					proxyInfo.ProxyId = nodes[i].ProxyId
					proxyInfo.syncTime = curTime
					proxyInfo.Version = nodes[i].Version

					// 更新
					machineIdInfo[nodes[i].MachineId] = proxyInfo
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
				machineIdInfo[nodes[i].MachineId] = newProxyInfo
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
			mapMachineValue := make(map[string]ProxyInfo)
			// 保存机器id map对应的数据
			mapMachineValue[nodes[i].MachineId] = newProxyInfo

			// 保存代理对应的机器id信息
			pc.proxyMapInfo[strNodeId] = mapMachineValue
		}
	}

	return true
}

/*
 * GetNodeIdProxy 根据节点地址和机器id获取对应的缓存代理信息，如果机器id传入空串，将会获取该地址对应的所有端的代理信息
 *
 * @param	nodeId		*AddressNet			客户端节点id
 * @param	machineId	string				客户端机器id, 如果传入空串, 将获取地址对应所有machineid的客户端地址
 * @return	exist		bool				是否存在该节点对应的代理信息
 * @return	res			[]*ProxyInfo		代理信息列表
 */
func (pc *ProxyCache) GetNodeIdProxy(nodeId *nodeStore.AddressNet, machineId string) (exist bool, res []*ProxyInfo) {
	// 1. 判断是否初始化过，如果没有初始化，则返回false
	if pc == nil {
		return false, nil
	}
	if nodeId == nil || len(*nodeId) == 0 {
		return false, nil
	}

	// 2. 加读锁
	pc.optLock.Lock()
	defer pc.optLock.Unlock()

	// 3. 获取目标节点地址的base58字符串
	strNodeId := utils.Bytes2string(*nodeId)

	// 4. 根据节点地址, 获取机器id对应的设备信息
	machineIdInfo, exist := pc.proxyMapInfo[strNodeId]
	if !exist {
		return false, nil
	}

	// 构建返回的参数
	res = make([]*ProxyInfo, 0)

	curTime := time.Now()

	// 5. 根据设备信息, 获取代理信息
	if machineId != "" {
		proxyInfo, exist := machineIdInfo[machineId]
		if !exist {
			return false, nil
		}

		// 判断代理信息的合法性
		if proxyInfo.ProxyId == nil || len(*proxyInfo.ProxyId) == 0 {
			return false, nil
		}

		// 如果长时间没有同步到数据，则删除对应的记录信息
		if curTime.Sub(proxyInfo.syncTime) >= CacheValidTime {
			delete(machineIdInfo, proxyInfo.MachineId)
			return false, nil
		}

		res = append(res, &proxyInfo)
	} else {
		// 获取节点地址所有的代理信息
		for k := range machineIdInfo {
			proxyInfo := machineIdInfo[k]
			if proxyInfo.ProxyId == nil || len(*proxyInfo.ProxyId) == 0 {
				continue
			}

			// 如果长时间没有同步到数据，则删除对应的记录信息
			if curTime.Sub(proxyInfo.syncTime) >= CacheValidTime {
				delete(machineIdInfo, proxyInfo.MachineId)
				return false, nil
			}

			res = append(res, &proxyInfo)
		}
	}

	// 6. 返回
	return true, res
}

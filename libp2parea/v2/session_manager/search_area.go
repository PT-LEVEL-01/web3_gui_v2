package session_manager

import (
	"bytes"
	ma "github.com/multiformats/go-multiaddr"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
迭代查找自己的域网络节点
*/
func (this *SessionManager) searchArea() {
	//this.Log.Info().Str("开始查找自己的域网络", "start").Send()
	//检查自己是否已经在指定域中
	in := this.checkAreaNetOnlineWAN()
	if in {
		//this.Log.Info().Str("开始查找自己的域网络", "end").Send()
		//在指定的域中，就没必要再查找了
		return
	}

	findNodeInfos := make(map[string]*nodeStore.NodeInfo) //保存已经配置的域名称
	total := 0
	for {
		//this.Log.Info().Str("开始查找自己的域网络", "11111111111").Send()
		newFindNodeInfos := make(map[string]*nodeStore.NodeInfo)
		//oldTotal := len(countNodes)
		//询问所有的连接，自己的逻辑域名称有哪些
		ss := this.sessionEngine.GetSessionAll()
		var ERRout utils.ERROR
		for _, sOne := range ss {
			nodesWAN, nodesLAN, ERR := this.searchAreaIP_net(sOne)
			if ERR.CheckFail() {
				ERRout = ERR
				continue
			}
			//查找目标域节点
			have, targetNodeInfo := utils.ContainSliceFunc(&nodesWAN, func(one *nodeStore.NodeInfo) bool {
				return bytes.Equal(one.AreaName, this.nodeManager.NodeSelf.AreaName)
			})
			//已经找到目标域节点
			if have {
				addrInfos := append(targetNodeInfo.GetMultiaddrWAN(), targetNodeInfo.GetMultiaddrLAN()...)
				addrs := make([]ma.Multiaddr, 0, len(addrInfos))
				for _, addrInfoOne := range addrInfos {
					addrs = append(addrs, addrInfoOne.Multiaddr)
				}
				//要找到节点，并且能连上才退出
				ok := this.syncConnectNet(addrs)
				if ok {
					return
				}
			}
			//未找到目标域
			//保存到新节点列表中
			for _, one := range append(nodesWAN, nodesLAN...) {
				newFindNodeInfos[utils.Bytes2string(one.IdInfo.Id.Data())] = &one
			}
		}

		//this.Log.Info().Str("开始查找自己的域网络", "11111111111").Send()
		//如果存在有的节点发送超时，则重新发送
		if ERRout.CheckFail() {
			continue
		}

		//看查找到的节点是否有变化
		newFindNodeInfosTemp := make(map[string]*nodeStore.NodeInfo)
		for k, v := range findNodeInfos {
			newFindNodeInfosTemp[k] = v
		}
		for k, v := range newFindNodeInfos {
			newFindNodeInfosTemp[k] = v
		}

		//this.Log.Info().Str("开始查找自己的域网络", "11111111111").Send()
		//对比变化
		equal := utils.EqualMapFunc(&newFindNodeInfosTemp, &findNodeInfos, func(a, b **nodeStore.NodeInfo) bool {
			return bytes.Equal((*a).IdInfo.Id.Data(), (*b).IdInfo.Id.Data())
		})
		if equal {
			//相同，无变化
			total++
			//this.Log.Info().Str("开始查找自己的域网络", "相同，无变化").Send()
		} else {
			//有变化，继续查找
			total = 0
			//this.Log.Info().Str("开始查找自己的域网络", "有变化").Send()
			//连接所有新节点
			for _, one := range newFindNodeInfos {
				addrInfos := append(one.GetMultiaddrWAN(), one.GetMultiaddrLAN()...)
				addrs := make([]ma.Multiaddr, 0, len(addrInfos))
				for _, addrInfoOne := range addrInfos {
					addrs = append(addrs, addrInfoOne.Multiaddr)
				}
				this.syncConnectNet(addrs)
			}
			findNodeInfos = newFindNodeInfosTemp
		}
		//未发现新节点多次后，需退出
		if total > 2 {
			break
		}
	}

	//this.Log.Info().Str("开始查找自己的域网络", "end").Send()
	return
}

/*
检查是否已经在指定域网络中
@return    bool    true=在网络中了;false=不在网络中;
*/
func (this *SessionManager) checkAreaNetOnlineWAN() bool {
	ss := this.sessionEngine.GetSessionAll()
	for _, sOne := range ss {
		//连接上，但还未注册的连接，不询问
		nodeInfo := GetNodeInfo(sOne)
		if nodeInfo == nil {
			//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
			continue
		}
		//获取可连接节点的公开端口
		addrInfo := GetSessionCanConnPort(sOne)
		if addrInfo == nil {
			continue
		}
		if bytes.Equal(nodeInfo.AreaName, this.nodeManager.NodeSelf.AreaName) {
			return true
		}
	}
	return false
}

/*
检查是否已经在指定域网络中
@return    bool    true=在网络中了;false=不在网络中;
*/
func (this *SessionManager) checkAreaNetOnlineLAN() bool {
	ss := this.sessionEngine.GetSessionAll()
	for _, sOne := range ss {
		//连接上，但还未注册的连接，不询问
		nodeInfo := GetNodeInfo(sOne)
		if nodeInfo == nil {
			//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
			continue
		}
		if bytes.Equal(nodeInfo.AreaName, this.nodeManager.NodeSelf.AreaName) {
			return true
		}
	}
	return false
}

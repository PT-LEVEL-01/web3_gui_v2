package virtual_node

import (
	"math/big"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

/*
添加文件共享
*/
func AddFileShare() {

}

/*
	发送搜索节点消息
*/
// func SendSearchAllMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) {
// 	// GetVnodeLogical()

// 	// msg, ok := message_center.SendP2pMsgHE(msgid, recvid, content)

// }

/*
找到最近的虚拟节点
这是给搜索磁力节点提供的方法，排除了Index为0的节点地址。
@nodeId         *AddressNetExtend        要查找的节点
@outId          *AddressNetExtend        排除一个节点
@includeSelf    bool                     是否包括自己
@includeIndex0	bool                     是否包含index为0(即发现节点)的节点信息
@blockAddr      []*nodeStore.AddressNet  黑名单地址
@return         查找到的节点id，可能为空
*/
func (this *VnodeManager) FindNearVnodeSearchVnode(nodeId, outId *AddressNetExtend, includeSelf, includeIndex0 bool, blockAddr map[string]int) AddressNetExtend {

	//获取本节点的所有逻辑节点
	vnodeMap := this.GetVnodeAll()
	//排除掉黑名单地址
	for k, v := range blockAddr {
		if v > engine.MaxRetryCnt {
			delete(vnodeMap, k)
		}
	}
	//包括自己，就添加自己的虚拟节点
	if includeSelf {
		if !includeIndex0 {
			//虚拟节点中index为0的是发现节点，SearchVnode操作时，排除这些节点
			for k, v := range vnodeMap {
				if v.Index == 0 {
					delete(vnodeMap, k)
				}
			}
		}
		vnodes := this.GetVnodeSelf()
		for _, v := range vnodes {
			if !includeIndex0 {
				//虚拟节点中index为0的是发现节点，SearchVnode操作时，排除这些节点
				if v.Index == 0 {
					continue
				}
			}
			vnodeMap[utils.Bytes2string(v.Vid)] = v
		}
	} else {
		//不包括自己，逻辑节点中有可能包括自己节点，则删除自己节点
		vnodes := this.GetVnodeSelf()
		for _, v := range vnodes {
			delete(vnodeMap, utils.Bytes2string(v.Vid))
		}
	}

	//有排除的节点，不添加
	if outId != nil {
		delete(vnodeMap, utils.Bytes2string(*outId))
	}

	//构建kad算法，添加逻辑节点
	kl := nodeStore.NewKademlia(len(vnodeMap))
	for _, v := range vnodeMap {
		// utils.Log.Info().Msgf("kad add:%s", v.Vid.B58String())
		kl.Add(new(big.Int).SetBytes(v.Vid))
	}

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	targetIdBs := targetId.Bytes()
	mh := AddressNetExtend(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	// utils.Log.Info().Msgf("kad find:%s", mh.B58String())
	return mh
}

/*
找到最近的虚拟节点
@nodeId         要查找的节点
@outId          排除一个节点
@includeSelf    是否包括自己
@return         查找到的节点id，可能为空
*/
func (this *VnodeManager) FindNearVnodeP2P(nodeId, outId *AddressNetExtend, includeSelf bool, blockAddr map[string]int) AddressNetExtend {

	//获取本节点的所有逻辑节点
	vnodeMap := this.GetVnodeAll()
	//排除掉黑名单地址
	for k, v := range blockAddr {
		if v > engine.MaxRetryCnt {
			delete(vnodeMap, k)
		}
	}
	//包括自己，就添加自己的虚拟节点
	if includeSelf {
		//虚拟节点中index为0的是发现节点，SearchVnode操作时，排除这些节点
		// for k, v := range vnodeMap {
		// 	if v.Index == 0 {
		// 		delete(vnodeMap, k)
		// 	}
		// }
		vnodes := this.GetVnodeSelf()
		for _, v := range vnodes {
			//虚拟节点中index为0的是发现节点，SearchVnode操作时，排除这些节点
			if v.Index == 0 {
				continue
			}
			vnodeMap[utils.Bytes2string(v.Vid)] = v
		}
	} else {
		//不包括自己，逻辑节点中有可能包括自己节点，则删除自己节点
		vnodes := this.GetVnodeSelf()
		for _, v := range vnodes {
			delete(vnodeMap, utils.Bytes2string(v.Vid))
		}
	}

	//有排除的节点，不添加
	if outId != nil {
		delete(vnodeMap, utils.Bytes2string(*outId))
	}

	//构建kad算法，添加逻辑节点
	kl := nodeStore.NewKademlia(len(vnodeMap))
	for _, v := range vnodeMap {
		// utils.Log.Info().Msgf("kad add:%s", v.Vid.B58String())
		kl.Add(new(big.Int).SetBytes(v.Vid))
	}

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	targetIdBs := targetId.Bytes()
	mh := AddressNetExtend(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	// utils.Log.Info().Msgf("kad find:%s", mh.B58String())
	return mh
}

/*
在自己的虚拟节点中找到最近的虚拟节点
@nodeId         要查找的节点
*/
func (this *VnodeManager) FindNearVnodeInSelf(nodeId *AddressNetExtend) *AddressNetExtend {
	vnodeinfo := this.GetVnodeSelf()

	//构建kad算法，添加逻辑节点
	kl := nodeStore.NewKademlia(len(vnodeinfo))
	for _, v := range vnodeinfo {
		kl.Add(new(big.Int).SetBytes(v.Vid))
	}

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}

	targetIdBs := targetId.Bytes()
	mh := AddressNetExtend(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	return &mh
}

/*
 * 在自己的虚拟节点中找到最近的虚拟节点，指定是否查找index为0的节点信息
 * @nodeId          要查找的节点
 * @includeIndex0	是否要包含index为0的节点进行查询标识
 */
func (this *VnodeManager) FindNearVnodeInSelfAppIndex0(nodeId *AddressNetExtend, includeIndex0 bool) AddressNetExtend {
	vnodeinfo := this.GetVnodeSelf()

	//构建kad算法，添加逻辑节点
	kl := nodeStore.NewKademlia(len(vnodeinfo))
	for _, v := range vnodeinfo {
		// 如果指定不包含index为0的查询，则进行排除
		if !includeIndex0 && v.Index == 0 {
			continue
		}
		kl.Add(new(big.Int).SetBytes(v.Vid))
	}

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}

	targetIdBs := targetId.Bytes()
	mh := AddressNetExtend(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	return mh
}

/*
 * 根据目标节点，返回排序后的虚拟节点地址列表
 * @nodeId         要查找的节点
 * @outId          排除一个节点
 * @includeSelf    是否包括自己
 * @includeIndex0  是否包含index为0(即发现节点)的节点信息
 * @return         查找到的节点id，可能为空
 */
func (this *VnodeManager) FindNearVnodesSearchVnode(nodeId, outId *AddressNetExtend, includeSelf, includeIndex0 bool) (res []AddressNetExtend) {
	//获取本节点的所有逻辑节点
	vnodeMap := this.GetVnodeAll()
	//包括自己，就添加自己的虚拟节点
	if includeSelf {
		if !includeIndex0 {
			//虚拟节点中index为0的是发现节点，SearchVnode操作时，排除这些节点
			for k, v := range vnodeMap {
				if v.Index == 0 {
					delete(vnodeMap, k)
				}
			}
		}
		vnodes := this.GetVnodeSelf()
		for _, v := range vnodes {
			if !includeIndex0 {
				//虚拟节点中index为0的是发现节点，SearchVnode操作时，排除这些节点
				if v.Index == 0 {
					continue
				}
			}
			vnodeMap[utils.Bytes2string(v.Vid)] = v
		}
	} else {
		//不包括自己，逻辑节点中有可能包括自己节点，则删除自己节点
		vnodes := this.GetVnodeSelf()
		for _, v := range vnodes {
			delete(vnodeMap, utils.Bytes2string(v.Vid))
		}
	}

	//有排除的节点，不添加
	if outId != nil {
		delete(vnodeMap, utils.Bytes2string(*outId))
	}

	//构建kad算法，添加逻辑节点
	kl := nodeStore.NewKademlia(len(vnodeMap))
	for _, v := range vnodeMap {
		// utils.Log.Info().Msgf("kad add:%s", v.Vid.B58String())
		kl.Add(new(big.Int).SetBytes(v.Vid))
	}

	var targetIds []*big.Int
	if nodeId != nil {
		targetIds = kl.Get(new(big.Int).SetBytes(*nodeId))
	} else {
		targetIds = kl.Get(new(big.Int).SetBytes(nil))
	}
	if len(targetIds) == 0 {
		return nil
	}

	res = make([]AddressNetExtend, 0, len(targetIds))

	// 依次处理逻辑地址
	for i := range targetIds {
		targetId := targetIds[i]
		if targetId == nil {
			continue
		}
		targetIdBs := targetId.Bytes()
		mh := AddressNetExtend(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
		res = append(res, mh)
		// utils.Log.Info().Msgf("kad find:%s", mh.B58String())
	}

	return
}

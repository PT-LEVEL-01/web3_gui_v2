package message_center

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/big"
	"sort"
	"time"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

/*
检查该消息是否是自己的
不是自己的则自动转发出去
@return    bool    是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    error   发送错误信息
*/
func (this *Message) Send(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, timeout time.Duration) (bool, error) {
	//失败发送节点列表
	var blockAddr = make(map[string]int)
	// 判断是否需要经过代理节点转发
	isProxy := this.routerProxy(version, nodeManager, sessionEngine, vnc, nil, timeout)
	if isProxy {
		return true, this.sendRaw(version, sessionEngine, this.Head.SenderProxyId, timeout, nodeManager.NodeSelf.IdInfo.Id, nil, nodeManager.AreaNameSelf)
	}

	routerType := this.checkRouterType()
	if routerType == version_node_search_onebyone {
		//0.判断是否由自己处理
		if this.Head.RecvId != nil && bytes.Equal(*this.Head.RecvId, nodeManager.NodeSelf.IdInfo.Id) {
			//utils.Log.Info().Msgf("++++++++++++++++++++++++由自己来处理， 自己addr %s", nodeManager.NodeSelf.IdInfo.Id.B58String())
			return false, nil
		}
		//1.解析消息内容
		if err := this.ParserContentProto(); err != nil {
			// fmt.Println(err)
			return true, err
		}
		//2.获取搜索节点个数
		searchNum := 1
		if this.Body.Content != nil && len(*this.Body.Content) != 0 {
			searchNum = int(binary.LittleEndian.Uint16(*this.Body.Content))
		}

		//3.查询目标多节点
		targetIds, err := this.routerSearchNodeOnebyone2(version, nodeManager, sessionEngine, vnc, blockAddr, searchNum)
		if err != nil {
			return true, err
		}

		//4.1自己是客户端节点的情况下
		if !nodeManager.NodeSelf.GetIsSuper() {
			for i := 0; i < len(targetIds); i++ {
				if !bytes.Equal(targetIds[i], nodeManager.NodeSelf.IdInfo.Id) {
					tId := targetIds[i]
					err = this.sendRaw(version, sessionEngine, &tId, timeout, nodeManager.NodeSelf.IdInfo.Id, nil, nodeManager.AreaNameSelf)
				}
				return true, err
			}

		}

		//4.2自己处理或异步发送给目标多节点
		var hasSelf bool
		for n := 0; n < len(targetIds); n++ {
			//不管最佳命中点是自己还是别人都照常转发
			tId := targetIds[n]
			if bytes.Equal(tId, nodeManager.NodeSelf.IdInfo.Id) {
				hasSelf = true
			} else {
				// utils.Log.Info().Msgf("@@@@@@@@@@@@ self %s to %s", nodeManager.NodeSelf.IdInfo.Id.B58String(), tId.B58String())
				go this.sendRaw(version, sessionEngine, &tId, timeout, nodeManager.NodeSelf.IdInfo.Id, nil, nodeManager.AreaNameSelf)
			}
		}
		if hasSelf {
			return false, nil
		}
	} else {
		allSessionCnt := sessionEngine.GetSessionCnt(nodeManager.AreaNameSelf) // 最大重试次数改为最低10次
		maxRouterCnt := allSessionCnt * engine.MaxRetryCnt
		if maxRouterCnt < 10 {
			maxRouterCnt = 10
		}
		// 统计尝试连接的数量
		var routerCnt int64

		for {
			// 增加重试次数
			routerCnt++
			// 如果尝试次数大于总连接数时, 直接返回
			if routerCnt > maxRouterCnt {
				return false, config.Error_retry_over
			}

			isOther, targetId, err, force := this.router(version, nodeManager, sessionEngine, vnc, nil, timeout, blockAddr, routerType)
			if err != nil {
				return true, err
			}
			if !isOther {
				return false, nil
			}

			if err := this.sendRaw(version, sessionEngine, targetId, timeout, nodeManager.NodeSelf.IdInfo.Id, nil, nodeManager.AreaNameSelf); err != nil {
				if err == config.ERROR_get_node_conn_fail {
					return false, err
				}
				// routerType := this.checkRouterType()
				// utils.Log.Warn().Msgf("转发消息出错:  t :%s type %d %s", targetId.B58String(), routerType, err.Error())
				// if targetId != nil {
				// 	utils.Log.Warn().Msgf("发送给 %s 消息出错:%s", targetId.B58String(), err.Error())
				// }
				if targetId != nil {
					blockAddr[utils.Bytes2string(*targetId)]++
				}

				// for _, v := range blockAddr {
				// 	utils.Log.Info().Msgf("黑名单！！！ %s self: %s", v.B58String(), nodeManager.NodeSelf.IdInfo.Id.B58String())
				// }
				//当为发送消息终点时，farce为true，即使发送出错也不继续中转
				if force {
					//utils.Log.Info().Msgf("推出！ 推出！ 发送给 %s 消息出错:%s", targetId.B58String(), err.Error())
					return false, err
				}
			} else {
				break
			}
		}
	}
	return true, nil
}

/*
路由搜索超级节点消息
@return    bool    是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    error   发送错误信息
*/
func (this *Message) sendRaw(version uint64, sessionEngine *engine.Engine, targetId *nodeStore.AddressNet, timeout time.Duration, from nodeStore.AddressNet, mh []byte, areaName []byte) error {
	if targetId == nil {
		return config.ERROR_get_node_conn_fail
	}

	sessions, ok := sessionEngine.GetSessionAll(areaName, utils.Bytes2string(*targetId))
	if ok {
		var mheadBs []byte
		if mh != nil {
			mheadBs = mh
		} else {
			mheadBs = this.Head.Proto()
		}
		mbodyBs, err := this.Body.Proto()
		if err != nil {
			return err
		}
		sendChan := make(chan error, len(sessions))
		var bNeedCheckMachineID bool
		if this.Head.Accurate && this.Head.RecvMachineID != "" && this.Head.RecvId != nil && bytes.Equal(*this.Head.RecvId, *targetId) {
			bNeedCheckMachineID = true
		}
		var sendCnt int
		for k := range sessions {
			if bNeedCheckMachineID && sessions[k].GetMachineID() != this.Head.RecvMachineID {
				// utils.Log.Error().Msgf("RecvMachineID 不相同, tId:%s, machineID:%s", targetId.B58String(), this.Head.RecvMachineID)
				continue
			}
			sendCnt++

			go func(k int) {
				err := sessions[k].Send(version, &mheadBs, &mbodyBs, timeout)
				sendChan <- err
			}(k)
		}

		for i := 0; i < sendCnt; i++ {
			err = <-sendChan
			if err == nil {
				break
			}
		}
		// utils.Log.Info().Msgf("发送消息 222222222222")
		return err
	}

	return config.ERROR_get_node_conn_fail
}

/*
检查该消息是否是自己的
不是自己的则自动转发出去
@return    bool    是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    error   发送错误信息
*/
func (this *Message) Forward(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration) (bool, error) {
	if this.DataPlus != nil && this.Head.RecvId != nil && !bytes.Equal(*this.Head.RecvId, nodeManager.NodeSelf.IdInfo.Id) {
		flood.RegisterRequest(utils.Bytes2string(*this.DataPlus))
		defer flood.RemoveRequest(utils.Bytes2string(*this.DataPlus))
		//utils.Log.Info().Msgf("111 开启等待转发消息回复 self : %s %v", nodeManager.NodeSelf.IdInfo.Id.B58String(), *this.DataPlus)
	}
	// if this.Head.OneByOne {
	//utils.Log.Info().Msgf("Forward self11: %s  %v", nodeManager.NodeSelf.IdInfo.Id.B58String(), *this.DataPlus)
	// }

	//失败节点列表
	var blockAddr = make(map[string]int)
	routerType := this.checkRouterType()
	if routerType == version_node_search_onebyone {
		//0.判断是否由自己处理
		if this.Head.RecvId != nil && bytes.Equal(*this.Head.RecvId, nodeManager.NodeSelf.IdInfo.Id) {
			return false, nil
		}
		//1.解析消息内容
		if err := this.ParserContentProto(); err != nil {
			// fmt.Println(err)
			return true, err
		}
		//2.获取搜索节点个数
		searchNum := 1
		if this.Body.Content != nil && len(*this.Body.Content) != 0 {
			searchNum = int(binary.LittleEndian.Uint16(*this.Body.Content))
		}

		//3.查询目标多节点
		targetIds, err := this.routerSearchNodeOnebyone2(version, nodeManager, sessionEngine, vnc, blockAddr, searchNum)
		if err != nil {
			return true, err
		}

		//4.1判断自己是否是第二次中继节点，如果是第二次中继节点，直接发给最终目标点
		if fNode := nodeManager.FindNode(from); fNode != nil {
			if fNode.GetIsSuper() {
				return false, nil
			}
		}

		//4.2自己处理或异步发送给目标多节点
		var hasSelf bool
		for n := 0; n < len(targetIds); n++ {
			//最佳命中点是自己，直接返回处理，不需要发给别人
			//命中的其他节点有是自己的直接跳过不必处理
			if bytes.Equal(targetIds[n], nodeManager.NodeSelf.IdInfo.Id) {
				hasSelf = true
				continue
			}

			//发给最终处理的点时候需要修改RecvId
			tId := targetIds[n]
			go this.forwardRaw(version, sessionEngine, &tId, timeout, *from, nil, nodeManager.AreaNameSelf)
		}
		if hasSelf {
			return false, nil
		}

	} else {
		allSessionCnt := sessionEngine.GetSessionCnt(nodeManager.AreaNameSelf) // 最大重试次数改为最低10次
		maxRouterCnt := allSessionCnt * engine.MaxRetryCnt
		if maxRouterCnt < 10 {
			maxRouterCnt = 10
		}
		// 统计尝试连接的数量
		var routerCnt int64

		for {
			// 增加重试次数
			routerCnt++
			// 如果尝试次数大于总连接数时, 直接返回
			if routerCnt > maxRouterCnt {
				return false, config.Error_retry_over
			}

			isOther, targetId, err, force := this.router(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr, routerType)
			if err != nil {
				return true, err
			}
			if !isOther {
				return false, nil
			}

			if err := this.forwardRaw(version, sessionEngine, targetId, timeout, nodeManager.NodeSelf.IdInfo.Id, nil, nodeManager.AreaNameSelf); err != nil {
				if err == config.ERROR_get_node_conn_fail {
					return false, config.ERROR_send_to_sender
				}

				if this.checkRouterType() == version_vnode_search_onebyone {
					if targetId != nil && this.Head.RecvId != nil {
						blockAddr[utils.Bytes2string(*this.Head.RecvId)]++
					}
				} else {
					if targetId != nil {
						blockAddr[utils.Bytes2string(*targetId)]++
					}
				}

				// routerType := this.checkRouterType()
				// utils.Log.Warn().Msgf("转发消息出错: self : %s t :%s type %d %s", from.B58String(), targetId.B58String(), routerType, err.Error())
				// for i, v := range blockAddr {
				// 	utils.Log.Info().Msgf("黑名单 %d: %s self : %s", i, v.B58String(), from.B58String())
				// }

				//当为发送消息终点时，force为true，即使发送出错也不继续中转
				if force {
					return false, config.ERROR_send_to_sender
				}
			} else {
				// if this.DataPlus != nil {
				// 	//utils.Log.Info().Msgf("222 消息发送成功，等待下一个节点回复 self : %s DatasPlus : %v", nodeManager.NodeSelf.IdInfo.Id.B58String(), *this.DataPlus)
				// 	//start := time.Now().Unix()
				// 	waitr := []byte(*targetId)
				// 	waitr = append(waitr, *this.DataPlus...)
				// 	_, err := flood.WaitResponseItr(utils.Bytes2string(waitr), 2*time.Second)
				// 	if err != nil && err == config.ERROR_not_in_waitRequest {
				// 		//utils.Log.Warn().Msgf(" %d self %s forward messge err: %s DataPlus : %v", time.Now().Unix()-start, nodeManager.NodeSelf.IdInfo.Id.B58String(), err.Error(), *this.DataPlus)
				// 		return true, nil
				// 	} else if err != nil {
				// 		if targetId != nil {
				// 			//utils.Log.Info().Msgf("===================> %s", err.Error())
				// 			blockAddr = append(blockAddr, targetId)
				// 		}
				// 	} else {
				// 		//utils.Log.Info().Msgf("333 收到下一个节点接收回复 self : %s DataPlus : %v", nodeManager.NodeSelf.IdInfo.Id.B58String(), *this.DataPlus)
				// 		return true, nil
				// 	}
				// } else {
				break
				//}
			}
		}
	}
	return true, nil
}

/*
路由搜索超级节点消息
@return    bool    是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    error   发送错误信息
*/
func (this *Message) forwardRaw(version uint64, sessionEngine *engine.Engine, targetId *nodeStore.AddressNet, timeout time.Duration, from nodeStore.AddressNet, mh []byte, areaName []byte) error {
	if targetId == nil {
		return config.ERROR_get_node_conn_fail
	}

	sessions, ok := sessionEngine.GetSessionAll(areaName, utils.Bytes2string(*targetId))
	if ok {
		var mheadBs []byte
		if mh != nil {
			mheadBs = mh
		} else {
			mheadBs = this.Head.Proto()
		}
		var err error
		sendChan := make(chan error, len(sessions))
		var bNeedCheckMachineID bool
		if this.Head.Accurate && this.Head.RecvMachineID != "" && this.Head.RecvId != nil && bytes.Equal(*this.Head.RecvId, *targetId) {
			bNeedCheckMachineID = true
		}
		var sendCnt int
		for k := range sessions {
			if bNeedCheckMachineID && sessions[k].GetMachineID() != this.Head.RecvMachineID {
				continue
			}
			sendCnt++

			go func(k int) {
				err := sessions[k].Send(version, &mheadBs, this.DataPlus, timeout)
				sendChan <- err
			}(k)
		}

		for i := 0; i < sendCnt; i++ {
			err = <-sendChan
			if err == nil {
				break
			}
		}
		// utils.Log.Info().Msgf("发送消息 222222222222 targetId:%s", targetId.B58String())
		return err
	} else {
		return config.ERROR_get_node_conn_fail
	}
}

/*
检查该消息是否是自己的
不是自己的则自动转发出去
@return    bool    是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    error   发送错误信息
*/
func (this *Message) router(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet,
	timeout time.Duration, blockAddr map[string]int, routerType int) (bool, *nodeStore.AddressNet, error, bool) {
	// defer fmt.Println("------------", version, this.Body.MessageId, this.Head)
	// utils.Log.Info().Msgf("1111111111111")
	// utils.Log.Info().Msgf("打印message:%+v", this)

	isOther := false                   //是否发送给其他人
	var targetId *nodeStore.AddressNet //
	var err error                      //
	var force bool                     //找到指定人时，出错时强制退出

	// bs, _ := this.Head.JSON()
	//utils.Log.Info().Msgf("路由消息类型:%d self:%s", routerType, vnc.GetVnodeDiscover().Vnode.Vid.B58String()) //this.Head.Sender.B58String()+"_"+this.Head.RecvId.B58String())
	switch routerType {
	case version_search_super:
		isOther, targetId, err, force = this.routerSearchSuper(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
	case version_p2p:
		isOther, targetId, err, force = this.routerP2P(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
	case version_vnode_search:
		// utils.Log.Info().Msgf("路由消息类型:%d self:%s %s", routerType, vnc.GetVnodeDiscover().Vnode.Vid.B58String(), string(bs))
		isOther, targetId, err, force = this.routerSearchVnode(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
	case version_vnode_p2pHE:
		// utils.Log.Info().Msgf("路由消息类型:%d self:%s %s", routerType, vnc.GetVnodeDiscover().Vnode.Vid.B58String(), string(bs))
		isOther, targetId, err, force = this.routerVnodeP2P(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
	case version_vnode_search_end:
		isOther, targetId, err, force = this.routerSearchVnodeEnd(version, nodeManager, sessionEngine, vnc, from, timeout, true, blockAddr)

	//#############给onebyone协议调用方法
	case version_p2p_onebyone:
		isOther, targetId, err, force = this.routerP2POnebyone(version, nodeManager, sessionEngine, vnc, from, blockAddr)
	case version_vnode_search_onebyone:
		for {
			isOther, targetId, err, force = this.routerSearchVnodeOnebyone(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
			if targetId != nil && !force && vnc.FindInVnodeinfoSelf(virtual_node.AddressNetExtend(*targetId)) != nil {
				//utils.Log.Info().Msgf("self : %s 找到跳转是自己 %s ,继续在router中路由", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
				continue
			}
			break
		}
	case version_node_search_onebyone:
		isOther, targetId, err, force = this.routerSearchNodeOnebyone(version, nodeManager, sessionEngine, vnc, from, blockAddr)
	}

	// if targetId != nil && this.Head.OneByOne {
	// 	utils.Log.Info().Msgf("\nzyz_路由 出 消息类型: %d self: %s  是否交其他节点处理: %t  targetId: %s \n   是否强制退出：%t\n    黑名单: %v\n onebyone %t", routerType, nodeManager.NodeSelf.IdInfo.Id.B58String(),
	// 		isOther, targetId.B58String(), force, blockAddr, this.Head.OneByOne)
	// } else if targetId == nil && this.Head.OneByOne {
	// 	utils.Log.Info().Msgf("\nzyz_路由 出 消息类型: %d self:%s  是否交其他节点处理: %t  targetId: nil \n   是否强制退出：%t\n    黑名单: %v\n onebyone %t", routerType, nodeManager.NodeSelf.IdInfo.Id.B58String(),
	// 		isOther, force, blockAddr, this.Head.OneByOne)
	// }
	// if err != nil && this.Head.OneByOne {
	// 	utils.Log.Info().Msgf("\nzyz_路由 出 消息类型: %d self: %s ===> err : %s", routerType, nodeManager.NodeSelf.IdInfo.Id.B58String(), err.Error())
	// }

	return isOther, targetId, err, force
}

/*
检查路由类型
*/
func (this *Message) checkRouterType() int {
	// version_neighbor            = 1  //邻居节点消息，本消息不转发
	// version_multicast           = 2  //广播消息
	// version_search_super        = 3  //搜索节点消息
	// version_search_all          = 4  //搜索节点消息
	// version_p2p                 = 5  //点对点消息
	// version_p2pHE               = 6  //点对点可靠传输加密消息
	// version_vnode_search        = 7  //搜索虚拟节点消息
	// version_vnode_p2pHE         = 8  //虚拟节点间点对点可靠传输加密消息
	// version_multicast_sync      = 9  //查询邻居节点的广播消息
	// version_multicast_sync_recv = 10 //查询广播消息的回应
	// version_vnode_search_end    = 11 //搜索最终接收者虚拟节点
	//#新增onebyone消息协议
	// version_p2p_onebyone              = 12  //点对点消息，onebyone规则路由
	// version_vnode_search_end_onebyone = 13  //搜索最终接收者虚拟节点，onebyone规则路由
	// version_vnode_p2pHE_onebyone      = 14  //虚拟节点间点对点可靠传输加密消息，onebyone规则路由
	// version_vnode_search_onebyone     = 15  //搜索虚拟节点消息，onebyone规则路由
	// version_search_super_onebyone     = 16  //搜索节点消息，onebyone规则路由
	if this.Head.SearchVnodeEndId != nil && len(*this.Head.SearchVnodeEndId) != 0 {
		return version_vnode_search_end
	}

	if this.Head.SenderVnode != nil && len(*this.Head.SenderVnode) != 0 &&
		this.Head.RecvVnode != nil && len(*this.Head.RecvVnode) != 0 {
		if this.Head.Accurate {
			return version_vnode_p2pHE
		} else {
			if this.Head.OneByOne {
				return version_vnode_search_onebyone
			} else {
				return version_vnode_search
			}
		}
	}

	//使用真实节点onebyone规则，查询节点
	if this.Head.OneByOne &&
		this.Head.Sender != nil && len(*this.Head.Sender) != 0 &&
		this.Head.RecvId != nil && len(*this.Head.RecvId) != 0 &&
		this.Head.SenderVnode == nil && this.Head.RecvVnode == nil {
		return version_node_search_onebyone
	}
	if this.Head.Sender != nil && len(*this.Head.Sender) != 0 &&
		this.Head.SenderSuperId != nil && len(*this.Head.SenderSuperId) != 0 &&
		this.Head.RecvId != nil && len(*this.Head.RecvId) != 0 &&
		this.Head.RecvSuperId != nil && len(*this.Head.RecvSuperId) != 0 {
		if this.Head.Accurate {
			return version_p2p
		} else {
			return version_search_super
		}
	}

	return 0
}

/*
路由搜索超级节点消息，带黑名单
@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    *nodeStore.AddressNet  根据规则选出的地址
@return    error                  发送错误信息
@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
*/
func (this *Message) routerSearchSuper(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet,
	timeout time.Duration, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {

	if from == nil {
		targetId := nodeManager.FindNearInSuper(this.Head.RecvId, nil, false, blockAddr)
		// utils.Log.Info().Msgf("转发给self:%s targetID:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
		return true, targetId, nil, false
	} else {
		targetId := nodeManager.FindNearInSuper(this.Head.RecvId, nil, true, blockAddr)
		if targetId != nil && bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *targetId) {
			// utils.Log.Info().Msgf("p2p收消息人就是自己self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
			return false, nil, nil, false
		}
		// utils.Log.Info().Msgf("转发给self:%s targetID:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
		return true, targetId, nil, false
	}

	// targetId := nodeManager.FindNearInSuper(this.Head.RecvId, from, true)
	// if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *targetId) {
	// 	// utils.Log.Info().Msgf("这个消息丢失了self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
	// 	// utils.Log.Info().Msgf("p2p收消息人就是自己self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
	// 	return false, nil, nil
	// }
	// utils.Log.Info().Msgf("转发给self:%s targetID:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
	// return true, targetId, nil
}

/*
路由搜索超级节点消息
@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    *nodeStore.AddressNet  选出的发送/中转地址
@return    error                  发送错误信息
@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
*/
func (this *Message) routerP2P(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	//p2p收消息人就是自己
	if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
		//		utils.Log.Info().Msgf("p2p收消息人就是自己self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
		return false, nil, nil, true
	}
	if this.Head.RecvProxyId != nil && len(*this.Head.RecvProxyId) != 0 {
		// 指定了接收者的代理节点
		if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvProxyId) {
			// 自己是接收者代理节点，因此需要确认本地是否持有接收者的连接信息
			// 接收者肯定不是自己，上面已做判断
			//utils.Log.Info().Msgf("qlw----我就是接收方的代理节点: self:%s, recv:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), this.Head.RecvId.B58String())
			_, ok := sessionEngine.GetSession(nodeManager.AreaNameSelf, utils.Bytes2string(*this.Head.RecvId))
			if ok {
				// 存在接收方的连接信息
				if blockAddr[utils.Bytes2string(*this.Head.RecvId)] <= engine.MaxRetryCnt {
					return true, this.Head.RecvId, nil, true
				}
			}

			// 没有接收方的连接信息，则报错，丢失消息
			return true, nil, errors.New("don't have recv node connection"), true
		}
		// 自己不是接收者的代理节点，查看自己是否有和代理节点直接连接的信息
		_, ok := sessionEngine.GetSession(nodeManager.AreaNameSelf, utils.Bytes2string(*this.Head.RecvProxyId))
		if ok {
			//utils.Log.Info().Msgf("p2p消息找到目标代理节点并发送出去了self:%s recv:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), this.Head.RecvProxyId.B58String())
			if blockAddr[utils.Bytes2string(*this.Head.RecvProxyId)] <= engine.MaxRetryCnt {
				return true, this.Head.RecvProxyId, nil, true
			}
		}
	} else {
		_, ok := sessionEngine.GetSession(nodeManager.AreaNameSelf, utils.Bytes2string(*this.Head.RecvId))
		if ok {
			// utils.Log.Info().Msgf("p2p消息找到目标节点并发送出去了self:%s recv:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), this.Head.RecvId.B58String())
			if blockAddr[utils.Bytes2string(*this.Head.RecvId)] <= engine.MaxRetryCnt {
				return true, this.Head.RecvId, nil, true
			}
		}
	}

	var targetId *nodeStore.AddressNet
	if from == nil {
		if this.Head.RecvProxyId == nil || len(*this.Head.RecvProxyId) == 0 {
			// 如果没有指定接收者代理节点，则直接根据接收方id获取当前最近的节点id
			targetId = nodeManager.FindNearInSuper(this.Head.RecvId, from, false, blockAddr)
		} else {
			// 指定了接收者代理节点，则根据接收方代理节点id获取最近的节点id
			targetId = nodeManager.FindNearInSuper(this.Head.RecvProxyId, from, false, blockAddr)
		}
		// utils.Log.Info().Msgf("转发给self:%s targetID:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
		return true, targetId, nil, false
	} else {
		if this.Head.RecvProxyId == nil || len(*this.Head.RecvProxyId) == 0 {
			// 如果没有指定接收者代理节点，则直接根据接收方id获取当前最近的节点id
			targetId = nodeManager.FindNearInSuper(this.Head.RecvId, from, true, blockAddr)
		} else {
			// 指定了接收者代理节点，则根据接收方代理节点id获取最近的节点id
			targetId = nodeManager.FindNearInSuper(this.Head.RecvProxyId, from, true, blockAddr)
		}
		if targetId != nil && bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *targetId) {
			// utils.Log.Info().Msgf("这个消息丢失了")
			// utils.Log.Info().Msgf("p2p收消息人就是自己self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
			return true, nil, nil, true
		}
		if targetId == nil {
			return true, nil, nil, true
		}
		// utils.Log.Info().Msgf("转发给self:%s targetID:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
		return true, targetId, nil, false
	}
}

/*
路由搜索超级节点消息，给onebyone规则调用
@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    *nodeStore.AddressNet  选出的发送/中转地址
@return    error                  发送错误信息
@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
*/
func (this *Message) routerP2POnebyone(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	//p2p收消息人就是自己
	if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
		//utils.Log.Info().Msgf("p2p收消息人就是自己self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
		return false, nil, nil, false
	}
	for _, one := range sessionEngine.GetAllDownUp(nodeManager.AreaNameSelf) {
		if bytes.Equal([]byte(one), *this.Head.RecvId) {
			//接收者地址时我的onebyone规则节点
			if blockAddr[utils.Bytes2string(*this.Head.RecvId)] <= engine.MaxRetryCnt {
				return true, this.Head.RecvId, nil, true
			}
		}
	}

	targetId := nodeManager.FindNearInSuperOnebyone(this.Head.RecvId, from, blockAddr, sessionEngine)
	return true, targetId, nil, false

}

func (this *Message) GetSortSession(allAddr []nodeStore.AddressNet, sessionEngine *engine.Engine) ([]nodeStore.AddressNet, int) {
	var addrKey []byte
	var sortedr []nodeStore.AddressNet
	var rec *big.Int
	sortLargeNum := 0

	//1. 给虚拟节点onebyone规则找 2. 给真实节点onebyone规则找
	if this.Head.RecvVnode != nil && len(*this.Head.RecvVnode) != 0 {
		rec = new(big.Int).SetBytes(*this.Head.RecvVnode)
	} else {
		rec = new(big.Int).SetBytes(*this.Head.RecvId)
	}

	sort.Sort(nodeStore.AddressBytes(allAddr))
	for i := range allAddr {
		addrKey = append(addrKey, allAddr[i]...)
	}
	addrkeyS := utils.Bytes2string(addrKey)

	if sorted, ok := sessionEngine.GetSortSessionByKey(addrkeyS); ok {
		for i := range sorted {
			if new(big.Int).SetBytes(sorted[i]).Cmp(rec) == 1 {
				sortLargeNum += 1
			}
			sortedr = append(sortedr, nodeStore.AddressNet(sorted[i]))
		}
		return sortedr, sortLargeNum
	}

	sortM := make(map[string]nodeStore.AddressNet)
	onebyone := new(nodeStore.IdDESC)
	var sortE []engine.AddressNet
	for _, one := range allAddr {
		sortM[one.B58String()] = one
		*onebyone = append(*onebyone, new(big.Int).SetBytes(one))
	}

	sort.Sort(onebyone)

	for i := 0; i < len(*onebyone); i++ {
		one := (*onebyone)[i]
		if one.Cmp(rec) == 1 {
			sortLargeNum += 1
		}
		IdBs := one.Bytes()
		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
		sortedr = append(sortedr, sortM[nodeStore.AddressNet(*IdBsP).B58String()])
		sortE = append(sortE, engine.AddressNet(sortM[nodeStore.AddressNet(*IdBsP).B58String()]))
	}
	sessionEngine.SetSortSessionByKey(addrkeyS, sortE)
	return sortedr, sortLargeNum
}

/*
搜索虚拟节点消息,用在虚拟节点中查询中转的方法
@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    *nodeStore.AddressNet  选出的发送/转发节点地址
@return    error                  处理中错误
@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
*/
func (this *Message) routerSearchVnodeOnebyone(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	var an []virtual_node.VnodeinfoS
	if vnodeinfo := vnc.FindInVnodeinfoSelf(*this.Head.RecvVnode); vnodeinfo != nil {
		return false, nil, nil, true
	}

	if this.Head.RecvId == nil || len(*this.Head.RecvId) == 0 {
		//TODO 优化消息第一次上onebyone的起始位置
		//onebyone正常模式
		//第一次上onebyone时，写死的第一个虚拟节点
		var nVnodeIndex uint64
		// 查找index最大的虚拟地址
		vnc.RLock()
		defer vnc.RUnlock()
		for k := range vnc.VnodeMap {
			if k > nVnodeIndex {
				nVnodeIndex = k
			} else if nVnodeIndex == 0 {
				nVnodeIndex = k
			}
		}

		if nVnodeIndex == 0 {
			an = vnc.VnodeMap[nVnodeIndex].GetDownVnodeInfo()
			an = append(an, vnc.VnodeMap[nVnodeIndex].GetUpVnodeInfo()...)
		} else {
			an = vnc.VnodeMap[nVnodeIndex].GetOnebyoneVnodeInfo()
		}
		//本地没有带虚拟空间的节点，把消息通过连接的逻辑节点传递出去
		if len(an) <= 0 {
			if from != nil && !bytes.Equal(*from, *this.Head.Sender) {
				return true, nil, nil, true
			}
			if bytes.Equal(*this.Head.Sender, nodeManager.NodeSelf.IdInfo.Id) {
				targetid := nodeManager.FindNearInSuper(&nodeManager.NodeSelf.IdInfo.Id, from, false, blockAddr)
				return true, targetid, nil, false
			} else {
				return true, nil, nil, true
			}
		}
		return this.searchMsgNextRecvVnode(vnc.VnodeMap[nVnodeIndex], &an)
	} else if vnode := vnc.FindVnodeInAllSelf(*this.Head.RecvId); vnode != nil {
		an := vnode.GetOnebyoneVnodeInfo()
		return this.searchMsgNextRecvVnode(vnode, &an)
	} else {
		vnc.RLock()
		defer vnc.RUnlock()
		for i, _ := range vnc.VnodeMap {
			if bytes.Equal(vnc.VnodeMap[i].Vnode.Vid, *this.Head.SelfVnodeId) {
				var vnodesInfo []virtual_node.VnodeinfoS
				//处理消息重试时，只用查询单方向连接的虚拟节点
				if new(big.Int).SetBytes(vnc.VnodeMap[i].Vnode.Vid).Cmp(new(big.Int).SetBytes(*this.Head.RecvVnode)) == 1 {
					an = vnc.VnodeMap[i].GetDownVnodeInfo()
				} else {
					an = vnc.VnodeMap[i].GetUpVnodeInfo()
				}
				for i, _ := range an {
					if blockAddr[utils.Bytes2string(an[i].Vid)] > engine.MaxRetryCnt {
						continue
					}
					vnodesInfo = append(vnodesInfo, an[i])
				}
				if len(vnodesInfo) == 0 {
					return true, nil, nil, true
				}
				return this.searchMsgNextRecvVnode(vnc.VnodeMap[i], &vnodesInfo)
			}
		}
	}
	return true, nil, nil, false
}

/*
搜索真实节点消息,用onebyone在真实节点中寻找的方法
修改版，选出 要求个数 的距离目标节点最近的点（自己 up down ...）

**********************************************
注意：此方法在消息路由到代理服务端节点以后只能再路由一次
**********************************************

@return    []nodeStore.AddressNet  选出的发送/转发节点地址数组
@return    error                  处理中错误
*/
func (this *Message) routerSearchNodeOnebyone2(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, blockAddr map[string]int, searchNum int) ([]nodeStore.AddressNet, error) {
	// utils.Log.Info().Msgf("发送routerSearchVnode self:%s from:%s to:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(),
	// 	this.Head.SenderVnode.B58String(), this.Head.RecvVnode.B58String())

	var allAddr []nodeStore.AddressNet
	rnode := make([]nodeStore.AddressNet, 0)

	//1.1正常模式下的判断和转发
	allAddr = append(allAddr, nodeManager.NodeSelf.IdInfo.Id)
	ud := sessionEngine.GetAllDownUp(nodeManager.AreaNameSelf)
	for _, v := range ud {
		if this.Head.RecvId != nil && utils.Bytes2string(*this.Head.RecvId) == v {
			nearId := nodeStore.AddressNet(v)
			rnode = append(rnode, nearId)
		}
		if v := blockAddr[v]; v > engine.MaxRetryCnt {
			continue
		}
		allAddr = append(allAddr, nodeStore.AddressNet(v))
	}

	//1.2自己不是服务端节点，给自己连接的非自己的节点发消息
	if !nodeManager.NodeSelf.GetIsSuper() {
		for i := range allAddr {
			if !bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, allAddr[i]) {
				target := allAddr[i]
				rnode = append(rnode, target)
				return rnode, nil
			}
		}
		return rnode, errors.New("未发现服务端节点")
	}

	//2.计算接受节点和当前onebyone连接节点状态
	sorted, times := this.GetSortSession(allAddr, sessionEngine)
	// utils.Log.Info().Msgf("_______________________________________ len(sorted) %d, times %d len(allAddr) %d", len(sorted), times, len(allAddr))
	//3.1当搜索地址落到比最大大 或者 比最小小 时候
	if times == 0 || times == len(sorted) {
		//当前6个节点，updown一共20连接情境下使用
		if times == 0 {
			for i := 0; i < searchNum; i++ {
				if i < len(sorted) {
					target := sorted[i]
					rnode = append(rnode, target)
				}
			}
		} else if times == len(sorted) {
			for i := 0; i < searchNum; i++ {
				if len(sorted)-1-i >= 0 {
					target := sorted[len(sorted)-1-i]
					rnode = append(rnode, target)
				}
			}
		}
		return rnode, nil
	}

	//3.2 当搜索地址在最大最小范围里的时候排序，根据要求个数返回
	for i := 1; i < len(sorted)+1; i++ {
		//找upsession，由近及远
		if u := times - i; u >= 0 {
			anode := sorted[u]
			rnode = append(rnode, anode)
		}
		//找downsession，由近及远
		if d := times + i - 1; d < len(sorted) {
			anode := sorted[d]
			rnode = append(rnode, anode)
		}
	}

	if len(rnode) > searchNum {
		rnode = rnode[:searchNum]
	}
	return rnode, nil
}

/*
搜索真实节点消息,用onebyone在真实节点中寻找的方法
@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    *nodeStore.AddressNet  选出的发送/转发节点地址
@return    error                  处理中错误
@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
*/
func (this *Message) routerSearchNodeOnebyone(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	// utils.Log.Info().Msgf("发送routerSearchVnode self:%s from:%s to:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(),
	// 	this.Head.SenderVnode.B58String(), this.Head.RecvVnode.B58String())

	var allAddr []nodeStore.AddressNet
	//判断nodeid是否是自己，是则自己处理
	if nodeManager.NodeSelf.GetIsSuper() && bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
		return false, nil, nil, true
	}

	//正常模式下的判断和转发
	allAddr = append(allAddr, nodeManager.NodeSelf.IdInfo.Id)
	ud := sessionEngine.GetAllDownUp(nodeManager.AreaNameSelf)
	for _, v := range ud {
		if blockAddr[v] > engine.MaxRetryCnt {
			continue
		}

		allAddr = append(allAddr, nodeStore.AddressNet(v))
	}

	//计算接受节点和当前onebyone连接节点状态
	sorted, times := this.GetSortSession(allAddr, sessionEngine)
	// sorted, times := this.GetSortSession(allAddr)
	if !nodeManager.NodeSelf.GetIsSuper() {
		for i := range sorted {
			if !bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, sorted[i]) {
				target := sorted[i]
				return true, &target, nil, false
			}
		}
		return true, nil, nil, true
	}
	if nodeManager.IsHeadTailModl {
		if times == len(sorted) {
			target := sorted[len(sorted)-1]
			//如果自己是最后一个点，则自己处理
			if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
				return false, nil, nil, true
			}
			return true, &target, nil, false
		}

		if times == 0 {
			target := sorted[0]
			//如果自己是第一个点，则自己处理
			if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
				return false, nil, nil, true
			}
			return true, &target, nil, false
		}
		target := sorted[times-1]
		if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
			//如果自己是紧挨着真实节点的点，则自己处理
			return false, nil, nil, true
		}

		return true, &target, nil, false
	} else {
		if times == 0 || times == len(sorted) {
			//说明不在自己onebyone范围里，需要路由走
			if times == 0 {
				target := sorted[0]
				if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
					//如果自己是紧挨着真实节点的点，则自己处理
					return false, nil, nil, false
				}
				return true, &target, nil, false
			} else if times == len(sorted) {
				target := sorted[len(sorted)-1]
				if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
					//如果自己是紧挨着真实节点的点，则自己处理
					return false, nil, nil, false
				}
				return true, &target, nil, false
			}
		}

		//utils.Log.Info().Msgf("寻找真实节点在自己onebyone范围里")
		target := sorted[times-1]
		//utils.Log.Info().Msgf("****%s", target.B58String())
		if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
			//如果自己是紧挨着真实节点的点，则自己处理
			return false, nil, nil, false
		}

		return true, &target, nil, true
	}

}

/*
搜索虚拟节点消息
@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    *nodeStore.AddressNet  选出的发送/转发节点地址
@return    error                  处理中错误
@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
*/
func (this *Message) routerSearchVnode(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	// utils.Log.Info().Msgf("发送routerSearchVnode self:%s from:%s to:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(),
	// 	this.Head.SenderVnode.B58String(), this.Head.RecvVnode.B58String())

	//判断vnodeid是否是自己，不是则路由出去
	vnodeinfo := vnc.FindInVnodeinfoSelf(*this.Head.RecvVnode)
	if vnodeinfo != nil {
		// utils.Log.Info().Msgf("找到vnode节点了from:%s to:%s self:%s", this.Head.SenderVnode.B58String(),
		// 	this.Head.RecvVnode.B58String(), vnodeinfo.Vid.B58String())
		return false, nil, nil, false
	}

	isOther := false
	var targetId *nodeStore.AddressNet
	var err error
	var force bool
	if this.Head.RecvId == nil || len(*this.Head.RecvId) == 0 {
		// utils.Log.Info().Msgf("1111111111")
		isOther, targetId, err, force = this.changeVnodeRecvidSearchVnode(version, nodeManager, sessionEngine, vnc, from, timeout, false, true, blockAddr)
		if err != nil || isOther {
			if err != nil {
				// utils.Log.Error().Msgf("出错了from:%s to:%s self:%s error:%s", this.Head.SenderVnode.B58String(),
				// 	this.Head.RecvVnode.B58String(), vnc.DiscoverVnodes.Vnode.Vid.B58String(), err.Error())
			}
			//return isOther, targetId, err
		}
		return isOther, targetId, err, force
	} else if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
		// utils.Log.Info().Msgf("22222222")
		//接收地址是自己
		isOther, targetId, err, force = this.changeVnodeRecvidSearchVnode(version, nodeManager, sessionEngine, vnc, from, timeout, true, true, blockAddr)
		if err != nil {
			// utils.Log.Error().Msgf("出错了from:%s to:%s self:%s error:%s", this.Head.SenderVnode.B58String(),
			// 	this.Head.RecvVnode.B58String(), vnc.DiscoverVnodes.Vnode.Vid.B58String(), err.Error())
			// return isOther, targetId, err
		} else if isOther {
			// utils.Log.Info().Msgf("转发vnode节点了from:%s to:%s self:%s", this.Head.SenderVnode.B58String(),
			// 	this.Head.RecvVnode.B58String(), vnc.DiscoverVnodes.Vnode.Vid.B58String())
			// return true, targetId, err
		} else {
			// utils.Log.Info().Msgf("找到vnode节点了from:%s to:%s", this.Head.SenderVnode.B58String(), this.Head.RecvVnode.B58String())
			// return isOther, targetId, err
		}
		return isOther, targetId, err, force
	} else {
		return this.routerP2P(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
	}
}

/*
修改Vnode消息中中转节点地址，带黑名单
@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    *nodeStore.AddressNet  选出的发送/转发节点地址
@return    error                  处理中错误
@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
*/
func (this *Message) changeVnodeRecvidSearchVnode(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration, includeSelf, includeIndex0 bool, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	//查询接收地址最近的虚拟节点
	targetVnodeId := vnc.FindNearVnodeSearchVnode(this.Head.RecvVnode, nil, includeSelf, true, blockAddr)
	if targetVnodeId == nil {
		//没有可用的邻居节点
		// utils.Log.Info().Msgf("1111111111111")
		// utils.Log.Info().Msgf("没有可用的邻居节点")
		return true, nil, config.ERROR_no_neighbor, true
	}

	// utils.Log.Info().Msgf("changeVnodeRecvidSearchVnode 11111 target:%s", targetVnodeId.B58String())
	//判断最近的接收地址是不是自己
	if vnc.FindInVnodeSelf(targetVnodeId) {
		// 如果最近的接收地址是发现节点的地址，则排出所有index为0的节点重新获取一次最近地址
		if includeIndex0 && bytes.Equal(vnc.DiscoverVnodes.Vnode.Nid, targetVnodeId) {
			// 排除index为0的所有节点，重新确认最终的目标虚拟节点
			targetVnodeId = vnc.FindNearVnodeSearchVnode(this.Head.RecvVnode, nil, includeSelf, false, blockAddr)
			if len(targetVnodeId) == 0 {
				// 防止不存在虚拟节点，因此只能包含发现节点查询，进而作为最终目标虚拟节点
				// targetVnodeId = vnc.DiscoverVnodes.Vnode.Vid
				targetVnodeId = vnc.FindNearVnodeSearchVnode(this.Head.RecvVnode, nil, includeSelf, true, blockAddr)
			}
			this.Head.SearchVnodeEndId = &targetVnodeId
			return this.routerSearchVnodeEnd(version, nodeManager, sessionEngine, vnc, from, timeout, false, blockAddr)
		}

		// vnodeinfo := vnc.FindInVnodeinfoSelf(targetId)
		// utils.Log.Info().Msgf("消息属于自己节点self:%s find:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), this.Head.RecvVnode.B58String())
		return false, nil, nil, false
	}

	// utils.Log.Info().Msgf("changeVnodeRecvidSearchVnode 22222 target:%s", targetVnodeId.B58String())
	//避免发送给不存在的节点或者节点不在线，产生消息死循环发送
	if !includeSelf && this.Head.RecvId != nil && len(*this.Head.RecvId) != 0 {

		// utils.Log.Info().Msgf("changeVnodeRecvidSearchVnode 33333 target:%s", targetVnodeId.B58String())
		targetVnodeIdTemp := vnc.FindNearVnodeSearchVnode(this.Head.RecvVnode, nil, true, true, blockAddr)
		if targetVnodeIdTemp == nil {
			//没有可用的邻居节点
			// utils.Log.Info().Msgf("没有可用的邻居节点")
			return true, nil, config.ERROR_no_neighbor, true
		}

		// utils.Log.Info().Msgf("changeVnodeRecvidSearchVnode 44444 target:%s", targetVnodeId.B58String())
		if !bytes.Equal(targetVnodeIdTemp, targetVnodeId) {
			// utils.Log.Info().Msgf("这个节点不存在 self:%s target:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetVnodeId.B58String())
			return true, nil, nil, true
		}
	}

	//查询vnodeinfo
	vnodeinfo := vnc.FindVnodeinfo(targetVnodeId)
	if vnodeinfo == nil {
		// utils.Log.Info().Msgf("没有可用的邻居节点:%s %s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(), targetVnodeId.B58String())
		return true, nil, config.ERROR_no_neighbor, true
	}
	// utils.Log.Info().Msgf("改变接收地址self:%s toNID:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(), vnodeinfo.Nid.B58String())
	this.Head.RecvId = &vnodeinfo.Nid
	this.Head.RecvSuperId = &vnodeinfo.Nid
	return this.routerP2P(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
}

/*
路由虚拟节点消息
@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    *nodeStore.AddressNet  选出的发送/转发节点地址
@return    error                  处理中错误
@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
*/
func (this *Message) routerVnodeP2P(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	// utils.Log.Info().Msgf("发送routerSearchVnode self:%s from:%s to:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(),
	// 	this.Head.SenderVnode.B58String(), this.Head.RecvVnode.B58String())
	// utils.Log.Info().Msgf("发送routerVnodeP2P self:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String())

	//判断vnodeid是否是自己，不是则路由出去
	vnodeinfo := vnc.FindInVnodeinfoSelf(*this.Head.RecvVnode)
	if vnodeinfo != nil {
		// utils.Log.Info().Msgf("找到vnode节点了from:%s to:%s self:%s", this.Head.SenderVnode.B58String(),
		// 	this.Head.RecvVnode.B58String(), vnodeinfo.Vid.B58String())
		return false, nil, nil, false
	}

	isOther := false
	var targetId *nodeStore.AddressNet
	var err error
	var force bool
	if this.Head.RecvId == nil || len(*this.Head.RecvId) == 0 {
		// utils.Log.Info().Msgf("222 1111111111")
		isOther, targetId, err, force = this.changeVnodeRecvidVnodeP2P(version, nodeManager, sessionEngine, vnc, from, timeout, false, blockAddr)
		if err != nil || isOther {
			if err != nil {
				// utils.Log.Error().Msgf("出错了from:%s to:%s self:%s error:%s", this.Head.SenderVnode.B58String(),
				// 	this.Head.RecvVnode.B58String(), vnc.DiscoverVnodes.Vnode.Vid.B58String(), err.Error())
			}
			if isOther {
				// utils.Log.Error().Msgf("出错了from:%s to:%s self:%s", this.Head.SenderVnode.B58String(),
				// 	this.Head.RecvVnode.B58String(), vnc.DiscoverVnodes.Vnode.Vid.B58String())
			}
			//return isOther, targetId, err
		}
		return isOther, targetId, err, force
	} else if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
		// utils.Log.Info().Msgf("222 22222222")
		//接收地址是自己
		isOther, targetId, err, force = this.changeVnodeRecvidVnodeP2P(version, nodeManager, sessionEngine, vnc, from, timeout, true, blockAddr)
		if err != nil {
			// utils.Log.Error().Msgf("出错了from:%s to:%s self:%s error:%s", this.Head.SenderVnode.B58String(),
			// 	this.Head.RecvVnode.B58String(), vnc.DiscoverVnodes.Vnode.Vid.B58String(), err.Error())
			// return isOther, targetId, err
		} else if isOther {
			// utils.Log.Info().Msgf("转发vnode节点了from:%s to:%s self:%s", this.Head.SenderVnode.B58String(),
			// 	this.Head.RecvVnode.B58String(), vnc.DiscoverVnodes.Vnode.Vid.B58String())
			// return true, targetId, err
		} else {
			// utils.Log.Info().Msgf("找到vnode节点了from:%s to:%s", this.Head.SenderVnode.B58String(), this.Head.RecvVnode.B58String())
			// return isOther, targetId, err
		}
		return isOther, targetId, err, force
	} else {
		// utils.Log.Info().Msgf("222 3333333333")
		return this.routerP2P(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
	}
}

/*
修改Vnode消息中中转节点地址
@return    bool    是否发送给其他人，true=成功发送给其他人;false=发送给自己;
@return    error   发送错误信息
*/
func (this *Message) changeVnodeRecvidVnodeP2P(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration, includeSelf bool, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	//查询接收地址最近的虚拟节点
	targetVnodeId := vnc.FindNearVnodeP2P(this.Head.RecvVnode, nil, includeSelf, blockAddr)
	if targetVnodeId == nil {
		//没有可用的邻居节点
		// utils.Log.Info().Msgf("1111111111111")
		// utils.Log.Info().Msgf("没有可用的邻居节点")
		return true, nil, config.ERROR_no_neighbor, true
	}

	// utils.Log.Info().Msgf("changeVnodeRecvid 11111 target:%s", targetVnodeId.B58String())
	//判断最近的接收地址是不是自己
	if vnc.FindInVnodeSelf(targetVnodeId) {
		// vnodeinfo := vnc.FindInVnodeinfoSelf(targetId)
		// utils.Log.Info().Msgf("消息属于自己节点self:%s find:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), this.Head.RecvVnode.B58String())
		if bytes.Equal(targetVnodeId, *this.Head.RecvVnode) {
			return false, nil, nil, false
		} else {
			// utils.Log.Info().Msgf("qlw------没有可用的邻居节点")
			return true, nil, config.ERROR_no_neighbor, true
		}
	}

	// utils.Log.Info().Msgf("changeVnodeRecvid 22222 target:%s", targetVnodeId.B58String())
	//避免发送给不存在的节点或者节点不在线，产生消息死循环发送
	if !includeSelf && this.Head.RecvId != nil && len(*this.Head.RecvId) != 0 {

		// utils.Log.Info().Msgf("changeVnodeRecvid 33333 target:%s", targetVnodeId.B58String())
		targetVnodeIdTemp := vnc.FindNearVnodeP2P(this.Head.RecvVnode, nil, true, blockAddr)
		if targetVnodeIdTemp == nil {
			//没有可用的邻居节点
			// utils.Log.Info().Msgf("没有可用的邻居节点")
			return true, nil, config.ERROR_no_neighbor, true
		}

		// utils.Log.Info().Msgf("changeVnodeRecvid 44444 target:%s", targetVnodeId.B58String())
		if !bytes.Equal(targetVnodeIdTemp, targetVnodeId) {
			// utils.Log.Info().Msgf("这个节点不存在 self:%s target:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetVnodeId.B58String())
			return true, nil, nil, true
		}
	}

	//查询vnodeinfo
	vnodeinfo := vnc.FindVnodeinfo(targetVnodeId)
	if vnodeinfo == nil {
		// utils.Log.Info().Msgf("没有可用的邻居节点:%s %s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(), targetVnodeId.B58String())
		return true, nil, config.ERROR_no_neighbor, true
	}
	// utils.Log.Info().Msgf("改变接收地址self:%s toNID:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(), vnodeinfo.Nid.B58String())
	this.Head.RecvId = &vnodeinfo.Nid
	this.Head.RecvSuperId = &vnodeinfo.Nid
	return this.routerP2P(version, nodeManager, sessionEngine, vnc, from, timeout, blockAddr)
}

// var debuf_msgid uint64 = 0

//var debuf_msgid uint64 = 1000
//var debuf_msgid uint64 = MSGID_TextMsg
//var debuf_msgid uint64 = gconfig.MSGID_findSuperID

/*
	检查该消息是否是自己的
	不是自己的则自动转发出去
	@return    bool    是否发送给其他人，true=成功发送给其他人;false=发送给自己;
	@return    error   发送错误信息
*/
// func (this *Message) SendOld(version uint64, nodeManager *nodeStore.NodeManager,
// 	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, timeout time.Duration) (bool, error) {
// 	//虚拟节点之间的路由
// 	if this.Head.SenderVnode != nil && this.Head.RecvVnode != nil && this.Head.RecvId == nil {
// 		// if isVnodeRouter {
// 		this.Head.Sender = &nodeManager.NodeSelf.IdInfo.Id
// 		//收消息人就是自己
// 		if vnc.FindInVnodeSelf(*this.Head.RecvVnode) {
// 			// utils.Log.Info().Msgf("1111111111111")
// 			return false, nil
// 		}

// 		//Accurate参数是否发送给指定的某个虚拟节点
// 		//区分查找节点协议和点对点通讯协议
// 		var targetId virtual_node.AddressNetExtend
// 		if this.Head.Accurate {
// 			targetId = vnc.FindNearVnode(this.Head.RecvVnode, nil, false)
// 		} else {
// 			targetId = vnc.FindNearVnode(this.Head.RecvVnode, nil, true)
// 		}
// 		//没有可用的邻居节点
// 		if targetId == nil {
// 			// utils.Log.Info().Msgf("1111111111111")
// 			utils.Log.Info().Msgf("没有可用的邻居节点")
// 			return true, config.ERROR_no_neighbor
// 		}

// 		//判断是否是自己的节点
// 		if vnc.FindInVnodeSelf(targetId) {
// 			// vnodeinfo := vnc.FindInVnodeinfoSelf(targetId)
// 			// utils.Log.Info().Msgf("消息属于自己节点:%s %d", vnodeinfo.Nid.B58String(), vnodeinfo.Index)
// 			return false, nil
// 		}

// 		// fmt.Println("打印地址", (targetId).B58String())
// 		vnodeinfo := vnc.FindVnodeinfo(targetId)
// 		if vnodeinfo == nil {
// 			utils.Log.Info().Msgf("没有可用的邻居节点:%s %s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(), targetId.B58String())
// 			return false, config.ERROR_no_neighbor
// 		}
// 		this.Head.RecvId = &vnodeinfo.Nid
// 		this.Head.RecvSuperId = &vnodeinfo.Nid
// 	}
// 	// utils.Log.Info().Msgf("发送消息1")
// 	return this.sendNormal(version, nodeManager, sessionEngine, timeout)
// }

/*
	发送给普通节点，最原始的消息
	@return    bool    是否发送给别人
	@return    error   错误
*/
// func (this *Message) sendNormal(version uint64, nodeManager *nodeStore.NodeManager, sessionEngine *engine.Engine, timeout time.Duration) (bool, error) {
// 	//安全协议不需buildhash
// 	// utils.Log.Info().Msgf("buildhash befor:%s %+v", hex.EncodeToString(this.Body.Hash), this.Body)
// 	this.BuildHash()
// 	// utils.Log.Info().Msgf("buildhash after:%s %+v", hex.EncodeToString(this.Body.Hash), this.Body)

// 	if nodeManager.NodeSelf.GetIsSuper() {
// 		// if version == debuf_msgid {
// 		// 	fmt.Println("-=-=- 111111111111")
// 		// }
// 		// utils.Log.Info().Msgf("11111111111:%+v", this.Head)
// 		//收消息人就是自己
// 		if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
// 			// utils.Log.Info().Msgf("发送消息1111111111")
// 			return false, nil
// 		}
// 		// if version == debuf_msgid {
// 		// 	fmt.Println("-=-=- 333333333333333")
// 		// }
// 		//查找代理节点
// 		// if _, ok := nodeStore.GetProxyNode(this.Head.RecvId.B58String()); ok {
// 		if _, ok := nodeManager.GetProxyNode(utils.Bytes2string(*this.Head.RecvId)); ok {
// 			//发送给代理节点
// 			// if session, ok := engine.GetSession(this.Head.RecvId.B58String()); ok {
// 			if session, ok := sessionEngine.GetSession(utils.Bytes2string(*this.Head.RecvId)); ok {
// 				// if version == debuf_msgid {
// 				// 	fmt.Println("-=-=- 4444444444444")
// 				// }
// 				// session.Send(version, this.Head.JSON(), this.Body.JSON(), false)

// 				mheadBs := this.Head.Proto()
// 				mbodyBs, err := this.Body.Proto()
// 				if err != nil {
// 					return true, err
// 				}
// 				err = session.Send(version, &mheadBs, &mbodyBs, timeout)
// 				if err != nil {
// 					utils.Log.Info().Msgf("错误:%s", err.Error())
// 					return true, err
// 				}
// 				// utils.Log.Info().Msgf("发送消息 222222222222")
// 			} else {
// 				// utils.Log.Info().Msgf("这个代理节点的链接断开了")
// 			}
// 			return true, nil
// 		}
// 		// if version == debuf_msgid {
// 		// 	fmt.Println("-=-=- 5555555")
// 		// }

// 		//		fmt.Println(string(*this.Head.JSON()))
// 		var targetId *nodeStore.AddressNet
// 		if this.Head.Accurate {
// 			targetId = nodeManager.FindNearInSuper(this.Head.RecvSuperId, nil, false)
// 		} else {
// 			targetId = nodeManager.FindNearInSuper(this.Head.RecvSuperId, nil, true)
// 		}
// 		// utils.Log.Info().Msgf("本节点的其他超级节点:%s", targetId.B58String())
// 		if targetId == nil {
// 			utils.Log.Info().Msgf("没有可用的邻居节点")
// 			return true, config.ERROR_no_neighbor
// 		}
// 		// if version == debuf_msgid {
// 		// 	fmt.Println("-=-=- 666666666666")
// 		// }
// 		//收消息人就是自己
// 		// if nodeStore.NodeSelf.IdInfo.Id.B58String() == targetId.B58String() {
// 		if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *targetId) {
// 			// utils.Log.Info().Msgf("1111111111111")
// 			return false, nil
// 		}
// 		//转发出去
// 		if session, ok := sessionEngine.GetSession(utils.Bytes2string(*targetId)); ok {
// 			mheadBs := this.Head.Proto()
// 			mbodyBs, err := this.Body.Proto()
// 			if err != nil {
// 				return true, err
// 			}
// 			err = session.Send(version, &mheadBs, &mbodyBs, timeout)
// 			if err != nil {
// 				return true, err
// 			} else {
// 				return true, nil
// 			}
// 		} else {
// 			// utils.Log.Info().Msgf("1111111111111")
// 			return true, config.ERROR_get_node_conn_fail
// 		}
// 	} else {
// 		if nodeManager.SuperPeerId == nil {
// 			// utils.Log.Info().Msgf("没有可用的超级节点")
// 			return true, config.ERROR_no_super
// 		}
// 		if session, ok := sessionEngine.GetSession(utils.Bytes2string(*nodeManager.SuperPeerId)); ok {
// 			mheadBs := this.Head.Proto()
// 			mbodyBs, err := this.Body.Proto()
// 			if err != nil {
// 				return true, err
// 			}
// 			err = session.Send(version, &mheadBs, &mbodyBs, timeout)
// 			if err != nil {
// 				// utils.Log.Info().Msgf("send message error:%s", err.Error())
// 				return true, err
// 			} else {
// 				return true, nil
// 			}
// 			// session.Send(version, this.Head.Proto(), this.Body.Proto(), false)
// 		} else {
// 			// utils.Log.Info().Msgf("超级节点的session未找到")
// 			return true, config.ERROR_get_node_conn_fail
// 		}
// 	}
// }

/*
	检查该消息是否是自己的
	不是自己的则自动转发出去
	@return    bool    是否发送给其他人。true=发送给其他人了;false=自己的消息;
*/
// func (this *Message) IsSendOther(from *nodeStore.AddressNet, messageCenter *MessageCenter, timeout time.Duration) (bool, error) {
// 	// utils.Log.Info().Msgf("IsSendOther打印Accurate:%v", this.Head.Accurate)
// 	//如果是虚拟节点之间的消息，则一定是指定某节点的
// 	// oldAccurate := this.Head.Accurate
// 	// if this.Head.SenderVnode != nil && this.Head.RecvVnode != nil {
// 	// 	this.Head.Accurate = true
// 	// }
// 	ok, err := IsSendToOtherSuperToo(this.Head, this.DataPlus, this.msgid, from, messageCenter.nodeManager, messageCenter.sessionEngine, timeout)
// 	if err != nil {
// 		return ok, err
// 	}

// 	// utils.Log.Info().Msgf("打印消息2 %t", ok)
// 	//将messageHead.Accurate参数恢复
// 	// messageHead.Accurate = oldAccurate

// 	//发送给自己并且是虚拟节点之间的消息
// 	if !ok && this.Head.SenderVnode != nil && this.Head.RecvVnode != nil {
// 		if len(messageCenter.vm.GetVnodeSelf()) <= 0 {
// 			return true, nil
// 		}

// 		// this.Head.Sender = &nodeStore.NodeSelf.IdInfo.Id
// 		//收消息人就是自己
// 		if messageCenter.vm.FindInVnodeSelf(*this.Head.RecvVnode) {
// 			return ok, nil
// 		}

// 		//Accurate参数是否发送给指定的某个虚拟节点
// 		//区分查找节点协议和点对点通讯协议
// 		var targetId virtual_node.AddressNetExtend
// 		if this.Head.Accurate {
// 			// utils.Log.Info().Msgf("11111111111")
// 			targetId = messageCenter.vm.FindNearVnode(this.Head.RecvVnode, nil, false)
// 		} else {
// 			// utils.Log.Info().Msgf("22222222222")
// 			targetId = messageCenter.vm.FindNearVnode(this.Head.RecvVnode, nil, true)
// 		}
// 		//没有可用的邻居节点
// 		if targetId == nil {
// 			return true, nil
// 		}
// 		//判断是否是自己的节点
// 		if messageCenter.vm.FindInVnodeSelf(targetId) {
// 			// vnodeinfo := messageCenter.vm.FindInVnodeinfoSelf(targetId)
// 			// utils.Log.Info().Msgf("消息属于自己节点:%s %d", vnodeinfo.Nid.B58String(), vnodeinfo.Index)
// 			return false, nil
// 		}
// 		// utils.Log.Info().Msgf("打印消息3 %v", this.Body)
// 		vnodeinfo := messageCenter.vm.FindVnodeinfo(targetId)
// 		this.Head.RecvId = &vnodeinfo.Nid
// 		this.Head.RecvSuperId = &vnodeinfo.Nid
// 		// bs :=
// 		if this.DataPlus == nil {
// 			// utils.Log.Info().Msgf("33333333333333")
// 			_, err = messageCenter.SendVnodeP2pMsgHE(this.Body.MessageId, this.Head.SenderVnode, this.Head.RecvVnode, nil)
// 		} else {
// 			// utils.Log.Info().Msgf("本虚拟节点ID:%s", messageCenter.vm.GetVnodeDiscover().Vnode.Vid.B58String())
// 			// utils.Log.Info().Msgf("要转发的节点ID:%s", this.Head.RecvVnode.B58String())
// 			// utils.Log.Info().Msgf("打印一哈:%v", this.Body)
// 			// _, err = messageCenter.SendVnodeP2pMsgHE(this.Body.MessageId, this.Head.SenderVnode, this.Head.RecvVnode, this.DataPlus)
// 			err = this.forward(messageCenter, timeout)
// 		}
// 		return true, err
// 	}
// 	return ok, nil
// }

/*
转发未解析的Body内容
*/
func (this *Message) forward(messageCenter *MessageCenter, timeout time.Duration) error {
	//转发出去
	if session, ok := messageCenter.sessionEngine.GetSession(messageCenter.areaName, utils.Bytes2string(*this.Head.RecvId)); ok {
		// utils.Log.Info().Msgf("forward打印Accurate:%v", this.Head.Accurate)
		mheadBs := this.Head.Proto()
		err := session.Send(this.msgid, &mheadBs, this.DataPlus, timeout)
		if err != nil {
			return err
		} else {
			return nil
		}
	} else {
		// utils.Log.Info().Msgf("1111111111111")
		return config.ERROR_get_node_conn_fail
	}
}

/*
* 搜索虚拟节点最终接收节点消息
* @checkSelf 是否需要检测消息的发送者是否是自己
* @return    bool    是否发送给其他人，true=成功发送给其他人;false=发送给自己;
* @return    error   发送错误信息
* @return    bool    是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
 */
func (this *Message) routerSearchVnodeEnd(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration, checkSelf bool, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
	// 不需要检查发送者是不是自己
	if !checkSelf {
		if this.Head.SearchVnodeEndId == nil || len(*this.Head.SearchVnodeEndId) == 0 {
			return true, nil, errors.New("searchVnodeEndId is nil"), true
		}

		// 如果自己就是虚拟节点的最终目标节点，则直接处理
		vnodeInfo := vnc.FindInVnodeinfoSelf(*this.Head.SearchVnodeEndId)
		if vnodeInfo != nil {
			return false, nil, nil, false
		}

		// 修改消息的接收者地址信息
		vnodeinfo := vnc.FindVnodeinfo(*this.Head.SearchVnodeEndId)
		if vnodeinfo == nil {
			return true, nil, config.ERROR_no_neighbor, true
		}
		this.Head.RecvId = &vnodeinfo.Nid
		this.Head.RecvSuperId = &vnodeinfo.Nid
	}

	// 需要检测发送者信息，如果发送者就是自己，则抛出错误信息
	if checkSelf && from != nil && bytes.Equal(vnc.GetVnodeDiscover().Vnode.Vid, *this.Head.SearchVnodeEndId) {
		return true, nil, errors.New("repeat message err"), true
	}

	// 如果自己就是虚拟节点的最终目标节点，则直接处理
	vnodeInfo := vnc.FindInVnodeinfoSelf(*this.Head.SearchVnodeEndId)
	if vnodeInfo != nil {
		return false, nil, nil, false
	}

	return this.routerP2P(version, nodeManager, sessionEngine, vnc, nil, timeout, blockAddr)
}

/*
 * 检查该消息是否需要经过代理节点转发出去
 * @return    bool    需要经过代理节点转发，true=需要经过代理;false=不需要经过代理;
 */
func (this *Message) routerProxy(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet,
	timeout time.Duration) bool {
	// 判断是否指定发送代理节点
	if this.Head.SenderProxyId == nil {
		return false
	}

	// 获取发送者信息
	var senderId nodeStore.AddressNet
	if this.Head.Sender != nil {
		senderId = *this.Head.Sender
	} else if this.Head.SenderVnode != nil {
		senderId = nodeStore.AddressNet(*this.Head.SenderVnode)
	} else {
		return false
	}

	// 判断自己是不是消息的发送者
	if !bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, senderId) {
		return false
	}

	// 判断自己是不是代理节点
	if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.SenderProxyId) {
		return false
	}

	return true
}

/*
给连接的节点排序(包括本节点)，结果相对target 由正到负 大到小，确认消息发送顺序
+@return    []nodeStore.AddressNet   由远及近排列的节点地址;
+@return    数组中有多少地址大于传入self地址；
+
*/
func GetSortSessionForTarget(ss []engine.Session, self nodeStore.AddressNet, target nodeStore.AddressNet) ([]nodeStore.AddressNet, int) {
	var targetAddrBI *big.Int
	var sortedS []nodeStore.AddressNet
	sortLargeNum := 0
	onebyone := new(nodeStore.IdDESC)

	if self != nil {
		targetAddrBI = new(big.Int).SetBytes(target)
	}

	//1.把自己和上下session放在排序器中等待排序
	*onebyone = append(*onebyone, new(big.Int).SetBytes(self))
	for _, one := range ss {
		*onebyone = append(*onebyone, new(big.Int).SetBytes(nodeStore.AddressNet([]byte(one.GetName()))))
	}

	//2.排序器排序
	sort.Sort(onebyone)

	for i := 0; i < len(*onebyone); i++ {
		one := (*onebyone)[i]
		if one.Cmp(targetAddrBI) == 1 {
			sortLargeNum += 1
		}

		IdBs := one.Bytes()
		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
		sortedS = append(sortedS, nodeStore.AddressNet(*IdBsP))
	}

	return sortedS, sortLargeNum
}

/*
+给连接的节点排序，结果相对self 由正到负 大到小，确认消息发送顺序
+@return    []engine.Session   由远及近排列的节点session数组;
+@return    数组中有多少地址大于传入self地址；
+
*/
func GetSortSession(ss []engine.Session, self nodeStore.AddressNet) ([]engine.Session, int) {
	var selfAddrBI *big.Int
	var sortedS []engine.Session
	sortLargeNum := 0
	sortM := make(map[string]engine.Session)
	onebyone := new(nodeStore.IdDESC)

	if self != nil {
		selfAddrBI = new(big.Int).SetBytes(self)
	}

	for _, one := range ss {
		sortM[nodeStore.AddressNet([]byte(one.GetName())).B58String()] = one
		*onebyone = append(*onebyone, new(big.Int).SetBytes(nodeStore.AddressNet([]byte(one.GetName()))))
	}

	sort.Sort(onebyone)

	for i := 0; i < len(*onebyone); i++ {
		// utils.Log.Info().Msgf("selfAddrBI: ", selfAddrBI, " one", one).
		one := (*onebyone)[i]
		if one.Cmp(selfAddrBI) == 1 {
			sortLargeNum += 1
		}

		IdBs := one.Bytes()
		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
		sortedS = append(sortedS, sortM[nodeStore.AddressNet(*IdBsP).B58String()])
	}

	return sortedS, sortLargeNum
}

// 搜索消息中转给的虚拟节点
func (this *Message) searchMsgNextRecvVnode(vnode *virtual_node.Vnode, an *[]virtual_node.VnodeinfoS) (bool, *nodeStore.AddressNet, error, bool) {
	rN, rV, isother := vnode.SearchAVnodeByOnebyone(an, nodeStore.AddressNet(*this.Head.RecvVnode))
	if !isother {
		rVe := virtual_node.AddressNetExtend(*rV)
		this.Head.SelfVnodeId = &vnode.Vnode.Vid
		this.Head.RecvVnode = &rVe
		this.Head.RecvId = rV
		return false, nil, nil, true
	}
	this.Head.SelfVnodeId = &vnode.Vnode.Vid
	this.Head.RecvId = rV
	return true, rN, nil, false
}

package superinfo

import (
	"sync"
	"time"

	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
)

// 超级节点信息
type SuperInfo struct {
	Area                         *libp2parea.Area                      // 节点信息
	SuperNodes                   *sync.Map                             // 存放超级节点信息
	initialized                  bool                                  // 是否初始化过
	superNodeOnlineCallbackFunc  []libp2parea.NodeEventCallbackHandler // 超级节点上线回调
	superNodeOfflineCallbackFunc []libp2parea.NodeEventCallbackHandler // 超级节点下线回调
}

/*
 * 创建超级节点信息
 *
 * @param	area					*Area						所属区域
 * @param	superOnlineCallback		NodeEventCallbackHandler	大区信息
 * @param	superOfflineCallback	NodeEventCallbackHandler	大区信息
 * @return	si						*SuperInfo					超级节点信息
 */
func NewSuperInfo(area *libp2parea.Area, superOnlineCallback, superOfflineCallback libp2parea.NodeEventCallbackHandler) (si *SuperInfo) {
	// 判断参数的合法性
	if area == nil {
		utils.Log.Error().Msgf("节点信息为空, 请传入有效的节点信息!!!!!!")
		return nil
	}

	// 构建超级节点信息
	si = &SuperInfo{Area: area, SuperNodes: new(sync.Map)}
	si.Register_superNodeOnlineCallback(superOnlineCallback)
	si.Register_superNodeOfflineCallback(superOfflineCallback)

	// 注册新连接回调方法
	area.Register_nodeNewConnCallback(si.newConnCallbackFunc)

	// 初始化群消息
	si.init()

	return
}

// 初始化
func (si *SuperInfo) init() {
	// 判断是否初始化过
	if si.initialized {
		return
	}
	si.initialized = true

	// 注册消息
	si.registerMsg()

	// 只有超级节点才会执行该操作
	go si.loopSendOnlineInfo()
	go si.clearSuperNode()
}

/*
 * 获取大区中记录的
 */
func (si *SuperInfo) loopSendOnlineInfo() {
	if si == nil {
		return
	}

	// 构建定时器
	timer := time.NewTicker(TIME_LOOP_SEND_ONLINE_INFO)
	defer timer.Stop()

	si.Area.WaitAutonomyFinish()

	// 定时获取一次存储的节点地址, 直到获取到为止
	for range timer.C {
		// 如果还没初始化, 则等待下一循环
		if !si.initialized {
			continue
		}

		bs, err := si.Area.NodeManager.NodeSelf.Proto()
		if err != nil {
			return
		}

		// utils.Log.Info().Msgf("发送节点在线广播消息: self:%s", si.Area.NodeManager.NodeSelf.IdInfo.Id.B58String())
		si.Area.SendMulticastMsg(MSGID_SUPER_INFO_MULTICAST_MSG, &bs)
	}
}

/*
 * 定期清理下线的超级节点信息
 */
func (si *SuperInfo) clearSuperNode() {
	ticker := time.NewTicker(TIME_LOOP_CHECK_CACHE)
	defer ticker.Stop()

	si.Area.WaitAutonomyFinish()

	for range ticker.C {
		if !si.Area.NodeManager.NodeSelf.GetIsSuper() {
			return
		}

		// 遍历需要清理的超级节点列表
		si.SuperNodes.Range(func(key, value any) bool {
			// 1. 获取节点信息
			node, ok := value.(*nodeStore.Node)
			if !ok || node == nil {
				utils.Log.Error().Msgf("遍历需要清理的超级节点列表 11111111111111")
				si.SuperNodes.Delete(key)
				return true
			}

			// 2. 判断最后更新在线时间是否超时
			lastOnlineTime := node.GetOnlineTime()
			if time.Since(lastOnlineTime).Seconds() <= TIME_CACHE_INVALID {
				return true
			}
			// subSeconds := time.Since(lastOnlineTime).Seconds()
			// utils.Log.Error().Msgf("[%s]间隔秒数为: %v", node.IdInfo.Id.B58String(), subSeconds)

			// 3. 判断本地是否存在连接信息
			if si.Area.NodeManager.FindNode(&node.IdInfo.Id) != nil {
				// 2.1 连接信息存在
				node.FlashOnlineTime()
				si.SuperNodes.Store(key, node)
				return true
			}

			// 4. 删除超级节点记录信息
			si.SuperNodes.Delete(key)

			// 5. 调用超级节点下线回调方法
			for _, h := range si.superNodeOfflineCallbackFunc {
				go h(node.IdInfo.Id, node.MachineIDStr)
			}

			return true
		})
	}
}

/*
 * 注册超级节点下线回调函数
 */
func (si *SuperInfo) Register_superNodeOnlineCallback(handler libp2parea.NodeEventCallbackHandler) {
	if si == nil || handler == nil {
		return
	}
	si.superNodeOnlineCallbackFunc = append(si.superNodeOnlineCallbackFunc, handler)
}

/*
 * 注册超级节点下线回调函数
 */
func (si *SuperInfo) Register_superNodeOfflineCallback(handler libp2parea.NodeEventCallbackHandler) {
	if si == nil || handler == nil {
		return
	}
	si.superNodeOfflineCallbackFunc = append(si.superNodeOfflineCallbackFunc, handler)
}

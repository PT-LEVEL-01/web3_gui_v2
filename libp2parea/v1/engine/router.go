package engine

import (
	"sync"

	"web3_gui/utils"
)

type MsgHandler func(c Controller, msg Packet)

type Router struct {
	handlersMapping map[string]map[uint64]MsgHandler
	lock            *sync.RWMutex
}

/*
 * AddRouter 添加引擎路由
 *
 * @param	areaName	[]byte	区域名称
 * @param	msgId		uint64	消息id
 *
 */
func (this *Router) AddRouter(areaName []byte, msgId uint64, handler MsgHandler) {
	// key := fmt.Sprintf("%s_%d", utils.Bytes2string(areaName), msgId)
	key := utils.Bytes2string(areaName)

	this.lock.Lock()

	// 根据区域名获取消息号对应的回调方法
	_, exist := this.handlersMapping[key]
	if !exist {
		// 不存在时
		this.handlersMapping[key] = make(map[uint64]MsgHandler)
	} else {
		// 存在, 判断之前是否存在消息号对应的回调是否存在
		_, exist = this.handlersMapping[key][msgId]
		if exist {
			// 已存在，打印错误日志信息
			utils.Log.Error().Msgf("protocol number [%d] covered!", msgId)
		}
	}

	// 添加区域名对应的消息号回调处理
	this.handlersMapping[key][msgId] = handler
	this.lock.Unlock()
}

/*
 * GetHandler 根据区域名和消息号获取回调方法
 *
 * @param	areaName	string		区域名
 * @param	msgId		uint64		消息号
 * @return	handler		MsgHandler	回调方法
 */
func (this *Router) GetHandler(areaName string, msgId uint64) MsgHandler {
	// key := fmt.Sprintf("%s_%d", areaName, msgId)
	key := utils.Bytes2string([]byte(areaName))

	this.lock.RLock()

	// 先获取区域对应的消息号回调map
	msgIdMap, exist := this.handlersMapping[key]
	if !exist || msgIdMap == nil {
		return nil
	}
	// 从消息号回调map中根据消息号，获取回调方法
	handler := msgIdMap[msgId]

	this.lock.RUnlock()

	return handler
}

/*
 * RemoveHandlers 移除区域名对应的所有消息号回调方法
 *
 * @param	areaName	string		区域名
 */
func (this *Router) RemoveHandlers(areaName string) {
	// key := fmt.Sprintf("%s_%d", areaName, msgId)
	key := utils.Bytes2string([]byte(areaName))

	this.lock.Lock()

	// 根据区域名删除所有的消息号回调注册信息
	delete(this.handlersMapping, key)

	this.lock.Unlock()
}

func NewRouter() *Router {
	return &Router{
		handlersMapping: make(map[string]map[uint64]MsgHandler),
		lock:            new(sync.RWMutex),
	}
}

package engine

import (
	"sync"
	"web3_gui/utils"
)

type Router struct {
	handlersMapping map[uint64]MsgHandler //
	lock            *sync.RWMutex         //
}

/*
添加引擎路由
@param	msgId		uint64	消息id
*/
func (this *Router) AddRouter(msgId uint64, handler MsgHandler) {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.handlersMapping[msgId]
	if ok {
		utils.Log.Error().Uint64("msgId existed", msgId).Send()
	}
	this.handlersMapping[msgId] = handler
}

/*
根据区域名和消息号获取回调方法
@param	msgId		uint64		消息号
@return	handler		MsgHandler	回调方法
*/
func (this *Router) GetHandler(msgId uint64) (MsgHandler, bool) {
	this.lock.RLock()
	// 从消息号回调map中根据消息号，获取回调方法
	handler, ok := this.handlersMapping[msgId]
	this.lock.RUnlock()
	return handler, ok
}

func NewRouter() *Router {
	return &Router{
		handlersMapping: make(map[uint64]MsgHandler),
		lock:            new(sync.RWMutex),
	}
}

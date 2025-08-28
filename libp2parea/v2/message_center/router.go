package message_center

import (
	"github.com/rs/zerolog"
	"sync"
	"web3_gui/utils"
)

type Router struct {
	lock     *sync.RWMutex         //
	handlers map[uint64]MsgHandler //key:uint64=消息版本号;value:MsgHandler=;
	log      **zerolog.Logger
}

func NewRouter(log **zerolog.Logger) *Router {
	r := Router{
		lock:     new(sync.RWMutex),
		handlers: make(map[uint64]MsgHandler),
		log:      log,
	}
	return &r
}

/*
注册消息
@param	version		消息号
@param	handler		消息处理回调方法
@param	isSysMsg	是否是系统消息
*/
func (this *Router) Register(msgId uint64, handler MsgHandler) {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.handlers[msgId]
	if ok {
		utils.Log.Error().Uint64("msgId existed", msgId).Send()
	}
	//(*this.log).Info().Uint64("注册消息", msgId).Send()
	this.handlers[msgId] = handler
}

func (this *Router) GetHandler(msgId uint64) (MsgHandler, bool) {
	this.lock.RLock()
	// 从消息号回调map中根据消息号，获取回调方法
	handler, ok := this.handlers[msgId]
	this.lock.RUnlock()
	//(*this.log).Info().Uint64("查找消息", msgId).Send()
	return handler, ok
}

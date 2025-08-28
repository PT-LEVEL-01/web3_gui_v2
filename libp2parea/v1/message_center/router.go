package message_center

import (
	"sync"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/utils"
)

type Router struct {
	handlers *sync.Map //key:uint64=消息版本号;value:MsgHandler=;
}

/*
 * Register 注册消息
 * @param	version		消息号
 * @param	handler		消息处理回调方法
 * @param	isSysMsg	是否是系统消息
 */
func (this *Router) Register(version uint64, handler MsgHandler, isSysMsg bool) {
	if !isSysMsg {
		// 用户自定义消息
		if version < config.MSGID_MAX_ID {
			utils.Log.Error().Msgf("!!!系统预留了[%d]以内的消息号, 请注意消息号的范围, 当前注册的消息号为: [%d]", config.MSGID_MAX_ID, version)
			return
		}
	} else if version >= config.MSGID_MAX_ID {
		// 系统消息
		utils.Log.Error().Msgf("!!!系统只预留了[%d]以内的消息号, 请注意消息号的范围, 当前注册的消息号为: [%d]", config.MSGID_MAX_ID, version)
		return
	}
	if _, exist := this.handlers.Load(version); exist {
		utils.Log.Error().Msgf("!!!重复注册消息, 该消息的处理将会被抛弃, 消息号: [%d]", version)
		return
	}
	this.handlers.Store(version, handler)
}

func (this *Router) GetHandler(msgid uint64) MsgHandler {
	value, ok := this.handlers.Load(msgid)
	if !ok {
		return nil
	}
	h := value.(MsgHandler)
	return h
}

func NewRouter() *Router {
	return &Router{
		handlers: new(sync.Map),
	}
}

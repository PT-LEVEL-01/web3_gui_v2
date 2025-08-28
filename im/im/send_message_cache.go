package im

import (
	"context"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

var sendMsgBackoffTimer = []time.Duration{time.Second * 5, time.Second * 10, time.Second * 20, time.Second * 40,
	time.Second * 80, time.Second * 160, time.Second * 320, time.Second * 640, time.Hour}

var msgChanMapLock = new(sync.RWMutex)                  //
var msgChanMap = make(map[string]*MessageEscortProgram) //未发送成功消息管道。key:string=用户地址;value:*chan model.MessageContent=;

type MessageEscortProgram struct {
	RemoteAddr nodeStore.AddressNet       //远端地址
	msgChan    chan *model.MessageContent //消息管道
	destroy    context.Context            //销毁信号
}

func NewMessageEscortProgram(addr nodeStore.AddressNet, destroy context.Context) *MessageEscortProgram {
	mep := MessageEscortProgram{
		RemoteAddr: addr,                                                      //远端地址
		msgChan:    make(chan *model.MessageContent, config.MsgChanMaxLength), //消息管道
		destroy:    destroy,                                                   //销毁信号
	}
	go mep.loopSend()
	return &mep
}

func (this *MessageEscortProgram) loopSend() {
	var ERR utils.ERROR
	var one *model.MessageContent
	for {
		select {
		case one = <-this.msgChan:
			ERR = sendMsg(one)
		case <-this.destroy.Done():
			return
		}
		if ERR.CheckSuccess() {
			continue
		}
		//发送失败
		timer := utils.NewBackoffTimerChan(sendMsgBackoffTimer...)
		var interval time.Duration
		for {
			interval = timer.Wait(this.destroy)
			if interval == 0 {
				return
			}
			ERR = sendMsg(one)
			if ERR.CheckSuccess() {
				break
			}
		}
	}
}

/*
添加一个未发送消息
*/
func (this *MessageEscortProgram) addMsg(msg *model.MessageContent) utils.ERROR {
	select {
	case this.msgChan <- msg:
		//保存到数据库

		return utils.NewErrorSuccess()
	default:
		return utils.NewErrorBus(config.ERROR_CODE_IM_too_many_undelivered_messages, "")
	}
}

func sendMsg(msg *model.MessageContent) utils.ERROR {
	msgBs, err := msg.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	np := utils.NewNetParams(config.NET_protocol_version_v1, *msgBs)
	bs, err := np.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	resultBS, ERR := Node.SendP2pMsgHEWaitRequest(config.MSGID_send_friend_message, &msg.To, bs, time.Second*10)
	if ERR.CheckFail() {
		//utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	if resultBS == nil || len(*resultBS) == 0 {
		//返回参数错误
		return utils.NewErrorBus(config.ERROR_CODE_params_format_return, "")
	}
	nr, err := utils.ParseNetResult(*resultBS)
	if err != nil {
		utils.Log.Info().Msgf("解析远端参数错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	if !nr.CheckSuccess() {
		return nr.ConvertERROR()
	}
	return utils.NewErrorSuccess()
}

/*
给好友发消息
*/
func SendMessageFirend(msg model.MessageContent) utils.ERROR {
	msgChanMapLock.Lock()
	mep, ok := msgChanMap[utils.Bytes2string(msg.To.GetAddr())]
	if !ok {
		mep = NewMessageEscortProgram(msg.To, context.WithoutCancel(context.Background()))
		msgChanMap[utils.Bytes2string(msg.To.GetAddr())] = mep
	}
	msgChanMapLock.Unlock()
	ERR := mep.addMsg(&msg)
	return ERR
}

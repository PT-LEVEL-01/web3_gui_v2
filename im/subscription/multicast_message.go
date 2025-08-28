package subscription

import (
	"time"
	"web3_gui/im/model"
)

/*
广播消息缓存
*/
var multicastMsgChan = make(chan *model.MessageContentVO, 1000)

/*
添加一个广播消息
*/
func AddMuticastMsg(msg *model.MessageContentVO) {
	//utils.Log.Info().Msgf("推送消息")
	//尝试往里面放
	select {
	case multicastMsgChan <- msg:
		//utils.Log.Info().Msgf("推送消息")
		return
	default:
		//放不进去就取出来再放
		select {
		case <-multicastMsgChan:
			//待取出后再次放入
			select {
			case multicastMsgChan <- msg:
			default:
			}
		default:
		}
	}
}

/*
获取一个广播消息
*/
func GetMuticastMsg(timeout time.Duration) *model.MessageContentVO {
	timer := time.NewTimer(timeout)
	//放不进去就取出来再放
	select {
	case msg := <-multicastMsgChan:
		//utils.Log.Info().Msgf("推送消息")
		timer.Stop()
		return msg
	case <-timer.C:
	}
	return nil
}

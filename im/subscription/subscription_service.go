package subscription

import (
	"time"
	"web3_gui/config"
	"web3_gui/gui/tray"
	"web3_gui/im/model"
	"web3_gui/libp2parea/v2/node_store"
)

type SubscriptionMsgItr interface {
	GetSubscription() uint64 //获取通知类型
	GetNoticeKey() string    //
}

/*
消息订阅服务器
*/
var subscriptionChan = make(chan SubscriptionMsgItr, 1000)

/*
添加一个订阅消息
*/
func AddSubscriptionMsg(msg SubscriptionMsgItr) {
	//处理发送者昵称
	msgOne := msg.(*model.MessageContentVO)
	if msgOne.From != "" {
		from := nodeStore.AddressFromB58String(msgOne.From)
		info := CacheGetUserInfo(&from)
		if info != nil {
			msgOne.Nickname = info.Nickname
		}
	}
	msg = msgOne

	//utils.Log.Info().Msgf("订阅消息")
	//尝试往里面放
	select {
	case subscriptionChan <- msg:
		//utils.Log.Info().Msgf("订阅消息")
		return
	default:
		//放不进去就取出来再放
		select {
		case <-subscriptionChan:
			//utils.Log.Info().Msgf("订阅消息")
			//待取出后再次放入
			select {
			case subscriptionChan <- msg:
				//utils.Log.Info().Msgf("订阅消息")
			default:
				//utils.Log.Info().Msgf("订阅消息")
			}
		default:
			//utils.Log.Info().Msgf("订阅消息")
		}
	}
}

/*
获取一个订阅消息
*/
func GetSubscriptionMsg(timeout time.Duration) SubscriptionMsgItr {
	timer := time.NewTimer(timeout)
	//放不进去就取出来再放
	select {
	case msg := <-subscriptionChan:
		//utils.Log.Info().Msgf("订阅消息")
		timer.Stop()
		if msg.GetSubscription() == config.SUBSCRIPTION_type_msg {
			tray.SendNotice(true, msg.GetNoticeKey())
		} else if msg.GetSubscription() == config.SUBSCRIPTION_type_addFriend {
			tray.SendNotice(true, msg.GetNoticeKey())
		}
		//utils.Log.Info().Msgf("订阅消息")
		return msg
	case msg := <-multicastMsgChan:
		// utils.Log.Info().Msgf("有新消息了:%+v", msg)
		timer.Stop()
		return msg
	case <-timer.C:
	}
	return nil
}

package security_signal

// import (
// 	"web3_gui/libp2parea/v1/message_center"
// 	"web3_gui/libp2parea/v1/message_center/flood"
// 	_ "web3_gui/libp2parea/v1/message_center/security_signal/doubleratchet"
// 	"web3_gui/libp2parea/v1/nodeStore"
// 	"errors"
// 	"fmt"
// )

// /*
// 	发送消息
// 	@addr       crypto.Address    目标地址
// 	@content    []byte            消息内容
// */
// func SendMessage(addr nodeStore.AddressNet, content []byte) error {
// 	//搜索节点，获取节点公钥信息
// 	// message_search.SearchAddress(addr)

// 	fmt.Println("SendMessage", string(content))

// 	message, ok := message_center.SendP2pMsg(MSGID_SearchAddr, &addr, &content)
// 	if ok {
// 		fmt.Println("SendMsg 1111111")
// 		bs := flood.WaitRequest(CLASS_Hello, message.Body.Hash.B58String())
// 		if bs == nil {
// 			fmt.Println("SendMsg 222222222")
// 			return errors.New("发送共享文件消息失败，可能超时")
// 		}
// 		fmt.Println("SendMsg 3333333333")
// 		fmt.Println("收到的node", string(*bs))
// 		// nodeid := nodeStore.Parse(*bs)

// 		return nil
// 	}
// 	return nil

// 	// mhead := message_center.NewMessageHead(&addr, &addr, true)
// 	// mbody := message_center.NewMessageBody(&content, "", nil, 0)
// 	// message := message_center.NewMessage(mhead, mbody)
// 	// if message.Send(MSGID_SearchAddr) {
// 	// 	fmt.Println("SendMsg 1111111")
// 	// 	bs := flood.WaitRequest(CLASS_Hello, message.Body.Hash.B58String())
// 	// 	if bs == nil {
// 	// 		fmt.Println("SendMsg 222222222")
// 	// 		return errors.New("发送共享文件消息失败，可能超时")
// 	// 	}
// 	// 	fmt.Println("SendMsg 3333333333")
// 	// 	fmt.Println("收到的node", string(*bs))
// 	// 	// nodeid := nodeStore.Parse(*bs)

// 	// 	return nil
// 	// }
// 	// fmt.Println("SendMsg 44444444444")
// 	// return nil
// }

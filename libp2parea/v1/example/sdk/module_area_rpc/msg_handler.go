package module_area_rpc

import (
	"net/http"

	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/sdk/jsonrpc2/model"
	"web3_gui/libp2parea/v1/virtual_node"
)

/*
发送一个新的广播消息
*/
func SendMulticastMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)
	err = Area.SendMulticastMsg(MSGID_TEST_MULTICAST, &content)
	if err != nil {
		return model.Errcode(model.Exist, err.Error())
	}

	res, err = model.Tojson(true)

	return
}

/*
发送一个新的查找超级节点消息
*/
func SendSearchSuperMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	msg, err := Area.SendSearchSuperMsg(MSGID_TEST_SEARCH_SUPER, &recvAddr, &content)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(map[string]interface{}{
		"content": string(*msg.Body.Content),
	})

	return
}

/*
发送一个新的查找超级节点消息
*/
func SendSearchSuperMsgWaitRequest(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	msgBody, err := Area.SendSearchSuperMsgWaitRequest(MSGID_TEST_SEARCH_SUPER_WAIT, &recvAddr, &content, DEFAULT_TIMEOUT)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}
	res, err = model.Tojson(map[string]interface{}{
		"content": string(*msgBody),
	})
	return
}

/*
发送一个新消息
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func SendP2pMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	msg, sendok, sendself, err := Area.SendP2pMsg(MSGID_TEST_P2P, &recvAddr, &content)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}
	res, err = model.Tojson(map[string]interface{}{
		"content":   string(*msg.Body.Content),
		"send_ok":   sendok,
		"send_self": sendself,
	})

	return
}

/*
给指定节点发送一个消息
@return    *[]byte      返回的内容
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func SendP2pMsgWaitRequest(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	msgBody, sendok, sendself, err := Area.SendP2pMsgWaitRequest(MSGID_TEST_P2P_WAIT, &recvAddr, &content, DEFAULT_TIMEOUT)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(map[string]interface{}{
		"content":   string(*msgBody),
		"send_ok":   sendok,
		"send_self": sendself,
	})

	return
}

/*
发送一个加密消息，包括消息头也加密
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func SendP2pMsgHE(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	msg, sendok, sendself, err := Area.SendP2pMsgHE(MSGID_TEST_P2P_HE, &recvAddr, &content)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(map[string]interface{}{
		"content":   string(*msg.Body.Content),
		"send_ok":   sendok,
		"send_self": sendself,
	})
	return
}

/*
发送一个加密消息，包括消息头也加密
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func SendP2pMsgHEWaitRequest(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	msgBody, sendok, sendself, err := Area.SendP2pMsgHEWaitRequest(MSGID_TEST_P2P_HE_WAIT, &recvAddr, &content, DEFAULT_TIMEOUT)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(map[string]interface{}{
		"content":   string(*msgBody),
		"send_ok":   sendok,
		"send_self": sendself,
	})
	return
}

/*
网络中查询一个逻辑节点地址的真实地址
*/
func SearchVnodeId(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	vnodeAddressItr, ok := rj.Get("vnode_address")
	if !ok {
		return model.Errcode(model.NoField, "vnode_address")
	}

	vnodeAddressStr := vnodeAddressItr.(string)
	vnodeAddress := virtual_node.AddressFromB58String(vnodeAddressStr)
	addrVnode, err := Area.SearchVnodeId(&vnodeAddress)
	if err != nil || addrVnode == nil || len(*addrVnode) == 0 {
		return model.Errcode(model.Nomarl, err.Error())
	}

	virtualNode := virtual_node.AddressNetExtend(*addrVnode)

	res, err = model.Tojson(virtualNode.B58String())
	return
}

/*
* 发送一个新的查找超级节点消息，可以指定接收端和发送端的代理节点
 */
func SendSearchSuperMsgProxy(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	recvProxyAddrItr, ok := rj.Get("recv_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_proxy_address")
	}

	recvProxyAddrddrStr := recvProxyAddrItr.(string)
	recvProxyAddr := nodeStore.AddressFromB58String(recvProxyAddrddrStr)

	senderProxyAddrItr, ok := rj.Get("sender_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "sender_proxy_address")
	}

	senderProxyAddrddrStr := senderProxyAddrItr.(string)
	senderProxyAddr := nodeStore.AddressFromB58String(senderProxyAddrddrStr)

	msg, err := Area.SendSearchSuperMsgProxy(MSGID_TEST_SEARCH_SUPER_PROXY, &recvAddr, &recvProxyAddr, &senderProxyAddr, &content)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}
	res, err = model.Tojson(map[string]interface{}{
		"content": string(*msg.Body.Content),
	})
	return
}

/*
* 发送一个新的查找超级节点消息，可以指定接收端和发送端的代理节点
 */
func SendSearchSuperMsgProxyWaitRequest(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	// 获取接收代理节点地址
	recvProxyAddrItr, ok := rj.Get("recv_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_proxy_address")
	}

	recvProxyAddrddrStr := recvProxyAddrItr.(string)
	recvProxyAddr := nodeStore.AddressFromB58String(recvProxyAddrddrStr)

	senderProxyAddrItr, ok := rj.Get("sender_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "sender_proxy_address")
	}

	senderProxyAddrddrStr := senderProxyAddrItr.(string)
	senderProxyAddr := nodeStore.AddressFromB58String(senderProxyAddrddrStr)

	msgBody, err := Area.SendSearchSuperMsgProxyWaitRequest(MSGID_TEST_SEARCH_SUPER_PROXY_WAIT, &recvAddr, &recvProxyAddr, &senderProxyAddr, &content, DEFAULT_TIMEOUT)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}
	res, err = model.Tojson(map[string]interface{}{
		"content": string(*msgBody),
	})

	return
}

/*
* 发送一个新消息，可以指定接收端和发送端的代理节点
*	@return    *Message     返回的消息
*	@return    bool         是否发送成功
*	@return    bool         消息是发给自己
 */
func SendP2pMsgProxy(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	// 获取接收代理节点地址
	recvProxyAddrItr, ok := rj.Get("recv_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_proxy_address")
	}

	recvProxyAddrddrStr := recvProxyAddrItr.(string)
	recvProxyAddr := nodeStore.AddressFromB58String(recvProxyAddrddrStr)

	senderProxyAddrItr, ok := rj.Get("sender_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "sender_proxy_address")
	}

	senderProxyAddrddrStr := senderProxyAddrItr.(string)
	senderProxyAddr := nodeStore.AddressFromB58String(senderProxyAddrddrStr)

	msg, sendok, sendself, err := Area.SendP2pMsgProxy(MSGID_TEST_P2P_PROXY, &recvAddr, &recvProxyAddr, &senderProxyAddr, &content)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}
	res, err = model.Tojson(map[string]interface{}{
		"content":   string(*msg.Body.Content),
		"send_ok":   sendok,
		"send_self": sendself,
	})

	return
}

/*
* 给指定节点发送一个消息，可以指定接收端和发送端的代理节点
*	@return    *[]byte      返回的内容
*	@return    bool         是否发送成功
*	@return    bool         消息是发给自己
 */
func SendP2pMsgProxyWaitRequest(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	// 获取接收代理节点地址
	recvProxyAddrItr, ok := rj.Get("recv_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_proxy_address")
	}

	recvProxyAddrddrStr := recvProxyAddrItr.(string)
	recvProxyAddr := nodeStore.AddressFromB58String(recvProxyAddrddrStr)

	senderProxyAddrItr, ok := rj.Get("sender_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "sender_proxy_address")
	}

	senderProxyAddrddrStr := senderProxyAddrItr.(string)
	senderProxyAddr := nodeStore.AddressFromB58String(senderProxyAddrddrStr)

	msgBody, sendok, sendself, err := Area.SendP2pMsgProxyWaitRequest(MSGID_TEST_P2P_PROXY_WAIT, &recvAddr, &recvProxyAddr, &senderProxyAddr, &content, DEFAULT_TIMEOUT)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(map[string]interface{}{
		"content":   string(*msgBody),
		"send_ok":   sendok,
		"send_self": sendself,
	})

	return
}

/*
* 发送一个加密消息，包括消息头也加密，可以指定接收端和发送端的代理节点
*	@return    *Message     返回的消息
*	@return    bool         是否发送成功
*	@return    bool         消息是发给自己
 */
func SendP2pMsgHEProxy(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	// 获取接收代理节点地址
	recvProxyAddrItr, ok := rj.Get("recv_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_proxy_address")
	}

	recvProxyAddrddrStr := recvProxyAddrItr.(string)
	recvProxyAddr := nodeStore.AddressFromB58String(recvProxyAddrddrStr)

	senderProxyAddrItr, ok := rj.Get("sender_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "sender_proxy_address")
	}

	senderProxyAddrddrStr := senderProxyAddrItr.(string)
	senderProxyAddr := nodeStore.AddressFromB58String(senderProxyAddrddrStr)

	msg, sendok, sendself, err := Area.SendP2pMsgHEProxy(MSGID_TEST_P2P_HE_PROXY, &recvAddr, &recvProxyAddr, &senderProxyAddr, &content)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(map[string]interface{}{
		"content":   string(*msg.Body.Content),
		"send_ok":   sendok,
		"send_self": sendself,
	})
	return
}

/*
* 发送一个加密消息，包括消息头也加密，可以指定接收端和发送端的代理节点
*	@return    *Message     返回的消息
*	@return    bool         是否发送成功
*	@return    bool         消息是发给自己
 */
func SendP2pMsgHEProxyWaitRequest(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	contentItr, ok := rj.Get("content")
	if !ok {
		return model.Errcode(model.NoField, "content")
	}

	contentStr := contentItr.(string)
	content := []byte(contentStr)

	recvAddrItr, ok := rj.Get("recv_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_address")
	}

	recvAddrddrStr := recvAddrItr.(string)
	recvAddr := nodeStore.AddressFromB58String(recvAddrddrStr)

	// 获取接收代理节点地址
	recvProxyAddrItr, ok := rj.Get("recv_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_proxy_address")
	}

	recvProxyAddrddrStr := recvProxyAddrItr.(string)
	recvProxyAddr := nodeStore.AddressFromB58String(recvProxyAddrddrStr)

	senderProxyAddrItr, ok := rj.Get("sender_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "sender_proxy_address")
	}

	senderProxyAddrddrStr := senderProxyAddrItr.(string)
	senderProxyAddr := nodeStore.AddressFromB58String(senderProxyAddrddrStr)

	msgBody, sendok, sendself, err := Area.SendP2pMsgHEProxyWaitRequest(MSGID_TEST_P2P_HE_PROXY_WAIT, &recvAddr, &recvProxyAddr, &senderProxyAddr, &content, DEFAULT_TIMEOUT)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(map[string]interface{}{
		"content":   string(*msgBody),
		"send_ok":   sendok,
		"send_self": sendself,
	})
	return
}

package module_area_rpc

import (
	"net/http"
	"strconv"
	"time"

	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/sdk/jsonrpc2/model"
	"web3_gui/libp2parea/v1/virtual_node"
)

/*
等待网络自治完成
*/
func WaitAutonomyFinish(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var finish = make(chan bool, 1)
	go func() {
		Area.WaitAutonomyFinish()
		finish <- true
	}()

	select {
	case <-finish:
		res, err = model.Tojson(true)
	case <-time.After(time.Second * 1): // 超时1秒
		res, err = model.Tojson(false)
	}

	return
}

/*
等待虚拟节点网络自治完成
*/
func WaitAutonomyFinishVnode(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var finish = make(chan bool, 1)
	go func() {
		Area.WaitAutonomyFinishVnode()
		finish <- true
	}()

	select {
	case <-finish:
		res, err = model.Tojson(true)
	case <-time.After(time.Second * 1): // 超时1秒
		res, err = model.Tojson(false)
	}

	return
}

/*
获取本节点地址
*/
func GetNetId(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addressnet := Area.GetNetId()
	res, err = model.Tojson(addressnet.B58String())
	return
}

/*
获取本节点虚拟节点地址
*/
func GetVnodeId(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addressNetExtend := Area.GetVnodeId()
	res, err = model.Tojson(addressNetExtend.B58String())
	return
}

/*
获取idinfo
*/
func GetIdInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	idInfo := Area.GetIdInfo()
	res, err = model.Tojson(idInfo)
	return
}

/*
获取NodeSelf
*/
func GetNodeSelf(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	node := Area.GetNodeSelf()
	res, err = model.Tojson(node)
	return
}

/*
关闭所有网络连接
移动端关闭移动网络切换到wifi网络时调用
*/
func CloseNet(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	Area.CloseNet()
	res, err = model.Tojson(true)
	return
}

/*
重新链接网络
*/
func ReconnectNet(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	ok := Area.ReconnectNet()
	res, err = model.Tojson(ok)
	return
}

/*
检查是否在线(链接有没有断开)
*/
func CheckOnline(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	ok := Area.CheckOnline()
	res, err = model.Tojson(ok)
	return
}

/*
添加一个地址到白名单
*/
func AddWhiteList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		return model.Errcode(model.NoField, "address")
	}

	addrStr := addrItr.(string)
	addressNet := nodeStore.AddressFromB58String(addrStr)
	ok = Area.AddWhiteList(addressNet)
	res, err = model.Tojson(ok)
	return
}

/*
删除一个地址到白名单
*/
func RemoveWhiteList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		return model.Errcode(model.NoField, "address")
	}

	addrStr := addrItr.(string)
	addressNet := nodeStore.AddressFromB58String(addrStr)
	ok = Area.RemoveWhiteList(addressNet)
	res, err = model.Tojson(ok)
	return
}

/*
添加一个连接
*/
func AddConnect(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var port int
	ipItr, ok := rj.Get("ip")
	if !ok {
		return model.Errcode(model.NoField, "ip")
	}
	ipStr := ipItr.(string)

	portItr, ok := rj.Get("port")
	if !ok {
		if !rj.VerifyType("portItr", "string") {
			res, err = model.Errcode(model.TypeWrong, "port")
			return
		}
	}
	portS := portItr.(string)

	if portS != "" {
		port, err = strconv.Atoi(portS)
		if err != nil {
			res, err = model.Errcode(model.TypeWrong, "port")
			return
		}
	}

	_, err = Area.AddConnect(ipStr, uint16(port))
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(true)
	return
}

/*
搜索磁力节点网络地址
*/
func SearchNetAddr(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		return model.Errcode(model.NoField, "address")
	}

	addrStr := addrItr.(string)
	addressNet := nodeStore.AddressFromB58String(addrStr)

	netAddr, err := Area.SearchNetAddr(&addressNet)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(netAddr.B58String())
	return
}

/*
搜索磁力虚拟节点网络地址
*/
func SearchNetAddrVnode(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		return model.Errcode(model.NoField, "address")
	}

	addrStr := addrItr.(string)
	addrExtend := virtual_node.AddressFromB58String(addrStr)
	addrVnode, err := Area.SearchNetAddrVnode(&addrExtend)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(addrVnode)
	return
}

/*
获取所有网络连接
*/
func GetNetworkInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	logicAddrs := Area.GetNetworkInfo()
	addrs := make([]string, 0, len(*logicAddrs))
	for i := range *logicAddrs {
		addrs = append(addrs, (*logicAddrs)[i].B58String())
	}
	res, err = model.Tojson(addrs)
	return
}

/*
 *获取本节点所有连接节点
 */
func GetAllConnectNode(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	allNodes := Area.NodeManager.GetAllNodes()
	addrs := make([]string, 0, len(allNodes))
	for i := range allNodes {
		addrs = append(addrs, allNodes[i].IdInfo.Id.B58String())
	}
	res, err = model.Tojson(addrs)
	return
}

/*
 * 根据目标ip地址及端口添加到白名单
 */
func AddAddrWhiteList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var port int
	ipItr, ok := rj.Get("ip")
	if !ok {
		return model.Errcode(model.NoField, "ip")
	}
	ipStr := ipItr.(string)

	portItr, ok := rj.Get("port")
	if !ok {
		if !rj.VerifyType("portItr", "string") {
			res, err = model.Errcode(model.TypeWrong, "port")
			return
		}
	}
	portS := portItr.(string)

	if portS != "" {
		port, err = strconv.Atoi(portS)
		if err != nil {
			res, err = model.Errcode(model.TypeWrong, "port")
			return
		}
	}

	_, ok = Area.AddAddrWhiteList(ipStr, uint16(port))
	res, err = model.Tojson(ok)
	return
}

/*
 * 设置区域上帝地址信息
 */
func SetAreaGodAddr(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var port int
	ipItr, ok := rj.Get("ip")
	if !ok {
		return model.Errcode(model.NoField, "ip")
	}
	ipStr := ipItr.(string)

	portItr, ok := rj.Get("port")
	if !ok {
		if !rj.VerifyType("portItr", "string") {
			res, err = model.Errcode(model.TypeWrong, "port")
			return
		}
	}
	portS := portItr.(string)

	if portS != "" {
		port, err = strconv.Atoi(portS)
		if err != nil {
			res, err = model.Errcode(model.TypeWrong, "port")
			return
		}
	}

	ok, _ = Area.SetAreaGodAddr(ipStr, port)
	res, err = model.Tojson(ok)
	return
}

/*
 * 搜索磁力节点网络地址，可以指定发送端的代理节点
 */
func SearchNetAddrProxy(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		return model.Errcode(model.NoField, "address")
	}

	addrStr := addrItr.(string)
	addressNet := nodeStore.AddressFromB58String(addrStr)

	recvProxyItr, ok := rj.Get("recv_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "recv_proxy_address")
	}

	recvProxyStr := recvProxyItr.(string)
	recvProxyAddr := nodeStore.AddressFromB58String(recvProxyStr)

	senderProxyItr, ok := rj.Get("sender_proxy_address")
	if !ok {
		return model.Errcode(model.NoField, "sender_proxy_address")
	}

	senderProxyStr := senderProxyItr.(string)
	senderProxyAddr := nodeStore.AddressFromB58String(senderProxyStr)

	netAddr, err := Area.SearchNetAddrProxy(&addressNet, &recvProxyAddr, &senderProxyAddr)
	if err != nil {
		return model.Errcode(model.Nomarl, err.Error())
	}

	res, err = model.Tojson(netAddr)
	return
}

/*
 * 根据目标节点，返回排序后的虚拟节点地址列表
 */
func FindNearVnodesSearchVnode(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		return model.Errcode(model.NoField, "address")
	}

	addrStr := addrItr.(string)
	addressNet := nodeStore.AddressFromB58String(addrStr)

	includeSelfItr, ok := rj.Get("include_self")
	if !ok {
		if !rj.VerifyType("includeSelfItr", "bool") {
			res, err = model.Errcode(model.TypeWrong, "include_self")
			return
		}
	}
	includeSelf := includeSelfItr.(bool)

	includeIndex0Itr, ok := rj.Get("include_index0")
	if !ok {
		if !rj.VerifyType("includeIndex0Itr", "bool") {
			res, err = model.Errcode(model.TypeWrong, "include_index0")
			return
		}
	}
	includeIndex0 := includeIndex0Itr.(bool)

	logicAddrs := Area.Vm.FindNearVnodesSearchVnode((*virtual_node.AddressNetExtend)(&addressNet), nil, includeSelf, includeIndex0)
	addrs := make([]string, 0, len(logicAddrs))
	for i := range logicAddrs {
		addrs = append(addrs, logicAddrs[i].B58String())
	}
	res, err = model.Tojson(addrs)
	return
}

/*
 * 得到所有连接的节点信息，不包括本节点
 */
func GetAllNodes(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	logicAddrs := Area.NodeManager.GetAllNodes()
	addrs := make([]string, 0, len(logicAddrs))
	for i := range logicAddrs {
		addrs = append(addrs, logicAddrs[i].IdInfo.Id.B58String())
	}
	res, err = model.Tojson(addrs)
	return
}

/*
 * 获取设备机器Id
 */
func GetMachineID(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	machineId := Area.GetMachineID()
	res, err = model.Tojson(machineId)
	return
}

package rpc

import (
	"net/http"
	"strconv"

	"web3_gui/chain/config"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
)

/*
获取本节点网络信息，包括节点地址，本节点是否是超级节点
*/
func NetworkInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	netAddr := config.Area.GetNetId().B58String() // nodeStore.NodeSelf.IdInfo.Id.B58String()
	isSuper := false                              // nodeStore.NodeSelf.IsSuper
	nodes := config.Area.NodeManager.GetLogicNodes()
	ids := make([]string, 0)
	for _, one := range nodes {
		ids = append(ids, one.B58String())
	}

	// otherNodes := nodeStore.GetOtherNodes()
	// oids := make([]string, 0)
	// for _, one := range otherNodes {
	// 	oids = append(oids, one.B58String())
	// }

	m := make(map[string]interface{})
	m["netaddr"] = netAddr
	m["issuper"] = isSuper
	m["superNodes"] = ids
	// m["otherNodes"] = oids
	m["webaddr"] = config.WebAddr + ":" + strconv.Itoa(int(config.WebPort))
	m["tcpaddr"] = config.Init_LocalIP + ":" + strconv.Itoa(int(config.Init_LocalPort))

	res, err = model.Tojson(m)
	return
}

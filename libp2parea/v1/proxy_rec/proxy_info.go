package proxyrec

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
)

/*
 * 代理信息
 */
type ProxyInfo struct {
	NodeId    *nodeStore.AddressNet `json:"Id"`        // 客户端地址
	ProxyId   *nodeStore.AddressNet `json:"-"`         // 代理地址
	MachineId string                `json:"machineid"` // 机器id
	Version   int64                 `json:"version"`   // 版本号
	syncTime  time.Time             `json:"-"`         // 最后同步时间
}

/*
 * 生成proto信息
 */
func (pi *ProxyInfo) Proto() ([]byte, error) {
	proxyInfo := go_protobuf.ProxyInfo{
		Id:        *pi.NodeId,
		ProxyId:   *pi.ProxyId,
		MachineID: pi.MachineId,
		Version:   pi.Version,
	}
	return proxyInfo.Marshal()
}

/*
 * 解析代理信息
 */
func ParseProxyProto(bs *[]byte) (*ProxyInfo, error) {
	if bs == nil {
		return nil, nil
	}
	proxyInfo := new(go_protobuf.ProxyInfo)
	err := proto.Unmarshal(*bs, proxyInfo)
	if err != nil {
		return nil, err
	}

	var res ProxyInfo
	res.NodeId = (*nodeStore.AddressNet)(&proxyInfo.Id)
	res.ProxyId = (*nodeStore.AddressNet)(&proxyInfo.ProxyId)
	res.MachineId = proxyInfo.MachineID
	res.Version = proxyInfo.Version

	return &res, nil
}

/*
 * 解析代理数组信息
 */
func ParseProxyesProto(bs *[]byte) ([]ProxyInfo, error) {
	res := make([]ProxyInfo, 0)
	if bs == nil {
		return res, nil
	}
	proxyInfoes := new(go_protobuf.ProxyRepeated)
	err := proto.Unmarshal(*bs, proxyInfoes)
	if err != nil {
		return nil, err
	}

	for i := range proxyInfoes.Proxys {
		var proxyInfo ProxyInfo
		proxyInfo.NodeId = (*nodeStore.AddressNet)(&proxyInfoes.Proxys[i].Id)
		proxyInfo.ProxyId = (*nodeStore.AddressNet)(&proxyInfoes.Proxys[i].ProxyId)
		proxyInfo.MachineId = proxyInfoes.Proxys[i].MachineID
		proxyInfo.Version = proxyInfoes.Proxys[i].Version

		res = append(res, proxyInfo)
	}

	return res, nil
}

package virtual_node

import (
	"bytes"
	"time"

	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/utils"
)

const (
	max_vnode_index = 100000000
)

type Vnodeinfo struct {
	Nid   nodeStore.AddressNet `json:"nid"`   //节点真实网络地址
	Index uint64               `json:"index"` //节点第几个空间，从1开始,下标为0的节点为实际节点。
	Vid   AddressNetExtend     `json:"vid"`   //vid，虚拟节点网络地址
	// lastContactTimestamp time.Time  //最后检查的时间戳
}

type VnodeinfoS struct {
	Nid   nodeStore.AddressNet `json:"nid"`   //节点真实网络地址
	Index uint64               `json:"index"` //节点第几个空间，从1开始,下标为0的节点为实际节点。
	Vid   AddressNetExtend     `json:"vid"`   //vid，虚拟节点网络地址
	// lastContactTimestamp time.Time  //最后检查的时间戳
	Addr     string `json:"addr"`     //真实节点IP地址
	TcpPort  uint64 `json:"tcpport"`  //真实节点tcp端口
	QuicPort uint64 `json:"quicport"` //真实节点quic端口
}

func (this *Vnodeinfo) Proto() ([]byte, error) {
	vnodeinfo := go_protobuf.Vnodeinfo{
		Nid:   this.Nid,   //节点真实网络地址
		Index: this.Index, //节点第几个空间，从1开始,下标为0的节点为实际节点。
		Vid:   this.Vid,   //vid，虚拟节点网络地址
	}
	return vnodeinfo.Marshal()
}

/*
验证节点id是否合法
*/
func (this *Vnodeinfo) Check() bool {

	var newAddressNetExtend AddressNetExtend

	if this.Index == 0 {
		newAddressNetExtend = AddressNetExtend(this.Nid)
	} else if this.Index > max_vnode_index {
		return false
	} else {
		buf := bytes.NewBuffer(utils.Uint64ToBytes(this.Index))
		buf.Write(this.Nid)
		hashBs := utils.Hash_SHA3_256(buf.Bytes())
		newAddressNetExtend = AddressNetExtend(hashBs)
	}

	if bytes.Equal(newAddressNetExtend, this.Vid) {
		return true
	}
	return false
}

func BuildNodeinfo(index uint64, addrNet nodeStore.AddressNet) *Vnodeinfo {
	vnodeInfo := Vnodeinfo{
		Nid:   addrNet, //
		Index: index,   //节点第几个空间，从0开始。
		// Vid:   addressNetExtend, //vid，虚拟节点网络地址
	}
	if index == 0 {
		vnodeInfo.Vid = AddressNetExtend(addrNet)
	} else if index > max_vnode_index {
		return nil
	} else {
		buf := bytes.NewBuffer(utils.Uint64ToBytes(index))
		buf.Write(addrNet)

		hashBs := utils.Hash_SHA3_256(buf.Bytes())
		addressNetExtend := AddressNetExtend(hashBs)
		vnodeInfo.Vid = addressNetExtend
	}
	return &vnodeInfo
}

/*
带超时的vnodeinfo
*/
type VnodeinfoTimeout struct {
	Vnodeinfo
	timeout time.Time
}

/*
创建一个带超时的Vnodeinfo
*/
func CreateVnodeinfoTimeout(vnodeinfo *Vnodeinfo) *VnodeinfoTimeout {
	return &VnodeinfoTimeout{
		Vnodeinfo: *vnodeinfo,
		timeout:   time.Now(),
	}
}

/*
刷新时间
*/
func (this *VnodeinfoTimeout) FlashTime() {
	this.timeout = time.Now()
}

/*
检查时间是否过期
@return    bool    是否过期。true=过期了;false=未过期。
*/
func (this *VnodeinfoTimeout) CheckTimeout(outtime time.Duration) bool {
	interval := time.Now().Sub(this.timeout)
	if outtime > interval {
		return false
	}
	return true
}

package imdatachain

import (
	"github.com/oklog/ulid/v2"
	"time"
	"web3_gui/im/protos/go_protos"
	"web3_gui/libp2parea/v2/node_store"
)

type DataChainClientItr interface {
	Proto() (*[]byte, error)                //格式化成proto字节
	GetClientCmd() int                      //获取客户端解析命令
	SetProxyItr(proxyItr DataChainProxyItr) //
	GetProxyItr() DataChainProxyItr         //
	CheckCmd() bool                         //检查命令是否正确
	GetAddrFrom() nodeStore.AddressNet      //
	GetAddrTo() nodeStore.AddressNet        //
}

type DataChainClientBase struct {
	proxyItr DataChainProxyItr    //
	Command  int                  //日志指令
	Random   []byte               //随机数
	AddrFrom nodeStore.AddressNet //发送者
	AddrTo   nodeStore.AddressNet //发送给谁
	Time     int64                //时间
}

func NewDataChainClientBase(cmd int, addrSelf, addrFriend nodeStore.AddressNet) *DataChainClientBase {
	return &DataChainClientBase{
		Command:  cmd,
		Random:   ulid.Make().Bytes(),
		AddrFrom: addrSelf,   //发送者
		AddrTo:   addrFriend, //发送给谁
		Time:     time.Now().Unix(),
	}
}

func (this *DataChainClientBase) GetProto() *go_protos.ImClientBase {
	base := go_protos.ImClientBase{
		Command:  uint64(this.Command),
		Random:   this.Random,
		AddrFrom: this.AddrFrom.GetAddr(),
		AddrTo:   this.AddrTo.GetAddr(),
		Time:     this.Time,
	}
	return &base
}

func (this *DataChainClientBase) Proto() (*[]byte, error) {
	base := go_protos.ImDataChainClientBase{
		Base: &go_protos.ImClientBase{
			Command:  uint64(this.Command),
			Random:   this.Random,
			AddrFrom: this.AddrFrom.GetAddr(),
			AddrTo:   this.AddrTo.GetAddr(),
			Time:     this.Time,
		},
	}
	bs, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
获取客户端解析命令
*/
func (this *DataChainClientBase) GetClientCmd() int {
	return this.Command
}

/*
设置代理节点
*/
func (this *DataChainClientBase) SetProxyItr(proxyItr DataChainProxyItr) {
	this.proxyItr = proxyItr
}

/*
获取代理节点
*/
func (this *DataChainClientBase) GetProxyItr() DataChainProxyItr {
	return this.proxyItr
}

/*
获取代理节点
*/
func (this *DataChainClientBase) GetAddrFrom() nodeStore.AddressNet {
	return this.AddrFrom
}

/*
获取代理节点
*/
func (this *DataChainClientBase) GetAddrTo() nodeStore.AddressNet {
	return this.AddrTo
}

func ConvertClientBase(basePro *go_protos.ImClientBase) *DataChainClientBase {
	base := DataChainClientBase{
		proxyItr: nil,                                        //
		Command:  int(basePro.Command),                       //日志指令
		Random:   basePro.Random,                             //随机数
		AddrFrom: *nodeStore.NewAddressNet(basePro.AddrFrom), //发送者
		AddrTo:   *nodeStore.NewAddressNet(basePro.AddrTo),   //发送给谁
		Time:     basePro.Time,                               //
	}
	return &base
}

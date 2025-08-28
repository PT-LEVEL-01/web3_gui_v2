package utils

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/utils/protos/go_protos"
)

/*
网络发送参数
*/
type NetParams struct {
	Version uint64 //协议版本号
	Data    []byte //参数
}

/*
创建一个网络发送参数对象
*/
func NewNetParams(version uint64, data []byte) *NetParams {
	return &NetParams{
		Version: version,
		Data:    data,
	}
}

func (this *NetParams) Proto() (*[]byte, error) {
	bhp := go_protos.NetResult{
		Version: this.Version,
		Data:    this.Data,
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
解析网络接收参数
*/
func ParseNetParams(bs []byte) (*NetParams, error) {
	if bs == nil {
		return nil, nil
	}
	sdi := new(go_protos.NetResult)
	err := proto.Unmarshal(bs, sdi)
	if err != nil {
		return nil, err
	}
	nr := NetParams{
		Version: sdi.Version,
		Data:    sdi.Data,
	}
	return &nr, nil
}

/*
网络返回参数
*/
type NetResult struct {
	Version uint64 //协议版本号
	Code    uint64 //错误编码
	Msg     string //错误信息
	Data    []byte //返回参数
}

/*
检查是否成功
*/
func (this *NetResult) CheckSuccess() bool {
	if this.Code == ERROR_CODE_success {
		return true
	}
	return false
}

/*
转化成错误
*/
func (this *NetResult) ConvertERROR() ERROR {
	if this.Code == ERROR_CODE_success {
		return NewErrorSuccess()
	}
	//对方的系统错误，就是自己的远端节点错误
	if this.Code == ERROR_CODE_system_error_self {
		return NewErrorSysRemote(this.Msg)
	}
	return NewErrorBus(this.Code, this.Msg)
}

func NewNetResult(version, code uint64, msg string, data []byte) *NetResult {
	nr := NetResult{
		Version: version,
		Code:    code,
		Msg:     msg,
		Data:    data,
	}
	return &nr
}

func (this *NetResult) Proto() (*[]byte, error) {
	bhp := go_protos.NetResult{
		Version: this.Version,
		Code:    this.Code,
		Msg:     this.Msg,
		Data:    this.Data,
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

/*
解析网络返回错误编号
*/
func ParseNetResult(bs []byte) (*NetResult, error) {
	if bs == nil {
		return nil, nil
	}
	sdi := new(go_protos.NetResult)
	err := proto.Unmarshal(bs, sdi)
	if err != nil {
		return nil, err
	}
	nr := NetResult{
		Version: sdi.Version,
		Code:    sdi.Code,
		Msg:     sdi.Msg,
		Data:    sdi.Data,
	}
	return &nr, nil
}

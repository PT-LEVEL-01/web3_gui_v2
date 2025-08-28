package utils

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"sort"
	"strconv"
	"sync"
	"web3_gui/utils/protos/go_protos"
)

var (
	ERROR_CODE_success             = RegErrCodeExistPanic(0, "成功")         //成功
	ERROR_CODE_system_error_remote = RegErrCodeExistPanic(1, "系统错误--远程节点") //系统错误--远程节点
	ERROR_CODE_system_error_self   = RegErrCodeExistPanic(2, "系统错误--自己节点") //系统错误--自己节点
)

type ERROR struct {
	Code uint64 `json:"c"` //业务错误编号
	Msg  string `json:"m"` //错误信息
}

func (this *ERROR) Proto() (*[]byte, error) {
	bhp := go_protos.ERROR{
		Code: this.Code,
		Msg:  this.Msg,
	}
	bs, err := bhp.Marshal()
	return &bs, err
}

func ParseERROR(bs []byte) (*ERROR, error) {
	if bs == nil {
		return nil, nil
	}
	sdi := new(go_protos.ERROR)
	err := proto.Unmarshal(bs, sdi)
	if err != nil {
		return nil, err
	}
	nr := ERROR{
		Code: sdi.Code,
		Msg:  sdi.Msg,
	}
	return &nr, nil
}

/*
检查这个错误类型是否是成功
*/
func (this *ERROR) CheckSuccess() bool {
	if this.Code == ERROR_CODE_success || this.Code == 0 {
		return true
	}
	return false
}

/*
检查这个错误类型是否是失败
*/
func (this *ERROR) CheckFail() bool {
	return !this.CheckSuccess()
}

/*
json字符串
*/
func (this *ERROR) String() string {
	bs, err := json.Marshal(this)
	if err != nil {
		return err.Error()
	}
	return string(bs)
}

/*
创建一个系统错误
*/
func NewErrorSysSelf(err error) ERROR {
	if err == nil {
		return ERROR{
			Code: ERROR_CODE_success,
		}
	}
	return ERROR{
		Code: ERROR_CODE_system_error_self,
		Msg:  err.Error(),
	}
}

/*
创建一个远端节点的系统错误
*/
func NewErrorSysRemote(msg string) ERROR {
	return ERROR{
		Code: ERROR_CODE_system_error_remote,
		Msg:  msg,
	}
}

/*
创建一个业务错误
*/
func NewErrorBus(code uint64, msg string) ERROR {
	return ERROR{
		Code: code,
		Msg:  msg,
	}
}

/*
创建一个返回成功
*/
func NewErrorSuccess() ERROR {
	return ERROR{
		Code: ERROR_CODE_success,
	}
}

var errorCodeMap = new(sync.Map) //已经注册的错误编号。key:uint64=错误编码;value:string=错误描述;

/*
注册错误编号，判断是否重复
*/
func RegisterErrorCode(code uint64, desc string) error {
	//minNumber := uint64(0)
	//if code <= minNumber {
	//	Log.Error().Uint64("The error number must be greater than", minNumber)
	//	//Log.Error("The error number must be greater than %d", minNumber)
	//	return errors.New("The error number must be greater than:" + strconv.Itoa(int(minNumber)))
	//}
	_, ok := errorCodeMap.LoadOrStore(code, desc)
	if ok {
		Log.Error().Uint64("error code repeat", code)
		//Log.Error("error code repeat:%d", code)
		return errors.New("error code repeat:" + strconv.Itoa(int(code)))
	}
	return nil
}

/*
注册错误编号，有重复就panic
*/
func RegErrCodeExistPanic(code uint64, desc string) uint64 {
	if err := RegisterErrorCode(code, desc); err != nil {
		panic(err)
	}
	return code
}

func GetErrorsList() []ErrorDesc {
	errorList := make([]ErrorDesc, 0)
	errorCodeMap.Range(func(key, value any) bool {
		code := key.(uint64)
		desc := value.(string)
		errorList = append(errorList, ErrorDesc{Code: code, Desc: desc})
		return true
	})
	sort.Slice(errorList, func(i, j int) bool {
		return errorList[i].Code < errorList[j].Code
	})
	return errorList
}

type ErrorDesc struct {
	Code uint64
	Desc string
}

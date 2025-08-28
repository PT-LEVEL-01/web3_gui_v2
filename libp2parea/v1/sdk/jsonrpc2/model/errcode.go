package model

import (
	"fmt"
	"strconv"
	"sync"
	"web3_gui/utils"
)

const (
	Success   = 2000 //成功
	NoMethod  = 4001 //没有这个方法
	TypeWrong = 5001 //参数类型错误
	NoField   = 5002 //缺少参数
	Nomarl    = 5003 //一般错误，请看错误提示信息
	Timeout   = 5004 //超时
	Exist     = 5005 //已经存在
	FailPwd   = 5006 //密码错误
	NotExist  = 5007 //不存在
)

var codes = new(sync.Map)

func init() {
	RegisterErrcode(Success, "success")
	RegisterErrcode(NoMethod, "no method")
	RegisterErrcode(TypeWrong, "param type wrong")
	RegisterErrcode(NoField, "no field")
	RegisterErrcode(Nomarl, "")
	RegisterErrcode(Timeout, "time out")
	RegisterErrcode(Exist, "exist")
	RegisterErrcode(NotExist, "not exist")
	RegisterErrcode(FailPwd, "fail password")
}

func RegisterErrcode(code int, message string) {
	_, ok := codes.LoadOrStore(code, message)
	if ok {
		utils.Log.Info().Msgf("register rpc Errcode exist:%s", code)
	}
}

func Errcode(code int, p ...string) (res []byte, err error) {
	res = []byte(strconv.Itoa(code))
	cItr, ok := codes.Load(code)
	// c, ok := codes[code]
	if ok {
		c := cItr.(string)
		if len(p) > 0 {
			if c == "" {
				err = fmt.Errorf("%s", p[0])
			} else {
				err = fmt.Errorf("%s: %s", p[0], c)
			}
		} else {
			err = fmt.Errorf("%s", c)
		}
	}
	return
}

package model

import (
	"reflect"
)

type RpcHandler interface {
	SetBody(data []byte)
	GetBody() []byte
	Out(data []byte)
	Err(code, data string)
	Validate() (msg string, ok bool)
}
type RpcJson struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func (rj *RpcJson) Get(key string) (interface{}, bool) {
	v, b := rj.Params[key]
	return v, b
}
func (rj *RpcJson) Type(key string) string {
	v, b := rj.Get(key)
	if !b {
		return ""
	}
	return reflect.TypeOf(v).String()
}
func (rj *RpcJson) VerifyType(key, types string) bool {
	if rj.Type(key) == types {
		return true
	} else {
		return false
	}
}

/*
验证必须存在
*/
func (this *RpcJson) VerifyExistMust(key, types string) (res []byte, err error, itr interface{}) {
	ok := this.VerifyType(key, types)
	if !ok {
		res, err = Errcode(TypeWrong, key)
		return
	}
	itr, ok = this.Get(key)
	if !ok {
		res, err = Errcode(TypeWrong, key)
		return
	}
	return nil, nil, itr
}

/*
验证可选存在
*/
func (this *RpcJson) VerifyExistOptional(key, types string) (res []byte, err error, itr interface{}) {
	var ok = false
	itr, ok = this.Get(key)
	if !ok {
		return
	}
	ok = this.VerifyType(key, types)
	if !ok {
		res, err = Errcode(TypeWrong, key)
		return
	}
	return nil, nil, itr
}

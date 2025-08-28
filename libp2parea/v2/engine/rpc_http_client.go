package engine

import (
	"fmt"
	"net/http"
	"strings"
)

/*
参数用body传递
*/
func Post(addr, rpcUser, rpcPwd, rpcMethod string, params map[string]interface{}) (*PostResult, error) {
	if params == nil {
		params = make(map[string]interface{})
	}
	method := "POST"
	client := &http.Client{}
	params[RPC_username] = rpcUser
	params[RPC_password] = rpcPwd
	params[RPC_method] = rpcMethod
	bs, err := json.Marshal(params)
	req, err := http.NewRequest(method, "http://"+addr+HTTP_URL_rpc_json, strings.NewReader(string(bs)))
	if err != nil {
		fmt.Println("创建request错误")
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求服务器错误", err.Error())
		return nil, err
	}
	result := new(PostResult)
	if resp.StatusCode == 200 {
		//buf := bytes.NewBuffer(nil)
		//buf.ReadFrom(resp.Body)
		////utils.Log.Info().Hex("body 内容", buf.Bytes()).Send()
		//decoder := json.NewDecoder(buf)
		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber()
		err = decoder.Decode(result)
		if err != nil {
			return nil, err
		}
	}
	result.HTTPCode = resp.StatusCode
	return result, nil
}

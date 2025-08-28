package main

import (
	"encoding/json"
	"fmt"
	"time"
	"web3_gui/chain/test/test_case/config"
	"web3_gui/chain/test/test_case/request"

	"web3_gui/libp2parea/adapter/engine"
)

// 通过超级节点给普通节点转账 然后普通节点发起质押demo
// 通过超级节点给普通节点转账 然后普通节点发起质押demo
// 通过超级节点给普通节点转账 然后普通节点发起质押demo
// 通过超级节点给普通节点转账 然后普通节点发起质押demo
// 通过超级节点给普通节点转账 然后普通节点发起质押demo
// 通过超级节点给普通节点转账 然后普通节点发起质押demo

//var superNode = "http://47.109.47.85:2080/rpc"
//var nodeList = []string{"http://47.109.42.175:2080/rpc", "http://47.109.38.162:2080/rpc", "http://47.108.165.236:2080/rpc", "http://47.108.56.49:2080/rpc"}

var superNode = "http://127.0.0.1:2080/rpc"
var nodeList = []string{"http://127.0.0.1:2082/rpc"}

// 获取指定节点的第一个地址
func GetAddress(url string) (addr string) {

	// 定义头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	mutlParams := map[string]interface{}{"page": 1, "page_size": 1}

	// 构建参数
	params := map[string]interface{}{"method": "listaccounts", "params": mutlParams}
	byteData, _ := json.Marshal(params)
	_, resultJson := request.Http(url, "POST", header, byteData)
	result, _ := request.JsonToStruct(resultJson)
	return result.Result.([]interface{})[0].(map[string]interface{})["AddrCoin"].(string)

}

// 发起转账
func SendAddress(superAddrMap, addrMap map[string]string) {
	//fmt.Println(superAddrMap)
	//fmt.Println(addrMap)
	//return

	engine.Log.Info("--------------------------给账号转账中--------------------------")

	// 定义头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	for _, v := range addrMap {

		addressPair := []string{superAddrMap[superNode], v}

		_, params := getParamsV1("sendtoaddress", addressPair, 100000000000000)

		err, resultJson := request.Http(superNode, "POST", header, params)
		if err != nil {
			engine.Log.Error("请求出错：%s", err.Error())
			return
		} else {
			result, err := request.JsonToStruct(resultJson)
			if err != nil {
				engine.Log.Error("请求出错：%s", err.Error())
				return
			}

			if result.Code != 2000 {
				engine.Log.Error("请求出错：%s", resultJson)
				return
			}
		}
	}
	engine.Log.Info("--------------------------账号转账结束--------------------------")
}

// 发起质押
func SendDepositin(addrMap map[string]string) {

	engine.Log.Info("--------------------------发起质押中--------------------------")

	// 定义头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	for k, _ := range addrMap {

		mutlParams := map[string]interface{}{
			"amount":  10000000000000,
			"gas":     10000,
			"pwd":     "xhy19liu21@",
			"payload": "",
			"rate":    10,
		}

		// 构建参数
		params := map[string]interface{}{"method": "depositin", "params": mutlParams}
		byteData, _ := json.Marshal(params)

		err, resultJson := request.Http(k, "POST", header, byteData)
		if err != nil {
			engine.Log.Error("请求出错：%s", err.Error())
			return
		} else {
			result, err := request.JsonToStruct(resultJson)
			if err != nil {
				engine.Log.Error("请求出错：%s", err.Error())
				return
			}

			if result.Code != 2000 {
				engine.Log.Error("请求出错：%s", resultJson)
				return
			}
			fmt.Println(result)
		}
	}
	engine.Log.Info("--------------------------发起质押结束--------------------------")
}

// 生成指定并发参数
func getParamsV1(method string, address []string, amount int) (newUrl string, byteData []byte) {

	// 构建参数
	paramsData := map[string]interface{}{
		"srcaddress":    address[0],
		"address":       address[1],
		"changeaddress": address[0],
		"amount":        amount,
		"gas":           100000,
		"frozen_height": 7,
		"pwd":           config.SupperNodePwd,
		"comment":       "test",
	}
	params := map[string]interface{}{"method": method, "params": paramsData}

	byteData, _ = json.Marshal(params)

	return
}

func main() {

	superAddrMap := map[string]string{}
	addrMap := map[string]string{}

	// 获取超级节点的地址
	addr := GetAddress(superNode)
	superAddrMap[superNode] = addr

	// 获取普通节点地址
	for _, v := range nodeList {
		addr = GetAddress(v)
		addrMap[v] = addr
	}

	// 发起转账
	SendAddress(superAddrMap, addrMap)
	time.Sleep(5 * time.Second)

	// 发起质押
	SendDepositin(addrMap)
}

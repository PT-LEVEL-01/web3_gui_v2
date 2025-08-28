package main

import (
	"encoding/json"
	"time"
	"web3_gui/chain/test/test_case/config"
	"web3_gui/chain/test/test_case/request"

	"web3_gui/libp2parea/adapter/engine"
)

// 新增账户成为社区节点demo
// 新增账户成为社区节点demo
// 新增账户成为社区节点demo
// 新增账户成为社区节点demo
// 新增账户成为社区节点demo
// 新增账户成为社区节点demo

var url = "http://127.0.0.1:2080/rpc"
var superAddress = "MMSPnuTb9HwUHTsr8FqrtjvCXTacPwHjoPef4"
var num = 200 // 创建账户数量

// 创建账户
func CreateAccount() (account []string) {

	engine.Log.Info("--------------------------创建账户中--------------------------")

	// 定义头部数据
	header := map[string]string{"user": "test", "password": "testp"}

	for i := 0; i < num; i++ {

		mutlParams := map[string]string{"password": "xhy19liu21@"}
		// 构建参数
		params := map[string]interface{}{"method": "getnewaddress", "params": mutlParams}
		byteData, _ := json.Marshal(params)

		err, resultJson := request.Http(url, "POST", header, byteData)

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
			account = append(account, result.Result.(map[string]interface{})["address"].(string))
		}
	}
	engine.Log.Info("--------------------------创建账户中--------------------------")
	return
}

// 生成指定并发参数
func getParamsV2(method string, address []string, amount int) (newUrl string, byteData []byte) {

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

// 发起转账
func SendAddress1(account []string) {
	//fmt.Println(superAddrMap)
	//fmt.Println(addrMap)
	//return

	engine.Log.Info("--------------------------给账号转账中--------------------------")

	// 定义头部数据
	header := map[string]string{"user": "test", "password": "testp"}
	// 防止太快减掉20
	limitNum := config.TransactionLimitNum - 20
	i := 0
	for _, v := range account {

		// 因为每个地址只能最对有64笔交易 为了防止出错就中断一会儿
		i++
		if i%limitNum == 0 {
			time.Sleep(5 * time.Second)
		}

		addressPair := []string{superAddress, v}

		_, params := getParamsV2("sendtoaddress", addressPair, 100000000000000)

		err, resultJson := request.Http(url, "POST", header, params)
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

// 发起投票
func Votein(account []string) {
	//fmt.Println(superAddrMap)
	//fmt.Println(addrMap)
	//return

	engine.Log.Info("--------------------------投票中--------------------------")

	// 定义头部数据
	header := map[string]string{"user": "test", "password": "testp"}

	for _, v := range account {

		mutlParams := map[string]interface{}{
			"votetype": 1,
			"address":  v,
			"witness":  superAddress,
			"amount":   100000000000,
			"gas":      100000000,
			"pwd":      "xhy19liu21@",
			"payload":  "",
			"rate":     10,
		}
		// 构建参数
		params := map[string]interface{}{"method": "votein", "params": mutlParams}
		byteData, _ := json.Marshal(params)

		err, resultJson := request.Http(url, "POST", header, byteData)

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
	engine.Log.Info("--------------------------投票结束--------------------------")
}

func main() {
	// 创建账号
	account := CreateAccount()

	// 转账
	SendAddress1(account)

	time.Sleep(5 * time.Second)

	// 成为社区节点
	Votein(account)

}

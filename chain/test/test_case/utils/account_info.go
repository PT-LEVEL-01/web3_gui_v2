package utils

import (
	"encoding/json"
	"web3_gui/chain/test/test_case/config"
	"web3_gui/chain/test/test_case/request"

	"web3_gui/libp2parea/adapter/engine"
)

// 请求信息
type AccountInfo struct {
	AccountAddress    []string // 账号地址总集合
	AccountNewAddress []string // 账号新增的地址集合 用于后面转账
}

// 创建账户
func (c *AccountInfo) CreateAccount(accountOld []string, num int) {

	engine.Log.Info("--------------------------创建测试账号地址中--------------------------")

	// 判断已有账户数量是否满足当前测试
	if len(accountOld) >= num {
		c.AccountAddress = append(c.AccountAddress, accountOld[:num]...)
		return
	} else {
		c.AccountAddress = append(c.AccountAddress, accountOld[:]...)
		num -= len(accountOld)
	}

	// 定义头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	for i := 0; i < num; i++ {
		// 构建参数
		paramsT := map[string]interface{}{"password": config.SupperNodePwd}
		params := map[string]interface{}{"method": "getnewaddress", "params": paramsT}
		byteData, _ := json.Marshal(params)

		err, resultJson := request.Http(config.Url, "POST", header, byteData)
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

			c.AccountAddress = append(c.AccountAddress, result.Result.(map[string]interface{})["address"].(string))
			c.AccountNewAddress = append(c.AccountNewAddress, result.Result.(map[string]interface{})["address"].(string))
		}
	}
}

// 根据并发数构建转账账户切片
func (c *AccountInfo) GetMinValueAccount() (account []string) {

	// 定义头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	// 构建参数
	paramsT := map[string]interface{}{"page": 1, "page_size": 1000}
	params := map[string]interface{}{"method": "listaccounts", "params": paramsT}
	byteData, _ := json.Marshal(params)

	err, resultJson := request.Http(config.Url, "POST", header, byteData)
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

		// 取出当前钱包大于指定余额的账户
		resultSlice := result.Result.([]interface{})[1:]
		for _, v := range resultSlice {
			tempData := v.(map[string]interface{})
			value, _ := tempData["Value"].(json.Number).Int64()
			index, _ := tempData["Index"].(json.Number).Int64()

			if value > int64(config.MinValue) && index != 0 {
				account = append(account, tempData["AddrCoin"].(string))
			}
		}
	}

	return
}

// 根据并发数构建转账账户切片
func (c *AccountInfo) CreateAccountSlice(repeatNum int) (accountSlice [][]string) {

	// 把账号分成两两一组
	tempSlice := [][]string{}
	tempSliceOne := []string{}
	i := 0
	for _, v := range c.AccountAddress {
		tempSliceOne = append(tempSliceOne, v)
		i++
		if i == 2 {
			tempSlice = append(tempSlice, tempSliceOne)
			tempSliceOne = []string{}
			i = 0
		}
	}

	// 限制一个账号只能接受60个转账交易
	limitNum := config.TransactionLimitNum
	countNum := 0
	for _, v := range tempSlice {
		for i := 0; i < limitNum; i++ {
			if countNum >= repeatNum {
				return
			}
			countNum++
			accountSlice = append(accountSlice, v)
		}
	}

	return
}

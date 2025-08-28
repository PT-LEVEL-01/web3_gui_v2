package main

import (
	"encoding/json"
	"flag"
	"math"
	"sync"
	"time"
	"web3_gui/chain/test/test_case/config"
	"web3_gui/chain/test/test_case/request"
	"web3_gui/chain/test/test_case/utils"

	"web3_gui/libp2parea/adapter/engine"
)

// 生成指定并发参数
func getParams(method string, address []string, amount int) (newUrl string, byteData []byte) {

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

// 给每个账号转100万
func SendToAccount(address []string) {

	engine.Log.Info("--------------------------给账号转账中--------------------------")

	// 定义头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	// 防止太快减掉20
	limitNum := config.TransactionLimitNum - 20
	i := 0

	for _, v := range address {

		// 因为每个地址只能最对有64笔交易 为了防止出错就中断一会儿
		i++
		if i%limitNum == 0 {
			time.Sleep(2 * time.Second)
		}
		addressPair := []string{config.SupperNodeAddress, v}

		_, params := getParams("sendtoaddress", addressPair, 100000000000000)

		err, resultJson := request.Http(config.Url, "POST", header, params)
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

func main() {

	// 初始化
	var repeatNum = flag.Int("num", 10, "并发数")
	var method = flag.String("method", "sendtoaddress", "请求函数")
	flag.Parse()

	// 根据并发数计算需要创建多少个账号 (并发数/config.TransactionLimitNum) * 2
	accountNum := int(math.Ceil(float64(*repeatNum)/float64(config.TransactionLimitNum))) * 2

	// 定义请求头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	// 创建指定个数账号
	a := new(utils.AccountInfo)

	// 获取当前钱包超过指定余额的账户地址
	accountOld := a.GetMinValueAccount()

	// 创建测试账户
	a.CreateAccount(accountOld, accountNum)

	// 给每个新增的指定账号转钱
	SendToAccount(a.AccountNewAddress)

	// 构建转账的账户组切片
	accountAddress := a.CreateAccountSlice(*repeatNum)

	// 转账需要等一会才能到账
	if len(a.AccountNewAddress) != 0 {
		engine.Log.Info("--------------------------等到到账中--------------------------")
		time.Sleep(10 * time.Second)
	}

	// 并发请求统计信息
	c := utils.StatisticsInfo{SuccessNum: 0, FailNum: 0}

	// 交易是否上链统计信息
	s := utils.TransactionInfo{SuccessNum: 0, FailNum: 0}

	//
	c.SetRepeatNum(*repeatNum)

	// 请求开始时间
	c.SetStartTime()

	var wg sync.WaitGroup
	dFunc := func(method string, i int) {
		defer wg.Done()

		_, params := getParams(method, accountAddress[i], 10000000000)

		err, resultJson := request.Http(config.Url, "POST", header, params)
		if err != nil {
			engine.Log.Error("请求出错：%s", err.Error())
			c.FailRequestsNumInc()
			return
		} else {
			result, err := request.JsonToStruct(resultJson)
			if err != nil {
				engine.Log.Error("请求出错：%s", err.Error())
				c.FailRequestsNumInc()
				return
			}

			c.SuccessRequestsInc()
			if result.Code != 2000 {
				engine.Log.Error("请求出错：%s", resultJson)
				c.FailNumInc()
				return
			}
			c.SuccessNumInc()

			// 存交易id 后面统计上链多少数量
			s.SetTxIds(result.Result.(map[string]interface{})["hash"].(string))
		}
	}

	wg.Add(*repeatNum)
	for i := 0; i < *repeatNum; i++ {
		//time.Sleep(1 * time.Millisecond)
		go dFunc(*method, i)
	}
	wg.Wait()

	// 请求结束时间
	c.SetEndTime()

	// 打印并发信息
	c.PrintInfo()

	// 延时 刚生成的交易不会立马上链
	time.Sleep(5 * time.Second)

	// 统计交易成功上链。并打印
	s.Statistics()

}

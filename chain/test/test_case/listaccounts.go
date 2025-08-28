package main

import (
	"encoding/json"
	"flag"
	"sync"
	"web3_gui/chain/test/test_case/config"
	"web3_gui/chain/test/test_case/request"
	"web3_gui/chain/test/test_case/utils"

	"web3_gui/libp2parea/adapter/engine"
)

func main() {
	// 初始化
	var repeatNum = flag.Int("num", 100, "并发数")
	var method = flag.String("method", "listaccounts", "请求函数")
	flag.Parse()

	// 定义头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	// 构建参数
	params := map[string]interface{}{"method": *method}
	byteData, _ := json.Marshal(params)

	c := utils.StatisticsInfo{SuccessNum: 0, FailNum: 0}

	//
	c.SetRepeatNum(*repeatNum)

	// 请求开始时间
	c.SetStartTime()

	var wg sync.WaitGroup
	dFunc := func() {
		defer wg.Done()

		err, resultJson := request.Http(config.Url, "POST", header, byteData)
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
		}
	}

	wg.Add(*repeatNum)
	for i := 0; i < *repeatNum; i++ {
		go dFunc()
	}
	wg.Wait()

	// 请求结束时间
	c.SetEndTime()

	// 打印信息
	c.PrintInfo()

}

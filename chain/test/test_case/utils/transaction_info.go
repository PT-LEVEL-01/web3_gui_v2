package utils

import (
	"encoding/json"
	"strings"
	"sync"
	"web3_gui/chain/test/test_case/config"
	"web3_gui/chain/test/test_case/request"

	"web3_gui/libp2parea/adapter/engine"
)

// 请求信息
type TransactionInfo struct {
	mu              sync.Mutex
	TxIds           []string // 交易id集合
	SuccessNum      int      // 成功请求数
	FailNum         int      // 失败请求数
	SuccessBlockNum int      // 成功上链数
	FailBlockNum    int      // 失败上链数
	SuccessTxIds    []string // 成功上链交易id集合
	FailTxIds       []string // 失败上链交易id集合
}

func (c *TransactionInfo) SuccessNumInc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SuccessNum++
}

func (c *TransactionInfo) FailNumInc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FailNum++
}

func (c *TransactionInfo) SuccessBlockNumInc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SuccessBlockNum++
}

func (c *TransactionInfo) FailBlockNumInc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FailBlockNum++
}

func (c *TransactionInfo) SetTxIds(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TxIds = append(c.TxIds, id)
}

func (c *TransactionInfo) SetSuccessTxIds(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SuccessTxIds = append(c.SuccessTxIds, id)
}

func (c *TransactionInfo) SetFailTxIds(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FailTxIds = append(c.FailTxIds, id)
}

func (c *TransactionInfo) Statistics() {

	RepeatNum := len(c.TxIds)

	// 定义头部数据
	header := map[string]string{"user": config.UserName, "password": config.Pwd}

	var wg sync.WaitGroup
	dFunc := func(txId string) {
		defer wg.Done()

		// 构建参数
		paramsT := map[string]interface{}{"txid": txId}
		params := map[string]interface{}{"method": "findtx", "params": paramsT}
		byteData, _ := json.Marshal(params)

		err, resultJson := request.Http(config.Url, "POST", header, byteData)
		if err != nil {
			engine.Log.Error("请求出错：%s", err.Error())
			c.FailNumInc()
			return
		} else {
			result, err := request.JsonToStruct(resultJson)
			if err != nil {
				engine.Log.Error("请求出错：%s", err.Error())
				c.FailNumInc()
				return
			}

			if result.Code != 2000 {
				engine.Log.Error("请求出错：%s", resultJson)
				engine.Log.Error("交易id：%s", txId)
				c.FailNumInc()
				c.FailBlockNumInc()
				c.SetFailTxIds(txId)
				return
			}
			// 成功请求数加1
			c.SuccessNumInc()
			// 成功上链数加1
			c.SuccessBlockNumInc()
			// 成功的交易id集合
			c.SetSuccessTxIds(txId)
		}
	}

	// 发起并发统计
	wg.Add(RepeatNum)
	for i := 0; i < RepeatNum; i++ {
		go dFunc(c.TxIds[i])
	}
	wg.Wait()

	// 打印统计信息
	c.PrintInfo()

}

// 打印测试信息
func (c *TransactionInfo) PrintInfo() {
	engine.Log.Info("--------------------------上链信息--------------------------")
	engine.Log.Info("总交易数：%d", len(c.TxIds))
	//engine.Log.Info("请求成功数：%d",c.SuccessNum)
	//engine.Log.Info("请求失败数：%d",c.FailNum)
	engine.Log.Info("成功上链数：%d", c.SuccessBlockNum)
	engine.Log.Info("失败上链数：%d", c.FailBlockNum)
	engine.Log.Info("失败交易id集合：%s", strings.Join(c.FailTxIds, ","))
	engine.Log.Info("--------------------------上链信息--------------------------")
}

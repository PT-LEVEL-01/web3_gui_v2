package utils

import (
	"fmt"
	"strconv"
	"sync"
	"time"
	"web3_gui/libp2parea/adapter/engine"
)

// 请求信息
type StatisticsInfo struct {
	mu                 sync.Mutex
	RepeatNum          int   // 当前测试并发数
	SuccessNum         int   // 成功数
	FailNum            int   // 失败数
	SuccessRequestsNum int   // 成功数
	FailRequestsNum    int   // 失败数
	StartTime          int64 // 程序开始时间(单位ms)
	EndTime            int64 // 程序结束时间
}

func (c *StatisticsInfo) SuccessNumInc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SuccessNum++
}

func (c *StatisticsInfo) FailNumInc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FailNum++
}

func (c *StatisticsInfo) SuccessRequestsInc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SuccessRequestsNum++
}

func (c *StatisticsInfo) FailRequestsNumInc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.FailRequestsNum++
}

func (c *StatisticsInfo) SetStartTime() {
	c.StartTime = time.Now().UnixNano() / 1e6
}

func (c *StatisticsInfo) SetEndTime() {
	c.EndTime = time.Now().UnixNano() / 1e6
}

// 获取程序运行总时长
func (c *StatisticsInfo) GetAllTime() int64 {
	return c.EndTime - c.StartTime
}

// 获取程序平均时长
func (c *StatisticsInfo) GetAverageTime() float64 {
	allTime := c.EndTime - c.StartTime
	average, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", float64(allTime)/float64(c.RepeatNum)), 64)
	return average
}

// 设置并发数
func (c *StatisticsInfo) SetRepeatNum(repeatNum int) {
	c.RepeatNum = repeatNum
}

// 打印测试信息
func (c *StatisticsInfo) PrintInfo() {
	engine.Log.Info("--------------------------并发信息--------------------------")
	engine.Log.Info("总请求并发数：%d", c.RepeatNum)
	engine.Log.Info("总耗时：%d ms", c.GetAllTime())
	engine.Log.Info("平均耗时：%.1f ms", c.GetAverageTime())
	engine.Log.Info("request请求成功数：%d", c.SuccessRequestsNum)
	engine.Log.Info("request请求失败数：%d", c.FailRequestsNum)
	engine.Log.Info("返回code=2000数：%d", c.SuccessNum)
	engine.Log.Info("返回code!=2000数：%d", c.FailNum)
	engine.Log.Info("--------------------------并发信息--------------------------")
}

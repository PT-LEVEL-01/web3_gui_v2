package tpstest

import (
	"sync"
	"sync/atomic"
	"time"

	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

/*
 * 测试条件:
 * 	1. 本地电脑进行测试
 *  2. 启动了6个节点
 *  3. 内存: 16G
 *  4. CPU: i5-12400 2.5GHZ 12核
 *  5. 系统: windows 10专业版
 *  6. 硬盘: HDD
 *  7. 协程: 10000-12000
 */

/*
 * 测试SearchNetAddr的tps数据
 *
 * 测试结果: 88000 - 92000
 */
func TPS_SearchNetAddr(area *libp2parea.Area, maxCnt int) {
	var wg sync.WaitGroup
	dstAddr := nodeStore.AddressFromB58String("FXjBPgu7rB3Hir9CRt16DRPXgLmUPcAanmfLSET2MVd5")
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := area.SearchNetAddr(&dstAddr)
			if err != nil {
				utils.Log.Error().Msgf("SearchNetAddr err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}

			atomic.AddInt32(&successCnt, 1)
			// utils.Log.Info().Msgf("SearchNetAddr addr:%s", getAddr.B58String())
		}()
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("SearchNetAddr TPS:%v", (successCnt*1000)/int32(useTime))
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 测试 SendSearchSuperMsgProxyWaitRequest 的tps数据
 *
 * 测试结果: 91000 - 98000
 */
func TPS_SendSearchSuperMsgProxyWaitRequest(area *libp2parea.Area, recvId, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, maxCnt int, msgId uint64) {
	var wg sync.WaitGroup
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := area.SendSearchSuperMsgProxyWaitRequest(msgId, recvId, recvProxyId, senderProxyId, content, 10*time.Second)
			if err != nil {
				utils.Log.Error().Msgf("SendSearchSuperMsgProxyWaitRequest err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}

			atomic.AddInt32(&successCnt, 1)
		}()
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("SearchNetAddr TPS:%v", (successCnt*1000)/int32(useTime))
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 测试SendSearchSuperMsgProxyWaitRequest的tps数据
 *
 * 测试结果: 84000 - 94000
 */
func TPS_SendSearchSuperMsgProxy(area *libp2parea.Area, recvId, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, maxCnt int, msgId uint64, resChan chan bool) {
	var wg sync.WaitGroup
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := area.SendSearchSuperMsgProxyWaitRequest(msgId, recvId, recvProxyId, senderProxyId, content, 10*time.Second)
			if err != nil {
				utils.Log.Error().Msgf("SendSearchSuperMsgProxyWaitRequest err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}

			select {
			case <-time.NewTimer(10 * time.Second).C:
				atomic.AddInt32(&failedCnt, 1)
			case <-resChan:
				atomic.AddInt32(&successCnt, 1)
			}
		}()
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("SearchNetAddr TPS:%v", (successCnt*1000)/int32(useTime))
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 测试 SendNeighborMsg 的tps数据
 *
 * 测试结果: 93000 - 106000
 */
func TPS_SendNeighborMsg(area *libp2parea.Area, recvId *nodeStore.AddressNet, content *[]byte, maxCnt int, msgId uint64, resChan chan bool) {
	var wg sync.WaitGroup
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := area.SendNeighborMsg(msgId, recvId, content)
			if err != nil {
				utils.Log.Error().Msgf("SendSearchSuperMsgProxyWaitRequest err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}

			select {
			case <-time.NewTimer(10 * time.Second).C:
				atomic.AddInt32(&failedCnt, 1)
			case <-resChan:
				atomic.AddInt32(&successCnt, 1)
			}
		}()
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("SearchNetAddr TPS:%v", (successCnt*1000)/int32(useTime))
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 测试 SendNeighborMsgWaitRequest 的tps数据
 *
 * 测试结果: 91000 - 98000
 */
func TPS_SendNeighborMsgWaitRequest(area *libp2parea.Area, recvId, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, maxCnt int, msgId uint64) {
	var wg sync.WaitGroup
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := area.SendSearchSuperMsgProxyWaitRequest(msgId, recvId, recvProxyId, senderProxyId, content, 10*time.Second)
			if err != nil {
				utils.Log.Error().Msgf("SendSearchSuperMsgProxyWaitRequest err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}

			atomic.AddInt32(&successCnt, 1)
		}()
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("SearchNetAddr TPS:%v", successCnt/int32(useTime)*1000)
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 测试 SearchNetAddrOneByOne 的tps数据
 *
 * 测试条件: 香港单台服务器部署6个节点, 一个客户端节点, 每次开启10000个协程, 请求10000次, 重复执行十次的结果
 * 测试结果:
 */
func TPS_SearchNetAddrOneByOne(area *libp2parea.Area, recvId, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, maxCnt int, msgId uint64, resChan chan bool) {
	var wg sync.WaitGroup
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			searchAddr := nodeStore.AddressFromB58String(GetRandomDomain())
			nodes, err := area.SearchNetAddrOneByOne(&searchAddr, 5)
			if err != nil || len(nodes) == 0 {
				utils.Log.Error().Msgf("area.SearchNetAddrOneByOne err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}

			atomic.AddInt32(&successCnt, 1)
		}()
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("SearchNetAddrOneByOne TPS:%v", (successCnt*1000)/int32(useTime))
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 测试 SearchNetAddrOneByOneProxy 的tps数据
 *
 * 测试条件: 香港单台服务器部署6个节点, 一个客户端节点, 每次开启10000个协程, 请求10000次, 重复执行十次的结果
 * 测试结果: 2561 4807 5002 4852 5076 3514 4492 4835 5005 5081
 */
func TPS_SearchNetAddrOneByOneProxy(area *libp2parea.Area, recvId, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, maxCnt int, msgId uint64, resChan chan bool) {
	var wg sync.WaitGroup
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			searchAddr := nodeStore.AddressFromB58String(GetRandomDomain())
			nodes, err := area.SearchNetAddrOneByOneProxy(&searchAddr, nil, senderProxyId, 5)
			if err != nil || len(nodes) == 0 {
				utils.Log.Error().Msgf("area.SearchNetAddrOneByOneProxy err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}

			atomic.AddInt32(&successCnt, 1)
		}()
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("SearchNetAddrOneByOneProxy TPS:%v", (successCnt*1000)/int32(useTime))
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 测试 SendP2pMsgProxyWaitRequest 的tps数据
 *
 * 测试条件: 香港单台服务器部署6个节点, 一个客户端节点, 每次开启10000个协程, 请求10000次, 重复执行十次的结果
 * 测试结果: 3979 9615 9756 10775 8000 6333 5219 4911 6146 4803
 */
func TPS_SendP2pMsgProxyWaitRequest(area *libp2parea.Area, recvId, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, maxCnt int, msgId uint64, resChan chan bool) {
	var wg sync.WaitGroup
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	serverAddrs := []string{
		// "E9TKqMNbiAjg4DgHHwqixdKamk4GPRwDGrEcrQ11HKn9",
		// "3Sd6PEDFqTcw4TzMjbqFsnbVPybx8xu9Dt3TXFvVEia1",
		// "7FN5DF6pnWZUepzDrD8Gp1DbyWk7B8QJFscBWbtTGF6E",
		// "6B75e6XFYhoyak8U67iRKvwxfvAyy17SmWGcfFdJkz5v",
		// "8JiTZHDsHwTeJQDcFUZaLRA4AMN1s6xe1nc9Gjq2N28p",
		// "BKYmFomN9bj8R74k7wVqUCifiiDZe4vdfnavwaBBGnbw",

		// 内网服务器地址
		"CFSMMxytmyyJsh6XYsMbpuKChgvUQQg9Cwb8eSc1XMRe",
		"9Vcxta7LTwk4seWC7KHm7CSiurzrPx4qCL1PAKvofNTs",
		"HE9FaiZBdbjgDjZxD285AQhPdiq4ZvSfMStSx6HRmVZ3",
		"9cLMUUKJM4jbBLqnz9Niwp8gUoaZFSgZCJvtrtHjf2iV",
		"CVtZh6mEWtibcePC2mNtL5TNdgnTAGYqdCoGFCUVfM2q",
		"6EbGTKsiTtJxSqBsi4dYWe19yGWGRbWSjQwnSmZrwgPr",
	}

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		ii := i % 6
		go func(nIndex int) {
			defer wg.Done()

			toAddr := nodeStore.AddressFromB58String(serverAddrs[nIndex])
			// _, err := area.SendSearchSuperMsgProxyWaitRequest(msgId, &nodes[ii], nil, senderProxyId, content, 10*time.Second)
			_, _, _, err := area.SendP2pMsgProxyWaitRequest(msgId, &toAddr, nil, senderProxyId, content, 10*time.Second)
			if err != nil {
				utils.Log.Error().Msgf("SendP2pMsgProxyWaitRequest err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}

			atomic.AddInt32(&successCnt, 1)
		}(ii)
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("SendP2pMsgProxyWaitRequest TPS:%v", (successCnt*1000)/int32(useTime))
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 测试 UnionApi 的tps数据
 *
 * 测试条件: 香港单台服务器部署6个节点, 一个客户端节点, 每次开启10000个协程, 请求10000次, 重复执行十次的结果
 * 测试结果: 2066 3238 3337 3537 3451 3598 3365 3332 2637 2886
 */
func TPS_UnionApi(area *libp2parea.Area, recvId, recvProxyId, senderProxyId *nodeStore.AddressNet, content *[]byte, maxCnt int, msgId uint64, resChan chan bool) {
	var wg sync.WaitGroup
	var successCnt, failedCnt int32

	startTime := time.Now().UnixMilli()

	for i := 0; i < maxCnt; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			searchAddr := nodeStore.AddressFromB58String(GetRandomDomain())
			nodes, err := area.SearchNetAddrOneByOneProxy(&searchAddr, recvProxyId, senderProxyId, 5)
			if err != nil {
				utils.Log.Error().Msgf("area.SearchNetAddrOneByOne err:%s", err)
				atomic.AddInt32(&failedCnt, 1)
				return
			}
			if len(nodes) > 0 {
				// _, err := area.SendSearchSuperMsgProxyWaitRequest(msgId, &nodes[ii], nil, senderProxyId, content, 10*time.Second)
				_, _, _, err := area.SendP2pMsgProxyWaitRequest(msgId, &nodes[0], nil, senderProxyId, content, 10*time.Second)
				if err != nil {
					utils.Log.Error().Msgf("SendSearchSuperMsgProxyWaitRequest err:%s", err)
					atomic.AddInt32(&failedCnt, 1)
					return
				}

				atomic.AddInt32(&successCnt, 1)
			}
		}()
	}

	wg.Wait()

	endTime := time.Now().UnixMilli()
	useTime := endTime - startTime

	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
	utils.Log.Warn().Msgf("cost time: %d毫秒", useTime)
	utils.Log.Warn().Msgf("success cnt:%d", successCnt)
	utils.Log.Warn().Msgf("failed cnt:%d", failedCnt)
	utils.Log.Warn().Msgf("UnionApi TPS:%v", (successCnt*1000)/int32(useTime))
	utils.Log.Info().Msgf("--------------------------------------------------")
	utils.Log.Info().Msgf("")
}

/*
 * 随机获取一个域名
 */
func GetRandomDomain() string {
	str := "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
	result := ""
	r := int64(0)
	for i := 0; i < 44; i++ {
		r = utils.GetRandNum(int64(57))
		result = result + str[r:r+1]
	}
	return result
}

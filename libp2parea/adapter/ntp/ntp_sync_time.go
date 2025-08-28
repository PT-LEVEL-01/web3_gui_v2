package ntp

import (
	"context"
	"github.com/beevik/ntp"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

/*
 * 通过ntp方式去同步时间
 *
 * CreateTime: 2023-05-19 17:31
 */

// ntpServers 用于ntp同步时间的服务器域名列表
var ntpServers = []string{
	"time.windows.com",
	"time.tencent.com",
	"ntp.tencent.com",
	"0.centos.pool.ntp.org",
	"0.rhel.pool.ntp.org",
	"0.debian.pool.ntp.org",
	"time.izatcloud.net",
	"time.gpsonextra.net",
	"0.beevik-ntp.pool.ntp.org",
	"0.cn.pool.ntp.org",
	"ntp.aliyun.com",
	"time.apple.com",
	"pool.ntp.org",
	"0.pool.ntp.org",
	"asia.pool.ntp.org",
	"cn.pool.ntp.org",
	"ntp0.ntp-servers.net",
	"time.pool.aliyun.com",
	"time.cloud.tencent.com",
	"time.asia.apple.com",
	"time.euro.apple.com",
	"time.google.com",
	"time.facebook.com",
	"time.cloudflare.com",
	"clock.sjc.he.net",
	"clock.fmt.he.net",
	"clock.nyc.he.net",
	"ntp.ix.ru",
	"ntp.jst.mfeed.ad.jp",
	"ntp.ntsc.ac.cn",
	"ntp1.nim.ac.cn",
	"stdtime.gov.hk",
	"time.hko.hk",
	"time.smg.gov.mo",
	"tick.stdtime.gov.tw",
	"time.kriss.re.kr",
	"ntp.nict.jp",
	"s2csntp.miz.nao.ac.jp",
	"time.nist.gov",
	"time.nplindia.org",
	"ntp.pagasa.dost.gov.ph",
	"ntp1.sirim.my",
	"ntp1.vniiftri.ru",
	"vniiftri.khv.ru",
	"master.ntp.mn",
	"slave.ntp.mn",
	"ntp.mn",
	"ntp.atlas.ripe.net",
	"time.ustc.edu.cn",
	"ntp.sjtu.edu.cn",
	"ntp.neu.edu.cn",
	"time.jmu.edu.cn",
	"ntp.neusoft.edu.cn",
	"time.edu.cn",
	"ntp.nc.u-tokyo.ac.jp",
	"ntp.kuins.kyoto-u.ac.jp",
	"ntp.tohoku.ac.jp",
}

// ntpOffSetTime 经过ntp同步后确认的时间差值
var ntpOffSetTime time.Duration

// ntp同步时间结果
var syncTimeCh = make(chan bool, 1)

const (
	ntp_Sync_Time           = time.Hour * 3    // 每3个小时同步一次时间
	max_Sync_Cnt            = 10               // 最大同步协程数
	check_system_time       = time.Second * 30 // 检查系统时间是否发生变化间隔时间
	system_time_valid_range = 200              // 系统时间误差时间范围
	max_retry_sync_cnt      = 5                // 最多重试同步次数
)

func init() {
	// 开协程，进行首次的同步时间操作
	go retrySyncNtpTime()
	// 开协程，定期做一次同步时间操作
	go loopSyncNtpTime()
	// 开协程，定期检查系统时间是否发生变化
	go loopCheckSystemTime()
}

// 重试同步时间
func retrySyncNtpTime() {
	//fmt.Println("111开始同步ntp时间-------------------")
	var bSyncTime bool
	for i := 0; i < max_retry_sync_cnt; i++ {
		// 同步时间操作, 同步到数据, 则直接退出
		if asyncNtpTime() {
			bSyncTime = true
			break
		}
	}

	// 写入同步结果
	select {
	case syncTimeCh <- bSyncTime:
	default:
	}
	//fmt.Println("结束同步ntp时间-------------------")
}

// 系统时间发生变化管道
var chSystemTimeChange = make(chan interface{}, 1)

// 定期检查系统时间是否发生变化
func loopCheckSystemTime() {
	// 获取当前的毫秒时间
	lastCheckTime := time.Now().UnixMilli()
	// 定时器
	ticker := time.NewTicker(check_system_time)
	// 检查次数
	var nCheckTime int64
	for {
		<-ticker.C
		// 获取当前时间
		nowTime := time.Now().UnixMilli()
		// 自增检查次数
		nCheckTime++
		// 获取期望时间
		nExpectTime := lastCheckTime + nCheckTime*check_system_time.Milliseconds()
		// 检查实际时间和期望时间的差值
		nDiffTime := nExpectTime - nowTime

		// 如果误差时间大于指定的值, 则会触发更新ntp时间操作, 同时会立刻更改ntp时间差值
		if math.Abs(float64(nDiffTime)) >= system_time_valid_range {
			// utils.Log.Error().Msgf("更改时间 diffTime:%d", nDiffTime)
			offsetTime := ntpOffSetTime.Milliseconds() + nDiffTime

			// 更新ntp时间差值
			atomic.StoreInt64((*int64)(&ntpOffSetTime), offsetTime*int64(time.Millisecond))

			// 更新最后检测时间及次数
			lastCheckTime = nowTime
			nCheckTime = 0

			// 触发同步时间操作
			select {
			case chSystemTimeChange <- struct{}{}:
			default:
			}
		}
	}
}

/*
 * loopSyncNtpTime 定时触发同步ntp时间操作
 */
func loopSyncNtpTime() {
	// 定时器
	ticker := time.NewTicker(ntp_Sync_Time)
	defer ticker.Stop()

	// 通过定时器触发同步
	for {
		// 等待定时器触发
		select {
		case <-ticker.C:
		case <-chSystemTimeChange:
			// 因为系统时间触发同步操作
			ticker.Reset(ntp_Sync_Time)
		}

		// 同步ntp服务器时间
		syncNtpTime()
	}
}

/*
 * syncNtpTime ntp同步服务器时间
 *
 * 依次遍历ntp服务器进行同步，只要同步到，就退出
 */
func syncNtpTime() {
	// 依次遍历ntp服务器，只要有一个能获取到，就会退出
	for i := 0; i < len(ntpServers); i++ {
		ntpResp, err := ntp.Query(ntpServers[i])
		if err != nil {
			// utils.Log.Error().Msgf("[%s] 无法访问了, err:%s", ntpServers[i], err)
			continue
		}
		// utils.Log.Info().Msgf("同步到时间[%s] %d", ntpServers[i], ntpResp.ClockOffset)

		// 获取到ntp信息，保存时间差值
		atomic.StoreInt64((*int64)(&ntpOffSetTime), int64(ntpResp.ClockOffset))
		break
	}
}

/*
 * asyncNtpTime 异步进行ntp同步服务器时间
 *
 * 开启10个协程, 同时去同步ntp时间, 只要有一个同步到就可以退出了
 */
func asyncNtpTime() bool {
	getValue := false                                       // 是否获取到值标识
	ctx, cancel := context.WithCancel(context.Background()) // cancel的context
	offSetCh := make(chan time.Duration, 1)                 // 获取时间差值管道

	// 启动协程，获取时间差值内容
	go func() {
		select {
		case offSetTime := <-offSetCh: // 获取到结果，更新ntp服务器时间差值
			// utils.Log.Info().Msgf("Get value!!! %s", offSetTime)
			atomic.StoreInt64((*int64)(&ntpOffSetTime), int64(offSetTime))
			getValue = true // 触发获取数据状态
			cancel()        // 触发context关闭
		case <-ctx.Done(): // 所有的请求都执行完，却没有获取到时间差值
			// utils.Log.Info().Msgf("启动协程，获取时间差值内容 ctx.Done()!!!")
		}
	}()

	taskCh := make(chan string, 1) // 任务管道
	var wg sync.WaitGroup
	// 启动指定的协程，从任务管道中竞争资源，进行获取ntp时间操作
	for i := 0; i < max_Sync_Cnt; i++ {
		wg.Add(1)
		go func(i int, ctx context.Context) {
			defer wg.Done()
			// defer utils.Log.Error().Msgf("11111 [goroutine%d] 退出!!!!\n", i)
			// var nIndex int

			for v := range taskCh {
				// nIndex++
				// utils.Log.Info().Msgf("11111 [goroutine%d] 开始处理第%d个数据", i, nIndex)

				var ntpResp *ntp.Response                                // ntp时间同步返回值
				var err error                                            // 获取错误信息
				ctx2, cancel2 := context.WithTimeout(ctx, time.Second*5) // 超时context
				chNetReq := make(chan interface{}, 1)                    // 获取连接结果管道
				// 启动协程，获取请求，拿到结果后，写入连接结果管道
				go func() {
					ntpResp, err = ntp.Query(v)
					select {
					case chNetReq <- struct{}{}:
					default:
					}
				}()

				// 等待结果、超时、或运行结束标识
				select {
				case <-chNetReq: // 等到请求结果返回
					cancel2()
					if err != nil || ntpResp == nil {
						// utils.Log.Info().Msgf("11111 [goroutine%d][%d] %s 无法访问了 %s", i, nIndex, v, err)
						continue
					}
					// utils.Log.Info().Msgf("11111 [goroutine%d][%d] %s 获取到数据", i, nIndex, v)

					// 结算时间差值
					offSetTime := ntpResp.ClockOffset

					// 尝试将时间差值写入结果中
					select {
					case offSetCh <- offSetTime:
						// utils.Log.Info().Msgf("11111 [goroutine%d][%d] %s 成功写入数据", i, nIndex, v)
					default:
					}
					return
				case <-ctx2.Done(): // 超时，或触发结束标识
					// utils.Log.Info().Msgf("11111 [goroutine%d][%d] %s 超时或触发结束标识", i, nIndex, v)
					cancel2()
				}
			}
		}(i, ctx)
	}

	// 依次将请求写入到任务管道中
	for i := 0; i < len(ntpServers); i++ {
		taskCh <- ntpServers[i]
		// utils.Log.Info().Msgf("22222 写入到channel中 %d", i)
		if getValue {
			cancel()
			// utils.Log.Info().Msgf("22222 get value!!! break for range")
			break
		}
	}
	// 关闭管道，从而使所有的协程处理完任务后，不再继续等待任务，退出执行协程
	close(taskCh)

	// 等待所有任务执行完毕
	wg.Wait()

	cancel() // 运行结束，关闭context

	return getValue
}

/*
 * GetNtpTime 获取ntp同步之后的当前时间
 *
 * @return	ntpNowTime	time.Time	同步ntp后的当前时间
 */
func GetNtpTime() time.Time {
	// 当前时间，加上同步的时间差，就可以得到同步的时间
	offSetTime := atomic.LoadInt64((*int64)(&ntpOffSetTime))
	t := time.Now()
	ntpNowTime := t.Add(time.Duration(offSetTime))

	return ntpNowTime
}

/*
 * GetNtpOffsetTime 获取ntp的时间差值
 *
 * @return	offSetTime	time.Duration	ntp时间差值
 */
func GetNtpOffsetTime() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&ntpOffSetTime)))
}

/*
 * TriggerNtpSync 主动触发ntp进行时间同步, 不需要等待是否执行成功的结果
 */
func TriggerNtpSync() {
	// 触发ntp时间同步操作
	go asyncNtpTime()
}

/*
 * TriggerNtpSyncWaitRes 主动触发ntp进行时间同步, 并返回触发是否成功标识
 *
 * @return res		bool	是否同步成功
 */
func TriggerNtpSyncWaitRes() bool {
	// 触发ntp时间同步操作
	return asyncNtpTime()
}

/*
 * WaitSyncTimeFinish 等待ntp时间同步结果
 *
 * @return	res		bool	是否同步成功
 */
func WaitSyncTimeFinish() bool {
	// 获取同步结果
	res := <-syncTimeCh

	// 写回同步结果
	select {
	case syncTimeCh <- res:
	default:
	}

	return res
}

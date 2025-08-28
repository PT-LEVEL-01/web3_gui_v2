/*
通过RPC基础函数包装成轻节点SDK
*/
package light

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/libp2parea/adapter/engine"
	jsonrpc2 "web3_gui/libp2parea/adapter/sdk/jsonrpc2"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
)

var rpcHandlers = &sync.Map{}

// 高频SDK和耗时统计，便于后期优化
var tmpSDKCallerCount = sync.Map{}

// 设置handlers映射
func SetupHanlders() {
	rpcHandlers = jsonrpc2.GetRPCHandlers()
}

// 组装轻节点SDK响应
func lightResponse(data []byte, err error) []byte {
	buf := bytes.NewBufferString(`{"jsonrpc":"2.0","code":`)
	if err != nil {
		buf.WriteString(fmt.Sprintf(`%s,"message":"%s"}`, string(data), err.Error()))
		return buf.Bytes()
	}
	buf.WriteString(fmt.Sprintf("%d,", model.Success))
	buf.Write(data[1:])
	return buf.Bytes()
}

// 轻节点就绪
func Ready(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			if config.Area != nil {
				// 网络自治完成
				config.Area.WaitAutonomyFinish()
				// 轻节点就绪
				if mining.GetLongChain() != nil {
					engine.Log.Info("============ 轻节点就绪 ============")
					return
				}
			}
		}
	}
}

// 调用轻节点SDK封装
func Handler(in *model.RpcJson) ([]byte, error) {
	startAt := config.TimeNow().UnixMilli()
	defer sdkCallerCount(in.Method, startAt)

	if rh, ok := rpcHandlers.Load(in.Method); ok {
		res, err := rh.(jsonrpc2.ServerHandler)(in, nil, nil)
		return lightResponse(res, err), nil
	}

	return lightResponse(model.Errcode(model.NoMethod, in.Method)), nil
}

type sdkCount struct {
	method      string
	count       int64
	maxusedtime int64
	minusedtime int64
	aveusedtime int64
}

func sdkCallerCount(method string, startAt int64) {
	usedtime := config.TimeNow().UnixMilli() - startAt
	if item, ok := tmpSDKCallerCount.Load(method); ok {
		c := item.(*sdkCount)
		if usedtime > c.maxusedtime {
			c.maxusedtime = usedtime
		}
		if usedtime < c.minusedtime {
			c.minusedtime = usedtime
		}

		c.count++
		c.aveusedtime = c.aveusedtime + (usedtime-c.aveusedtime)/(c.count)
	} else {
		tmpSDKCallerCount.Store(method, &sdkCount{
			method:      method,
			count:       1,
			aveusedtime: usedtime,
			minusedtime: usedtime,
			maxusedtime: usedtime,
		})
	}

	if config.TimeNow().Unix()%100 == 0 {
		res := strings.Builder{}
		tmpSDKCallerCount.Range(func(key, value any) bool {
			v := value.(*sdkCount)
			res.WriteString(fmt.Sprintf("方法: %s, 调用次数: %d, 平均耗时(ms): %d, 单次最大耗时: %d, 单次最小耗时: %d\n", v.method, v.count, v.aveusedtime, v.maxusedtime, v.minusedtime))
			return true
		})
		engine.Log.Error("SDK高频活动统计:\n%s", res.String())
	}
}

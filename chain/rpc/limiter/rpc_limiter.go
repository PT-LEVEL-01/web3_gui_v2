package limiter

import (
	"sync"
	"web3_gui/chain/rpc/limiter/token2"

	"web3_gui/chain/config"
	"web3_gui/chain/rpc/limiter/limiter"
	"web3_gui/chain/rpc/limiter/token"
	"web3_gui/libp2parea/adapter/engine"
)

type ReqLimiter struct {
	mm sync.Map
	mu sync.Mutex
}

var RpcReqLimiter *ReqLimiter

// 合约额外消耗Token量
var ContractExtraUsed = uint64(20000)

type RpcMethod struct {
	Method string
	Limit  uint64
}

type RpcLimiter struct {
	*RpcMethod
	Limiter limiter.Limiter
}

func InitRpcLimiter() {
	engine.Log.Info("limiter token total：%d", config.HandleTxTokens)

	RpcReqLimiter = &ReqLimiter{mm: sync.Map{}, mu: sync.Mutex{}}

	//生成限流器
	for k, v := range config.ReqLimiterMap {
		RpcReqLimiter.mm.Store(k, createRpcLimiter(k, &limiter.LimitConf{v.Limit, v.Burst}))

	}
}

func createRpcLimiter(method string, conf *limiter.LimitConf) *RpcLimiter {
	var limit limiter.Limiter

	switch config.DriverDefault {
	case config.DriverToken:
		limit = token.NewLimiter(conf.Burst)
	case config.DriverToken2:
		limit = token2.NewLimiter(token2.Limit(conf.Limit), int(conf.Burst))

	}

	return &RpcLimiter{&RpcMethod{method, conf.Limit}, limit}

}

func (l *ReqLimiter) Allow(method string) bool {
	return l.AllowN(method, 1)
}

// 添加令牌
func (l *ReqLimiter) AllowN(method string, n uint64) bool {
	if n <= 0 {
		return false
	}

	rl := l.getLimiter(method)
	if rl == nil {
		return false
	}

	return rl.Limiter.AllowN(n)
}

func (l *ReqLimiter) Recover(method string) bool {
	return l.RecoverN(method, 1)
}

// 回收令牌
func (l *ReqLimiter) RecoverN(method string, n uint64) bool {
	if n <= 0 {
		return false
	}

	rl := l.getLimiter(method)
	if rl == nil {
		return false
	}

	return rl.Limiter.RecoverN(n)
}

func (l *ReqLimiter) IsExists(method string) bool {
	_, ok := config.ReqLimiterMap[method]
	return ok
}

// 使用数量
func (l *ReqLimiter) Len(method string) uint64 {
	rl := l.getLimiter(method)
	if rl == nil {
		return 0
	}

	return rl.Limiter.Len()
}

// 剩余数量
func (l *ReqLimiter) Surplus(method string) uint64 {
	rl := l.getLimiter(method)
	if rl == nil {
		return 0
	}

	return rl.Limiter.Surplus()
}

// 获取对应的限流器
func (l *ReqLimiter) getLimiter(method string) *RpcLimiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.IsExists(method) {
		method = config.MethodDefault
	}

	v, ok := l.mm.Load(method)
	if !ok {
		return nil
	}

	return v.(*RpcLimiter)
}

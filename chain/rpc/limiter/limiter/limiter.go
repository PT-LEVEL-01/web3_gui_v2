package limiter

type Limiter interface {
	Allow() bool
	AllowN(uint64) bool
	Recover() bool
	RecoverN(uint64) bool
	Len() uint64
	Surplus() uint64
}

type LimitConf struct {
	Limit uint64
	Burst uint64
}

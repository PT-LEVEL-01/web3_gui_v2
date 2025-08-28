package config

const (
	Url = "http://127.0.0.1:2080/rpc"
	UserName = "test"
	Pwd = "testp"
	SupperNodeAddress = "MMS3vpuD3LmHSwhMCqimeT3nQQqwZFskoirP4"
	SupperNodePwd	= "xhy19liu21@"
	TransactionLimitNum = 60 // 一个账号只能有64笔交易滞留，这里设置60以防万一
	MinValue = 90000000000000 // 提取出账号大于当前的账号地址 就不用创建了
)

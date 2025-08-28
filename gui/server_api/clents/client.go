package clents

type PeerClient interface {
	GetNewAddress(pwd string) error
	SendToAddress(srcAddr string, toAddr string, pwd string, amount float64, gas float64) (*RespSendToAddress, error)
	Depositin(amount float64, gas float64, rate float64, pwd string, payload string) error
	Depositout(witnessAddr string, amount float64, gas float64, pwd string) error
	VoteIn(votetype uint16, address string, witness string, amoount float64, gas float64, rate float64, pwd string, payload string) error
	VoteOut(votetype uint16, address string, amount float64, gas float64, rate float64, pwd string, payload string) error
	CommunityDistribute(srcaddress string, gas float64, pwd string) error
	NameIn(address string, amount float64, gas float64, pwd string, name string, netids string, addrcoins string, comment string) error
	NameOut(address string, amount float64, gas float64, pwd string, name string, comment string) error
	GetNames() ([]*NameinfoVO, error)
	PeerPolling
}

// 轮询的接口
type PeerPolling interface {
	ListAccounts() ([]*RespListAccounts, error)
	GetInfo() (*RespGetInfo, error)
	GetAccount(addr string) (*RespGetAccount, error)
}

type PostResult struct {
	Jsonrpc string      `json:"jsonrpc"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

type RespListAccounts struct {
	Index       int    //排序
	AddrCoin    string //收款地址
	Value       uint64 //可用余额
	ValueFrozen uint64 //冻结余额
	ValueLockip uint64 //
	Type        int    // 1 见证人 2 社区节点 3 轻节点 4 什么都不是
}

type RespGetInfo struct {
	Balance        uint64 `json:"balance"`
	BalanceFrozen  uint64 `json:"BalanceFrozen"`  //冻结
	BalanceLockup  uint64 `json:"BalanceLockup"`  //锁仓
	Group          uint64 `json:"group"`          //区块组高度
	StartingBlock  uint64 `json:"StartingBlock"`  //区块开始高度
	StartBlockTime uint64 `json:"StartBlockTime"` //创建区块出块时间
	HighestBlock   uint64 `json:"HighestBlock"`   //所链接的节点的最高高度
	CurrentBlock   uint64 `json:"CurrentBlock"`   //已同步到的区块高度
	PulledStates   uint64 `json:"PulledStates"`   //正在同步的区块高度
	SnapshotHeight uint64 `json:"SnapshotHeight"` //快照高度
	WitnessPeer    uint64 `json:"WitnessPeer"`    //见证人押金数量
}

type RespSendToAddress struct {
	GasUsed uint64  `json:"GasUsed"`
	Vot     []*Vout `json:"vot"`
}

type Vout struct {
	Value uint64 `json:"value"` //输出金额 = 实际金额* 100000000
}

type RespGetAccount struct {
	Balance       uint64 `json:"Balance"`
	BalanceFrozen uint64 `json:"BalanceFrozen"` //冻结
}

type Nameinfo struct {
	Name           string   `json:"name"`           //域名
	Owner          string   `json:"owner"`          //拥有者
	NetIds         []string `json:"netIds"`         //节点地址
	AddrCoins      []string `json:"addrCoins"`      //钱包收款地址
	Height         uint64   `json:"height"`         //注册区块高度，通过现有高度计算出有效时间
	NameOfValidity uint64   `json:"nameOfValidity"` //有效块数量
	Deposit        uint64   `json:"deposit"`        //冻结金额
}

type NameinfoVO struct {
	Name           string   `json:"name"`           //域名
	Owner          string   `json:"owner"`          //拥有者
	NetIds         []string `json:"netIds"`         //节点地址
	AddrCoins      []string `json:"addrCoins"`      //钱包收款地址
	Height         uint64   `json:"height"`         //注册区块高度，通过现有高度计算出有效时间
	NameOfValidity uint64   `json:"nameOfValidity"` //有效块数量
	Deposit        float64  `json:"deposit"`        //冻结金额
}

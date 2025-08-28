## 快速开始

### 源码编译

#### 1 windows平台

进入项目目录 icom_chain/example/peer_root,运行如下命令

构建程序名: peer_root.exe
```shell
go build -x -v -ldflags "-s -w" -o peer_root.exe firstPeer.go
```

#### 2 linux平台

进入项目目录 icom_chain/example/peer_root,运行如下命令

构建程序名: peer_root
```shell
GOOS=linux GOARCH=amd64 go build -x -v -ldflags "-s -w" -o peer_root firstPeer.go
```

### 节点运行

#### 1 配置文件
配置有3个文件
- config.json
- nodeEntry.json
- config_extra.json(可选)

##### 1.1 文件 config.json 参数
```text
  //节点ip地址
  "ip": "127.0.0.1",
  //节点端口
  "port": 19981,
  //节点名称
  "AreaName": "testAreaName",
  //节点RPC接口地址
  "WebAddr": "0.0.0.0",
  //节点RPC端口地址
  "WebPort": 2080,
  //网页静态文件路径
  "WebStatic": "./static",
  //网页模板文件路径
  "WebViews": "./views",
  //rpc请求账号密码
  "RpcUser": "test",
  "RpcPassword": "testp",
  //本节点是否是矿工
  "miner": true,
  //网络类型:正式网络release/测试网络not release
  "NetType": "not release",
  //收款地址前缀
  "AddrPre": "IM",
  //用于验证的创始区块hash
  "CheckStartBlockHash": "",
  //是否禁用快照
  "DisableSnapshot": false,
  //是否开启虚拟机模式
  "EvmRewardEnable": false
  //是否开启自重启功能
  "EnableRestart": false
```

##### 1.2 文件 config_extra.json 参数
这个文件可以没有,没有则按以下默认配置运行

```go
Witness_backup_max        = 21                   //备用见证人排名靠前的最多数量，之后的人依然是选举中的候选见证人。
Witness_backup_reward_max = 40                   //有奖励的最大见证人数量
Witness_Ave_Ratio         = 30                   //出块见证人+候选见证人平均分奖励
Mining_Reward_Interval    = uint64(600)          //主链发奖励间隔,默认600块高度(等于1及时到账),真实交易中是含有每个见证人的奖励的,但奖励会首先放入各自见证人的奖励池中.待时间间隔到了累计的奖励全部到账
Mining_deposit            = uint64(100000 * 1e8) //见证人押金最少金额
Mining_vote               = uint64(1000 * 1e8)   //社区节点投票押金最少金额
Mining_light_min          = uint64(10 * 1e8)     //轻节点押金最少金额
CancelVote_Interval       = uint64(10)           //社区取消质押/轻节点取消投票间隔,默认10块高度
Max_Community_Count       = 1000                 //全网最大社区数量,默认1000个
Max_Light_Count           = 1000                 //每个社区下的最大轻节点数量,默认1000个
```

##### 1.3 文件 nodeEntry.json 参数
配置服务发现节点
```json
{"172.28.0.115:19981":"","172.28.0.116:19981":"",,"172.28.0.117:19981":""}
```


##### 1.4 程序可调参数释义
```go
MinCPUCores               = uint32(3)            //最低cpu核数,仅在启动时检查,0=不检查,大于0就按给定值进行检查,不符合就终止程序
MinFreeMemory             = uint32(8 * 1024)     //最低可用内存(MB),仅在启动时检查,0=不检查,大于0就按给定值进行检查,不符合就终止程序
MinFreeDisk               = uint32(6 * 1024)     //最低可用磁盘空间(MB),仅在启动时检查,0=不检查,大于0就按给定值进行检查,不符合就终止程序
MinNetworkBandwidth       = uint32(10)           //最低带宽(MB/s),仅在启动时检查,0=不检查,大于0就按给定值进行检查,不符合发出警告
```

#### 2 节点运行

节点目录结构:
```shell
.
├── conf
│   ├── config_extra.json(可选)
│   ├── config.json
│   └── nodeEntry.json
└── peer_root.exe
```

##### 2.1 单节点运行
```shell
#创世节点首次运行
peer_root init
```

```shell
#节点拉起
peer_root load
```


##### 2.2 多节点运行

- 整个网络节点首次运行
```shell
#创世节点运行
peer_root init
#其余节点不加任何命令启动
peer_root
```

- 整个网络节点部分停止的情况拉起节点
```shell
#不加任何命令启动
peer_root
```

- 整个网络节点全部停止的情况拉起节点
```shell
#首个节点拉起
peer_root load
#其余节点不加任何命令启动
peer_root
```
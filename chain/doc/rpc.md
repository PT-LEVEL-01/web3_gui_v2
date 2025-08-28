# RPC接口

## 一. 接口说明

### 1. 接口用例

```curl
curl --request POST --url http://192.168.0.2:2081/rpc --header 'password: testp'  --header 'user: test' --data '{"method":"getinfo"}'
```

#### url

`http://192.168.0.2:2080/rpc`

#### method

`POST`

#### Content-Type

`application/json`

#### header参数

| 参数  | 值   |
|------|------|
| user|test|
|password|testp|

### 2.返回结构

```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
  }
}
```

### 3.错误码

| 参数  | 值   | 描述   |
|------|------|------|
|Success |2000|成功|
|NoMethod|4001|没有这个方法|
|TypeWrong|5001|参数类型错误|
|NoField|5002|缺少参数|
|Nomarl|5003|一般错误，请看错误提示信息|
|Timeout|5004|超时|
|Exist|5005|已经存在|
|FailPwd|5006|密码错误|
|NotExist|5007|不存在|
|NotEnough|5008|余额不足|
|ContentIncorrectFormat|5009|参数格式不正确|
|AmountIsZero|5010|转账不能为0|
|RuleField|5011|地址角色不正确|
|BalanceNotEnough|5012|余额不足|
|VoteExist|5013|投票已经存在|
|VoteNotOpen|5014|投票功能还未开放|
|RewardCountSync|5015|轻节点奖励异步执行中|
|CommentOverLengthMax|5016|备注信息字符串超过最大长度|
|GasTooLittle|5017|交易手续费太少|

## 二. 接口明细

#### 1. 获取区块高度、余额等相关信息

###### 请求

```json
{
  "method": "getinfo"
}
```

###### 返回说明
```json
{
    "netid": "WkhD",
    "TotalAmount": 300000000000000000, //发行总量
    "balance": 999999900000000,      //可用余额
    "BalanceFrozen": 0,              //冻结余额
    "BalanceLockup": 0,              //锁定余额
    "testnet": true,                 //是否是测试网络
    "blocks": 370758,
    "group": 0,
    "StartingBlock": 370703,         //起始高度
    "HighestBlock": 370759,          //最新高度
    "CurrentBlock": 370758,
    "PulledStates": 370759,          //同步高度
    "BlockTime": 10,
    "LightNode": 1000000000,         //轻节点最少押金
    "CommunityNode": 100000000000,   //社区节点最少押金
    "WitnessNode": 1000000000000,    //见证人节点最少押金
    "NameDepositMin": 100000000,     //域名注册最少押金
    "AddrPre": "ZHC",                //地址前缀，单位
    "TokenBalance": [
        {
            "TokenId": "CgAAAAAAAAAmxc6MbyDrbgEEydoIIGjAMhlUOanpwo9woQxTJ97vLw==", //Token合约地址
            "Name": "通贝宝",                                                       //Token名称全称
            "Symbol": "TBB",                                                     //Token单位
            "Supply": 1500000000000000,                                         //Token发行总量
            "Balance": 1500000000000000,                                        //可用余额  
            "BalanceFrozen": 0,                                                 //冻结余额
            "BalanceLockup": 0                                                  //锁定余额
        },
        {
            "TokenId": "CgAAAAAAAAAjO0dZlSuAZ55ogOwYqhFCmNUtXcMgFLvEGDo9ugAwxQ==",
            "Name": "bitebi",
            "Symbol": "BTC",
            "Supply": 2100000000000000,
            "Balance": 2100000000000000,
            "BalanceFrozen": 0,
            "BalanceLockup": 0
        }
    ]
}
```
#### 2. 创建新地址

###### 请求

```json
{
  "method": "getnewaddress",
  "params": {
    "password": "123456"    //支付密码
  }
}
```
#### 3. 帐号列表

###### 请求

```json
{
  "method": "listaccounts"
}
```

###### 返回说明
```json
"result": [
    {
        "Index": 0,                                            //索引
        "AddrCoin": "ZHCDq7xc8QiZ7VAuGy2XibbZxagKqQvPf3iX4",   //地址
        "Value": 0,                                            //可用余额
        "ValueFrozen": 0,                                      //冻结余额
        "ValueLockup": 0,                                      //锁仓余额
        "Type": 4                                              //地址类型 1=见证人;2=社区节点;3=轻节点;4=什么也不是;
    },
    {
        "Index": 1,
        "AddrCoin": "ZHCHrH6NEDe57WzWVWZbd2XNJdbAb5XdAoAj4",
        "Value": 999999400000000,
        "ValueFrozen": 0,
        "ValueLockup": 0,
        "Type": 4
    }
]
```

#### 4. 获取某一帐号余额

###### 请求

```json
{
  "method": "getaccount",
  "params": {
    "address": "1AX9mfCRZkdEg5Ci3G5SLcyGgecj6GTzLo"       //查询目标地址
  }
}
```

###### 返回说明
```json
"result": {
    "Balance": 0,           //账号余额
    "BalanceFrozen": 0      //锁定余额
}
```

#### 5. 验证地址合法性

###### 请求

```json
{
  "method": "validateaddress",
  "params": {
    "address": "12EUY1EVnLJe4Ejb1VaL9NbuDQbBEV"     //查询目标地址
  }
}
```

###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": false
}
```

#### 6. 导入钱包私钥

###### 请求

```json
{
  "method": "import",
  "params": {
    "password": "123456789",      //支付密码
    "seed": "xxxxxxxxxxxxxx"      //随机种子
  }
}
```
#### 7. 导出钱包

###### 请求

```json
{
  "method": "export",
  "params": {
    "password": "123456789"       //支付密码
  }
}
```

#### 8. 转账

###### 请求

```json
{
  "method": "sendtoaddress",
  "params": {
    "srcaddress": "",   //转出地址
    "address": "ZHCMqZE3UFSgf9t1r69QT6hYnG2dApnpmv1q4",//转入地址
    "changeaddress": "ZHCMqZE3UFSgf9t1r69QT6hYnG2dApnpmv1q4",//指定找零地址
    "amount": 1000000000000000,//转账金额
    "gas": 1,//手续费
    "frozen_height": 7,//锁定高度
    "pwd": "123456789",//支付密码
    "comment": "test"//备注
  }
}
```
#### 9. 给多个地址转账

###### 请求

```json
{
  "method": "sendtoaddressmore",
  "params": {
    "addresses": [
      {
        "address": "ZHC6iA3DTyy4iz4nkZrbdUKbgJtyvfNDebW84",//目标地址
        "amount": 10000000000000 //转账金额
      },
      {
        "address": "ZHCL8b8YUCGYVCKeBTipT9iZ9n5NhSrU5DiW4",
        "amount": 10000000000000
      }
    ],
    "gas": 10000000,
    "pwd": "123456789",
    "comment": "test"
  }
}
```
#### 10.修改支付密码

###### 请求

```json
{
  "method": "updatepwd",
  "params": {
    "oldpwd": "123456",
    "newpwd": "222222"
  }
}
```

#### 11.域名注册，续费，修改

###### 请求

```json
{
  "method": "namesin",
  "params": {
    "address": "",//拥有者，空位本节点地址
    "amount": 500000000,//押金金额
    "gas": 0,//手续费
    "pwd": "123456789",//支付密码
    "name": "testweb",//域名名称
    "netids": [//网络地址
      "3Dsjd9qaor3bS8NTwsDGbTdGPA3tXiXGKdGJHWDhe6M8"
    ],
    "addrcoins": [//收款地址
      "16BXbn97j97jUJrCoicwYuvm1v8gGoG467"
    ]
  }
}
```
#### 12.域名注销，退还押金

###### 请求

```json
{
  "method": "namesout",
  "params": {
    "address": "押金地址",
    "amount": 1000000,
    "gas": 1000,
    "pwd": "123456",
    "name": "域名：abc"
  }
}
```
#### 13.获取自己注册的域名列表

###### 请求

```json
{
  "method": "getnames"
}
```
###### 返回说明
```json
"result": [
    {
        "Name": "testweb",//域名
        "NetIds": [//网络地址
          "3Dsjd9qaor3bS8NTwsDGbTdGPA3tXiXGKdGJHWDhe6M8"
        ],
        "AddrCoins": [//收款地址
            "16BXbn97j97jUJrCoicwYuvm1v8gGoG467"
        ],
        "Height": 378725,                     //注册时间
        "NameOfValidity": 3153600,            //有效期
        "Deposit": 500000000                  //押金
    }
]
```

#### 14.查询域名

###### 请求

```json
{
  "method": "findname",
  "params": {
    "name": "域名"
  }
}
```
###### 返回说明
```json
"result": {
    "AddrCoins": [//收款地址
      "TU1TIjUJRPZcZzRcwxFCGs4oatG5WMc1aW0QAw=="
    ],
    "Deposit": 1000000000,//押金
    "Height": 840,//注册时间
    "Name": "aaa",//域名
    "NameOfValidity": 0,//有效期
    "NetIds": [//网络地址
      "JJkZH0i5WlP2rz2Ftf484bxShWNByBffpL9nja2cNm8="
    ],
    "Txid": "BQAAAAAAAAA2I8f8AugpiKZrrAMm51pOwI9TvlCqCKyyuCqfbitzHg=="//交易id
}
```

#### 15.获得转账交易历史记录

###### 请求

```json
{
  "method": "gettransactionhistory",
  "params": {
    "id": "0",//查询记录id
    "total": 10 //查询数量
  }
}
```
###### 返回说明
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": [
        {
            "GenerateId": "0",//自增长最高id，保存最新生成的id
            "IsIn": true,//资金转入转出方向，true=转入;false=转出;
            "Type": 1,//交易类型
            "InAddr": [],//输入地址
            "OutAddr": [//输出地址
                "MMS47sWYUKuVqytYPzwjrJG3y25bEsFMthVZ4"
            ],
            "Value": 242164181942014, //交易金额
            "Txid": "0100000000000000da01f0db576a5fac04420e363ac5cb6f9407a83c5d7055714dbc22a0aa188dad",//交易id
            "Height": 1,//区块高度
            "Payload": ""//data域
        }
    ]
}
```
#### 16.查询见证人状态

###### 请求

```json
{
  "method": "getwitnessinfo"
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "IsCandidate": true,//是否是候选见证人
    "IsBackup": true,//是否是备用见证人
    "IsKickOut": false,//是否是没有按时出块，已经被踢出局，只有退还押金，重新缴纳押金成为候选见证人
    "Addr": "MMS47sWYUKuVqytYPzwjrJG3y25bEsFMthVZ4",//见证人地址
    "Payload": "first_witness",//data域
    "Value": 10000000000000 //押金
  }
}
```

#### 17.获得候选见证人列表

###### 请求

```json
{
  "method": "getcandidatelist"
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": [
    {
      "Addr": "MMS47sWYUKuVqytYPzwjrJG3y25bEsFMthVZ4",//见证人地址
      "Payload": "first_witness",//data域
      "Score": 10000000000000,//押金
      "Vote": 0,//投票值
      "CreateBlockTime": 0 //预计出块时间
    }
  ]
}
```

#### 18.获取社区节点列表

###### 请求

```json
 {
  "method": "getcommunitylist"
}
```
#### 19.获得自己投过票的列表

###### 请求

```json
{
  "method": "getvotelist"
}
```

#### 20.通过区块高度查询一个区块信息

###### 请求

```json
{
  "method": "findblock",
  "params": {
    "height": 2 //查询高度
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "Hash": "756166cb238351d7db71d402fdba3fda32f61238b58f9b93c46f6f76dd81057d",//区块hash
    "Height": 59,//区块高度
    "GroupHeight": 59,//见证人组高度
    "GroupHeightGrowth": 0,//组高度增长量。默认0为自动计算增长量（兼容之前的区块）
    "Previousblockhash": "b7ac7bd7de0a1dd888682e11f69bc2a64490055a86ea9e56c7d3aa4267885179",//上一个区块头hash
    "Nextblockhash": "6a77721bf7558321bd5635466562b1cdbf1b8023eee325f63ce1eb3e59ec267c",//下一个区块头hash
    "NTx": 1,//交易数量
    "MerkleRoot": "9e7bcbe8cafec3eae8fd1e18be7cf25fa39966d52f25dc1aba43f0892efff878",//交易默克尔树根hash
    "Tx": [//本区块包含的交易id
      "0100000000000000db7f6636599d06c6b88e57b33ad6de8b0bfac89a50e19e89ab64e401f89340ae"
    ],
    "Time": 1667190491266337700,//出块时间，unixNano时间戳
    "Witness": "MMS47sWYUKuVqytYPzwjrJG3y25bEsFMthVZ4",//此块见证人地址
    "Sign": "015b97dc01adb9c57939d68efcd21fea7f954c705728f4915cfd2f41d381bc5a294c147727824daf42489ed110f48aa945851a2bf2bd2c185ffa00cb951ac70e" //见证人对块签名
  }
}
```

#### 21.发布一个token

###### 请求

```json
{
  "method": "tokenpublish",
  "params": {
    "gas": 0,//手续费
    "pwd": "123456789",//支付密码
    "name": "test",//Token名称全称
    "symbol": "TEST",//Token单位，符号
    "supply": 19000000000,//发行总量
    "owner": "",//所有者
    "comment": "123" //备注
  }
}
```
#### 22.使用token支付

###### 请求

```json
{
  "method": "tokenpay",
  "params": {
    "address": "SELF9kEwJFPX8WjCDgMgiXKiXBddkChtN89Md5",//押金冻结的地址，空为本节点地址
    "amount": 2,//转账金额
    "gas": 0,//手续费
    "pwd": "123456789",//支付密码
    "txid": "0800000000000000045a02d7110508ab9295b5b865ceabeae16bff9ae3cc15bf6abb8816741400c1",//发布token的交易id
    "comment": "123" //备注
  }
}
```
#### 23.使用token支付给多个地址

###### 请求

```json
{
  "method": "tokenpaymore",
  "params": {
    "addresses": [
      {
        "srcaddress": "",//扣款源地址,空位节点本地址
        "address": "ZHCMqZE3UFSgf9t1r69QT6hYnG2dApnpmv1q4",//收款地址
        "amount": 11 //转账金额
      },
      {
        "address": "ZHC6iA3DTyy4iz4nkZrbdUKbgJtyvfNDebW84",
        "amount": 11
      }
    ],
    "gas": 0,//手续费
    "pwd": "123456789",//支付密码
    "txid": "0a00000000000000819785a0dbbbed8bcfd8ecb0a301abb3531ce32a2e47443bcae448ecca271bbe",//交易id
    "comment": "123" //备注
  }
}
```

#### 24.见证人押金

###### 请求

```json
{
  "method": "depositin",
  "params": {
    "amount": 1000000000000,//金额
    "gas": 0,//手续费
    "pwd": "123456789",//支付密码
    "payload": "" //data域
  }
}
```

#### 25.见证人取消押金

###### 请求

```json
{
  "method": "depositout",
  "params": {
    "witness": "ZHCDq7xc8QiZ7VAuGy2XibbZxagKqQvPf3iX4",//见证者地址
    "amount": 1000000000000,//金额
    "gas": 0,//手续费
    "pwd": "123456789" //支付密码
  }
}
```
#### 26.获取自己投票的列表

###### 请求

```json
{
  "method": "getvotelist",
  "params": {
    "votetype": 1
  }
}
```
#### 27.获取社区节点列表

###### 请求

```json
{
  "method": "getcommunitylist"
}
```
#### 28.获取社区节点奖励信息

###### 请求

```json
{
  "method": "getcommunityreward",
  "params": {
    "address": "1AmmKe1Jizjj2r8f1fafRSEk4nWJJekPfU"
  }
}
```
#### 29.投票

###### 请求

```json
{
  "method": "votein",
  "params": {
    "votetype": 1,//投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
    "address": "",//目标账户地址
    "witness": "1H3dYBJxsqdzb3HgpSfHbNuDhtY9TtXckR",//见证人地址
    "amount": 100000000000,//金额
    "gas": 1,
    "pwd": "123456789",
    "payload": ""
  }
}
```

#### 30.取消投票

###### 请求

```json
{
  "method": "voteout",
  "params": {
    "votetype": 1,//投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
    "address": "ZHCHrH6NEDe57WzWVWZbd2XNJdbAb5XdAoAj4",//目标账户地址
    "witness": "1H3dYBJxsqdzb3HgpSfHbNuDhtY9TtXckR",//见证人地址
    "amount": 100000000000,
    "gas": 0,
    "pwd": "123456789"
  }
}
```
#### 31.查询一个交易是否上链，以及交易详细信息

###### 请求

```json
{
  "method": "findtx",
  "params": {
    "txid": "0b0000000000000073b00655d03a6ef44be65eba68158a8194d795a86ef30a2410a7934d7266fe05"
  }
}
```

###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "txinfo": {
      "hash": "0400000000000000863f6886110b60de0d53434677bb4a1292df13a3274f11097b53b428cd743414",//交易hash
      "type": 4,//交易类型
      "vin_total": 1,//输入数
      "vin": [
        {
          "puk": "0f4d5ef7f8b48c62a2f93447605898fd04dc334e4a73f649a99a7361042c352f",//公钥
          "n": "1",//转出账户nonce
          "sign": "a2e701f5489050fcdd9e62b7c543477806101c607f97cf1eff64e6bc65fbec8119d2b2e2d261709ffd203ef535b09d5041198baa6c4abc7e608a0d31c19f2b0a" //输入签名
        }
      ],
      "vout_total": 1,//输出数
      "vout": [
        {
          "value": 20000000000000,//金额
          "address": "MMSNVdcn7LYbHdCiNxpHScsCYonvEH1n9tSK4",//收款地址
          "frozen_height": 2 //冻结高度
        }
      ],
      "gas": 1000000,//手续费
      "lock_height": 36,//锁定高度
      "payload": "test",//data域
      "blockhash": "" //区块hash
    },
    "upchaincode": 2 //upchaincode:1=未确认；2=成功；3=失败；
  }
}
#### 31.查询一个交易是否上链，以及交易详细信息

###### 请求

```json
{
  "method": "findtx",
  "params": {
    "txid": "0b0000000000000073b00655d03a6ef44be65eba68158a8194d795a86ef30a2410a7934d7266fe05"
  }
}
```

###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "txinfo": {
      "hash": "0400000000000000863f6886110b60de0d53434677bb4a1292df13a3274f11097b53b428cd743414",//交易hash
      "type": 4,//交易类型
      "vin_total": 1,//输入数
      "vin": [
        {
          "puk": "0f4d5ef7f8b48c62a2f93447605898fd04dc334e4a73f649a99a7361042c352f",//公钥
          "n": "1",//转出账户nonce
          "sign": "a2e701f5489050fcdd9e62b7c543477806101c607f97cf1eff64e6bc65fbec8119d2b2e2d261709ffd203ef535b09d5041198baa6c4abc7e608a0d31c19f2b0a" //输入签名
        }
      ],
      "vout_total": 1,//输出数
      "vout": [
        {
          "value": 20000000000000,//金额
          "address": "MMSNVdcn7LYbHdCiNxpHScsCYonvEH1n9tSK4",//收款地址
          "frozen_height": 2 //冻结高度
        }
      ],
      "gas": 1000000,//手续费
      "lock_height": 36,//锁定高度
      "payload": "test",//data域
      "blockhash": "" //区块hash
    },
    "upchaincode": 2 //upchaincode:1=未确认；2=成功；3=失败；
  }
}

```
###### 交易类型
```
	Wallet_tx_type_start          = 0 //
	Wallet_tx_type_mining         = 1 //挖矿所得
	Wallet_tx_type_deposit_in     = 2 //备用见证人押金输入，余额锁定
	Wallet_tx_type_deposit_out    = 3 //备用见证人押金输出，余额解锁
	Wallet_tx_type_pay            = 4 //普通支付
	Wallet_tx_type_account        = 5 //申请名称
	Wallet_tx_type_account_cancel = 6 //注销名称
	Wallet_tx_type_vote_in        = 7 //参与见证人投票输入，余额锁定
	Wallet_tx_type_vote_out       = 8 //参与见证人投票输出，余额解锁
	// Wallet_tx_type_deposit_out_force = 9 //见证人3次未出块，强制退还押金

	// Wallet_tx_type_register_store   = 20 //注册成为存储服务提供方
	// Wallet_tx_type_unregister_store = 21 //注册成为存储服务提供方
	// Wallet_tx_type_resources        = 20 //购买存储资源下载权限
	// Wallet_tx_type_resources_upload = 21 //上传资源付费

	Wallet_tx_type_token_publish = 10 //token发布
	Wallet_tx_type_token_payment = 11 //token支付

	Wallet_tx_type_spaces_mining_in  = 12 //存储挖矿押金输入，余额锁定
	Wallet_tx_type_spaces_mining_out = 13 //存储挖矿押金输出，余额解锁
	Wallet_tx_type_spaces_use_in     = 14 //用户存储空间押金输入，余额锁定
	Wallet_tx_type_spaces_use_out    = 15 //用户存储空间押金输出，余额解锁
	Wallet_tx_type_voting_reward     = 16 //社区节点给轻节点分发奖励
	Wallet_tx_type_nft               = 17 //nft交易

	Wallet_tx_type_end = 100 //
```

#### 32.通过区块Hash查询一个区块信息

###### 请求

```json
{
  "method": "findblockbyhash",
  "params": {
    "hash": "756166cb238351d7db71d402fdba3fda32f61238b58f9b93c46f6f76dd81057d" //区块hash
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "Hash": "756166cb238351d7db71d402fdba3fda32f61238b58f9b93c46f6f76dd81057d",//区块hash
    "Height": 59,//区块高度
    "GroupHeight": 59,//见证人组高度
    "GroupHeightGrowth": 0,//组高度增长量。默认0为自动计算增长量（兼容之前的区块）
    "Previousblockhash": "b7ac7bd7de0a1dd888682e11f69bc2a64490055a86ea9e56c7d3aa4267885179",//上一个区块头hash
    "Nextblockhash": "6a77721bf7558321bd5635466562b1cdbf1b8023eee325f63ce1eb3e59ec267c",//下一个区块头hash
    "NTx": 1,//交易数量
    "MerkleRoot": "9e7bcbe8cafec3eae8fd1e18be7cf25fa39966d52f25dc1aba43f0892efff878",//交易默克尔树根hash
    "Tx": [//本区块包含的交易id
      "0100000000000000db7f6636599d06c6b88e57b33ad6de8b0bfac89a50e19e89ab64e401f89340ae"
    ],
    "Time": 1667190491266337700,//出块时间，unixNano时间戳
    "Witness": "MMS47sWYUKuVqytYPzwjrJG3y25bEsFMthVZ4",//此块见证人地址
    "Sign": "015b97dc01adb9c57939d68efcd21fea7f954c705728f4915cfd2f41d381bc5a294c147727824daf42489ed110f48aa945851a2bf2bd2c185ffa00cb951ac70e" //见证人对块签名
  }
}
```
#### 33.分配奖励

###### 请求

```json
{
  "method": "sendcommunityreward",
  "params": {
    "address": "756166cb238351d7db71d402fdba3fda32f61238b58f9b93c46f6f76dd81057d", //轻节点地址
    "gas": 1,//手续费
    "pwd": "",//钱包密码
    "startheight": 1,//开始高度
    "endheight": 10 //结束高度
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "Hash": "756166cb238351d7db71d402fdba3fda32f61238b58f9b93c46f6f76dd81057d",//区块hash
    "Height": 59,//区块高度
    "GroupHeight": 59,//见证人组高度
    "GroupHeightGrowth": 0,//组高度增长量。默认0为自动计算增长量（兼容之前的区块）
    "Previousblockhash": "b7ac7bd7de0a1dd888682e11f69bc2a64490055a86ea9e56c7d3aa4267885179",//上一个区块头hash
    "Nextblockhash": "6a77721bf7558321bd5635466562b1cdbf1b8023eee325f63ce1eb3e59ec267c",//下一个区块头hash
    "NTx": 1,//交易数量
    "MerkleRoot": "9e7bcbe8cafec3eae8fd1e18be7cf25fa39966d52f25dc1aba43f0892efff878",//交易默克尔树根hash
    "Tx": [//本区块包含的交易id
      "0100000000000000db7f6636599d06c6b88e57b33ad6de8b0bfac89a50e19e89ab64e401f89340ae"
    ],
    "Time": 1667190491266337700,//出块时间，unixNano时间戳
    "Witness": "MMS47sWYUKuVqytYPzwjrJG3y25bEsFMthVZ4",//此块见证人地址
    "Sign": "015b97dc01adb9c57939d68efcd21fea7f954c705728f4915cfd2f41d381bc5a294c147727824daf42489ed110f48aa945851a2bf2bd2c185ffa00cb951ac70e" //见证人对块签名
  }
}
```
#### 34.查询地址投票情况

###### 请求

```json
{
  "method": "getvoteaddr",
  "params": {
    "address": "MMSNVdcn7LYbHdCiNxpHScsCYonvEH1n9tSK4"
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "balance": 0,//余额
    "balance_f": 148388368961,//冻结余额
    "role": 2,//角色 1见证人，2社区，3轻节点
    "depositin": 1000000000000,//押金
    "voteaddr": "MMS47sWYUKuVqytYPzwjrJG3y25bEsFMthVZ4",//给哪个社区节点地址投票
    "votein": 0,//投票金额
    "votenum": 3000000000 //获得的投票数量
  }
}
```

#### 35.查询见证者节点详情

###### 请求

```json
{
   "method": "getwitnessnodedetail",
  "params": {
    "address": "SELFLSLp9AwMLM5cYc9QKboYZD29xPeFLsqZx5",
    "page": 1,
    "page_size": 10
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "vote": 1000000000000,   // 投票数量
    "deposit": 10000000000000, // 质押量
    "add_block_count": 355, // 出块数
    "add_block_reward": 242695383252248, // 出块奖励
    "community_count": 1, // 社区数量
    "community_node": [  // 社区详情
      {
        "name": "", // 社区名字
        "addr": "SELFA4d5qsGYF6iB3FRobkoWp54GxjShWa9ah5", //社区地址
        "deposit": 1000000000000, // 社区节点质押量
        "reward": 21841647762, // 社区节点奖励
        "light_num": 1, // 轻节点数量
        "vote_num": 3000000000 // 轻节点质押数量
      }
    ]
  }
}
```

#### 36.获取账户交易历史

###### 请求

```json
{
  "method": "getaddresstx",
  "params": {
    "address": "iCom7MU1nVsstAXGuXog32EQufPrktkAbxKtj5",
    "contractaddress":"", //也可以过滤代币
    "page":1,
    "page_size":100
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "count": 1192, // 出块总数
    "data": [
      {
        "blockhash": "0f91f3a04644b04cdc3477696c9fe83e6c351dd56331fbec9f2b55e89b04d8c9", // 区块hash
        "blockheight": 384, // 区块高度
        "timestamp": 1670901465, // 区块时间
        "txinfo": {
          "hash": "01000000000000003fbc26a22bc3da196682f1d9f96a104a6784b611bbf594b20375ea159adcf3a3", //交易hash
          "type": 1, //交易类型  1 挖矿 2 见证人缴纳押金 3 见证人押金退出 4 普通支付 5 申请名称 6 注销名称 7 参与见证人投票输入 8 参与见证人投票输出  10 token发布 11 token支付 12 存储挖矿押金输入，余额锁定 13 存储挖矿押金输出，余额解锁 14 用户存储空间押金输入 15 用户存储空间押金输出,
          //        16 社区节点给轻节点分发奖励 17 nft交易 18 两个nft相交换 19 nft销毁 20 合约交易
          "vin_total": 1, // vin交易笔数
          "vin": [
            {
              "addr": "MMS9zuZcoLPsXjARM2msuZYvaWkXw4utbBS44", // 地址
              "puk": "829b0c3ba7e630adf98685a98c71c6957ccf92d4777f1e7a52d4239a1732e294", // 公钥
              "n": "0",
              "sign": "08f38adaa566861523b40436d0b40b84cb84afd9590b5a1be113fad41b9f78cfa7e535053f6fe9041341349f35f3284a3c2cdca73da23956c2a9328077742809"
            }
          ],
          "vout_total": 2, // vout交易笔数
          "vout": [
            {
              "value": 1054814649, // 交易值  需要除以10e8
              "address": "SELFLSLp9AwMLM5cYc9QKboYZD29xPeFLsqZx5", // 地址
              "frozen_height": 0
            },
            {
              "value": 445632931,
              "address": "SELFA4d5qsGYF6iB3FRobkoWp54GxjShWa9ah5",
              "frozen_height": 0
            }
          ],
          "gas": 0,
          "lock_height": 384,
          "payload": "",
          "blockhash": "",
          "vt": 1, //投票类型 type为7时为输入 为8时为输出. 1=给见证人投票；2=给社区节点投票；3=轻节点押金；
          "v":"SELFA4d5qsGYF6iB3FRobkoWp54GxjShWa9ah5" // 见证人地址
        },
        "upchaincode": 2 // 1=未确认；2=成功；3=失败；
      }
    ]
  }
}
```

#### 37.获取所有交易历史

###### 请求

```json
{
  "method": "getalltx",
  "params": {
    "page": 1,
    "page_size": 2   
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "count": 1192, // 出块总数
    "data": [
      {
        "blockhash": "0f91f3a04644b04cdc3477696c9fe83e6c351dd56331fbec9f2b55e89b04d8c9", // 区块hash
        "blockheight": 384, // 区块高度
        "timestamp": 1670901465, // 区块时间
        "txinfo": {
          "hash": "01000000000000003fbc26a22bc3da196682f1d9f96a104a6784b611bbf594b20375ea159adcf3a3", //交易hash
          "type": 1, //交易类型  1 挖矿 2 见证人缴纳押金 3 见证人押金退出 4 普通支付 5 申请名称 6 注销名称 7 参与见证人投票输入 8 参与见证人投票输出  10 token发布 11 token支付 12 存储挖矿押金输入，余额锁定 13 存储挖矿押金输出，余额解锁 14 用户存储空间押金输入 15 用户存储空间押金输出,
          //        16 社区节点给轻节点分发奖励 17 nft交易 18 两个nft相交换 19 nft销毁 20 合约交易
          "vin_total": 1, // vin交易笔数
          "vin": [
            {
              "addr": "MMS9zuZcoLPsXjARM2msuZYvaWkXw4utbBS44", // 地址
              "puk": "829b0c3ba7e630adf98685a98c71c6957ccf92d4777f1e7a52d4239a1732e294", // 公钥
              "n": "0",
              "sign": "08f38adaa566861523b40436d0b40b84cb84afd9590b5a1be113fad41b9f78cfa7e535053f6fe9041341349f35f3284a3c2cdca73da23956c2a9328077742809"
            }
          ],
          "vout_total": 2, // vout交易笔数
          "vout": [
            {
              "value": 1054814649, // 交易值  需要除以10e8
              "address": "SELFLSLp9AwMLM5cYc9QKboYZD29xPeFLsqZx5", // 地址
              "frozen_height": 0
            },
            {
              "value": 445632931,
              "address": "SELFA4d5qsGYF6iB3FRobkoWp54GxjShWa9ah5",
              "frozen_height": 0
            }
          ],
          "gas": 0,
          "lock_height": 384,
          "payload": "",
          "blockhash": "",
          "vt": 1, //投票类型 type为7时为输入 为8时为输出. 1=给见证人投票；2=给社区节点投票；3=轻节点押金；
          "v":"SELFA4d5qsGYF6iB3FRobkoWp54GxjShWa9ah5" // 见证人地址
        },
        "upchaincode": 2 // 1=未确认；2=成功；3=失败；
      }
    ]
  }
}
```

#### 38.获取轻节点详情

###### 请求

```json
{
  "method": "getlightnodedetail",
  "params": {
    "address": "SELFA6SMdseQ4np1kwr9f1GBudjqWfCJfs1YP5"    
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "community_addr": "SELFA4d5qsGYF6iB3FRobkoWp54GxjShWa9ah5",
    "deposit": 1000000000,
    "light_addr": "SELFA6SMdseQ4np1kwr9f1GBudjqWfCJfs1YP5",
    "reward": 0,
    "vote": 3000000000
  }
}
```


#### 39.获取获取见证者列表

###### 请求

```json
{
  "method": "getwitnesslistforminer",
  "params": {
    "page": 1,
    "page_size": 2
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "count": 1,  // 总数量
    "data": [
      {
        "addr": "SELFLSLp9AwMLM5cYc9QKboYZD29xPeFLsqZx5", // 地址
        "payload": "first_witness", // 名字
        "score": 10000000000000, // 质押量
        "vote": 0, // 总票数
        "add_block_count": 255, // 出块数
        "add_block_reward": 242545301561246, // 出块奖励
        "ratio": 1 // 奖励比例
      }
    ]
  }
}
```

#### 40.获取社区节点列表(带分页)

###### 请求

```json
{
  "method": "getcommunitylistforminer",
  "params": {
    "page": 1,
    "page_size": 2
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "count": 1,  // 总数量
    "data": [
      {
        "addr": "SELFLSLp9AwMLM5cYc9QKboYZD29xPeFLsqZx5", // 地址
        "payload": "first_witness", // 名字
        "score": 10000000000000, // 质押量
        "vote": 0, // 总票数
        "add_block_count": 255, // 出块数
        "add_block_reward": 242545301561246, // 出块奖励
        "reward_ratio": 1, // 奖励比例
        "dis_ratio": 1 // 分配比例
      }
    ]
  }
}
```


#### 41.获取出块详情

###### 请求

```json
{
  "method": "getminerblock",
  "params": {
    "address": "SELFLSLp9AwMLM5cYc9QKboYZD29xPeFLsqZx5",
    "page":1,
    "page_size": 1
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "count": 1192, // 出块总数
    "data": [
      {
        "block_reward": 1500738612, // 出块奖励
        "blockhash": "2cf52a91923a401a264ab3ae01599d4bc0985e8b37c654ebc1bb5be0fa6e809a", // 区块hash
        "blockheight": 1193, // 区块高度
        "destroy": 1, // 销毁数量
        "previous_hash": "812fc0a7b71130b80cd538f4c47ba0ac784f2a6d1fb2f45d26dbbc5071075bf0", // 父块hash
        "timestamp": 1670837311, // 出块时间
        "tx_count": 1 // 交易数量
      }
    ]
  }
}
```


#### 42.获取社区节点详情

###### 请求

```json
{
  "method": "getcommunitynodedetail",
  "params": {
    "address": "SELFLSLp9AwMLM5cYc9QKboYZD29xPeFLsqZx5",
    "page": 1,
    "page_size": 1
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "deposit": 0,  // 质押量
    "vote": 0, // 票数
    "light_count": 0, // 轻节点数量 
    "reward_ratio": 0, // 奖励比例
    "reward": 0, // 奖励
    "witness_name": "", // 见证者节点名字
    "witness_addr": "", // 见证者地址
    "light_node": [
      {
        "addr": "1", // 投票地址
        "reward": 0, // 奖励
        "reward_ratio": 0, // 奖励比例
        "vote_num": 0 // 票数
      }
    ]
  }
}
```




#### 43.获取交易手续费

###### 请求

```json
{
  "method": "gettxgas",
  "params": {
    "type": 4  // 交易类型  可有不传  默认为4 普通转账交易
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "fast": 12500,  // 快速
    "low": 10000,  // 慢速
    "normal": 10000 // 正常
  }
}
```




#### 101.CreateOfflineTx:创建离线交易

###### 请求

[//]: # (keyStorePath *C.char, srcaddress *C.char, address *C.char, pwd *C.char, comment *C.char, amount *C.char, gas *C.char, frozenHeight *C.char, nonce *C.char, currentHeight *C.char, domain *C.char, domainType *C.char)
```json
{
    "keyStorePath": "配置文件地址",
    "srcaddress": "MMSJWu5zw2YwCPmbJGr6hsNgJsiAQJR6sycT4",
    "address": "MMSEyCk9gMc76dpUwXPLhZuQLTWKzY7Cwi124",
    "pwd":"xhy19liu21@",
    "comment": "",
    "amount": 100000000000,
    "gas": 100000000,
    "frozenHeight": 7,
    "nonce":0,
    "currentHeight":2,
    "domain":"1",
    "domain_type":1
}
```
###### 返回说明
```json
{
  "code": 200,
  "data": "CsoBCigEAAAAAAAAAB495KEwaP208wQMh7rBtBtcLG2JRyaK8qb2s9cMpNfdEAQYASJnCiBFVyBx72zBTHTuDamWl+cRHHNrDbSwSQkCznCxQiPGcBJAPEFmsvVtDX/MjQBEYPuPzvEub+xjNwXeYzxJU3BpJVLndMcCxUun8k4+s9uw/r8wfapIJsylp1hDGtgL8h9dCBoBBygBMicIgNDbw/QCEhxNTVPSK5/Hwvf0d6UMfovUwlP6DP6jw6SaT/gDGAc4gMLXL0C3KQ=="

}
```


#### 102.CreateOfflineContractTx:创建离线合约交易

###### 请求

[//]: # (keyStorePath *C.char, srcaddress *C.char, address *C.char, pwd *C.char, comment *C.char, amount *C.char, gas *C.char, frozenHeight *C.char, gasPrice *C.char, nonce *C.char, currentHeight *C.char, domain *C.char, domainType *C.char, abi *C.char, source *C.char)
```json
{
  "keyStorePath": "配置文件地址",
    "srcaddress": "MMSJWu5zw2YwCPmbJGr6hsNgJsiAQJR6sycT4",
    "address": "",
    "pwd":"xhy19liu21@",
  "comment": "",
    "amount": 0,
    "gas": 1000000,
    "frozenHeight": 7,
    "gasPrice":1,
  "nonce":0,
  "currentHeight":2,
  "domain":"1",
  "domain_type":1,
  "abi": "abi",
  "source": "source"
}
```
###### 返回说明
```json
{
  "code": 200,
  "data": "CsoBCigEAAAAAAAAAB495KEwaP208wQMh7rBtBtcLG2JRyaK8qb2s9cMpNfdEAQYASJnCiBFVyBx72zBTHTuDamWl+cRHHNrDbSwSQkCznCxQiPGcBJAPEFmsvVtDX/MjQBEYPuPzvEub+xjNwXeYzxJU3BpJVLndMcCxUun8k4+s9uw/r8wfapIJsylp1hDGtgL8h9dCBoBBygBMicIgNDbw/QCEhxNTVPSK5/Hwvf0d6UMfovUwlP6DP6jw6SaT/gDGAc4gMLXL0C3KQ=="

}
```


#### 103.GetComment:获取comment

###### 请求

```json
{
    "tag": "LaunchDomain",
  "jsonData": "{\"len\":2,\"openTime\":1672727374,\"foreverPrice\":10000,\"price\":10,\"reNewPrice\":10,\"tiers\":[[3,9000]]}"
}

{
  "tag": "TransferStock",
  "jsonData": "{\"stocks\":[[\"aa\",1]],\"src\":\"如果stocks为空,绑定合约的钱包地址就是100%股权\",}"
}

{
  "tag": "DelayBaseRegistar",
  "jsonData": "{\"src\":\"如果stocks为空,绑定合约的钱包地址就是100%股权\",\"ens\":\"MMS1111111111111111111G2Pgos4\",\"ensPool\":\"iCom1111111111111111111goqach5\",\"name\":\"pl\",\"stocks\":[[\"aa\",1]]}"
}

{
  "tag": "AddDomain",
  "jsonData": "{\"name\":\"11\",\"open_time\":1672727374,\"forever_price\":10000,\"price\":10,\"reNewPrice\":10,\"tiers\":[[3,9000]]}"
}

{
  "tag": "SetDomainManger",
  "jsonData": "{\"registar\":\"MMSCipLzsJQgqYTyMKZQHhtA2F5ub2CaU3Tr4\",\"name\":\"pl\"}"
}

{
  "tag": "RegisterDomain",
  "jsonData": "{\"owner\":\"MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4\",\"duration\":31536000,\"forever\":false,\"name\":\"aaa\"}"
}

{
  "tag": "ReNewDomain",
  "jsonData": "{\"duration\":31536000,\"forever\":false,\"name\":\"aaa\"}"
}
{
  "tag": "SetDomainImResolver",
"jsonData": "{\"coin_type\":1,\"sub\":\"aa\",\"root\":\"\",\"im_address\":\"MMSFd6524tJjymNX4toTYkrNrMkUUtuV6Hj14\"}"
}
{
  "tag": "SetDomainOtherResolver",
  "jsonData": "{\"coin_type\":1,\"sub\":\"pl\",\"root\":\"\",\"other_address\":\"5B38Da6a701c568545dCfcB03FcB875f56beddC4\"}"
}

{
  "tag": "DomainWithDraw",
  "jsonData": ""
}

{
  "tag": "DomainTransfer",
  "jsonData": "{\"to\":\"MMSL3J1bRfi21iTLyRJU2eKtDKGhFMBZf2sN4\"}"
}
{
"tag": "DomainTransferEns",
"jsonData": "{\"to\":\"MMSL3J1bRfi21iTLyRJU2eKtDKGhFMBZf2sN4\",\"name\":\"pl\"}"
}
{
  "tag": "SetLockDomain",
  "jsonData": "{\"contractaddress\":\"MMS1111111111111111111kin8GP4\",\"names\":[[\"aa\",\"bb\"]],\"is_root\":true}"
}
{
  "tag": "DelDomainImResolver",
  "jsonData": "{\"sub\":\"pl\",\"root\":\"\",\"coin_type\":0}"
}

{
  "tag": "UnLockDomain",
  "jsonData": "{\"contractaddress\":\"MMSErfnkkcHvzCTGoAGB7caD2nWpFYsyFVpM4\",\"names\":[\"aa\",\"bb\"],\"is_root\":true}"
}

{
  "tag": "ModifyLaunchDomain",
  "jsonData": "{\"len\":2,\"open_time\":1672727374,\"forever_price\":10000,\"price\":10,\"reNewPrice\":10,\"tiers\":[[3,9000]]}"
}

{
  "tag": "AbortLaunchDomain",
  "jsonData": "{\"len\":100}"
}

{
  "tag": "ModifyAddDomain",
  "jsonData": "{\"name\":\"pl\",\"open_time\":1672727374,\"forever_price\":10000,\"price\":10,\"reNewPrice\":10,\"tiers\":[[3,9000]]}"
}

{
  "tag": "AbortAddDomain",
  "jsonData": "{\"name\":\"pl\"}"
}

{
"tag": "SetTypeName",
"jsonData": "{\"typeName\":\"字母\",\"name\":[\"aa\",\"bb\"]}"
}

{
"tag": "LaunchDomainType",
"jsonData": "{\"typeName\":\"字母\",\"openTime\":1672727374,\"foreverPrice\":10000,\"price\":10,\"reNewPrice\":10,\"tiers\":[[3,9000]]}"
}

{
"tag": "LaunchDomainTypeAbort",
"jsonData": "{\"typeName\":\"字母\"}"
}

{
"tag": "LaunchDomainTypeModify",
"jsonData": "{\"typeName\":\"字母\",\"openTime\":1672727374,\"foreverPrice\":10000,\"price\":10,\"reNewPrice\":10,\"tiers\":[[3,9000]]}"
}
{
"tag": "SetMasterLock",
"jsonData": "{\"tag\":\"字母\",\"name\":[\"aa\",\"bb\"]}"
}
{
"tag": "LockMasterDomain",
"jsonData": "{\"tags\":[\"aa\",\"bb\"]}"
}
{
"tag": "UnLockMasterDomain",
"jsonData": "{\"tags\":[\"aa\",\"bb\"]}"
}
{
"tag": "SetOperator", 
"jsonData": "{\"addrs\":[\"iComCAfwku18pJDYNRpDpUpi4hFdeDpKkosgK5\", \"iComLSDsWznbDjcWRESGwb7mVa8uqZP1e7gav5\", \"iComJJK5xUVfL7DRH3E6256La7Ce2pUGNEAoz5\", \"iComCanieydNc3if8HgZ3awpYVZ56S89zdiZg5\", \"iCom4Ms4aXMcEtmDbmAeWFdvNjfRtPacVxyVP5\"]}"
}
{
  "tag": "DelOperator",
  "jsonData": "{\"addrs\":[\"iComCAfwku18pJDYNRpDpUpi4hFdeDpKkosgK5\", \"iComLSDsWznbDjcWRESGwb7mVa8uqZP1e7gav5\", \"iComJJK5xUVfL7DRH3E6256La7Ce2pUGNEAoz5\", \"iComCanieydNc3if8HgZ3awpYVZ56S89zdiZg5\", \"iCom4Ms4aXMcEtmDbmAeWFdvNjfRtPacVxyVP5\"]}"
}
{
  "tag": "ResetTiers",
  "jsonData": "{\"isRenew\":false,\"tiers\":[[3,9000],[4,8000]]}"
}


```

#### tag参数

| 值                              | 类型                              |
|--------------------------------|---------------------------------|
| LaunchDomain                   | 投放域名(长度)                        |
| DelayBaseRegistar              | 部署某个域名的注册器                      |
| DelayController                | 没用                              |
| BuildDelayPublicResolver       |                                 |
| BuildDelayPublicReversResolver |                                 |
| AddDomain                      | 投放域名（名称）                        |
| SetDomainManger                | * 初始化设置注册表的注册器 * 创世节点初始化时候运行下接口 |
| RegisterDomain                 | 注册域名                            |
| ReNewDomain                    | 域名续费                            |
| SetDomainImResolver            | 解析域名到主链地址                       |
| SetDomainOtherResolver         | 解析域名到其他币种地址                     |
| DomainWithDraw                 | 域名提现                            |
| DomainTransfer                 | 域名转让                            |
| SetLockDomain                  | .锁定域名                           |
| SetReverseResolverName         |                                 |
| DelDomainImResolver            | 删除解析                            |
| UnLockDomain                   | 解锁域名                            |
| ModifyLaunchDomain             | 修改投放域名（长度）                      |
| AbortLaunchDomain              | 终止投放域名（长度）                      |
| ModifyAddDomain                | 修改投放域名（名称）                      |
| AbortAddDomain                 | 终止投放域名（名称）                      |
| TransferStock                  | 转让股权                            |
| SetTypeName                    | 初始化类型投放 类型和名字                   |
| LaunchDomainType               | 构建域名类型投放                        |
| LaunchDomainTypeAbort          | 构建终止域名类型投放                      |
| LaunchDomainTypeModify         | 构建域名类型投放修改                      |
| SetMasterLock                  | 初始化官方锁定 tag和域名                  |
| LockMasterDomain               | 锁定官方                            |
| UnLockMasterDomain             | 解除锁定官方                          |
| SetOperator                    | 设置根注册器操作员(最大5个)                 |
| DelOperator                    | 删除根注册器操作员                       |
| ResetTiers                    | 重置折扣率                           |


###### 返回说明
```json
{
  "code": 200,
  "data": "CsoBCigEAAAAAAAAAB495KEwaP208wQMh7rBtBtcLG2JRyaK8qb2s9cMpNfdEAQYASJnCiBFVyBx72zBTHTuDamWl+cRHHNrDbSwSQkCznCxQiPGcBJAPEFmsvVtDX/MjQBEYPuPzvEub+xjNwXeYzxJU3BpJVLndMcCxUun8k4+s9uw/r8wfapIJsylp1hDGtgL8h9dCBoBBygBMicIgNDbw/QCEhxNTVPSK5/Hwvf0d6UMfovUwlP6DP6jw6SaT/gDGAc4gMLXL0C3KQ=="
}
```


#### 104.MultDeal:同时处理comment、合约和push

###### 请求

```json
{
  "method": "multDeal",
  "params": {
    "tag": "LaunchDomain",
    "jsonData": "{\"len\":101,\"openTime\":1672727374,\"foreverPrice\":10000,\"price\":10}",
    "srcaddress": "MMS22TkFXtaivPEN29urwWqDMhY5HAs8gLpx4",
    "address": "MMS1111111111111111111PnLDCa4",
    "amount": 0,
    "gas": 100000,
    "frozen_height": 7,
    "gas_price": 1,
    "pwd": "xhy19liu21@",
    "nonce": 1,
    "currentHeight": 2
  }
}
```
###### 返回说明
```json
{
  "jsonrpc": "2.0",
  "code": 5003,
  "message": "预执行合约失败,退出码:REVERT,失败原因:the len domain has been add"
}
```

#### 105.createContract:部署合约
###### 请求
```json lines
{
    "method": "createContract",
    "params": {
        "srcaddress": "iCom7bia52M2vGnVqbGBWMnXPmv9LKq5MHFSQ5",
        "amount": 0,
        "pwd": "123456789",
        "comment": "",
        "gas": 100000,
        "source": ""
    }
}
//srcaddress:钱包地址
//amount:部署时转账金额
//pwd:钱包密码
//comment：合约bytecode
//gas：手续费
//source：合约源码
```
###### 返回说明
```json lines
//result.contract_address：部署合约的地址
//result.hash：部署合约交易的hash
```

#### 106.getContractInfo:获取合约信息
###### 请求
```json lines
{
  "method": "getContractInfo",
  "params": {
    "address": "iCom1111111111111111111A3DxBN5"
  }
}
//address:合约地址
```
###### 返回说明
```json lines
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "is_contract": true,
    "contract_status": 1,
    "contract_code": "",
    "name": "",
    "compiler_version": "0.4.25+commit.59dbf8f1",
    "abi": "",
    "source": "",
    "bin": ""
  }
}
//is_contract：是否是合约
//contract_status：合约状态 0无效,1正常，2自毁合约
//contract_code：合约code，部署后保存的code码
//name：合约名，只有奖励合约有值
//compiler_version：编译器版本，只有奖励合约有值
//abi：合约的abi结构
//source：合约的源码
//bin：合约的bytecode
```

#### 107.staticCallContract:调用合约
* 合约查询类方法用
* 直接返回查询结果
###### 请求
```json lines
{
  "method": "staticCallContract",
  "params": {
    "srcaddress": "iComJVWNgfcL8P919g77AfvikxRCAx17bSoyR5",
    "contractaddress":"iCom6ckfJiwAPtsLRYGQaiRtvRPrp13TDD7fp5",
    "gas":10000,
    "amount":0,
    "comment": "06fdde03"
  }
}
//srcaddress:钱包地址
//contractaddress：合约地址
//amount:转账金额，固定0
//gas：手续费
//comment：合约调用input
```
###### 返回说明
```json lines
//result:合约查询结果
```

#### 108.callContract:调用合约
* 会生成一笔交易，写类方法用
* 返回交易信息
###### 请求
```json lines
{
  "method": "callContract",
  "params": {
    "srcaddress": "iComNLtbKrd3BWYbLZau1kkBPFueGYg8VWbbX5",
    "contractaddress":"iCom1111111111111111111A3DxBN5",
    "amount":0,
    "gas":10000,
    "gas_price":1,
    "pwd":"123456789",
    "comment": "4fb2e45d0000000000000000000000004fdd5b6099a411d24cea34b7c2e1d63eea95cacc"
  }
}
//srcaddress:钱包地址
//contractaddress：合约地址
//amount:转账金额，固定0
//gas：手续费
//gas_price:gas价格,
//pwd：钱包密码
//comment：合约调用input
```
###### 返回说明
```json lines
//result.hash:生成合约交易的hash
```

#### 109.getcontractevent:获取合约交易事件
###### 请求
```json lines
{
  "method": "getcontractevent",
  "params": {
    "height": 3,
    "hash": "01000000000000006b25fb7df88d3ebf1f6230b214664b0124bf77660509eadaf8c743f93b2b1633"
  }
}
//height:区块高度
//hash：交易hash
```
###### 返回说明
```json lines
{
  "jsonrpc": "2.0",
  "code": 2000,
  "result": {
    "contract_events": [
      {
        "block_height": 3,
        "topic": "900928d4a48773c713a00e26efa83f1f9fe5a1e65bb937e6d8773daf54112983",
        "tx_id":"",
        "contract_address": "iCom1111111111111111111A3DxBN5",
        "event_data": [
          "000000000000000000000000486199327f23daedef799ce8e0e8209d0cf62e3f",
          "0000000000000000000000000000000000000000000000000000000000000000",
          "0000000000000000000000000000000000000000000000000000000036fd76a5000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000"
        ]
      }
  ]}
}
//topic:合约函数签名
//tx_id：交易hash
//contract_address：合约地址
//event_data：合约事件，数组最后一个元素为普通事件参数的并集，其他元素为事件的indexed标记元素
```
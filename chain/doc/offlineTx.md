## 离线交易测试 
#### 请求
```json lines
{
  "method": "createOfflineTxV1",
  "params": {
        "pwd": "123456789",
        "nonce": 0,
        "currentHeight": 2,
        "frozen_height": 2,
        "domain": "",
        "domain_type": 1,
        "key_store_path": "",
        "tag": "transfer",
        "jsonData": "{\"srcaddress\":\"iCom8AecPEDoRbpDZt4vDLm82Zp9XVxXjZYjm5\",\"address\":\"iCom8HHZJfjZatiESDdLRyiDzShHBr3iWueFw5\",\"amount\":20000000000023,\"gas\":1000001,\"comment\":\"test\"}"
  }
}
/****通用参数****/
// pwd:钱包密码
// nonce:实时获取地址nonce
// currentHeight:实时获取当前高度
// frozen_height:冻结高度，可和currentHeight高度一致
// domain:域名
// domainType:域名类型
// key_store_path:keystore路径，空则默认为：conf/keystore.key

/****可变参数****/
// tag:交易类型
// jsonData:交易数据
```
#### 返回
```json lines
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": {
        "isContract": false,
        "contractAddress": "",
        "hash": "0400000000000000682276539dba7750c8ee989b9ce6028c95c82bfa89971a149dfb377f31179a9c",
        "tx": "CtEBCigEAAAAAAAAAGgidlOdundQyO6Ym5zmAoyVyCv6iZcaFJ37N38xF5qcEAQYASJnCiAdtaBz5uZ5IcBAXG2SQrLxvJk5QHFt7gdEiswF9QHcLhJAhaIa3mRwX6rol2dC0CNMxf/rKmmQwa+B4PEw0b7QJ9UVvG1UmyOmqD5oRlDKhaQhZuZ/JQNMWx7ShTY0RvsXBxoBASgBMikIl4CV54nGBBIdaUNvbU/dW2CZpBHSTOo0t8Lh1j7qlcrMkLecJgQYAjjBhD1ArgJKBHRlc3Q="
    }
}
// isContract:是否合约调用
// contractAddress:调用合约地址
// hash:交易hash
// tx:交易base64字符串
```
#### 推送交易rpc
* 离线构建交易成功后，返回`isContract`：是否合约调用
* true：表示合约调用，推送交易使用rpc：`pushContractTx`
* false：普通转账，推送交易使用rpc：`pushtx`

```json lines
// 转账交易推送rpc
{
  "method": "pushtx",
  "params": {
    "tx":"CtEBCigEAAAAAAAAAEgC66z0aNRapfz8E75Adi977lUMMOh1E7cH9UO4+tlkEAQYASJnCiAdtaBz5uZ5IcBAXG2SQrLxvJk5QHFt7gdEiswF9QHcLhJAPpibZdEeQpmJKe2q3bFHEjJTaSo49ur/qXnx2Uzn5KfuT1c/38pCpQeShOTqtdwvo08xj4Cn5qUeKtgjz23gCBoBAigBMikIl4CV54nGBBIdaUNvbU/dW2CZpBHSTOo0t8Lh1j7qlcrMkLecJgQYATjBhD1ArQJKBHRlc3Q="
  }
}

// 合约交易推送rpc
{
  "method": "pushContractTx",
  "params": {
    "tx":"CqoDCigUAAAAAAAAAGa+8W7+/S9GrlB0JQk3rh2njI5seAzOLaZvjI7Sj5knEBQYASJnCiA1KuA5hAUn8WG0YM/6DfohTQdd04IfvPqMhW7Cvp/wIRJAFWFkrYreL77V77yG1gB5cX9QzyDeRgkKKmkiHJgcb2JbxKNbsgAjUIS40HqudtTv29f5kr4YLM9jcpI+1RzBBRoBGSgBMiESHWlDb20AAAAAAAAAAAAAAAAAAAAAAAAAAiNN09wEKAE4gIl6QKkISuQBlHEHwH29nnc7EicKEiDLDqIN5i7NWQbOI+0toZacaH91ctoEAAAAAAAAAAAAAAAAj0iDiNRAAPqbiOvJSDQ7lxaVncEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABwgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAKAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA2NndgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEgRjYWxsKAE="
  }
}
// tx:交易base64字符串
```

#### tag / jsonData 数据示例

###### 转账 `transfer`
```json lines
{
	"srcaddress": "iCom8AecPEDoRbpDZt4vDLm82Zp9XVxXjZYjm5",
	"address": "iCom8HHZJfjZatiESDdLRyiDzShHBr3iWueFw5",
	"amount": 20000000000023,
	"gas": 1000001,
	"comment": ""
}
//srcaddress：转账from地址
//address：转账to地址
//amount：金额
//gas：手续费
//comment:交易payload
```

###### 成为社区 `addCommunity`
```json lines
{
  "srcaddress": "iCom8HHZJfjZatiESDdLRyiDzShHBr3iWueFw5",
  "witnessAddress": "iCom8AecPEDoRbpDZt4vDLm82Zp9XVxXjZYjm5",
  "amount": 100000000000,
  "gas": 1000001,
  "gasPrice": 1,
  "rate": 10,
  "name": "shequ"
}
//srcaddress：调用者钱包地址
//witnessAddress：见证人地址
//amount：质押金额，固定 100000000000
//gas：手续费
//gasPrice：gas单价，固定 1
//rate：发奖比例，0-100
//name：设置社区名
```
###### 取消社区 `cancelCommunity`
```json lines
{
  "srcaddress": "iCom8HHZJfjZatiESDdLRyiDzShHBr3iWueFw5",
  "amount": 100000000000,
  "gas": 1000001,
  "gasPrice": 1
}
//srcaddress：调用者钱包地址
//amount：取消质押金额，固定 100000000000
//gas：手续费
//gasPrice：gas单价，固定 1
```

###### 成为轻节点 `addLight`
```json lines
{
  "srcaddress": "iComNLtbKrd3BWYbLZau1kkBPFueGYg8VWbbX5",
  "amount": 1000000000,
  "gas": 1000001,
  "gasPrice": 1,
  "name": "qing"
}
//srcaddress：调用者钱包地址
//amount：质押金额，固定 1000000000
//gas：手续费
//gasPrice：gas单价，固定 1
//name：设置轻节点名
```

###### 取消轻节点 `cancelLight`
```json lines
{
  "srcaddress": "iComNLtbKrd3BWYbLZau1kkBPFueGYg8VWbbX5",
  "amount": 1000000000,
  "gas": 1000001,
  "gasPrice": 1
}
//srcaddress：调用者钱包地址
//amount：取消质押金额，固定 1000000000
//gas：手续费
//gasPrice：gas单价，固定 1
```

###### 投票 `addVote`
```json lines
{
  "srcaddress": "iComNLtbKrd3BWYbLZau1kkBPFueGYg8VWbbX5",
  "communityAddress": "iCom8HHZJfjZatiESDdLRyiDzShHBr3iWueFw5",
  "amount": 10,
  "gas": 1000001,
  "gasPrice": 1
}
//srcaddress：调用者钱包地址
//communityAddress：社区节点地址
//amount：投票金额
//gas：手续费
//gasPrice：gas单价，固定 1
```

###### 取消投票 `cancelVote`
```json lines
{
  "srcaddress": "iComNLtbKrd3BWYbLZau1kkBPFueGYg8VWbbX5",
  "communityAddress": "iComCAfwku18pJDYNRpDpUpi4hFdeDpKkosgK5",
  "amount": 10,
  "gas": 1000001,
  "gasPrice": 1
}
//srcaddress：调用者钱包地址
//communityAddress：社区节点地址
//amount：投票金额
//gas：手续费
//gasPrice：gas单价，固定 1
```

###### 代币转账 `transferErc20`
```json lines
{
  "srcaddress": "iCom8AecPEDoRbpDZt4vDLm82Zp9XVxXjZYjm5",
  "contractAddress": "iComK7fS4Uu7Zyp12N9VY49PNBum17MqBQVui5",
  "toAddress": "iCom8HHZJfjZatiESDdLRyiDzShHBr3iWueFw5",
  "amount": "10",
  "gas": 1000001,
  "gasPrice": 1,
  "decimal":8
}
//srcaddress：调用者转账from地址
//contractAddress:调用合约地址
//toAddress:转账to地址
//amount：代币转账金额，字符串类型，不使用精度计算 如："1.234"
//gas：手续费
//gasPrice：gas单价，固定 1
//decimal：erc20代币精度
```

###### 奖励提现 `rewardWithdraw`
```json lines
{
  "srcaddress": "iCom8HHZJfjZatiESDdLRyiDzShHBr3iWueFw5",
  "gas": 1000001,
  "gasPrice": 1
}
//srcaddress：调用者钱包地址
//gas：手续费
//gasPrice：gas单价，固定 1
```


###### SDK构建离线签 `BuildOfflineTx`
#### 参数
| 值        | 类型     | 注释                                 |
|----------|--------|------------------------------------|
| keyStorePath    | string | keystore路径，空则默认为：conf/keystore.key |
| pwd    |     string   | 钱包密码                               |
| nonce    |    string    | 实时获取地址nonce                        |
| currentHeight   |  string      | 实时获取当前高度                           |
| frozenHeight     |   string     | 冻结高度，可和currentHeight高度一致           |
| domainType    |     string   | 域名类型                               |
| domain    |      string  | 域名                                 |
| tag     |    string    | 交易类型                               |
| jsonData    |   string     | 交易数据(json串)                        |


#### jsonData参数
| 值        | 类型     | 注释             |
|----------|--------|----------------|
| srcaddress    | string | from地址         |
| address    |     string   | to地址           |
| amount    |     string   | 金额             |
| gas    |     string   | 油费             |
| comment    |     string   | 通过getComment获取 |

```params demo```
```json lines
{
    "keyStorePath": "",
    "pwd": "123456789",
    "nonce": "0",
    "currentHeight": "2",
    "frozenHeight": "2",
    "domainType": "1",
    "domain": "",
    "tag": "transfer",
    "jsonData": "{\"srcaddress\":\"iCom8AecPEDoRbpDZt4vDLm82Zp9XVxXjZYjm5\",\"address\":\"iCom8HHZJfjZatiESDdLRyiDzShHBr3iWueFw5\",\"amount\":20000000000023,\"gas\":1000001,\"comment\":\"test\"}"
}

```
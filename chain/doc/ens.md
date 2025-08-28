## 初始化
* 初始化设置注册表的注册器
* 创世节点初始化时候运行下接口
```json
{
    "method": "setDomainManger",
    "params": {
    	"srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    	"contractaddress":"MMS1111111111111111111G2Pgos4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "name":"",
        "registar":"MMS1111111111111111111PnLDCa4"
    }
}
//srcaddress:创世节点见证人地址
//gas：交易燃料费
//pwd: 钱包密码
//frozen_height: 冻结高度,
//contractaddress：注册表地址，固定MMS1111111111111111111G2Pgos4
//name：根，固定“”空
//registar：注册表的注册器合约，固定MMS1111111111111111111PnLDCa4
```

## 锁定
#### 锁定域名

```json
{
    "method": "setLockDomain",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "contractaddress": "MMS1111111111111111111kin8GP4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "is_root": true,
        "names": [
            "aa",
            "bb"
        ]
    }
}
//srcaddress:钱包地址，需要为注册器的所有者地址才可以成功
//gas：交易燃料费
//pwd: 钱包密码
//frozen_height: 冻结高度,
//contractaddress：锁定合约地址（全局锁定器合约/域名注册器合约），MMS1111111111111111111kin8GP4为全局锁定器地址，全局锁定会包含所有的域名注册器限制
//names：需要锁定的关键词数组
//is_root: contractaddress为全局锁定器时生效(其他锁定器可忽略)，true为设置全局根域名，false为设置全局子域名
```
#### 解锁域名
```json
{
    "method": "unLockDomain",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "contractaddress": "MMSErfnkkcHvzCTGoAGB7caD2nWpFYsyFVpM4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "is_root": true,
        "names": [
            "aa",
            "bb"
        ]
    }
}
//srcaddress:钱包地址，需要为注册器的所有者地址才可以成功
//gas：交易燃料费
//pwd: 钱包密码
//frozen_height: 冻结高度,
//contractaddress：锁定合约地址（全局锁定器合约/域名注册器合约），MMS1111111111111111111kin8GP4为全局锁定器地址，全局锁定会包含所有的域名注册器限制
//names：需要锁定的关键词数组
//is_root: contractaddress为全局锁定器时生效(其他锁定器可忽略)，true为设置全局根域名，false为设置全局子域名
```
#### 获取锁定列表
`request`
```json
{
    "method": "getLockDomains",
    "params": {
        "contractaddress": "MMS1111111111111111111kin8GP4",
        "is_root": false
    }
}
//contractaddress：锁定合约地址（全局锁定器合约/域名注册器合约），MMS1111111111111111111kin8GP4为全局锁定器地址，全局锁定会包含所有的域名注册器限制
//is_root: contractaddress为全局锁定器时生效(其他锁定器可忽略)，true为设置全局根域名，false为设置全局子域名
```
`response`
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": {
        "lockNameList": [
            "cc",
            "dd",
            "ab",
            "abc"
        ],
        "unLockNameList": []
    }
}
//lockNameList：锁定的域名列表
//unLockNameList：解锁（全局锁定器设置的子域名）的域名列表，此字段查询子域名才会返回数据
```
## 投放
#### 投放域名(长度)

```json
{
    "method": "launchDomain",
    "params": {
    	"srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    	"contractaddress":"MMS1111111111111111111PnLDCa4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "len":2,
        "price":100,
        "open_time":1672727374,
        "forever_price":10000
    }
}
//srcaddress:钱包地址，需要为注册器的所有者地址才可以成功
//gas：交易燃料费
//pwd: 钱包密码
//frozen_height: 冻结高度,
//contractaddress 投放合约地址，MMS1111111111111111111PnLDCa4为根域名的投放
//len：域名字符长度 
//price：每年价格  
//open_time：开放时间  
//forever_price：永久价格  
```

#### 修改投放域名（长度）
```json
{
    "method": "modifyLaunchDomain",
    "params": {
    	"srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    	"contractaddress":"MMS1111111111111111111PnLDCa4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "len":2,
        "price":300,
        "open_time":1672727374,
        "forever_price":30000
    }
}
```
#### 终止投放域名（长度）
```json
{
    "method": "abortLaunchDomain",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "contractaddress": "MMS1111111111111111111PnLDCa4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "len": 2
    }
}
```
#### 投放域名（名称）

```json
{
  "method": "addDomain",
  "params": {
    "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    "contractaddress":"MMS1111111111111111111PnLDCa4",
    "gas": 100000,
    "frozen_height": 1,
    "pwd": "xhy19liu21@",
    "name":"aaa",
    "price":100,
    "open_time":1672727374,
    "forever_price":10000
  }
}
//srcaddress:钱包地址，需要为注册器的所有者地址才可以成功
//contractaddress 投放合约地址，MMS1111111111111111111PnLDCa4为根域名的投放
//gas：交易燃料费
//pwd: 钱包密码
//frozen_height: 冻结高度,
//name 为指定投放的名称，不受`投放域名（长度）`影响
//price：每年价格  
//open_time：开放时间  
//forever_price：永久价格  
```
#### 修改投放域名（名称）
```json
{
    "method": "modifyAddDomain",
    "params": {
    	"srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    	"contractaddress":"MMS1111111111111111111PnLDCa4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "name":"pl",
        "price":300,
        "open_time":1672727374,
        "forever_price":30000
    }
}
```
#### 终止投放域名（名称）
```json
{
    "method": "abortAddDomain",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "contractaddress": "MMS1111111111111111111PnLDCa4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "name": "pl"
    }
}
```
#### 4.获取域名长度价格

```json
{
    "method": "getDomainCost",
    "params": {
        "name": ""
    }
}
//name 查询的域名，空为获取根域名的价格
```
result

```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": [
        {
            "len": 2,
            "price": 100,
            "open_time": 1672727374,
            "forever_price": 10000,
            "status": true
        }
    ]
}
//len：域名字符长度 
//price：每年价格  
//open_time：开放时间  
//forever_price：永久价格  
//status:域名状态
```

#### 5.域名详情

```json
{
    "method": "getDomainDetail",
    "params": {
        "root": "",
        "sub": "pl"
    }
}
//root:根域名
//sub：子域名
```
result
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": {
        "can_register": false,
        "domain_exp": "永久",
        "manager": "MMS111111111111111111112phvcJ4",
        "owner": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4"
    }
}
//can_register：能否注册
//domain_exp：到期时间/永久
//manager：域名管理员地址，（普通账户/合约）
//owner：域名所有者地址
```

#### 6.获取放开的域名列表
```json
{
  "method": "getDomains",
  "params": {
    "name": ".pl",
    "page": 1,
    "page_size": 10
  }
}
//name：查询的域名，例：.xx or xx.xx
//page:页数
//page_size：也大小
```
result
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": {
        "list": [
            {
                "Name": "pl",
                "Price": 100,
                "OpenTime": 1672727374,
                "ForeverPrice": 10000,
                "Status": true,
                "BaseAddr": "",
                "ControllerAddr": ""
            },
            {
                "Name": "aa",
                "Price": 100,
                "OpenTime": 1672727374,
                "ForeverPrice": 10000,
                "Status": true,
                "BaseAddr": "",
                "ControllerAddr": ""
            },
            {
                "Name": "aaa",
                "Price": 100,
                "OpenTime": 1672727374,
                "ForeverPrice": 10000,
                "Status": true,
                "BaseAddr": "",
                "ControllerAddr": ""
            }
        ],
        "total": 3
    }
}
//Name：名称 
//Price：每年价格  
//OpenTime：开放时间  
//ForeverPrice：永久价格  
//Status:域名状态
//BaseAddr:暂不用
//ControllerAddr：暂不用
```
#### 7.获取域名的持有人
```json
{
    "method": "getDomainOwner",
    "params": {
        "root":"",
        "sub":"pl"
    }
}
//root:根域名
//sub：子域名
```
result
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4"
}
```

#### 8.获取名下的根域名
```json
{
    "method": "getMyRootDomain"
}
```
result
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": [
        {
            "Name": "pl",
            "Price": 100,
            "OpenTime": 1672727374,
            "ForeverPrice": 10000,
            "Status": true,
            "BaseAddr": "",
            "ControllerAddr": "",
            "Root": "",
            "Owner": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
            "DomainExp": "永久",
            "Manager": "MMS111111111111111111112phvcJ4"
        },
        {
            "Name": "aa",
            "Price": 100,
            "OpenTime": 1672727374,
            "ForeverPrice": 10000,
            "Status": true,
            "BaseAddr": "",
            "ControllerAddr": "",
            "Root": "",
            "Owner": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
            "DomainExp": "永久",
            "Manager": "MMS111111111111111111112phvcJ4"
        },
        {
            "Name": "aaa",
            "Price": 100,
            "OpenTime": 1672727374,
            "ForeverPrice": 10000,
            "Status": true,
            "BaseAddr": "",
            "ControllerAddr": "",
            "Root": "",
            "Owner": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
            "DomainExp": "永久",
            "Manager": "MMS111111111111111111112phvcJ4"
        }
    ]
}
//Name：名称 
//Price：每年价格  
//OpenTime：开放时间  
//ForeverPrice：永久价格  
//Status:域名状态
//BaseAddr:暂不用
//ControllerAddr：暂不用
//Root:根域名
//Owner：所有者地址
//DomainExp：过期时间/永久
//Manager：管理员地址
```

#### 9.部署某个域名的注册器
```json
{
    "method": "delayBaseRegistar",
    "params": {
    	"srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "amount": 0,
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "ens":"MMS1111111111111111111G2Pgos4",
        "name":"pl"
    }
}
//srcaddress:钱包地址
//gas：交易燃料费
//pwd: 钱包密码
//frozen_height: 冻结高度,
//ens：注册表合约地址
```
#### 10.设置域名的管理员
```json
{
    "method": "setDomainManger",
    "params": {
    	"srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    	"contractaddress":"MMS1111111111111111111G2Pgos4",
        "amount": 0,
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "name":"pl",
        "registar":"MMSCipLzsJQgqYTyMKZQHhtA2F5ub2CaU3Tr4"
    }
}
//srcaddress:钱包地址，需要为注册器的所有者地址才可以成功
//gas：交易燃料费
//pwd: 钱包密码
//frozen_height: 冻结高度,
//contractaddress：注册表地址
//name：设置的域名
//registar：管理员地址 （普通账户/合约地址）
```

#### 11.注册域名
```json
{
    "method": "registerDomain",
    "params": {
    	"srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    	"contractaddress":"MMS1111111111111111111PnLDCa4",
        "amount": 10000,
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "name":"aaa",
        "forever":false,
        "duration":31536000,
        "owner":"MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4"
    }
}
//srcaddress:钱包地址
//gas：交易燃料费
//pwd: 钱包密码
//frozen_height: 冻结高度,
//contractaddress:域名注册器，MMS1111111111111111111PnLDCa4表示注册根域名，其他则是注册子域名
//name:域名名称
//forver：是否永久
//duration：持续时间单位秒
//owner：域名管理员,（普通账户/注册器合约地址）
```

## 域名解析

#### 1.解析域名到主链地址
```json
{
    "method": "setDomainImResolver",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "amount": 0,
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "root":"",
        "sub":"pl",
        "im_address":"MMSFd6524tJjymNX4toTYkrNrMkUUtuV6Hj14"
    }
}
//root:根域名
//sub：子域名
//im_address：解析目标钱包地址
```
#### 2.获取主链域名的解析记录
`request`
```json
{
    "method": "getDomainResolver",
    "params": {
        "root":"",
        "sub":"pl"
    }
}
//root:根域名
//sub：子域名
```
`response`
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": "MMSFd6524tJjymNX4toTYkrNrMkUUtuV6Hj14"
}
//result:域名解析目标地址
```
#### 解析域名到其他币种地址
`request`
```json
{
    "method": "setDomainOtherResolver",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "amount": 0,
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "root":"",
        "sub":"pl",
        "coin_type":1,
        "other_address":"5B38Da6a701c568545dCfcB03FcB875f56beddC4"
    }
}
//root:根域名
//sub：子域名
//coin_type：币类型
//other_address：解析目标钱包地址
```
#### 获取域名的解析记录其他币种
`request`
```json
{
    "method": "getDomainOtherResolver",
    "params": {
        "root": "",
        "sub": "pl",
        "coin_type": 1
    }
}
```
`response`
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": "MMSFd6524tJjymNX4toTYkrNrMkUUtuV6Hj14"
}
//result:域名解析目标地址
```
#### 删除解析
`request`
```json
{
    "method": "delDomainImResolver",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "amount": 0,
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "root": "",
        "sub": "pl",
        "coin_type": 0
    }
}
//coin_type: 0为主链默认im地址，>0 为其他链币
```
## 转让
#### 域名转让
`request`
```json
{
    "method": "domainTransfer",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    	"contractaddress":"MMS1111111111111111111PnLDCa4",
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "from":"MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "to":"MMSL3J1bRfi21iTLyRJU2eKtDKGhFMBZf2sN4",
        "name":"pl"
    }
}
//contractaddress:转让域名的根注册器，根域名固定MMS1111111111111111111PnLDCa4
//from：原owner
//to：新owner
//name：转让的域名
```
## 续费
#### 域名续费
`request`
```json
{
    "method": "renewDomain",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
    	"contractaddress":"MMSErfnkkcHvzCTGoAGB7caD2nWpFYsyFVpM4",
        "amount": 100,
        "gas": 100000,
        "frozen_height": 1,
        "pwd": "xhy19liu21@",
        "name":"aa",
        "forever":false,
        "duration":31536000
    }
}
//forever:是否永久,是则忽略duration
//duration：续费时间，单位秒
```
## 收入
#### 获取域名收入
`request`
```json
{
    "method": "getRootIncome",
    "params": {
        "root": "pl"
    }
}
```
`response`
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": {
        "all_income": 0,
        "curr_withdraw_income": 0,
        "now_income": 0
    }
}
//根域名，为空，表示获取平台售卖根域名的收入
//all_income累计总收入
//curr_withdraw_income当前累计提现
//now_income剩余未提现收入
```
#### 域名提现
`request`
```json
{
    "method": "domainWithDraw",
    "params": {
        "srcaddress": "MMSAFx4zqRsvE8upNqfDzQAY2fDM96ZTR4pa4",
        "contractaddress": "MMSErfnkkcHvzCTGoAGB7caD2nWpFYsyFVpM4",
        "gas": 100000,
        "amount":0,
        "frozen_height": 1,
        "pwd": "xhy19liu21@"
    }
}
//amount：自定义提现金额，0为全部提现
```
## 其他
#### 获取域名基础合约地址
`request`
```json
{
    "method": "getEnsAddr"
}
```
`response`
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": {
        "ens": "MMS1111111111111111111G2Pgos4",
        "lockname": "MMS1111111111111111111kin8GP4",
        "registar": "MMS1111111111111111111PnLDCa4",
        "resolver": "MMS1111111111111111111gBAqbN4"
    }
}
//ens:注册表
//lockname：锁定器
//registar：平台注册器
//resolver：解析器
```
#### 获取热度高的根域名
`request`
```json
{
    "method": "getHotRootDomain",
    "params": {
        "num": 9
    }
}
//num:获取的排名数量
```
`response`
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": [
        {
            "name": "pl",
            "count": 1
        }
    ]
}
//name:根域名
//count：子域名数量
```
#### 获取域名到期时间
`request`
```json
{
    "method": "getDomainExp",
    "params": {
        "root":"",
        "sub":"pl"
    }
}
```
`response`
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": 317037723831
}
//result:到期时间，秒
```
#### 单个域名价格
`request`
```json
{
    "method": "getSingleDomainCost",
    "params": {
        "root": "",
        "sub": "pl"
    }
}
```
`response`
```json
{
    "jsonrpc": "2.0",
    "code": 2000,
    "result": {
        "Name": "pl",
        "Price": 100,
        "OpenTime": 1672727374,
        "ForeverPrice": 10000,
        "Status": true,
        "BaseAddr": "",
        "ControllerAddr": ""
    }
}
//Name：名称 
//Price：每年价格  
//OpenTime：开放时间  
//ForeverPrice：永久价格  
//Status:域名状态
//BaseAddr:暂不用
//ControllerAddr：暂不用
```
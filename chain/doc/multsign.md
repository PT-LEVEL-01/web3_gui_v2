### 多签

#### 1. 查询本节点公钥

*Method:* POST

*请求参数:*
```json
{
    "method": "getpublickey",
    "params": {
        "address": "TESTHa8poK2Bu2k3bjpTNJTkBJc832fxoNEU15"
    }
}
```

#### 2. 创建多签地址

*Method:* POST

*请求参数:*
```json
{
    "method": "createmultsign",
    "params": {
        "srcaddress":"IM8PdTQ6Y2HCwwg8RjrdXWFvfWM1UFACUnV3",
        "gas":100000,
        "pwd":"123456789",
        "puks": [
            "9e71635ed27249ee1238f5aa599ee31259802ffa6f0fc733e8a16e8c7d5a1ba7",
            "207f01fb9eae580b5ecb204b29423787a83dd46cc0ac9d30f6537d00e8edccff",
            "b7c0cb0865a72ee9defc085c022510abe5fd1961f761c3a282333b0038a2cda2"
        ]
    }
}
```

#### 3. 多签转入

*Method:* POST

*请求参数:*
```json
{
    "method": "sendtoaddress",
    "params": {
        "srcaddress": "IM8PdTQ6Y2HCwwg8RjrdXWFvfWM1UFACUnV3",
        "address": "IM6yp3n18UWjWKGc3Ufwc3hqG2LdiJewa5h3",
        "changeaddress": "IM8PdTQ6Y2HCwwg8RjrdXWFvfWM1UFACUnV3",
        "amount": 100000000,
        "gas": 1000000,
        "frozen_height": 7,
        "pwd": "123456789",
        "comment": "test"
    }
}
```

#### 4. 发起多签转出

*Method:* POST

*请求参数:*
```json
{
    "method": "multsignsendtoaddress",
    "params": {
        "multaddress": "IM6yp3n18UWjWKGc3Ufwc3hqG2LdiJewa5h3",
        "address": "IMAnogGgmRYGuYZgXt9UjgJC9rRaTYbcD5r3",
        "amount": 33,
        "gas": 100000,
        "frozen_height": 7,
        "pwd": "123456789"
    }
}
```

#### 5. 查询待签名交易

*Method:* POST

*请求参数:*
```json
{
    "method": "getrequestmultsigns"
}
```

#### 6. 节点签名

*Method:* POST

*请求参数:*
```json
{
    "method": "signmultsign",
    "params": {
        "txid": "2c00000000000000ec50ad31291d7029179ad59afde5f1f4fec979bc24b33ba77666547a83526508",
        "address": "IM2si2uWbSHcmJLPyS4BWm4yDGxRhDVidCt3",
        "pwd": "123456789"
    }
}
```

#### 7. 多签账号列表

*Method:* POST

*请求参数:*
```json
{
    "method": "listmultaccounts"
}
```

#### 8. 多账号信息

*Method:* POST

*请求参数:*
```json
{
    "method": "multiaccounts",
    "params": {
        "addresses": [
            "IM6yp3n18UWjWKGc3Ufwc3hqG2LdiJewa5h3",
            "IMAnogGgmRYGuYZgXt9UjgJC9rRaTYbcD5r3"
        ]
    }
}
```


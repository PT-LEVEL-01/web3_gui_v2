### 代币

#### 1. 发币

*Method:* POST

*请求参数:*
```json
{
    "method": "tokenpublish",
    "params": {
        "srcaddress": "TESTGiuKEgeDxvaZ25izhM3JPU7vANXTVFbpA5",
        "gas": 10000,
        "pwd": "123456789",
        "name": "TokenA",
        "symbol": "TA",
        "supply": "100",
        "owner": "TESTEYqUbFmNY2aYMrHLpus8Nw4fJ4GSQGXPd5",
        "frozen_height": 0
    }
}
```

#### 2. 支付

*Method:* POST

*请求参数:*
```json
{
    "method": "tokenpay",
    "params": {
        "srcaddress": "TESTEYqUbFmNY2aYMrHLpus8Nw4fJ4GSQGXPd5",
        "address": "TESTGiuKEgeDxvaZ25izhM3JPU7vANXTVFbpA5",
        "gas": 100000,
        "pwd": "123456789",
        "amount": 77,
        "txid": "0a000000000000005572830e7e9de1b72ae6e346cfc9f3e805c66a008afc9915ec8b5549a98942ff",
        "frozen_height": 0
    }
}
```

#### 3. 多个支付

*Method:* POST

*请求参数:*
```json
{
    "method": "tokenpaymore",
    "params": {
        "srcaddress": "TESTEYqUbFmNY2aYMrHLpus8Nw4fJ4GSQGXPd5",
        "addresses": [
            {
                "address": "TEST9YFksCfq9ZTAWG8tWu6dyeppJ7dRJihGA5",
                "amount": 100,
                "frozen_height": 2
            },
            {
                "address": "TEST9YFksCfq9ZTAWG8tWu6dyeppJ7dRJihGA5",
                "amount": 100,
                "frozen_height": 2
            }
        ],
        "gas": 10000,
        "pwd": "123456789",
        "amount": 10,
        "txid": "0a000000000000001f2c7bdf956a68daafadddc7bb4a8bbf9e5319271b9e9f37429b6164c69920d5",
        "frozen_height": 0
    }
}
```

#### 4. 地址详情

*Method:* POST

*请求参数:*
```json
{
    "method": "infoaccount",
    "params": {
        "address": "IMJyxVp3TeHtSDr4jqtYBDgeUzZ21XkgPCt3"
    }
}
```

#### 5. 多账号信息

*Method:* POST

*请求参数:*
```json
{
    "method": "multiaccounts",
    "params": {
        "addresses": [
            "TESTGiuKEgeDxvaZ25izhM3JPU7vANXTVFbpA5",
            "TESTEYqUbFmNY2aYMrHLpus8Nw4fJ4GSQGXPd5",
            "TEST9YFksCfq9ZTAWG8tWu6dyeppJ7dRJihGA5"
        ],
        "token_id": "0a000000000000005572830e7e9de1b72ae6e346cfc9f3e805c66a008afc9915ec8b5549a98942ff"
    }
}
```

#### 6. 代币信息

*Method:* POST

*请求参数:*
```json
{
    "method": "tokeninfo",
    "params": {
        "token_id": "0a00000000000000ce6508bc52e59578ef5fa640a3e1d89d895f4fcafa226627c2cc5d537dcbdd47"
    }
}
```

#### 7. 代币信息列表

*Method:* POST

*请求参数:*
```json
{
    "method": "tokenlist"}
}
```


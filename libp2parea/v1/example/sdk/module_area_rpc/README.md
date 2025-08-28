# RPC接口

## 一. 接口说明

### 1. 接口用例

```curl
curl --request POST --url http://127.0.0.1:2081/rpc --header 'password: testp'  --header 'user: test' --data '{"method":"waitautonomyfinish"}'
```

#### url

`http://127.0.0.1:2081/rpc`

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


## 二. 接口明细

#### 1. 等待网络自治完成

###### 请求

```json
{
    "method": "waitautonomyfinish"
}
```

#### 2. 等待网络自治完成

###### 请求

```json
{
    "method": "waitautonomyfinishvnode"
}
```

#### 3. 等待网络自治完成

###### 请求

```json
{
    "method": "getnetid"
}
```

#### 4. 获取本节点虚拟节点地址

###### 请求

```json
{
    "method": "getvnodeid"
}
```

#### 5. 获取idinfo

###### 请求

```json
{
    "method": "getidinfo"
}
```

#### 6. 获取NodeSelf

###### 请求

```json
{
    "method": "getnodeself"
}
```

#### 7. 关闭所有网络连接

###### 请求

```json
{
    "method": "closenet"
}
```

#### 8. 重新链接网络

###### 请求

```json
{
    "method": "reconnectnet"
}
```

#### 9. 检查是否在线

###### 请求

```json
{
    "method": "checkonline"
}
```

#### 10. 添加一个地址到白名单

###### 请求

```json
{
    "method": "addwhitelist",
    "params": {
        "address": "54oeEHBNB9yAnw6xN8YKTdv2BHrMseyRHEH24gt5Wt7n"
    }
}
```

#### 11. 删除一个地址到白名单

###### 请求

```json
{
    "method": "removewhitelist",
    "params": {
        "address": "54oeEHBNB9yAnw6xN8YKTdv2BHrMseyRHEH24gt5Wt7n"
    }
}
```

#### 12. 添加一个连接

###### 请求

```json
{
    "method": "addconnect",
    "params": {
        "ip": "127.0.0.1",
        "port": 2082
    }
}
```

#### 13. 搜索磁力节点网络地址

###### 请求

```json
{
    "method": "searchnetaddr",
    "params": {
        "address": "54oeEHBNB9yAnw6xN8YKTdv2BHrMseyRHEH24gt5Wt7n"
    }
}
```

#### 14. 搜索磁力虚拟节点网络地址

###### 请求

```json
{
    "method": "searchnetaddrvnode",
    "params": {
        "address": "OMko7LyxGwZFwZ2eTR/mvbnenU+5SUhnt5AvUO45hB8="
    }
}
```

#### 15. 获取所有连接

###### 请求

```json
{
    "method": "networkinfolist"
}
```


#### 16. 发送一个新的广播消息

###### 请求

```json
{
    "method": "sendmulticastmsg",
    "params":{
        "content":"hello world"
    }
}
```

#### 17. 发送一个新的查找超级节点消息

###### 请求

```json
{
    "method": "sendsearchsupermsg",
    "params": {
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello"
    }
}
```

#### 18. 发送一个新的查找超级节点消息

###### 请求

```json
{
    "method": "sendsearchsupermsgwaitrequest",
    "params": {
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello"
    }
}
```

#### 19. 发送一个新消息

###### 请求

```json
{
    "method": "sendp2pmsg",
    "params": {
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello"
    }
}
```

#### 20. 给指定节点发送一个消息

###### 请求

```json
{
    "method": "sendp2pmsgwaitrequest",
    "params": {
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello world"
    }
}
```

#### 21. 发送一个加密消息，包括消息头也加密

###### 请求

```json
{
    "method": "sendp2pmsghe",
    "params": {
        "recv_address": "7DcSuUeVWGLGr4RG2zVS7iR8rS4uqSeHtouJh955zEAz",
        "content": "hello world"
    }
}
```

#### 22. 发送一个加密消息，包括消息头也加密

###### 请求

```json
{
    "method": "sendp2pmsghewaitrequest",
    "params": {
        "recv_address": "7DcSuUeVWGLGr4RG2zVS7iR8rS4uqSeHtouJh955zEAz",
        "content": "hello world"
    }
}
```

#### 21. 网络中查询一个逻辑节点地址的真实地址

###### 请求

```json
{
    "method": "searchvnodeid",
    "params": {
        "vnode_address": "7DcSuUeVWGLGr4RG2zVS7iR8rS4uqSeHtouJh955zEAz"
    }
}
```

#### 22. 发送一个新的查找超级节点消息，可以指定接收端和发送端的代理节点

###### 请求

```json
{
    "method": "sendsearchsupermsgproxy",
    "params": {
        "sender_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello world"
    }
}
```

#### 23. 发送一个新的查找超级节点消息，可以指定接收端和发送端的代理节点

###### 请求

```json
{
    "method": "sendsearchsupermsgproxywaitrequest",
    "params": {
        "sender_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello world"
    }
}
```

#### 24. 发送一个新消息，可以指定接收端和发送端的代理节点

###### 请求

```json
{
    "method": "sendp2pmsgproxy",
    "params": {
        "sender_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello world"
    }
}
```

#### 25. 给指定节点发送一个消息，可以指定接收端和发送端的代理节点

###### 请求

```json
{
    "method": "sendp2pmsgproxywaitrequest",
    "params": {
        "sender_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello world"
    }
}
```

#### 26. 发送一个加密消息，包括消息头也加密，可以指定接收端和发送端的代理节点

###### 请求

```json
{
    "method": "sendp2pmsgheproxy",
    "params": {
        "sender_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello world"
    }
}
```

#### 27. 发送一个加密消息，包括消息头也加密，可以指定接收端和发送端的代理节点

###### 请求

```json
{
    "method": "sendp2pmsgheproxywaitrequest",
    "params": {
        "sender_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "recv_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "content": "hello world"
    }
}
```

#### 28. 根据目标ip地址及端口添加到白名单

###### 请求

```json
{
    "method": "addaddrwhitelist",
    "params": {
        "ip": "127.0.0.1",
        "port": "19990"
    }
}
```

#### 29. 设置区域上帝地址信息

###### 请求

```json
{
    "method": "setareagodaddr",
    "params": {
        "ip": "127.0.0.1",
        "port": "19990"
    }
}
```

#### 30. 搜索磁力节点网络地址，可以指定接收端和发送端的代理节点

###### 请求

```json
{
    "method": "searchnetaddrproxy",
    "params": {
        "address": "OMko7LyxGwZFwZ2eTR/mvbnenU+5SUhnt5AvUO45hB8=",
        "recv_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm",
        "sender_proxy_address": "FtZdqgsQst2YFqQyMC7qF2WmtUuWiMxoFYofAv2sCWNm"
    }
}
```

#### 31. 根据目标节点，返回排序后的虚拟节点地址列表

###### 请求

```json
{
    "method": "findnearvnodessearchvnode",
    "params": {
        "address": "OMko7LyxGwZFwZ2eTR/mvbnenU+5SUhnt5AvUO45hB8=",
        "include_self": true,
        "include_index0": true
    }
}
```

#### 32. 得到所有连接的节点信息，不包括本节点

###### 请求

```json
{
    "method": "getallnodes"
}
```

#### 32. 获取设备机器Id

###### 请求

```json
{
    "method": "getmachineid"
}
```

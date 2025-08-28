
一、将项目编译成可执行程序
1.项目main函数文件在example/peer_root/firstPeer.go
2.在peer_root目录下执行：go build -o peer_root.exe -ldflags "-w -s" -buildmode=exe
3.上一步编译成功后生成peer_root.exe可执行文件。


二、创始节点部署
1.创建一个新文件夹peer1，将peer_root.exe文件复制到peer1文件夹中。
2.在peer1目录中新建文件夹conf，用于存放配置文件。
3.进入conf文件夹，创建文本文件：config.json
4.config.json文件内容复制如下：
{
"AreaName":"test",
"ip":"127.0.0.1",
"port":39981,
"WebAddr":"0.0.0.0",
"WebPort":3081,
"WebStatic":"./static",
"WebViews":"./views",
"RpcServer":true,
"RpcUser":"test",
"RpcPassword":"testp",
"miner":true,
"NetType":"not release",
"AddrPre":"TEST",
"end":0
}
5.配置文件参数说明：
AreaName：域网络名称，多节点互相连接时，名称要设置为相同，不相同则不能联网。
ip：本地地址。当部署到外网服务器时，这里要配置为公网IP，这样才能被访问。
port：p2p网络端口，给其他节点连接用。
WebAddr：rpc连接用ip。
WebPort：rpc连接用端口。
RpcUser：rpc调用用户名
RpcPassword：rpc调用密码
NetType：部署到外网时，所有节点使用release。部署到内网时，所有节点使用not release。
AddrPre：链端收款地址前缀。
6.单节点启动命令：./peer_root.exe init
7.带init参数是创始节点启动，请注意，会覆盖之前的数据。
8.会在conf目录中生成一个keystore.key文件，这是节点密钥文件，保存节点地址身份。
9.peer1文件夹中会生成wallet文件夹，保存着区块数据，此节点作为见证人节点，负责出块。
10.这时候单节点就启动完成了，可以使用postman调用节点的rpc试试。


三、RPC调用
1.打开postman。
2.新建一个POST请求：127.0.0.1:3081/rpc
3.Headers中添加rpc用户名和密码键值对：
user:test
password：testp
4.Body中数据格式选择row，输入参数：{"method":"getinfo"}
5.点击send，开始rpc请求。


四、双节点部署
1.完成以上“创始节点部署”步骤后，保持创始节点运行状态。
2.将创始节点peer1文件夹复制一份改名为peer2，删除peer2中的wallet文件夹和conf中keystore.key文件。
3.在peer2/conf文件夹中创建文本文件nodeEntry.json，复制一下内容：
{"127.0.0.1:39981":"","127.0.0.1:39982":""}
4.配置文件中地址是创始节点地址及端口，当此节点启动后，会去连接这个地址，实现区块同步。
5.此文件中可以配置多个节点地址，对应节点此时无法连接不要紧，只要其中一个地址能连接上就行。
6.启动节点：./peer_root.exe
7.从节点启动不带任何参数，peer2文件夹中会生成wallet文件夹，生成conf/keystore.key文件。
8.此节点作为从节点，不出块，随时关闭不影响网络。要配置多个节点，可以按照此步骤再做一次。
备注：可以带密码参数启动，设置自定义密码  -walletpwd=123456789

六、节点拉起
1.此操作条件为，全网所有节点都关闭状态下，希望重启节点，并且延用之前的区块数据。
2.选择一个能被其他节点连接的见证人节点，正式网络一定是公网服务器上的节点。
3.一般选择见证人节点作为拉起节点。
4.使用命令：./peer_root.exe load
5.以上命令启动节点，其中参数load为拉起一个节点。
6.从节点中，nodeEntry.json文件配置中要有已经在运行的节点。
7.节点被拉起后，其他节点无论是否见证人，启动命令都不带任何参数。



五、节点热升级
1.此操作条件为，多节点网络中单个节点重启升级。
2.保留一个及以上能被访问的节点运行状态。
3.启动之前，检查nodeEntry.json文件中地址包含一直运行，并且能连接上的节点地址。
4.无论待重启的节点是创始节点还是从节点，或是见证人节点，启动命令都一样。
5.启动命令：./peer_root.exe
6.以上命令启动节点，不带任何参数。



六、节点清理数据，并重启。
1.关闭所有节点。
2.删除所有节点中wallet文件夹，此文件夹中保存区块数据。
3.可以保留conf中keystore.key文件。
4.按之前的步骤启动网络。


七、节点启动命令简介：
1.创始节点启动加init参数，会覆盖本节点wallet中区块数据，从1高度重新出块。
2.所有节点都关闭状态，想要继续使用之前的数据，选择一个节点使用load参数启动，其他节点不使用任何参数启动。
3.节点重启，无论什么节点，都不带任何参数启动。




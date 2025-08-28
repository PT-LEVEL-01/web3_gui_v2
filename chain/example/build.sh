#!/bin/bash
# 上面中的 #! 是一种约定标记, 它可以告诉系统这个脚本需要什么样的解释器来执行;

go build peer_root/firstPeer.go

/bin/cp -rf firstPeer /store/polarcloud/peer_init/peer_init

/bin/rm -rf /store/polarcloud/peer_init/nohup.out
/bin/rm -rf /store/polarcloud/peer_init/mem.prof
/bin/rm -rf /store/polarcloud/peer_init/wallet
/bin/rm -rf /store/polarcloud/peer_init/store

/bin/cp -rf firstPeer /store/polarcloud/peer_1/peer1

/bin/rm -rf /store/polarcloud/peer_1/nohup.out
/bin/rm -rf /store/polarcloud/peer_1/mem.prof
/bin/rm -rf /store/polarcloud/peer_1/wallet
/bin/rm -rf /store/polarcloud/peer_1/store
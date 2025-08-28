package mining

import (
	"encoding/hex"
	"sync"
)

var TempBlockAck *BlockAck

var Once_BlockAck sync.Once

type BlockAck struct {
	mu sync.Mutex
	mm map[string]int
}

func init() {
	Once_BlockAck.Do(func() {
		TempBlockAck = &BlockAck{sync.Mutex{}, map[string]int{}}
	})
}

/*
*
验证ack是否满足条件
2/3F+1
*/
// func (b *BlockAck) CheckAck(hash *[]byte) bool {
// 	whitelistNodes := Area.NodeManager.GetWhiltListNodes()
// 	return b.GetBlockAck(hash) >= int(checkSignRatio*(float64(len(whitelistNodes))+1))
// }

/*
*
删除ack缓存
*/
func (b *BlockAck) DelBlockAck(hash *[]byte) {
	b.mu.Lock()
	delete(b.mm, hex.EncodeToString(*hash))
	b.mu.Unlock()
}

/*
*
设置ack缓存
*/
func (b *BlockAck) SetBlockAck(hash *[]byte) {
	b.mu.Lock()

	////删掉之前高度的缓存
	//b.DelBlockAck(height - 1)

	value, _ := b.mm[hex.EncodeToString(*hash)]
	value++
	b.mm[hex.EncodeToString(*hash)] = value
	b.mu.Unlock()
}

/*
*
读取ack缓存
*/
func (b *BlockAck) GetBlockAck(hash *[]byte) int {
	b.mu.Lock()
	value, _ := b.mm[hex.EncodeToString(*hash)]
	b.mu.Unlock()
	return value
}

/*
*
请求block ack
*/
// func ReqBlockAck(hash *[]byte, timeout time.Duration) bool {
// 	hashStr := hex.EncodeToString(*hash)
// 	engine.Log.Info("去确认ack,hash:%s", hashStr)

// 	cs := make(chan bool, config.CPUNUM)
// 	group := new(sync.WaitGroup)

// 	whitelistNodes := Area.NodeManager.GetWhiltListNodes()
// 	for _, v := range whitelistNodes {
// 		cs <- false
// 		group.Add(1)
// 		utils.Go(func() {
// 			res, _, _, err := Area.SendP2pMsgWaitRequest(config.MSGID_p2p_block_ack, &v, hash, timeout)
// 			if err == nil && res != nil {
// 				TempBlockAck.SetBlockAck(hash)
// 			}

// 			<-cs
// 			group.Done()
// 		})
// 	}
// 	group.Wait()

// 	engine.Log.Info("hash：%s，是否满足条件：%t", hashStr, TempBlockAck.CheckAck(hash))
// 	return TempBlockAck.CheckAck(hash)
// }

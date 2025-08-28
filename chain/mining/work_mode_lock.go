package mining

import (
	"sync"

	"web3_gui/libp2parea/adapter/engine"
)

// 全局工作模式锁
var workModeLockStatic = NewWorkModeLock()

/*
节点工作模式锁
签名、导入区块、出块，三种工作模式之间切换。
*/
type WorkModeLock struct {
	lock            *sync.RWMutex //
	signGroupHeight uint64        //已签名组高度
	signBlockHeight uint64        //已签名块高度
}

func NewWorkModeLock() *WorkModeLock {
	return &WorkModeLock{
		lock: new(sync.RWMutex), //
	}
}

/*
带锁的签名
*/
func (this *WorkModeLock) CheckBlockSignLock(bhvo *BlockHeadVO, chain *Chain) bool {
	this.lock.Lock()
	err := this.CheckBlockSign(bhvo, chain)
	this.lock.Unlock()
	return err
}

/*
判断是否应该给某个区块签名
小于当前组高度或小于当前区块高度，不给签名
同一组高度和同一区块高度只给签名一次
签名高度必须连续
@return    bool    true=签名成功;false=签名失败;
*/
func (this *WorkModeLock) CheckBlockSign(bhvo *BlockHeadVO, chain *Chain) bool {
	// this.lock.Lock()
	// defer this.lock.Unlock()
	//在见证人组中找到这个见证人
	currentWitness, _ := chain.WitnessChain.FindWitnessForBlockOnly(bhvo)
	if currentWitness == nil {
		//找不到这个见证人
		engine.Log.Info("CheckSign 找不到这个见证人 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		return false
	}
	//已经签过名了
	if currentWitness.SignExist {
		engine.Log.Info("CheckSign 已经签过名了 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		return false
	}
	// importGroupHeight := chain.GetCurrentGroupHeight()
	importBlockHeight := chain.GetCurrentBlock()

	//高度不能小于自己已经统计的高度
	if bhvo.BH.Height <= importBlockHeight {
		engine.Log.Info("CheckSign 高度太小 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		return false
	}

	//查找前置区块是否存在，并且合法
	preWitness := chain.WitnessChain.FindPreWitnessForBlock(bhvo.BH.Previousblockhash)
	if preWitness == nil || preWitness.Block == nil {
		engine.Log.Info("CheckSign 查找前置区块失败 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		return false
	}
	//高度必须连续
	if preWitness.Block.Height+1 != bhvo.BH.Height {
		engine.Log.Info("CheckSign 高度不连续 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		return false
	}
	//
	//判断是否相同组，相同组内不能有相同高度的区块
	if currentWitness == preWitness {
		for _, one := range currentWitness.Group.Witness {
			if one.Block == nil {
				continue
			}
			if one.Block.Height == bhvo.BH.Height {
				engine.Log.Info("CheckSign 相同组内不能有相同高度的区块 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
				return false
			}
		}
	}
	// else {
	// //不同组也不能有相同高度的区块
	// //group:1195 block:2499 prehash:66d738aa4ccaef4f2eba63f05fc2c32518689b05c71a8889b306c1fcbfaeeab9 hash:77834994f87a964c8c0728968c04ec084339b7013e5f08870dc9df523f0aba3f
	// //group:1195 block:2500 prehash:77834994f87a964c8c0728968c04ec084339b7013e5f08870dc9df523f0aba3f hash:ce1c98913b7d219af20413c4f80168b837fb540d81f5bddbb3ad2d92babe218f
	// //group:1195 block:2501 prehash:ce1c98913b7d219af20413c4f80168b837fb540d81f5bddbb3ad2d92babe218f hash:26f78fdd73056df6596db487174cd9203d979bca600ca7ef05b9a5f1dc6fb114
	// //group:1196 block:2501 prehash:ce1c98913b7d219af20413c4f80168b837fb540d81f5bddbb3ad2d92babe218f hash:b6179c2fcd8b18e1eb6d0e3a1ea2201290ebe8f30776258802bb8e53ecee4500
	// for _, one := range preWitness.Group.Witness {
	// 	if one.Block == nil {
	// 		continue
	// 	}
	// 	if one.Block.Height == bhvo.BH.Height {
	// 		engine.Log.Info("CheckSign 不同组内不能有相同高度的区块 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// 		return false
	// 	}
	// }
	// }
	return true

	// //不同组，则看上一组是否合法的多人出块，如果是合法组，则不能有相同高度，必须高度加1
	// ok, group := preWitness.Group.CheckBlockGroup(nil)
	// if !ok && group != nil {
	// }

	// //判断本组人数
	// // if currentWitness

	// //判断这个组是否多人出块
	// ok, group := currentWitness.Group.CheckBlockGroup(nil)
	// //判断之前的组是否多人出块

	// tempGroupHeight := this.importGroupHeight
	// if this.signGroupHeight > this.importGroupHeight {
	// 	tempGroupHeight = this.signGroupHeight
	// }
	// tempBlockHeight := this.importBlockHeight
	// if this.signBlockHeight > this.importBlockHeight {
	// 	tempBlockHeight = this.signBlockHeight
	// }

	// //一个组只有一个块，未确认，下一组还是出相同高度的区块的情况。
	// if blockHeight == this.signBlockHeight && groupHeight+1 > tempGroupHeight {
	// 	engine.Log.Info("CheckSign 00000 %d %d", tempGroupHeight, tempBlockHeight)

	// 	return true
	// }
	// if tempBlockHeight+1 != blockHeight {
	// 	engine.Log.Info("CheckSign 11111 %d %d", tempBlockHeight, blockHeight)

	// 	return false
	// }
	// if groupHeight < tempGroupHeight {
	// 	engine.Log.Info("CheckSign 22222 %d %d", groupHeight, tempGroupHeight)

	// 	return false
	// }
	// engine.Log.Info("CheckSign 00000 %d %d", tempGroupHeight, tempBlockHeight)
	// this.signGroupHeight = groupHeight
	// this.signBlockHeight = blockHeight
	// return true
}

/*
获取锁
*/
func (this *WorkModeLock) GetLock() {
	this.lock.Lock()
}

/*
归还锁
*/
func (this *WorkModeLock) BackLock() {
	this.lock.Unlock()
}

// /*
// 获取出块锁
// */
// func (this *WorkModeLock) GetPackageLock() {
// 	this.lock.Lock()
// }

// /*
// 归还出块锁
// */
// func (this *WorkModeLock) BackPackageLock() {
// 	this.lock.Unlock()
// }

// /*
// 获取导入区块锁
// */
// func (this *WorkModeLock) GetImportBlockLock() {
// 	this.lock.Lock()
// }

// /*
// 归还导入区块锁成功
// */
// func (this *WorkModeLock) BackImportBlockLockSuccess(groupHeight, blockHeight uint64) {
// 	// this.importGroupHeight = groupHeight
// 	// this.importBlockHeight = blockHeight
// 	this.lock.Unlock()
// }

// /*
// 归还导入区块锁失败
// */
// func (this *WorkModeLock) BackImportBlockLockFail() {
// 	this.lock.Unlock()
// }

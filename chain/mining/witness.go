package mining

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/evm"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
见证人链
投票竞选出来的见证人加入链中，但是没有加入组
当交了押金后，被分配到见证人组
*/
type WitnessChain struct {
	chain           *Chain             //所属链
	witnessBackup   *WitnessBackup     //
	WitnessGroup    *WitnessSmallGroup //当前正在出块的见证人组
	witnessNotGroup []*Witness         //未分配组的见证人列表
	// firstWitnessNotGroup *Witness       //首个未分配组的见证人引用
	// lastWitnessNotGroup  *Witness       //最后一个未分配组的见证人引用
}

func NewWitnessChain(wb *WitnessBackup, chain *Chain) *WitnessChain {
	return &WitnessChain{
		chain:         chain,
		witnessBackup: wb,
	}
}

/*
见证人组
*/
type WitnessSmallGroup struct {
	Task         bool               //是否已经定时出块
	PreGroup     *WitnessSmallGroup //上一个组
	NextGroup    *WitnessSmallGroup //下一个组
	Height       uint64             //见证人组高度
	Witness      []*Witness         //本组见证人列表
	BlockGroup   *Group             //构建出来的合法区块组，为空则是没有合法组
	IsBuildGroup bool               //是否已经构建这个组
	tag          bool               //备用见证人队列标记，一次从备用见证人评选出来的备用见证人为一个队列标记，需要保证备用见证人组中有两个队列
	IsCount      bool               //是否已经统计本组合法区块中的交易。
	// check      bool          //这个见证人组是否多半人出块，多半人出块则合法有效
	CollationLock *sync.Mutex //整理关系锁
}

/*
见证人
*/
type Witness struct {
	Group           *WitnessSmallGroup  //属于哪个出块小组
	PreWitness      *Witness            //上一个见证人
	NextWitness     *Witness            //下一个见证人
	Addr            *crypto.AddressCoin //见证人地址
	Puk             []byte              //见证人公钥
	Block           *Block              //见证人生产的块
	Score           uint64              //见证人自己的押金
	CommunityVotes  []*VoteScore        //社区节点投票
	Votes           []*VoteScore        //轻节点投票和押金
	VoteNum         uint64              //投票数量
	StopMining      chan bool           //停止出块命令
	BlockHeight     uint64              //预计出块高度
	CreateBlockTime int64               //预计出块时间
	WitnessBigGroup *WitnessBigGroup    //候选见证人组，属于哪个大组
	CheckIsMining   bool                //检查是否已经验证出块标记，用于多次未出块的见证人，踢出列表
	syncBlockOnce   *sync.Once          //定时同步区块，只执行一次
	SignExist       bool                //是否已经给这个见证人签名了
	createTime      time.Time           //
	sleepTime       int64               //等待时间
}

/*
构建指定高度组之前的所有区块组，包括指定高度
@groupHeight    uint64    组高度
*/
func (this *WitnessChain) BuildBlockGroupForGroupHeight(groupHeight uint64, blockhash *[]byte) {
	BuildWitnessGroupLock.Lock()
	defer BuildWitnessGroupLock.Unlock()
	// engine.Log.Info("Current outgoing block group before building group %d", this.witnessGroup.Height)
	//找到参数指定组高度
	currentGroup := this.WitnessGroup
	for currentGroup != nil {
		if currentGroup.Height < groupHeight {
			if currentGroup.NextGroup == nil {
				break
			}
			currentGroup = currentGroup.NextGroup
			continue
		}
		if currentGroup.Height > groupHeight {
			if currentGroup.PreGroup == nil {
				return
			}
			currentGroup = currentGroup.PreGroup
			continue
		}
		if currentGroup.Height == groupHeight {
			break
		}
	}
	if currentGroup.Height >= groupHeight {
		return
	}
	have := false
	for currentGroup.Height < groupHeight {
		have = true
		currentGroup = this.AdditionalWitnessBackup()
	}
	if have {
		this.PrintWitnessList()
	}
	return

	// engine.Log.Info("找到的组高度 %d", currentGroup.Height)
	// engine.Log.Info("Group height found %d", currentGroup.Height)
	//从当前出块的组开始统计，一直统计到跳过之后的组
	for {
		// engine.Log.Info("当前见证人组高度 %d %d", this.witnessGroup.Height, currentGroup.Height)
		// engine.Log.Info("Current witness group height %d %d", this.witnessGroup.Height, currentGroup.Height)
		if (this.WitnessGroup.Height > currentGroup.Height) || (this.WitnessGroup.NextGroup == nil) {
			break
		}
		// if this.witnessGroup.BlockGroup != nil {
		// 	lastBlocks := this.witnessGroup.BlockGroup.Blocks
		// 	if lastBlocks[len(lastBlocks)-1].Height > this.chain.GetCurrentBlock()+1 {
		// 		this.witnessGroup.BuildGroup(blockhash)
		// 		this.chain.CountBlock(this.witnessGroup)
		// 	}
		// }
		this.BuildWitnessGroup(false, true)
		// this.witnessGroup.BuildGroup(blockhash)
		// this.chain.CountBlock(this.witnessGroup)
		// engine.Log.Info("移动高度3333333333333333 %d", this.witnessGroup.Height)
		this.WitnessGroup = this.WitnessGroup.NextGroup
		// if this.witnessGroup.Height >= currentGroup.Height {
		// 	break
		// }
	}
	// engine.Log.Info("构建组之后的当前出块组 %d", this.witnessGroup.Height)
	// engine.Log.Info("Current outgoing block group after building group %d", this.witnessGroup.Height)
}

/*
将见证人组中出的块构建为区块组
*/
func (this *WitnessChain) BuildBlockGroup(bhvo *BlockHeadVO, preWitness *Witness) {
	// engine.Log.Info("Height of block group constructed this time %d", bhvo.BH.GroupHeight)

	// start := config.TimeNow()

	if bhvo.BH.GroupHeight == config.Mining_group_start_height {
		// engine.Log.Info("11111 witnessGroup")
		this.WitnessGroup.BuildGroup(nil)
		this.BuildWitnessGroup(false, false)
		this.WitnessGroup = this.WitnessGroup.NextGroup
		this.BuildWitnessGroup(false, true)
		return
	}
	// engine.Log.Info("构建区块组 111 耗时 %s", config.TimeNow().Sub(start))
	if preWitness == nil {
		// engine.Log.Info("not find pre witness!")
		return
	}
	//查询当前出块的见证人
	witness, _ := this.FindWitnessForBlock(bhvo)
	if witness == nil {
		return
	}

	//---新版本统计-----------------
	// engine.Log.Info("构建区块组 222 耗时 %s", config.TimeNow().Sub(start))
	isNewGroup := false
	//是新的组出块，则构建前面的组
	if bhvo.BH.GroupHeight != preWitness.Group.Height {
		isNewGroup = true
		// if bhvo.BH.Height < config.FixBuildGroupBUGHeightMax {
		// 	preWitness.Group.BuildGroup(&bhvo.BH.Previousblockhash)
		// 	wg := preWitness.Group.BuildGroup(&bhvo.BH.Previousblockhash)
		// 	//找到上一组到本组的见证人组，开始查找没有出块的见证人
		// 	if wg != nil {
		// 		for _, one := range wg.Witness {
		// 			if !one.CheckIsMining {
		// 				if one.Block == nil {
		// 					this.chain.WitnessBackup.AddBlackList(*one.Addr)
		// 				} else {
		// 					this.chain.WitnessBackup.SubBlackList(*one.Addr)
		// 				}
		// 				one.CheckIsMining = true
		// 			}
		// 		}
		// 	}
		// 	this.chain.CountBlock(preWitness.Group)
		// }
		// engine.Log.Info("构建区块组 555 耗时 %s", config.TimeNow().Sub(start))
		if bhvo.BH.GroupHeight != config.Mining_group_start_height+1 {
			// engine.Log.Info("222222222222")
			this.WitnessGroup = witness.Group
		}
	}
	this.BuildWitnessGroup(false, isNewGroup)

	//----下面的代码是以前的版本-------------------------------
	// isNewGroup := false
	// //是新的组出块，则构建前面的组
	// for i := this.witnessGroup.Height; i < bhvo.BH.GroupHeight; i++ {
	// 	isNewGroup = true
	// 	// engine.Log.Info("本次构建的区块组高度   222222222222222 %d", i)
	// 	// engine.Log.Info("构建区块组 333 耗时 %s", config.TimeNow().Sub(start))

	// 	wg := this.witnessGroup.BuildGroup(&bhvo.BH.Previousblockhash)

	// 	//找到上一组到本组的见证人组，开始查找没有出块的见证人
	// 	if wg != nil {
	// 		for tempGroup := wg; tempGroup != nil && tempGroup.Height < i; tempGroup = tempGroup.NextGroup {
	// 			for _, one := range wg.Witness {
	// 				if !one.CheckIsMining {
	// 					if one.Block == nil {
	// 						this.chain.witnessBackup.AddBlackList(*one.Addr)
	// 					} else {
	// 						this.chain.witnessBackup.SubBlackList(*one.Addr)
	// 					}
	// 					one.CheckIsMining = true
	// 				}
	// 			}
	// 		}
	// 	}
	// 	// engine.Log.Info("构建区块组 444 耗时 %s", config.TimeNow().Sub(start))

	// 	this.chain.CountBlock()

	// 	// engine.Log.Info("构建区块组 555 耗时 %s", config.TimeNow().Sub(start))

	// 	if bhvo.BH.GroupHeight != config.Mining_group_start_height+1 {
	// 		// engine.Log.Info("本次构建的区块组高度   3333333333333333 %d", i)
	// 		this.witnessGroup = this.witnessGroup.NextGroup
	// 	}
	// 	//构建一个新的见证人组
	// 	// engine.Log.Info("构建区块组 666 耗时 %s", config.TimeNow().Sub(start))
	// }
	// this.BuildWitnessGroup(false, isNewGroup)

}

/*
则构建这一组见证人的所有出块，保存关系
@return    *WitnessSmallGroup    指向的前一个小组
*/
func (this *WitnessSmallGroup) BuildGroup(blockhash *[]byte) *WitnessSmallGroup {
	// fmt.Println("统计并修改组", this.Height)
	if blockhash == nil || len(*blockhash) <= 0 {
		engine.Log.Info("count group:%d", this.Height)
	} else {
		engine.Log.Info("count group:%d preblockhash:%s", this.Height, hex.EncodeToString(*blockhash))
	}

	//已经构建了则退出，避免重复构建浪费计算资源
	if this.IsBuildGroup {
		// engine.Log.Info("Already built.")
		return nil
	}

	//判断这个组是否多人出块
	ok, group := this.CheckBlockGroup(blockhash)
	if !ok {
		engine.Log.Info("The number of people in this group is too small and unqualified. group:%d", this.Height)
		// this.IsBuildGroup = true
		return nil
	}

	// str := ""
	// for _, one := range group.Blocks {
	// 	str = str + " " + strconv.Itoa(int(one.Height))
	// }
	// engine.Log.Info("本组出块数量:%d group:%d 包含:%s", len(group.Blocks), this.Height, str)
	// for _, one := range group.Blocks {
	// 	engine.Log.Info("出块组信息：组高度 %d 块高度 %d", one.Group.Height, one.Height)
	// }

	// engine.Log.Info("this.BlockGroup:%v", group)
	this.BlockGroup = group
	this.IsBuildGroup = true
	// for _, one := range group.Blocks {
	// 	engine.Log.Info("构建成功的块one :%d %s", one.Height, hex.EncodeToString(one.Id))
	// }

	// fmt.Println("BuildGroup  1111111111111")

	//找到上一组
	beforeGroup := this.PreGroup
	for beforeGroup = this.PreGroup; beforeGroup != nil; beforeGroup = beforeGroup.PreGroup {
		ok, _ = beforeGroup.CheckBlockGroup(blockhash)
		if ok {
			break
		}
	}

	if beforeGroup == nil {
		engine.Log.Info("The last group found is empty")
		return nil
	}

	// engine.Log.Info("确定本组被确认 ======================== %d", this.Height)
	//修改组引用
	beforeGroup.BlockGroup.NextGroup = this.BlockGroup
	this.BlockGroup.PreGroup = beforeGroup.BlockGroup

	//修改块引用
	beforeBlock := beforeGroup.BlockGroup.Blocks[len(beforeGroup.BlockGroup.Blocks)-1]
	//这里决定了保存到数据库中的链是已经确认的链，没有保存分叉
	beforeBlock.NextBlock = this.BlockGroup.Blocks[0]
	// engine.Log.Info("222 update block nextblockhash,this height:%d blockid:%s nextid:%s", beforeBlock.Height, hex.EncodeToString(beforeBlock.Id), hex.EncodeToString(this.BlockGroup.Blocks[0].Id))
	// beforeBlock.FlashNextblockhash()
	return beforeGroup
}

/*
判断是否多半人出块，只判断，不作修改和保存
@return    bool    是否多人出块
@return    *Group  评选出来的多人出块组
*/
func (this *WitnessSmallGroup) CheckBlockGroup(blockhash *[]byte) (bool, *Group) {
	// engine.Log.Info("统计本组合法的区块组:%d %v", this.Height, blockhash)

	//已经统计过了，就不需要再统计了，直接返回
	if this.BlockGroup != nil {
		// engine.Log.Info("统计过了，就不需要再统计了:%d", len(this.BlockGroup.Blocks))
		return true, this.BlockGroup
	}

	// engine.Log.Info("CheckBlockGroup SelectionChain")
	group := this.SelectionChain(blockhash)
	//出块数量为0
	if group == nil {
		// engine.Log.Info("评选出来的组为空")
		return false, nil
	}
	totalWitness := len(this.Witness)
	totalHave := len(group.Blocks)
	if config.ConfirmMinNumber(totalWitness) > totalHave {
		return false, group
	}
	//包含只有两个见证人的情况，一人出块既是成功
	// if (totalHave * 2) <= totalWitness {
	// 	// engine.Log.Info("出块人太少 不合格 %d %d", totalHave, totalWitness)
	// 	return false, group
	// }
	// fmt.Println("出块人多数 合格")
	// engine.Log.Info("groupheight:%d 出块人多数 %d 合格", this.Height, totalHave)
	return true, group
}

/*
选出这个见证人组中出块最多块的链
*/
func (this *WitnessSmallGroup) SelectionChain(blockhash *[]byte) *Group {
	this.CollationLock.Lock()
	defer this.CollationLock.Unlock()
	// if blockhash == nil || len(*blockhash) == 0 {
	// 	engine.Log.Info("整理本组区块前后关系:%d", this.Height)
	// } else {
	// 	engine.Log.Info("整理本组区块前后关系:%d hash:%s", this.Height, hex.EncodeToString(*blockhash))
	// }
	this.CollationRelationship(blockhash)
	//开始评选出块最多的组
	groupMap := make(map[string]*Group) //key:string=每个组的最后一个快hash;value:*Group=分叉组;
	for _, one := range this.Witness {
		if one.Block == nil {
			continue
		}
		if one.Block.Group == nil {
			// engine.Log.Info("开始评选出块最多的组 这里退出了")
			continue
		}
		blocks := one.Block.Group.Blocks
		lastBlock := blocks[len(blocks)-1]
		if blockhash != nil && len(*blockhash) > 0 && !bytes.Equal(lastBlock.Id, *blockhash) {
			continue
		}
		groupMap[utils.Bytes2string(lastBlock.Id)] = one.Block.Group
	}
	//评选出块多的组
	var group *Group            //这是无后接区块评选出来的组，两组相同，取前组
	var groupByBlockhash *Group //这是根据后接区块hash查找到的组，只有一个结果，查找不到则为空。
	for _, v := range groupMap {
		if blockhash != nil && groupByBlockhash == nil {
			for _, blockOne := range v.Blocks {
				if bytes.Equal(blockOne.Id, *blockhash) {
					// engine.Log.Info("找到这个块了:%s %d", hex.EncodeToString(blockOne.Id), len(blockOne.Group.Blocks))
					groupByBlockhash = v
					break
				}
			}
		}
		if group == nil {
			group = v
			continue
		}
		if len(v.Blocks) > len(group.Blocks) {
			group = v
		}
	}
	if blockhash != nil {
		group = groupByBlockhash
	}
	if group == nil || group.Blocks == nil || len(group.Blocks) <= 0 {
		return group
	}

	// str := ""
	// for _, one := range group.Blocks {
	// 	str = str + " " + strconv.Itoa(int(one.Height))
	// }
	// engine.Log.Info("评选出块多的组:%d group:%d 包含:%s", len(group.Blocks), this.Height, str)

	//给组内的区块通过修改NextBlock，从新建立关系
	for _, lastBlock := range group.Blocks {
		for _, preBlock := range group.Blocks {
			if bytes.Equal(lastBlock.PreBlockID, preBlock.Id) {
				preBlock.NextBlock = lastBlock
				lastBlock.PreBlock = preBlock
				break
			}
		}
	}

	//找到本大组第一个区块和上一个大组最后一个区块，修改NextBlock，把两个大组连接起来
	firstBlock := group.Blocks[0]
	for preWitnessGroup := this.PreGroup; preWitnessGroup != nil; preWitnessGroup = preWitnessGroup.PreGroup {
		have := false
		for _, witnessOne := range preWitnessGroup.Witness {
			if witnessOne.Block == nil {
				continue
			}
			if bytes.Equal(witnessOne.Block.Id, firstBlock.PreBlockID) {
				witnessOne.Block.NextBlock = firstBlock
				firstBlock.PreBlock = witnessOne.Block
				have = true
				break
			}
		}
		if have {
			break
		}
	}
	return group
}

/*
清理删除本组中的组关系
*/
func (this *WitnessSmallGroup) CleanRelationship() {
	for _, one := range this.Witness {
		if one.Block == nil {
			continue
		}
		one.Block.Group = nil
	}
}

/*
查找本组中已出区块之间的前后关系，从后往前查找，给他们创建出块组，用于统计出块最多的组
分叉处理：当参数为空时，则随意选择一组。用作出块时[选择]前置区块
分叉处理：找到参数指定的区块，创建分组。用作导入区块[寻找]前置区块
@blockhash    *[]byte    从指定区块hash开始，从后往前建立关系
*/
func (this *WitnessSmallGroup) CollationRelationship(blockhash *[]byte) {
	if blockhash == nil || len(*blockhash) == 0 {
		// engine.Log.Info("整理关系 11111")
		for i := len(this.Witness); i > 0; i-- {
			witness := this.Witness[i-1]
			if witness.Block == nil {
				continue
			}
			newBlock := witness.Block
			//找到连续的下一个块的见证人
			for witnessTemp := witness.NextWitness; witnessTemp != nil; witnessTemp = witnessTemp.NextWitness {
				if witnessTemp.Group.Height != witness.Group.Height {
					//未找到同组其他区块，自己创建一个组
					newGroup := new(Group)
					newGroup.Height = witness.Group.Height
					newGroup.Blocks = []*Block{newBlock}
					newBlock.Group = newGroup
					break
				}
				if witnessTemp.Block == nil {
					continue
				}
				if bytes.Equal(witnessTemp.Block.PreBlockID, newBlock.Id) {
					//找到这个见证人了，在组中添加这个区块
					// nextWitness = witnessTemp
					//从后往前查找，则把本块放在出块组的前面
					witnessTemp.Block.Group.Blocks = append([]*Block{newBlock}, witnessTemp.Block.Group.Blocks...)
					witness.Block.Group = witnessTemp.Block.Group
					newBlock.NextBlock = witnessTemp.Block
					// engine.Log.Info("333 update block nextblockhash,this %d blockid:%s nextid:%s", newBlock.Height,
					// 	hex.EncodeToString(newBlock.Id), hex.EncodeToString(witnessTemp.Block.Id))
					witnessTemp.Block.PreBlock = newBlock
					break
				}
			}
			//创始区块
			if witness.Block.Height == config.Mining_block_start_height {
				// engine.Log.Info("关系   888888888888888")
				newGroup := new(Group)
				newGroup.Height = witness.Group.Height // bhvo.BH.GroupHeight
				newGroup.Blocks = []*Block{newBlock}
				// newGroup.NextGroup = make([]*Group, 0)
				newBlock.Group = newGroup
				continue
			}
		}
	} else {
		// engine.Log.Info("整理关系 22222")
		var findBlock *Block
		for i := len(this.Witness); i > 0; i-- {
			witness := this.Witness[i-1]
			if witness.Block == nil {
				// engine.Log.Info("关系   2222222222")
				continue
			}
			// engine.Log.Info("关系区块 %d %s", witness.Block.Height, witness.Addr.B58String())
			newBlock := witness.Block
			if findBlock == nil {
				//当指定的区块未寻找到，则继续寻找
				if bytes.Equal(newBlock.Id, *blockhash) {
					// engine.Log.Info("-------- 创建一个区块组 %d", newBlock.Height)
					newGroup := new(Group)
					newGroup.Height = witness.Group.Height
					newGroup.Blocks = []*Block{newBlock}
					newBlock.Group = newGroup
					findBlock = newBlock
				}
				continue
			}
			//继续寻找同组的其他区块
			if bytes.Equal(witness.Block.Id, findBlock.PreBlockID) {
				//找到这个见证人了，在组中添加这个区块
				// nextWitness = witnessTemp
				//从后往前查找，则把本块放在出块组的前面
				findBlock.Group.Blocks = append([]*Block{newBlock}, findBlock.Group.Blocks...)
				// engine.Log.Info("-------- 找到区块组，添加一个区块:%d %d", newBlock.Height, len(findBlock.Group.Blocks))
				witness.Block.Group = findBlock.Group
				findBlock = witness.Block
				// break
				continue
			}

			//创始区块
			if witness.Block.Height == config.Mining_block_start_height {
				// engine.Log.Info("关系   888888888888888")
				newGroup := new(Group)
				newGroup.Height = witness.Group.Height // bhvo.BH.GroupHeight
				newGroup.Blocks = []*Block{newBlock}
				// newGroup.NextGroup = make([]*Group, 0)
				newBlock.Group = newGroup
				continue
			}
		}
	}
}

/*
选出这个见证人组中出块最多块的链
*/
func (this *WitnessSmallGroup) SelectionChainOld() *Group {
	this.CollationRelationshipOld()

	// if this.Height > logblockheight || this.Height < 145137 {
	// engine.Log.Info("开始评选出块最多的组 group height %d", this.Height)
	// }

	//开始评选出块最多的组
	groupMap := make(map[string]*Group) //key:string=每个组的首个快hash;value:*Group=分叉组;
	for _, one := range this.Witness {
		// if this.Height > logblockheight || this.Height < 145137 {
		// engine.Log.Info("开始评选出块最多的组 111111111111111")
		// }
		// totalWitness++
		if one.Block == nil {
			// if this.Height > logblockheight || this.Height < 145137 {
			// engine.Log.Info("开始评选出块最多的组 2222222222222222222222")
			// }
			continue
		}
		if one.Block.Group == nil {
			// engine.Log.Info("开始评选出块最多的组 这里退出了")
			continue
		}
		// engine.Log.Info("开始评选出块最多的组 3333333333333333333 %v", one.Block.Group)
		// engine.Log.Info("开始评选出块最多的组 3333333333333333333 group:%d total:%d", this.Height, len(one.Block.Group.Blocks))
		groupMap[utils.Bytes2string(one.Block.Group.Blocks[0].Id)] = one.Block.Group
	}
	// engine.Log.Info("开始评选出块最多的组 444444444444444444444")
	//评选出块多的组
	var group *Group
	for _, v := range groupMap {
		// engine.Log.Info("开始评选出块最多的组 555555555555555555")
		if group == nil {
			// engine.Log.Info("开始评选出块最多的组 66666666666666666")
			group = v
			continue
		}
		// engine.Log.Info("开始评选出块最多的组 77777777777777777777")
		if len(v.Blocks) > len(group.Blocks) {
			// engine.Log.Info("开始评选出块最多的组 88888888888888888888")
			group = v
		}
		// engine.Log.Info("开始评选出块最多的组 99999999999999999")
	}
	// engine.Log.Info("开始评选出块最多的组 99999 %v", group)
	return group
}

/*
整理本组中已出区块之间的前后关系
*/
func (this *WitnessSmallGroup) CollationRelationshipOld() {
	// engine.Log.Info("整理本组中区块的前后关系 group height %d", this.Height)
	//先将这个组的块分成组
	for _, witness := range this.Witness {
		// engine.Log.Info("关系   1111111111")
		if witness.Block == nil {
			// engine.Log.Info("关系   2222222222")
			continue
		}
		// engine.Log.Info("关系   3333333333333 %d", witness.Block.Height)
		newBlock := witness.Block
		//寻找前置区块的见证人
		var beforeWitness *Witness
		for witnessTemp := witness.PreWitness; witnessTemp != nil; witnessTemp = witnessTemp.PreWitness {
			// beforeWitness = nil
			// engine.Log.Info("关系   44444444444 %d %s", witnessTemp.Group.Height, witnessTemp.Addr.B58String())
			if witnessTemp.Block == nil {
				// engine.Log.Info("关系   55555555555555")
				continue
			}
			//往前查找n个组，查不到跳出循环
			// if witnessTemp.Group.Height+config.Witness_backup_group <= witness.Group.Height {
			if witnessTemp.Group.Height+uint64(config.Witness_backup_group) <= witness.Group.Height {
				//刚好本见证人的前面的一个见证人没有出块，则同步这个块
				for syncWitness := witness.PreWitness; syncWitness != nil &&
					witness.Group.Height == syncWitness.Group.Height+1; syncWitness = syncWitness.PreWitness {
					if syncWitness.Block != nil {
						continue
					}
					//					bfw := BlockForWitness{
					//						GroupHeight: syncWitness.Group.Height, //见证人组高度
					//						Addr:        *syncWitness.Addr,        //见证人地址
					//					}
					//					//开始同步
					//					go syncWitness.syncBlock(5, time.Second/5, &bfw)
				}
				// break
			}

			// fmt.Println("--", newBlock)
			// engine.Log.Info("查看上一个区块 %d hash %s %s %v", witnessTemp.Block.Height,
			// hex.EncodeToString(witnessTemp.Block.Id), hex.EncodeToString(newBlock.PreBlockID), newBlock)
			if bytes.Equal(witnessTemp.Block.Id, newBlock.PreBlockID) {
				// engine.Log.Info("关系   66666666666666")
				// newBlock.PreBlock = witnessTemp.Block
				// if witnessTemp.Group.Height == witness.Group.Height {
				// 	fmt.Println("关系   777777777777777")
				// 	witnessTemp.Block.NextBlock = newBlock
				// }
				beforeWitness = witnessTemp
				break
			}
		}
		// engine.Log.Info("关系   7777777777777777")
		//创始区块
		if witness.Block.Height == config.Mining_block_start_height {
			// engine.Log.Info("关系   888888888888888")
			newGroup := new(Group)
			newGroup.Height = witness.Group.Height // bhvo.BH.GroupHeight
			newGroup.Blocks = []*Block{newBlock}
			// newGroup.NextGroup = make([]*Group, 0)
			newBlock.Group = newGroup
			continue
		}
		// engine.Log.Info("关系   9999999999999999")
		//没有前置区块的，直接不分成组了
		if beforeWitness == nil {
			// engine.Log.Info("关系   ----------------")
			continue
		}
		// engine.Log.Info("关系   1111  111111111111")
		//有前置区块，判断前置区块组高度是否相同
		if beforeWitness.Group.Height == witness.Group.Height {
			// engine.Log.Info("关系   1111  22222222222")
			//去重复
			have := false
			for _, one := range beforeWitness.Block.Group.Blocks {
				if newBlock.Witness == one.Witness {
					have = true
					break
				}
			}
			if !have {
				// engine.Log.Info("update block nextblockhash,this blockid:%s nextid:%s", hex.EncodeToString(beforeWitness.Block.Id), hex.EncodeToString(newBlock.Id))
				//组高度相同
				beforeWitness.Block.Group.Blocks = append(beforeWitness.Block.Group.Blocks, newBlock)
				newBlock.Group = beforeWitness.Block.Group
				newBlock.PreBlock = beforeWitness.Block
				beforeWitness.Block.NextBlock = newBlock
			}
		} else {
			// engine.Log.Info("关系   1111  3333333333333")
			//组高度不同
			if newBlock.Group == nil {
				// engine.Log.Info("关系   1111  444444444444")
				newBlock.Group = new(Group)
				newBlock.Group.Height = witness.Group.Height // bhvo.BH.GroupHeight
				newBlock.Group.Blocks = []*Block{newBlock}
				// newGroup.NextGroup = make([]*Group, 0)
			}
			newBlock.PreBlock = beforeWitness.Block
		}
		// beforeWitness.Block.NextBlock = append(beforeWitness.Block.NextBlock, newBlock)
		// engine.Log.Info("关系   1111  5555555555555555 %d", len(newBlock.Group.Blocks))
		// for _, one := range newBlock.Group.Blocks {
		// 	engine.Log.Info("关系   1111  666666666 %d", one.Height)
		// }

	}
}

/*
	构建首个见证人组
*/
// func (this *WitnessChain) BuildFirstGroup(block *Block) {
// 	this.BuildWitnessGroup(true, block)
// 	group := this.backupGroupLast
// 	for group.PreGroup != nil {
// 		group = group.PreGroup
// 	}
// 	// this.witness = group.Witness[0]
// 	this.witnessGroup = group

// }

/*
*
计算剩余的见证组数量
*/
func (this *WitnessChain) calcBackupGroupNum(totalBackupGroup uint64) uint64 {
	backupGroup := uint64(config.Witness_backup_group)

	total := this.witnessBackup.GetBackupWitnessTotal()

	used := uint64(math.Ceil(float64(totalBackupGroup) / float64(total)))

	surplus := backupGroup - used

	return totalBackupGroup + surplus*total
}

var buildWitnessGroupLock sync.RWMutex

/*
依次获取n个未分配组的见证人，构建一个新的见证人组
@isPrint    bool    是否打印日志
@witness    *Witness    当前出块的见证人
*/
// func (this *WitnessChain) BuildWitnessGroupNew(first, isPrint bool, witness *Witness) {
// 	buildWitnessGroupLock.Lock()
// 	defer buildWitnessGroupLock.Unlock()

// 	backupGroup := uint64(config.Witness_backup_group)

// 	var preBigGroupLastGroup *WitnessSmallGroup
// 	if witness == nil {
// 		preBigGroupLastGroup = this.witnessGroup
// 	} else {
// 		preBigGroupLastGroup = witness.Group
// 	}
// 	lastGroupHeight := uint64(config.Mining_group_start_height)
// 	last := witness
// 	for {
// 		if last == nil || last.PreWitness == nil {
// 			break
// 		}
// 		last = last.PreWitness

// 		if last.WitnessBigGroup == witness.WitnessBigGroup {
// 			continue
// 		}
// 		preBigGroupLastGroup = last.Group
// 		break
// 	}
// 	// if preBigGroupLastGroup != nil {
// 	// 	engine.Log.Info("preBigGroupLastGroup ：%d ", preBigGroupLastGroup.Height)
// 	// }
// 	//判断备用见证人组数量是否足够，不够则创建
// 	totalBackupGroup := 0
// 	tag := false
// 	var lastGroup *WitnessSmallGroup
// 	findWitness := false //找到当前出块的见证人
// 	for lastGroup = preBigGroupLastGroup; lastGroup != nil; lastGroup = lastGroup.NextGroup {
// 		if !findWitness && witness != nil && witness.Group.Height == lastGroup.Height {
// 			for _, witnessOne := range lastGroup.Witness {
// 				if bytes.Equal(*witnessOne.Addr, *witness.Addr) {
// 					findWitness = true
// 					break
// 				}
// 			}
// 		}
// 		if findWitness {
// 			totalBackupGroup++
// 		}
// 		tag = lastGroup.tag
// 		lastGroupHeight = lastGroup.Height
// 		if lastGroup.NextGroup == nil {
// 			break
// 		}
// 	}

// 	// engine.Log.Info("-------构建见证人组  3333333333333333333 %d %d", totalBackupGroup, backupGroup)

// 	if first {
// 		backupGroup = 0
// 	}
// 	//else {
// 	//	total := this.witnessBackup.GetBackupWitnessTotal()
// 	//	groupNum := total / config.Mining_group_max
// 	//	if groupNum > backupGroup {
// 	//		backupGroup = groupNum
// 	//	}
// 	//}
// 	//engine.Log.Info("totalBackupGroup:%d,backupGroup:%d", totalBackupGroup, backupGroup)
// 	//if lastGroup != nil {
// 	//	engine.Log.Info("New group height %d %d", lastGroupHeight, lastGroup.Height)
// 	//}
// 	// engine.Log.Info("-------构建见证人组  444444444444444 %d %d", totalBackupGroup, backupGroup)

// 	groupNum := uint64(0)
// 	if totalBackupGroup > 0 {
// 		groupNum = this.calcBackupGroupNum(uint64(totalBackupGroup))
// 		if groupNum > backupGroup {
// 			backupGroup = groupNum
// 		}
// 		// engine.Log.Info("totalBackupGroup:%d,backupGroup:%d", totalBackupGroup, backupGroup)
// 	}

// 	for i := uint64(totalBackupGroup); i <= backupGroup; i++ {
// 		newGroup := this.AdditionalWitnessBackup(lastGroup, lastGroupHeight, tag)
// 		if newGroup == nil {
// 			break
// 		}
// 		lastGroupHeight++
// 		tag = !tag
// 		lastGroup = newGroup

// 		if this.witnessGroup == nil {
// 			// engine.Log.Info("3333333333333 %d", newGroup.Height)
// 			this.witnessGroup = newGroup
// 		}
// 	}

// 	// engine.Log.Info("剩余见证人：%d", len(this.witnessNotGroup))

// 	// for _, v := range this.witnessNotGroup {
// 	// 	engine.Log.Info("剩余见证人：%p  %s", v.WitnessBigGroup, v.Addr.B58String())
// 	// }

// 	leaveLen := len(this.witnessNotGroup)
// 	for i := 0; i < leaveLen; i++ {
// 		newGroup := this.AdditionalWitnessBackup(lastGroup, lastGroupHeight, tag)
// 		if newGroup == nil {
// 			break
// 		}
// 		lastGroupHeight++
// 		tag = !tag
// 		lastGroup = newGroup

// 		if this.witnessGroup == nil {
// 			// engine.Log.Info("3333333333333 %d", newGroup.Height)
// 			this.witnessGroup = newGroup
// 		}
// 	}
// 	if isPrint {
// 		this.PrintWitnessList()
// 	}
// }

/*
构建足够数量的备用见证人组
构建组数量分为两种：
一种是连续出块，则
@first    bool    链刚启动，首次构建见证人组
*/
var BuildWitnessGroupLock = new(sync.Mutex)

func (this *WitnessChain) BuildWitnessGroup(first, isPrint bool) {
	BuildWitnessGroupLock.Lock()
	defer BuildWitnessGroupLock.Unlock()
	//计算备用见证人组数量，获取最后的组高度
	// totalBackupGroup := uint64(0)                               //备用见证人组数量
	// lastGroupHeight := uint64(config.Mining_group_start_height) //最后的组高度
	// tag := false
	// var lastGroup *WitnessSmallGroup
	// for lastGroup = this.witnessGroup; lastGroup != nil; lastGroup = lastGroup.NextGroup {
	// 	totalBackupGroup++
	// 	tag = lastGroup.tag
	// 	lastGroupHeight = lastGroup.Height
	// 	if lastGroup.NextGroup == nil {
	// 		break
	// 	}
	// }

	//补充多少个组
	total := 0
	if first {
		//初始构建1个组
		total = 1
	} else {
		//计算应该构建多少个备用见证人组
		witnessBackupGroupNum := config.Witness_backup_group
		totalWitness := this.witnessBackup.GetBackupWitnessTotal()
		groupNum := int(totalWitness) / config.Mining_group_max
		if groupNum > witnessBackupGroupNum {
			witnessBackupGroupNum = groupNum
		}
		// engine.Log.Info("起始组高度:%d", this.witnessGroup.Height)
		//判断是否跳过太多的组没有出块
		count := 0
		var lastGroup *WitnessSmallGroup
		for lastGroup = this.WitnessGroup; lastGroup != nil; lastGroup = lastGroup.NextGroup {
			if lastGroup.NextGroup == nil {
				break
			}
			count++
		}
		//向前计算多少组未出块
		// over := true
		// count := 0
		// for i := 0; i <= witnessBackupGroupNum*2; i++ {
		// 	count++
		// 	if lastGroup.BlockGroup != nil || lastGroup.PreGroup == nil {
		// 		engine.Log.Info("截至到高度:%d", lastGroup.Height)
		// 		// over = false
		// 		break
		// 	}
		// 	lastGroup = lastGroup.PreGroup
		// }
		//超过了N组未出块，则补一个组
		if count > witnessBackupGroupNum {
			total = 1
			// engine.Log.Info("开始构建备用见证人数量:%d-%d %d", witnessBackupGroupNum, count, total)
			// } else if over {
			// 	total = 1
			// 	engine.Log.Info("开始构建备用见证人数量:%d", total)
		} else {
			total = witnessBackupGroupNum - count
			// engine.Log.Info("开始构建备用见证人数量:%d-%d=%d", witnessBackupGroupNum, count, total)
		}
	}

	//开始构建足够数量的备用见证人组
	for i := 0; i < total; i++ {
		newGroup := this.AdditionalWitnessBackup()
		if this.WitnessGroup == nil {
			this.WitnessGroup = newGroup
		}
	}
	if !first {
		//判断最后一组是否超时，如果超时则补足够的组
		// this.BuildWitnessGroupSupplemental()
	}
	if isPrint && total != 0 {
		this.PrintWitnessList()
	}
	return

	// //计算应该构建多少个备用见证人组
	// // engine.Log.Info("-------构建见证人组  3333333333333333333 %d %d", totalBackupGroup, backupGroup)
	// backupGroup := uint64(config.Witness_backup_group)
	// if first {
	// 	backupGroup = 0
	// 	totalBackupGroup = 0
	// } else {
	// 	total := this.witnessBackup.GetBackupWitnessTotal()
	// 	groupNum := total / config.Mining_group_max
	// 	if groupNum > backupGroup {
	// 		backupGroup = groupNum
	// 	}
	// }

	// // engine.Log.Info("-------构建见证人组  444444444444444 %d %d", totalBackupGroup, backupGroup)
	// engine.Log.Info("补充组数量:%d %d %d", backupGroup-totalBackupGroup, backupGroup, totalBackupGroup)
	// //开始构建足够数量的备用见证人组
	// for i := uint64(totalBackupGroup); i <= backupGroup; i++ {
	// 	newGroup := this.AdditionalWitnessBackup()
	// 	if newGroup == nil {
	// 		break
	// 	}
	// 	lastGroupHeight++
	// 	tag = !tag
	// 	lastGroup = newGroup

	// 	if this.witnessGroup == nil {
	// 		// engine.Log.Info("3333333333333 %d", newGroup.Height)
	// 		this.witnessGroup = newGroup
	// 	}
	// }
	// if isPrint {
	// 	this.PrintWitnessList()
	// }
}

/*
补充构建备用见证人组，只比当前时间多一个组
*/
func (this *WitnessChain) BuildWitnessGroupSupplemental() {
	BuildWitnessGroupLock.Lock()
	defer BuildWitnessGroupLock.Unlock()
	//计算备用见证人组数量，获取最后的组高度
	var lastGroup = this.WitnessGroup
	var witness *Witness
	have := false
	for {
		for ; lastGroup != nil; lastGroup = lastGroup.NextGroup {
			if lastGroup.NextGroup == nil {
				break
			}
		}
		witness = lastGroup.Witness[len(lastGroup.Witness)-1]
		if witness.CreateBlockTime > config.TimeNow().Unix() {
			engine.Log.Info("未超时")
			//未超时
			break
		}
		have = true
		this.AdditionalWitnessBackup()
		// this.BuildWitnessGroup(false, true)
		// this.witnessGroup = this.witnessGroup.NextGroup
	}
	if have {
		this.PrintWitnessList()
	}
}

/*
从最后一组备用见证人组追加一组备用见证人
*/
func (this *WitnessChain) AdditionalWitnessBackup() *WitnessSmallGroup {
	//找到最后一个备用见证人组
	var lastGroup *WitnessSmallGroup
	for lastGroup = this.WitnessGroup; lastGroup != nil; lastGroup = lastGroup.NextGroup {
		if lastGroup.NextGroup == nil {
			break
		}
	}

	// engine.Log.Info("开始构建组高度AdditionalWitnessBackup %d", lastGroupHeight)
	//从候选见证人中评选出备用见证人来
	//把评选出来的所有备用见证人分组成为见证人组。
	// if lastGroup != nil {
	// 	engine.Log.Info("lastgroup :%d 获取见证人ssssssssssssssss", lastGroup.Height)
	// }
	witnessGroup := this.GetOneGroupWitness()
	//engine.Log.Info("获取的见证人组中，见证人数量 %d", len(witnessGroup))
	if witnessGroup == nil || len(witnessGroup) < config.Mining_group_min {
		engine.Log.Info("Too few witnesses current:%d min:%d", len(witnessGroup), config.Mining_group_min)
		return nil
	}
	// if lastGroup != nil {
	// 	engine.Log.Info("lastgroup :%d  witness:%d %s", lastGroup.Height, len(witnessGroup), witnessGroup[0].Addr.B58String())
	// }
	//找到上一个见证人的出块时间
	startTime := int64(0)
	if lastGroup != nil {
		w := lastGroup.Witness[len(lastGroup.Witness)-1]
		startTime = w.CreateBlockTime
	}
	//pl time
	//给见证人计算出块时间
	for _, one := range witnessGroup {
		//startTime = startTime + config.Mining_block_time
		startTime = startTime + int64(config.Mining_block_time.Seconds())
		one.CreateBlockTime = startTime
	}
	//
	tag := false
	lastGroupHeight := uint64(1)
	if lastGroup != nil {
		lastGroupHeight = lastGroup.Height + 1
		tag = !lastGroup.tag
	}
	newGroup := &WitnessSmallGroup{
		PreGroup: lastGroup,       //备用见证人最后一个组
		Height:   lastGroupHeight, //
		Witness:  witnessGroup,    //本组见证人列表
		tag:      tag,             //
		// check:    false,                //这个见证人组是否多半人出块，多半人出块则合法有效
		CollationLock: new(sync.Mutex),
	}

	for i, _ := range newGroup.Witness {
		newGroup.Witness[i].Group = newGroup
	}

	if lastGroup != nil {
		lastGroup.Witness[len(lastGroup.Witness)-1].NextWitness = newGroup.Witness[0]
		newGroup.Witness[0].PreWitness = lastGroup.Witness[len(lastGroup.Witness)-1]
		lastGroup.NextGroup = newGroup
		newGroup.PreGroup = lastGroup
	}

	//engine.Log.Info("创建新的组:%d", newGroup.Height)
	if config.LOG_print_witness_group {
		printPre := "----"
		printStr := printPre + "\n"
		for _, one := range newGroup.Witness {
			printStr += fmt.Sprintf("wg:%d %s %s %d\n", newGroup.Height, one.Addr.B58String(),
				utils.FormatTimeToSecond(time.Unix(one.CreateBlockTime, 0)), one.VoteNum)
		}
		printStr = printStr + printPre
		engine.Log.Info("%s", printStr)
	}

	return newGroup
}

/*
拉起链端后，根据当前时间补偿备用见证人组
*/
func (this *WitnessChain) CompensateWitnessGroup() {
	this.BuildWitnessGroupSupplemental()

	// engine.Log.Info("开始打印备用见证人组信息")
	this.PrintWitnessList()
	// engine.Log.Info("开始查询是否是备用见证人")
	//查询是否是备用见证人
	_, isBackup, _, _, _ := GetWitnessStatus()
	if isBackup {
		config.AlreadyMining = true
	}
	// engine.Log.Info("开始暂停所有出块")
	this.StopAllMining()
	// engine.Log.Info("开始构建出块时间")
	this.BuildMiningTime()
	// engine.Log.Info("构建出块时间完成")
	this.chain.SyncBlockFinish = true
	engine.Log.Info("加载链端，补组完成")
}

/*
拉起链端后，根据当前时间补偿备用见证人组
*/
// func (this *WitnessChain) CompensateWitnessGroupOld() {
// 	//计算备用见证人组数量，获取最后的组高度

// 	//判断当前见证人组出块时间是否超时
// 	backupGroup := uint64(config.Witness_backup_group)
// 	witness := lastGroup.Witness[len(lastGroup.Witness)-1]
// 	for {
// 		witness = lastGroup.Witness[len(lastGroup.Witness)-1]
// 		if witness.CreateBlockTime > config.TimeNow() {
// 			//未超时
// 			break
// 		}
// 		//
// 		this.BuildBlockGroupForGroupHeight(lastGroup.Height, nil)
// 		totalBackupGroup = 0
// 		newGroup := this.AdditionalWitnessBackup(lastGroup, lastGroupHeight, tag)
// 		if newGroup == nil {
// 			break
// 		}
// 		lastGroupHeight++
// 		tag = !tag
// 		lastGroup = newGroup
// 		// engine.Log.Info("新组高度 %d %d", lastGroupHeight, lastGroup.Height)
// 		engine.Log.Info("New group height %d %d", lastGroupHeight, lastGroup.Height)
// 		// engine.Log.Info("预计出块时间 %d %s", witness.Group.Height, time.Unix(witness.CreateBlockTime, 0).Format("2006-01-02 15:04:05"))
// 		engine.Log.Info("Estimated block out time %d %s", witness.Group.Height, time.Unix(witness.CreateBlockTime, 0).Format("2006-01-02 15:04:05"))
// 		continue
// 	}

// 	total := this.witnessBackup.GetBackupWitnessTotal()
// 	groupNum := total / config.Mining_group_max
// 	if groupNum > backupGroup {
// 		backupGroup = groupNum
// 	}
// 	// engine.Log.Info("补充组数量:%d %d %d", backupGroup-totalBackupGroup, backupGroup, totalBackupGroup)
// 	// engine.Log.Info("-------构建见证人组  444444444444444 %d %d", totalBackupGroup, backupGroup)
// 	for i := uint64(totalBackupGroup); i <= backupGroup; i++ {
// 		newGroup := this.AdditionalWitnessBackup(lastGroup, lastGroupHeight, tag)
// 		if newGroup == nil {
// 			break
// 		}
// 		lastGroupHeight++
// 		tag = !tag
// 		lastGroup = newGroup
// 		if this.witnessGroup == nil {
// 			// engine.Log.Info("44444444444444444 %d", newGroup.Height)
// 			this.witnessGroup = newGroup
// 		}
// 	}

// 	// engine.Log.Info("剩余见证人：%d", len(this.witnessNotGroup))

// 	// //for _, v := range this.witnessNotGroup {
// 	// //	engine.Log.Info("剩余见证人：%p  %s", v.WitnessBigGroup, v.Addr.B58String())
// 	// //}

// 	// leaveLen := len(this.witnessNotGroup)
// 	// for i := 0; i < leaveLen; i++ {
// 	// 	newGroup := this.AdditionalWitnessBackup(lastGroup, lastGroupHeight, tag)
// 	// 	if newGroup == nil {
// 	// 		break
// 	// 	}
// 	// 	lastGroupHeight++
// 	// 	tag = !tag
// 	// 	lastGroup = newGroup

// 	// 	if this.witnessGroup == nil {
// 	// 		// engine.Log.Info("3333333333333 %d", newGroup.Height)
// 	// 		this.witnessGroup = newGroup
// 	// 	}
// 	// }
// 	// engine.Log.Info("开始打印备用见证人组信息")
// 	this.PrintWitnessList()
// 	// engine.Log.Info("开始查询是否是备用见证人")
// 	//查询是否是备用见证人
// 	_, isBackup, _, _, _ := GetWitnessStatus()
// 	if isBackup {
// 		config.AlreadyMining = true
// 	}
// 	// engine.Log.Info("开始暂停所有出块")
// 	this.StopAllMining()
// 	// engine.Log.Info("开始构建出块时间")
// 	this.BuildMiningTime()
// 	// engine.Log.Info("构建出块时间完成")
// 	this.chain.SyncBlockFinish = true
// }

/*
根据导入区块高度补偿备用见证人组
*/
func (this *WitnessChain) CompensateWitnessGroupByGroupHeight(groupHeight uint64) {
	BuildWitnessGroupLock.Lock()
	defer BuildWitnessGroupLock.Unlock()
	engine.Log.Info("CompensateWitnessGroupByGroupHeight build group height:%d", groupHeight)

	//判断备用见证人组数量是否足够，不够则创建
	totalBackupGroup := 0
	tag := false
	lastGroupHeight := uint64(config.Mining_group_start_height)
	var lastGroup *WitnessSmallGroup
	for lastGroup = this.WitnessGroup; lastGroup != nil; lastGroup = lastGroup.NextGroup {

		// engine.Log.Info("-------构建见证人组  222222222222222222222")
		totalBackupGroup++
		tag = lastGroup.tag
		lastGroupHeight = lastGroup.Height
		if lastGroup.NextGroup == nil {
			break
		}
	}

	// engine.Log.Info("已有组高度 %d", lastGroupHeight)

	//判断当前见证人组出块时间是否超时
	for i := lastGroupHeight; i < groupHeight; i++ {
		engine.Log.Info("start build group height:%d", lastGroupHeight)

		totalBackupGroup = 0
		newGroup := this.AdditionalWitnessBackup()
		if newGroup == nil {
			break
		}
		lastGroupHeight++
		tag = !tag
		lastGroup = newGroup

		// engine.Log.Info("新组高度 %d %d", lastGroupHeight, lastGroup.Height)

		// engine.Log.Info("预计出块时间 %d %s", witness.Group.Height, time.Unix(witness.CreateBlockTime, 0).Format("2006-01-02 15:04:05"))
		continue
	}

	this.PrintWitnessList()
}

/*
暂停所有出块
*/
func (this *WitnessChain) StopAllMining() {
	// engine.Log.Info("-----开始暂停所有出块-----")
	//判断自己出块顺序的时间
	addr := Area.Keystore.GetCoinbase()
	// witnessTemp := this.witness
	witnessTemp := this.WitnessGroup.Witness[0]
	for {
		witnessTemp = witnessTemp.NextWitness
		if witnessTemp == nil || witnessTemp.Group == nil {
			// engine.Log.Info("11111111111111")
			break
		}
		if !witnessTemp.Group.Task {
			// engine.Log.Info("222222222222222")
			continue
		}

		if !bytes.Equal(*witnessTemp.Addr, addr.Addr) {
			// engine.Log.Info("33333333333333333")
			continue
		}

		select {
		case witnessTemp.StopMining <- false:
			// engine.Log.Info("444444444444444444")
		default:
			// engine.Log.Info("5555555555555555555")
		}
		witnessTemp.Group.Task = false
	}
}

/*
给备用见证人添加定时任务，定时出块
*/
func (this *WitnessChain) BuildMiningTime() error {

	//engine.Log.Info("=========开始构建所有出块=======")

	//判断自己出块顺序的时间
	addr := Area.Keystore.GetCoinbase()

	for witnessGroup := this.WitnessGroup; witnessGroup != nil; witnessGroup = witnessGroup.NextGroup {

		if witnessGroup.Task {
			//engine.Log.Warn("33333333333333333333333333333333333333333333333")
			continue
		}
		for _, witnessTemp := range witnessGroup.Witness {
			if witnessTemp.Block != nil {
				//engine.Log.Warn("2222222222222222222222222222222222222222222222")
				continue
			}

			//自己是见证人的话，永远不停止导入
			atomic.StoreUint32(&this.chain.StopSyncBlock, 0)
			future := int64(0) //出块时间，也可以记录上一个块的出块时间
			//pl time
			now := config.TimeNow().UnixNano()
			//engine.Log.Debug("构建出块时间 预约时间 %s 现在时间 %s", time.Unix(witnessTemp.CreateBlockTime, 0), time.Unix(now, 0))
			createBlockTiem := time.Unix(witnessTemp.CreateBlockTime, 0).UnixNano()
			if createBlockTiem > now {
				future = createBlockTiem - now
			} else if witnessTemp.CreateBlockTime == now {
				future = 0
			} else {
				difference := now - createBlockTiem
				if difference < config.Mining_block_time.Nanoseconds() {
					future = 0
				} else {
					//最后一组的最后一个见证人需要补充见证人组
					witnessTemp.Group.Task = true
					go witnessTemp.SupplementBuildGroup(future)
					//engine.Log.Warn("111111111111111111111111111111111111111111111111")
					continue
				}

			}

			//时间太少，来不及出块
			if !config.InitNode && future <= 20 {
				//最后一组的最后一个见证人需要补充见证人组
				go witnessTemp.SupplementBuildGroup(future)
				engine.Log.Warn("There's too little time for a block %s %s", time.Unix(witnessTemp.CreateBlockTime, 0), time.Unix(now, 0))
				continue
			}

			witnessTemp.Group.Task = true
			// utils.Go(witnessTemp.SyncBuildBlock(int64(future)))
			if bytes.Equal(*witnessTemp.Addr, addr.Addr) {
				//出块定时任务
				go witnessTemp.SyncBuildBlock(future)
			}
			//给不是自己的见证人设置一个定时同步区块的方法

			//最后一组的最后一个见证人需要补充见证人组
			go witnessTemp.SupplementBuildGroup(future)

		}

	}

	return nil
}

/*
获取一组新的见证人组
从未分配组的见证人中按顺序获取一个组的见证人
*/
func (this *WitnessChain) GetOneGroupWitness() []*Witness {
	//当前组见证人数量由候选见证人数量来定
	groupNum := config.Mining_group_min
	total := this.witnessBackup.GetBackupWitnessTotal()
	if total > config.Mining_group_max {
		groupNum = config.Mining_group_max
	} else if total < config.Mining_group_min {
		groupNum = config.Mining_group_min
	} else {
		groupNum = int(total)
	}
	//engine.Log.Info("本次计算出见证人数量 %d %d", groupNum, total)
	//备用见证人数量太少，则从候选见证人中选一批新的备用见证人
	if len(this.witnessNotGroup) < groupNum {
		// engine.Log.Info("数量太少，则评选出新的备用见证人 %d", len(this.witnessNotGroup))
		witness := this.witnessBackup.CreateWitnessGroup()
		if witness == nil {
			return nil
		}
		this.witnessNotGroup = append(this.witnessNotGroup, witness...)

		// engine.Log.Info("witnessNotGroup length : %d", len(this.witnessNotGroup))
		// for _, temp := range this.witnessNotGroup {
		// 	engine.Log.Info("print backup witness list %p %s", temp.WitnessBigGroup, temp.Addr.B58String())
		// }
		// engine.Log.Info("print end")

	}
	//保存重复的见证人，需要向后面移动
	index := 0
	moveWitness := make([]*Witness, 0)
	witnessGroup := make([]*Witness, 0)
	for i, tempWitness := range this.witnessNotGroup {
		index = i
		//判断组中是否有重复的见证人
		isHave := false
		for _, one := range witnessGroup {
			if bytes.Equal(*tempWitness.Addr, *one.Addr) {
				// engine.Log.Info("有重复 %s", one.Addr.B58String())
				//把重复的见证人保存下来
				// moveOne := tempWitness
				moveWitness = append(moveWitness, tempWitness)
				// tempWitness = tempWitness.NextWitness
				// moveOne.NextWitness = nil
				// moveOne.PreWitness = nil
				isHave = true
				break
			}
		}
		//有重复的，跳出循环
		if isHave {
			continue
		}
		witnessGroup = append(witnessGroup, tempWitness)
		if len(witnessGroup) >= groupNum {
			break
		}
	}
	newWitnessNotGroup := this.witnessNotGroup[index+1:]
	this.witnessNotGroup = make([]*Witness, 0)
	//有重复的向后移动，从新排序
	for _, one := range moveWitness {
		this.witnessNotGroup = append(this.witnessNotGroup, one)
	}
	for _, one := range newWitnessNotGroup {
		this.witnessNotGroup = append(this.witnessNotGroup, one)
	}
	//从新建立引用关系
	tempWitness := witnessGroup[0]
	for i, _ := range witnessGroup {
		if i == 0 {
			continue
		}
		tempWitness.NextWitness = witnessGroup[i]
		witnessGroup[i].PreWitness = tempWitness
		tempWitness = witnessGroup[i]
	}
	//engine.Log.Info("最后的见证人数量 %d", len(witnessGroup))
	// for i, v := range this.witnessNotGroup {
	// 	engine.Log.Info("未分配完的见证人%d-%d:%s", len(this.witnessNotGroup), i, v.Addr.B58String())
	// }
	return witnessGroup

}

/* 添加见证人，依次添加*/
// func (this *WitnessChain) AddWitness(newwitness []*Witness) {

// 	if this.firstWitnessNotGroup == nil {
// 		this.firstWitnessNotGroup = newwitness
// 	} else {
// 		//查找到最后一个见证人
// 		lastWitnessNotGroup := this.firstWitnessNotGroup
// 		for lastWitnessNotGroup.NextWitness != nil {
// 			lastWitnessNotGroup = lastWitnessNotGroup.NextWitness
// 		}
// 		newwitness.PreWitness = lastWitnessNotGroup
// 		lastWitnessNotGroup.NextWitness = newwitness
// 	}

// }

/*
打印见证人列表
*/
func (this *WitnessChain) PrintWitnessList() {
	//打印未分组的见证人列表
	// this.witnessBackup.PrintWitnessBackup()
	return

	firstGroup := false
	//count := 0
	group := this.WitnessGroup
	for group != nil {
		if !firstGroup || group.NextGroup == nil || group.NextGroup.NextGroup == nil {
			engine.Log.Info("--------------")
			firstGroup = true
			for _, one := range group.Witness {
				// gp, _ := fmt.Printf("%p", group)
				engine.Log.Info("witness group:%d %s %s %d", group.Height, one.Addr.B58String(),
					time.Unix(one.CreateBlockTime, 0), one.VoteNum)
			}
		}
		//这里返回，只打印一组见证人
		//count++
		//if count >= 2 {
		//	return
		//}

		group = group.NextGroup
	}
	engine.Log.Info("--------------")
}

/*
在已经确认的区块中查找到对应的区块
@return    bool    是否找到
*/
func (this *WitnessChain) FindBlockInCurrent(bh *BlockHead) bool {
	//新的区块只能从未确认的区块和之前的一个已经确认的组开始往后链接区块
	//第一步：找到组高度对应的组
	currentGroup := this.WitnessGroup
	for currentGroup != nil {
		if currentGroup.Height < bh.GroupHeight {
			currentGroup = currentGroup.NextGroup
			continue
		}
		if currentGroup.Height > bh.GroupHeight {
			currentGroup = currentGroup.PreGroup
			continue
		}
		break
	}
	if currentGroup == nil {
		return false
	}
	//
	if currentGroup.BlockGroup == nil {
		return false
	}
	// engine.Log.Info("找到的组 %+v", currentGroup)

	//第二步：找出区块
	for _, one := range currentGroup.Witness {
		if one.Block == nil {
			continue
		}
		if bytes.Equal(one.Block.Id, bh.Hash) {
			return true
		}
	}

	// engine.Log.Info("find the group %+v", currentGroup)
	return false
}

/*
查找新区块的前置区块是否存在，并且合法
合法就是：前置区块必须是已确认的组最后一个区块，不能是已确认的组之前的区块
@return    *Witness    找到的见证人
*/

func (this *WitnessChain) FindPreWitnessForBlock(preBlockHash []byte) *Witness {

	//新的区块只能从未确认的区块和之前的一个已经确认的组开始往后链接区块
	//第一步：找到未确认区块之前的一个已经确认的组
	// engine.Log.Info("当前的组高度 %d", this.witnessGroup.Height)
	currentGroup := this.WitnessGroup //当前未确认的组
	for {
		if currentGroup.BlockGroup != nil {
			break
		}
		if currentGroup.PreGroup == nil {
			engine.Log.Info("find the group:%+v", currentGroup)
			return nil
		}
		currentGroup = currentGroup.PreGroup
	}
	// engine.Log.Info("找到的组 %+v", currentGroup)

	//第二步：找出前置区块
	var witness *Witness
	for have := false; have == false && currentGroup != nil; {
		for _, one := range currentGroup.Witness {
			if one.Block == nil {
				continue
			}
			if bytes.Equal(one.Block.Id, preBlockHash) {
				// engine.Log.Info("找到的见证人:%+v", one)
				witness = one
				have = true
				break
				// return one
			}
		}
		if have {
			break
		}
		currentGroup = currentGroup.NextGroup
	}

	//当找到的组已经被统计，那么前置区块hash只能是在统计中的区块中的那一个。
	if witness != nil {
		// engine.Log.Info("找到的见证人:%d", witness.Block.Height)
		if currentGroup.BlockGroup != nil {
			height := uint64(0)
			hash := []byte{}
			isCount := false
			for _, one := range currentGroup.BlockGroup.Blocks {
				// engine.Log.Info("找到的块:%d %t", one.Height, one.isCount)
				if one.isCount {
					isCount = true
				}
				if one.Height > height {
					height = one.Height
					hash = one.Id
				}
			}
			if isCount && !bytes.Equal(preBlockHash, hash) {
				// engine.Log.Info("找到的hash不相等:%s %s", hex.EncodeToString(preBlockHash), hex.EncodeToString(hash))
				return nil
			}
		}
	}
	// engine.Log.Info("find the group:%+v", currentGroup)
	return witness
}

/*
检查
*/
// func (this *WitnessChain) CheckWitnessBuildBlockTime(bhvo *BlockHeadVO) bool {
// 	return true
// }

/*
	是否超出了已确定的备用见证人组高度
	@return    bool    是否超出了。true:=超出了;false:=未超出;
*/
// func (this *WitnessChain) IsFutureWitnessGroupChain(bhvo *BlockHeadVO) bool {

// }

/*
通过新区块，在未出块的见证人组中找到这个见证人
@return    *Witness    找到的见证人
@return    bool        是否跳过的组太多，而未查到此组的见证人。true:=跳过太多了;false:=并没有;
*/
func (this *WitnessChain) FindWitnessForBlockOnly(bhvo *BlockHeadVO) (*Witness, bool) {
	// engine.Log.Info("find group height:%d this.witnessGroup:%d", bhvo.BH.GroupHeight, this.witnessGroup.Height)
	//找到组高度对应的组
	currentGroup := this.WitnessGroup
	for currentGroup != nil {
		if currentGroup.Height < bhvo.BH.GroupHeight {
			// engine.Log.Info("设置已出的块 333333333333333333 %d", group.Height)
			currentGroup = currentGroup.NextGroup
			continue
		}
		if currentGroup.Height > bhvo.BH.GroupHeight {
			if currentGroup.PreGroup != nil && currentGroup.PreGroup.BlockGroup == nil {
				currentGroup = currentGroup.PreGroup
				continue
			}
			engine.Log.Info("组高度不同 %d %d", currentGroup.Height, bhvo.BH.GroupHeight)
			return nil, false
		}
		break
	}

	//跳过的组高度太多，则不能在备用见证人组中查找到
	if currentGroup == nil {
		engine.Log.Info("跳过的组高度太多，则不能在备用见证人组中查找到")
		return nil, true
	}
	// engine.Log.Info("出的块:%d :%d 见证人组高度:%d", bhvo.BH.GroupHeight, bhvo.BH.Height, currentGroup.Height)
	// if bhvo.BH.Height < config.WitnessOrderCorrectEnd && bhvo.BH.Height > config.WitnessOrderCorrectStart {
	// 	// engine.Log.Info("在范围内")
	// 	//找到组中对应的见证人，此高度内，随便分配个见证人给他
	// 	for _, one := range currentGroup.Witness {
	// 		// engine.Log.Info("见证人one:%+v", one)
	// 		if one.Block != nil {
	// 			// engine.Log.Info("本组有块了 group:%d height:%d", one.Block.Group.Height, one.Block.Height)
	// 			continue
	// 		}
	// 		// engine.Log.Info("返回这个见证人见证人one:%+v", one)
	// 		return one, false
	// 	}
	// }

	//找到组中对应的见证人
	for _, one := range currentGroup.Witness {
		// engine.Log.Info("设置已出的块，对比一下 %s %s", bhvo.BH.Witness.B58String(), one.Addr.B58String())
		if bytes.Equal(bhvo.BH.Witness, *one.Addr) {
			//找到这个见证人了
			return one, false
		}
	}
	engine.Log.Info("跳过的组高度太多，则不能在备用见证人组中查找到")
	return nil, false
}

/*
检查区块是否已经导入过了
@return    bool        是否已经导入过了
*/
func (this *WitnessChain) CheckRepeatImportBlock(bhvo *BlockHeadVO) bool {
	//找到组高度对应的组
	currentGroup := this.WitnessGroup
	for currentGroup != nil {
		if currentGroup.Height < bhvo.BH.GroupHeight {
			currentGroup = currentGroup.NextGroup
			continue
		}
		if currentGroup.Height > bhvo.BH.GroupHeight {
			currentGroup = currentGroup.PreGroup
			continue
		}
		break
	}
	if currentGroup == nil {
		// engine.Log.Info("未导入过")
		return false
	}
	//如果该组已经构建过了，则判定为已经导入过了
	if currentGroup.IsBuildGroup {
		engine.Log.Info("this group is build: %d", currentGroup.Height)
		return true
	}

	// if bhvo.BH.Height < config.WitnessOrderCorrectEnd && bhvo.BH.Height > config.WitnessOrderCorrectStart {
	// 	// engine.Log.Info("在范围内")
	// 	//找到组中对应的见证人，此高度内，随便分配个见证人给他
	// 	for _, one := range currentGroup.Witness {
	// 		// engine.Log.Info("见证人one:%+v", one)
	// 		if one.Block != nil {
	// 			// engine.Log.Info("本组有块了 group:%d height:%d", one.Block.Group.Height, one.Block.Height)
	// 			continue
	// 		}
	// 		// engine.Log.Info("返回这个见证人见证人one:%+v", one)
	// 		return false
	// 	}
	// }
	// engine.Log.Info("当前见证人组高度:%d", currentGroup.Height)
	for _, one := range currentGroup.Witness {
		// engine.Log.Info("对比见证人地址：%s %s %+v", one.Addr.B58String(), bhvo.BH.Witness.B58String(), one.Block)
		if bytes.Equal(*one.Addr, bhvo.BH.Witness) {
			if one.Block != nil {
				// engine.Log.Info("this block is exist groupHeight:%d blockGroup:%d blockHeight:%d", currentGroup.Height, one.Block.Group.Height, one.Block.Height)
				return true
			} else {
				return false
			}
		}
	}
	return false
}

/*
检查区块是否分叉
@return    *Witness    从此见证人开始分叉
@return    bool        是否有分叉
*/
func (this *WitnessChain) CheckBifurcationBlock(groupHeight, blockHeight uint64, preBlockHash []byte) (*Witness, bool, bool) {
	return nil, false, false
}

/*
通过新区块，在未出块的见证人组中找到这个见证人
@return    *Witness    找到的见证人
@return    bool        是否需要同步
*/
func (this *WitnessChain) FindWitnessForBlock(bhvo *BlockHeadVO) (*Witness, bool) {
	// engine.Log.Info("FindWitnessForBlock:%+v", bhvo)
	var witness *Witness
	for group := this.WitnessGroup; group != nil; group = group.NextGroup {
		// engine.Log.Info("设置已出的块 2222222222222222222 %d", group.Height)
		if group.Height < bhvo.BH.GroupHeight {
			// engine.Log.Info("设置已出的块 333333333333333333 %d", group.Height)
			if group.NextGroup != nil {
				continue
			}
			//pl time
			if config.TimeNow().Unix() < bhvo.BH.Time {
				//if config.TimeNow().UnixNano() < time.Unix(bhvo.BH.Time, 0).UnixNano() {
				continue
			}
			// engine.Log.Info("开始统计之前的区块 %v %v", this.chain.SyncBlockFinish, bhvo.FromBroadcast)
			//如果当前同步还没完成，并且收到广播的区块，不能导入
			// if !this.chain.SyncBlockFinish && !bhvo.FromBroadcast {
			// 	//先统计之前的区块
			// 	for buildGroup := group; buildGroup != nil && buildGroup.BlockGroup == nil; buildGroup = buildGroup.PreGroup {
			// 		buildGroup.BuildGroup()
			// 	}
			// 	this.CompensateWitnessGroupByGroupHeight(bhvo.BH.GroupHeight)
			// }
		}
		if group.Height > bhvo.BH.GroupHeight {
			// engine.Log.Info("设置已出的块 4444444444444444444 %d", group.Height)
			// engine.Log.Warn("不能导入之前已经确认的块")
			return nil, false
		}

		// if bhvo.BH.Height < config.WitnessOrderCorrectEnd && bhvo.BH.Height > config.WitnessOrderCorrectStart {
		// 	// engine.Log.Info("在范围内")
		// 	//找到组中对应的见证人，此高度内，随便分配个见证人给他
		// 	for _, one := range group.Witness {
		// 		// engine.Log.Info("见证人one:%+v", one)
		// 		if one.Block != nil {
		// 			// engine.Log.Info("本组有块了 group:%d height:%d", one.Block.Group.Height, one.Block.Height)
		// 			continue
		// 		}
		// 		// engine.Log.Info("返回这个见证人见证人one:%+v", one)
		// 		//找到这个见证人了
		// 		witness = one
		// 		// engine.Log.Info("找到这个见证人了")
		// 		break
		// 	}
		// } else {
		// engine.Log.Info("设置已出的块 55555555555555555 %d", group.Height)
		for _, one := range group.Witness {
			//engine.Log.Info("设置已出的块，对比一下 %s %s", bhvo.BH.Witness.B58String(), one.Addr.B58String())
			if !bytes.Equal(bhvo.BH.Witness, *one.Addr) {
				// fmt.Println("-=-=-=-=-=对比下一个1", one.Block, one.BlockHeight, one.Group.Height, bhvo.BH.Witness.B58String(), one.Addr.B58String())
				continue
			}
			now := config.TimeNow().Unix()
			//pl time
			//now := config.TimeNow().UnixNano()
			//是未来的一个时间，直接退出
			//if one.CreateBlockTime > now+config.Mining_block_time {
			if one.CreateBlockTime > now+int64(config.Mining_block_time.Seconds()) {

				engine.Log.Warn("是未来的一个时间，直接退出 预约%s 出块%s 当前%s", time.Unix(one.CreateBlockTime, 0).Format("2006-01-02 15:04:05"),
					time.Unix(bhvo.BH.Time, 0).Format("2006-01-02 15:04:05"), time.Unix(now, 0).Format("2006-01-02 15:04:05"))

				break
			}

			// engine.Log.Info("设置已出的块 666666666666666 %d", group.Height)

			//找到这个见证人了
			witness = one
			// engine.Log.Info("找到这个见证人了")
			break
		}
		// }

		if witness != nil {
			// engine.Log.Info("找到这个见证人了 退出")
			break
		}
	}
	return witness, false
}

/*
设置见证人生成的块
只能设置当前组，不能设置其他组
当本组所有见证人都出块了，将当前组见证人的变量指针修改为下一组见证人
@return    bool    是否设置成功
*/
func (this *WitnessChain) SetWitnessBlock(bhvo *BlockHeadVO) bool {

	// start := config.TimeNow()
	// engine.Log.Info("Set blocks 1111111111111111111 %d %d %s", bhvo.BH.GroupHeight, bhvo.BH.Height, time.Unix(bhvo.BH.Time, 0).Format("2006-01-02 15:04:05"))

	//避免最新见证人组中区块未确认，导致导入区块不连续，因此在导入区块之前，先统计之前的区块组
	// if !this.chain.SyncBlockFinish {

	// }

	// engine.Log.Info("设置已出的块 111 %+v", bhvo.BH)
	// engine.Log.Info("此区块见证人 %s", bhvo.BH.Witness.B58String())
	//找到这个出块的见证人
	witness, needSync := this.FindWitnessForBlock(bhvo)
	if witness != nil && witness.Block != nil {
		//已经设置了就不需要重复设置了
		engine.Log.Warn("You don't need to set it again if it's already set")
		return false
	}
	// engine.Log.Info("Witness group found:%d height:%d", witness.Group.Height, witness.BlockHeight)
	// engine.Log.Info("SetWitnessBlock 11111111 %s", config.TimeNow().Sub(start))
	//找到见证人了，区块高度又是连续的，如果组高度增加，则统计之前的组

	// engine.Log.Info("开始设置见证人出块 33333333333")
	if witness == nil {
		//没有找到这个见证人
		engine.Log.Warn("No witness found")

		if needSync {
			//从邻居节点同步
			this.chain.NoticeLoadBlockForDB()
		}
		return false
	}
	// engine.Log.Info("开始设置见证人出块 444444444444")

	if !bhvo.BH.CheckBlockHead(witness.Puk) {
		//区块验证不通过，区块不合法
		engine.Log.Warn("Block verification failed, block is illegal group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		// this.chain.NoticeLoadBlockForDB(false)
		return false
	}

	// engine.Log.Info("SetWitnessBlock 22222222222 %s", config.TimeNow().Sub(start))

	// engine.Log.Info("开始设置见证人出块 55555555555555")
	if bhvo.BH.Height != config.Mining_block_start_height {
		//查询排除的交易
		// excludeTx := make([]config.ExcludeTx, 0)
		// for i, one := range config.Exclude_Tx {
		// 	if bhvo.BH.Height == one.Height {
		// 		excludeTx = append(excludeTx, config.Exclude_Tx[i])
		// 	}
		// }

		//查找出未确认的块
		preBlock, blocks := witness.CheckUnconfirmedBlock(&bhvo.BH.Previousblockhash)

		//检查当前区块高度是否连续
		if bhvo.BH.Height-1 != preBlock.Height {
			engine.Log.Warn("block height not continuity height:%d inport height:%d", preBlock.Height, bhvo.BH.Height)
			engine.Log.Warn("%+v %+v", witness, bhvo.BH)
			return false
		}
		// engine.Log.Info("SetWitnessBlock 3333333 %s", config.TimeNow().Sub(start))
		//检查区块中的交易是否正确
		for _, one := range bhvo.Txs {
			//排除的交易不验证
			// for _, two := range excludeTx {
			// 	if bytes.Equal(two.TxByte, *one.GetHash()) {
			// 		continue
			// 	}
			// }

			err := one.CheckLockHeight(bhvo.BH.Height)
			if err != nil {
				engine.Log.Error("Illegal transaction 111 %s %s", hex.EncodeToString(*one.GetHash()), err.Error())
				//交易不合法
				return false
			}
			// err = one.CheckFrozenHeight(bhvo.BH.Height, bhvo.BH.Time)
			// if err != nil {
			// 	engine.Log.Error("Illegal transaction 222 %s %s", hex.EncodeToString(*one.GetHash()), err.Error())
			// 	//交易不合法
			// 	return false
			// }

			//自己的未打包交易，是已经验证过的合法交易，已经验证过就不需要重复验证了
			// _, ok := this.chain.transactionManager.unpacked.Load(hex.EncodeToString(*one.GetHash()))
			// _, ok := this.chain.transactionManager.unpacked.Load(utils.Bytes2string(*one.GetHash()))
			// if ok {
			// 	continue
			// }
			if this.chain.TransactionManager.unpackedTransaction.ExistTxByAddrTxid(one) {
				continue
			}

			// engine.Log.Info("SetWitnessBlock 444444444 %s", config.TimeNow().Sub(start))

			if bhvo.BH.Height > config.Mining_block_start_height+config.Mining_block_start_height_jump {

				err = one.CheckSign()
				// 	runtime.GC()
				if err != nil {
					engine.Log.Error("Illegal transaction 333 %s %s", hex.EncodeToString(*one.GetHash()), err.Error())
					//交易不合法
					return false
				}
			}

			// engine.Log.Info("SetWitnessBlock 55555555 %s", config.TimeNow().Sub(start))
		}
		// if len(bhvo.Txs) > 0 {
		// 	runtime.GC()
		// }
		// engine.Log.Info("SetWitnessBlock 66666666666 %s", config.TimeNow().Sub(start))

		//检查未确认的块中的交易是否正确
		//未确认的交易
		unacknowledgedTxs := make([]TxItr, 0)
		//排除已经打包的交易
		exclude := make(map[string]string)
		for _, one := range blocks {
			// engine.Log.Info("设置见证人生成的块 SetWitnessBlock")
			_, txs, err := one.LoadTxs()
			if err != nil {
				engine.Log.Warn("not find transaction %s", err.Error())
				//找不到这个交易
				return false
			}
			for _, txOne := range *txs {
				// exclude[hex.EncodeToString(*txOne.GetHash())] = ""
				exclude[utils.Bytes2string(*txOne.GetHash())] = ""
				unacknowledgedTxs = append(unacknowledgedTxs, txOne)
			}
		}

		// engine.Log.Info("SetWitnessBlock 44444444444444 %s", config.TimeNow().Sub(start))

		//预执行交易evm
		vmRun := evm.NewCountVmRun(nil)
		vmRun.SetStorage(nil)

		gasTotal := uint64(0)
		sizeTotal := uint64(0) //保存区块所有交易大小
		for i, one := range bhvo.Txs {
			//判断重复的交易
			if !one.CheckRepeatedTx(unacknowledgedTxs...) {
				engine.Log.Warn("Transaction verification failed")
				//交易验证不通过
				return false
			}

			//判断合约交易并预执行
			if one.Class() == config.Wallet_tx_type_contract {
				from := (*one.GetVin())[0].GetPukToAddr()
				to := (*one.GetVout())[0].Address
				vmRun.SetTxContext(*from, to, *one.GetHash(), bhvo.BH.Time, bhvo.BH.Height, nil, nil)
				if err := one.(*Tx_Contract).PreExecV1(vmRun); err != nil {
					engine.Log.Error("Contract Transaction preExec failed %s %s", hex.EncodeToString(*one.GetHash()), err.Error())
					return false
				}
			}

			//验证域名
			if !one.CheckDomain() {
				engine.Log.Error("Transaction check domain failed %s", hex.EncodeToString(*one.GetHash()))
				return false
			}

			gasTotal += one.GetGasUsed()

			unacknowledgedTxs = append(unacknowledgedTxs, bhvo.Txs[i])
			if one.Class() == config.Wallet_tx_type_mining {
				continue
			}
			sizeTotal = sizeTotal + uint64(len(*one.Serialize()))
		}
		//判断交易总大小
		if sizeTotal > config.Block_size_max {
			engine.Log.Error("此包大小 %d 交易总量 %d", sizeTotal, len(bhvo.Txs))
			engine.Log.Warn("Transaction over size %d", sizeTotal)
			//这里用continue是因为，排在后面的交易有可能占用空间小，可以打包到区块中去，使交易费最大化。
			return false
		}

		//判断gas总大小
		if gasTotal > config.BLOCK_TOTAL_GAS {
			engine.Log.Warn("Transaction over gas total%d", gasTotal)
			return false
		}

		// engine.Log.Info("SetWitnessBlock 555555555555 %s", config.TimeNow().Sub(start))

		//检查区块奖励是否正确
		if bhvo.BH.Height > config.Mining_block_start_height+config.Mining_block_start_height_jump {

			if witness.WitnessBigGroup != preBlock.Witness.WitnessBigGroup {
				//对比vouts中分配比例是否正确
				haveReward := false //标记是否有区块奖励
				for _, one := range bhvo.Txs {
					if one.Class() != config.Wallet_tx_type_mining {
						continue
					}
					if haveReward {
						//如果一个块里有多个奖励交易，则不合法
						engine.Log.Warn("Illegal if there are multiple reward transactions in a block")
						return false
					}
					haveReward = true

					//对比奖励是否正确
					m := make(map[string]uint64) //key:string=奖励地址;value:uint64=奖励金额;
					for _, one := range *one.GetVout() {
						m[utils.Bytes2string(one.Address)+strconv.Itoa(int(one.FrozenHeight))] = one.Value
					}

					//虚拟机使能
					//if config.EVM_Reward_Enable {
					//走合约
					//vouts, _ = preBlock.witness.WitnessBigGroup.BuildRewardContractVouts(blocks, bhvo.BH.Height, &bhvo.BH.Previousblockhash, preBlock)
					//} else {
					totalReward, allVote, witnesses, witRatios := preBlock.Witness.WitnessBigGroup.CalculateBlockRewardAndWitnesses(blocks, bhvo.BH.Height, &bhvo.BH.Previousblockhash, preBlock)
					//vouts = preBlock.witness.WitnessBigGroup.BuildRewardVoutsV2(totalReward, allVote, witnesses, witRatios)
					//奖励分配验证
					vouts, tmpVouts := preBlock.Witness.WitnessBigGroup.BuildRewardVoutsV3(this.chain, totalReward, allVote, witnesses, witRatios)
					//}

					vouts = MergeVouts(&vouts)

					for _, one := range vouts {
						// engine.Log.Info("%+v", one)
						value, ok := m[utils.Bytes2string(one.Address)+strconv.Itoa(int(one.FrozenHeight))]
						if !ok {
							//没有这个人的奖励，则验证不通过
							engine.Log.Warn("Without this person's reward, the verification fails %s", one.Address.B58String())
							return false
						}
						if value != one.Value {
							//奖励数额不正确，则验证不通过
							engine.Log.Warn("If the reward amount is incorrect, the verification fails %s %d want:%d", one.Address.B58String(), value, one.Value)
							return false
						}
					}
					witAndRoleRewardSet.Store(utils.Bytes2string(bhvo.BH.Hash), tmpVouts)
				}
				if !haveReward {
					//如果没有区块奖励，则区块不合法
					engine.Log.Warn("If there is no block reward, the block is illegal")
					return false
				}

			} else {
				//判断每组不能有多个区块奖励交易
				for _, one := range bhvo.Txs {
					if one.Class() == config.Wallet_tx_type_mining {
						engine.Log.Warn("每组不能有多个区块奖励交易")
						return false
					}
				}
			}
		}
	}
	// engine.Log.Info("SetWitnessBlock 6666666666666 %s", config.TimeNow().Sub(start))

	// engine.Log.Info("开始设置见证人出块 66666666666666")
	//找到见证人了，不管这个见证人有没有开始出块，给他发送一个停止出块的信号
	select {
	case witness.StopMining <- false:
	default:
	}
	//创建新的块
	newBlock := new(Block)
	newBlock.Id = bhvo.BH.Hash
	newBlock.Height = bhvo.BH.Height
	newBlock.PreBlockID = bhvo.BH.Previousblockhash
	// engine.Log.Info("创建新的区块:%s height:%d", hex.EncodeToString(newBlock.Id), newBlock.Height)

	//找到了见证人，将见证人标记为已经出块
	witness.Block = newBlock
	witness.Block.Witness = witness

	//整理已出区块之间的前后关系
	// witness.Group.CollationRelationship(nil)
	witness.Group.SelectionChainOld()

	// engine.Log.Info("SetWitnessBlock 7777777777777 %s", config.TimeNow().Sub(start))

	// engine.Log.Info("找到这个见证人了 222 %d %s %v", witness.Group.Height, witness.Addr.B58String(), witness.Block)

	//如果是创始区块，则设置见证人的出块时间
	if bhvo.BH.Height == config.Mining_block_start_height {
		witness.CreateBlockTime = bhvo.BH.Time
	}

	this.chain.SetPulledStates(bhvo.BH.Height)
	// engine.Log.Info("开始设置见证人出块 777777777777777")
	return true

}

/*
	构建本组中的见证人出块奖励
	按股权分配
	只有见证人方式出块才统计
	组人数乘以每块奖励，再分配给实际出块的人
*/
// func (this *WitnessSmallGroup) CountReward(blockHeight uint64) *Tx_reward {

// 	vouts := make([]*Vout, 0)

// 	//统计本组股权和交易手续费
// 	witnessPos := uint64(0) //见证人押金
// 	votePos := uint64(0)    //投票者押金
// 	allPos := uint64(0)     //股权数量
// 	allGas := uint64(0)     //计算交易手续费
// 	allReward := uint64(0)  //本组奖励数量
// 	// txs := make([]TxItr, 0)
// 	for _, one := range this.Witness {
// 		if one.Block == nil {
// 			continue
// 		}

// 		//计算交易手续费
// 		// engine.Log.Info("构建本组中的见证人出块奖励 CountReward")
// 		_, txs, _ := one.Block.LoadTxs()
// 		for _, one := range *txs {
// 			allGas = allGas + one.GetGas()
// 		}

// 		//计算股权
// 		allPos = allPos + (one.Score * 2) //计算股权的时候，见证人的股权要乘以2
// 		witnessPos = witnessPos + one.Score
// 		for _, vote := range one.Votes {
// 			allPos = allPos + vote.Scores
// 			votePos = votePos + vote.Scores
// 		}

// 		//计算区块奖励，第一个块产出80个币
// 		//每增加一定块数，产出减半，直到为0
// 		//最多减半9次，第10次减半后产出为0
// 		//		oneReward := uint64(config.Mining_reward)
// 		//		if one.Block.Height <= config.Mining_lastblock_reward {
// 		//			n := one.Block.Height / config.Mining_block_cycle
// 		//			for i := uint64(0); i < n; i++ {
// 		//				oneReward = oneReward / 2
// 		//			}
// 		//		} else {
// 		//			oneReward = 0
// 		//		}

// 		//按照发行总量及减半周期计算出块奖励
// 		oneReward := config.ClacRewardForBlockHeight(one.Block.Height)
// 		allReward = allReward + oneReward
// 	}

// 	//--------------所有交易手续费分给云存储节点---------------
// 	// cloudReward := uint64(0)
// 	nameinfo := name.FindNameToNet(config.Name_store)
// 	if nameinfo != nil && len(nameinfo.AddrCoins) != 0 {
// 		addrCoin := nameinfo.AddrCoins[utils.GetRandNum(int64(len(nameinfo.AddrCoins)))]
// 		// cloudReward = uint64(float64(allReward) * 0.8)
// 		vout := Vout{
// 			Value:   allGas,
// 			Address: addrCoin,
// 		}
// 		vouts = append(vouts, &vout)
// 		// allReward = allReward - cloudReward
// 	}

// 	//---------------------------------------------------

// 	// allReward = allReward + allGas

// 	//计算见证人奖励
// 	// witnessRatio := int64(config.Mining_reward_witness_ratio * 100)
// 	// witnessReward := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(witnessRatio))
// 	// witnessReward = new(big.Int).Div(witnessReward, big.NewInt(100))
// 	countReward := uint64(0)
// 	for _, one := range this.Witness {
// 		//分配奖励是所有见证人组成员都要分配
// 		temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(one.Score*2)))
// 		value := new(big.Int).Div(temp, big.NewInt(int64(allPos)))
// 		//奖励为0的矿工交易不写入区块
// 		if value.Uint64() <= 0 {
// 			continue
// 		}
// 		vout := Vout{
// 			Value:   value.Uint64(),
// 			Address: *one.Addr,
// 		}
// 		vouts = append(vouts, &vout)
// 		countReward = countReward + value.Uint64()
// 		//给投票者分配奖励
// 		for _, two := range one.Votes {
// 			temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(two.Scores)))
// 			value := new(big.Int).Div(temp, big.NewInt(int64(allPos)))
// 			//奖励为0的矿工交易不写入区块
// 			if value.Uint64() <= 0 {
// 				continue
// 			}
// 			vout := Vout{
// 				Value:   value.Uint64(),
// 				Address: *two.Addr,
// 			}
// 			vouts = append(vouts, &vout)
// 			countReward = countReward + value.Uint64()
// 		}
// 	}
// 	//平均数不能被整除时候，剩下的给最后一个出块的见证人
// 	if len(vouts) > 0 {
// 		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (allReward - countReward)
// 	}

// 	// //计算投票者的奖励
// 	// voteReward := allReward - witnessReward.Uint64()
// 	// countReward = uint64(0)
// 	// for _, one := range this.Witness {
// 	// 	for _, two := range one.Votes {
// 	// 		temp := new(big.Int).Mul(big.NewInt(int64(voteReward)), big.NewInt(int64(two.Score)))
// 	// 		value := new(big.Int).Div(temp, big.NewInt(int64(votePos)))
// 	// 		//奖励为0的矿工交易不写入区块
// 	// 		if value.Uint64() <= 0 {
// 	// 			continue
// 	// 		}
// 	// 		vout := Vout{
// 	// 			Value:   value.Uint64(),
// 	// 			Address: *two.Addr,
// 	// 		}
// 	// 		vouts = append(vouts, vout)
// 	// 		countReward = countReward + value.Uint64()
// 	// 	}
// 	// }
// 	// //平均数不能被整除时候，剩下的给最后一个投票者
// 	// vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (voteReward - countReward)

// 	//构建输入
// 	baseCoinAddr := keystore.GetCoinbase()
// 	// puk, ok := keystore.GetPukByAddr(baseCoinAddr)
// 	// if !ok {
// 	// 	return nil
// 	// }
// 	vins := make([]*Vin, 0)
// 	vin := Vin{
// 		Puk:  baseCoinAddr.Puk, //公钥
// 		Sign: nil,              //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
// 	}
// 	vins = append(vins, &vin)

// 	var txReward *Tx_reward
// 	for i := uint64(0); i < 10000; i++ {
// 		base := TxBase{
// 			Type:       config.Wallet_tx_type_mining, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
// 			Vin_total:  1,
// 			Vin:        vins,
// 			Vout_total: uint64(len(vouts)), //输出交易数量
// 			Vout:       vouts,              //交易输出
// 			LockHeight: blockHeight + i,    //锁定高度
// 			//		CreateTime: config.TimeNow().Unix(),            //创建时间
// 		}
// 		txReward = &Tx_reward{
// 			TxBase: base,
// 		}

// 		//给输出签名，防篡改
// 		for i, one := range txReward.Vin {
// 			for _, key := range keystore.GetAddrAll() {

// 				puk, ok := keystore.GetPukByAddr(key.Addr)
// 				if !ok {
// 					return nil
// 				}

// 				if bytes.Equal(puk, one.Puk) {
// 					_, prk, _, err := keystore.GetKeyByAddr(key.Addr, config.Wallet_keystore_default_pwd)
// 					// prk, err := key.GetPriKey(pwd)
// 					if err != nil {
// 						return nil
// 					}
// 					sign := txReward.GetSign(&prk, one.Txid, one.Vout, uint64(i))
// 					//				sign := pay.GetVoutsSign(prk, uint64(i))
// 					txReward.Vin[i].Sign = *sign
// 				}
// 			}
// 		}

// 		txReward.BuildHash()
// 		if txReward.CheckHashExist() {
// 			txReward = nil
// 			continue
// 		} else {
// 			break
// 		}
// 	}
// 	return txReward
// }

/*
判断是否是本组首个见证人出块
*/
func (this *WitnessSmallGroup) FirstWitness() bool {
	for _, one := range this.Witness {
		if one.Block != nil {
			return false
		}
	}
	return true
}

/*
分配出块奖励
*/
func (this *WitnessSmallGroup) DistributionRewards() {

}

/*
查询见证人是否在备用见证人列表中
*/
func (this *WitnessChain) FindWitness(addr crypto.AddressCoin) bool {
	if this.WitnessGroup == nil {
		return false
	}
	witnessTemp := this.WitnessGroup.Witness[0]
	for {
		witnessTemp = witnessTemp.NextWitness
		if witnessTemp == nil || witnessTemp.Group == nil {
			break
		}
		if bytes.Equal(*witnessTemp.Addr, addr) {
			return true
		}
	}
	return false
}

func (this *WitnessChain) FindWitnessByAddr(addr crypto.AddressCoin) *Witness {
	if this.WitnessGroup == nil {
		return nil
	}
	witnessTemp := this.WitnessGroup.Witness[0]
	for {
		witnessTemp = witnessTemp.NextWitness
		if witnessTemp == nil || witnessTemp.Group == nil {
			break
		}
		if bytes.Equal(*witnessTemp.Addr, addr) {
			return witnessTemp
		}
	}
	return nil
}

/*
见证人定时同步区块
*/
func (this *Witness) syncBlockTiming() {
	//同步区块没有完成则不定时同步
	if !forks.GetLongChain().SyncBlockFinish {
		return
	}
	if this.syncBlockOnce != nil {
		return
	}
	this.syncBlockOnce = new(sync.Once)
	var syncBlock = func() {
		utils.Go(func() {
			goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
			_, file, line, _ := runtime.Caller(0)
			engine.AddRuntime(file, line, goroutineId)
			defer engine.DelRuntime(file, line, goroutineId)
			// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
			// engine.Log.Info("add Goroutine:%s", goroutineId)
			// defer engine.Log.Info("del Goroutine:%s", goroutineId)
			bfw := BlockForWitness{
				GroupHeight: this.Group.Height, //见证人组高度
				Addr:        *this.Addr,        //见证人地址
			}
			//pl time
			now := config.TimeNow().Unix()
			//now := config.TimeNow().UnixNano()
			// engine.Log.Info("查看时间 %d %d %d", now, this.CreateBlockTime, now-this.CreateBlockTime)

			if this.CreateBlockTime < now {
				intervalTime := now - this.CreateBlockTime
				//if intervalTime > config.Mining_block_time*2 {
				if intervalTime > int64(config.Mining_block_time.Seconds())*2 {
					// engine.Log.Info("时间太久远了，就不需要添加定时同步区块了")
					return
				}
			}

			waitTime := this.CreateBlockTime - config.TimeNow().Unix()
			//waitTime := this.CreateBlockTime - config.TimeNow().UnixNano()
			//
			if waitTime < 0 {
				waitTime = 0
			}

			//给等待同步时间设置一个随机，不要所有节点同一时间开始同步
			//pl time
			//delayTime_min := time.Duration(config.Mining_block_time * time.Second / 3) //延迟同步最小时间为出块时间的1/3
			//delayTime_max := time.Duration(config.Mining_block_time * time.Second / 2) //延迟同步最长时间为出块时间的一半
			delayTime_min := config.Mining_block_time / 3 //延迟同步最小时间为出块时间的1/3
			delayTime_max := config.Mining_block_time / 2 //延迟同步最长时间为出块时间的一半
			delayTime := delayTime_min + time.Duration(utils.GetRandNum(int64(delayTime_max-delayTime_min)))
			//delayTime = (time.Duration(waitTime) * time.Second) + delayTime
			delayTime = time.Duration(waitTime) + delayTime

			engine.Log.Info("Groups %d Time to wait for synchronization %d %s", this.Group.Height, waitTime, delayTime)
			// time.Sleep((time.Duration(waitTime) * time.Second) + (time.Second * 4)) //加n秒，n秒钟后再同步
			time.Sleep(delayTime) //加n秒，n秒钟后再同步

			intervalTime := time.Second //同步时间间隔
			//intervalTotal := ((config.Mining_block_time * time.Second) - delayTime) / intervalTime //间隔次数
			intervalTotal := (config.Mining_block_time - delayTime) / intervalTime //间隔次数
			// engine.Log.Info("%d 间隔时间 %d   间隔次数 %d", this.Group.Height, intervalTime, intervalTotal)

			this.syncBlock(int(intervalTotal), intervalTime, &bfw)

			// engine.Log.Info("%d 同步 end", this.Group.Height)

		}, nil)
	}
	this.syncBlockOnce.Do(syncBlock)
}

/*
见证人同步区块
@total           uint64           总共同步多少次
@intervalTime    time.Duration    同步失败间隔时间
*/
func (this *Witness) syncBlock(total int, intervalTime time.Duration, bfw *BlockForWitness) {
	bs, err := bfw.Proto()
	engine.Log.Info("qlw-----bs length:%d", len(*bs))
	// bs, err := json.Marshal(bfw)
	if err != nil {
		return
	}
	for i := int64(0); i < int64(total); i++ {
		if this.Block != nil {
			// engine.Log.Info("%d 这个区块已经有了，不需要同步了", this.Group.Height)
			return
		}
		//开始从邻居节点同步区块
		broadcasts := append(Area.NodeManager.GetLogicNodes(), Area.NodeManager.GetProxyAll()...)
		// engine.Log.Info("%d 邻居节点个数 %d", this.Group.Height, len(broadcasts))
		for j, _ := range broadcasts {
			if this.Block != nil {
				// engine.Log.Info("%d 这个区块已经有了，不需要同步了", this.Group.Height)
				return
			}
			engine.Log.Info("%d Synchronize blocks from neighbor nodes %s", this.Group.Height, broadcasts[j].B58String())
			// engine.Log.Info("%d 555555555555555555555", this.Group.Height)
			//pl time
			//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getblockforwitness, &broadcasts[j], bs, config.Mining_block_time*time.Second)
			bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getblockforwitness, &broadcasts[j], bs, config.Mining_block_time)
			if err != nil {
				continue
			}
			// message, err := message_center.SendNeighborMsg(config.MSGID_getblockforwitness, &broadcasts[j], bs)
			// if err != nil {
			// 	// engine.Log.Info("这个邻居节点消息发送不成功")
			// 	continue
			// }
			// // engine.Log.Info("%d 66666666666666666666", this.Group.Height)
			// // bs := flood.WaitRequest(config.CLASS_wallet_getblockforwitness, hex.EncodeToString(message.Body.Hash), 1)
			// bs, _ := flood.WaitRequest(config.CLASS_wallet_getblockforwitness, utils.Bytes2string(message.Body.Hash), 1)
			if bs == nil {
				// engine.Log.Warn("%d 收到查询区块回复消息超时  %s", this.Group.Height, broadcasts[j].B58String())
				// lock.Unlock()
				continue
			}
			// engine.Log.Info("%d 收到查询区块回复消息  %s", this.Group.Height, broadcasts[j].B58String())
			//导入区块
			bhVO, err := ParseBlockHeadVOProto(bs)
			// bhVO, err := ParseBlockHeadVO(bs)
			if err != nil {
				// engine.Log.Warn("%d 收到查询区块回复消息 error: %s", this.Group.Height, err.Error())
				continue
			}
			bhVO.FromBroadcast = true
			forks.GetLongChain().AddBlockOther(bhVO)
			// engine.Log.Info("%d 8888888888888888888", this.Group.Height)
			//等待导入
			//			time.Sleep(intervalTime)
			//			break
		}

		time.Sleep(intervalTime)
	}
}

/*
回收内存，将前n个见证人之前的见证人链删除。
n不是一个精确的数，大致是config.Mining_block_hash_count参数的3倍
*/
func (this *WitnessChain) GCWitnessOld() {

	//做见证人的节点，不做社区节点，社区节点不能删除之前的区块，见证人为了减少内存，需要删除之前的区块
	//新版本：见证人节点,社区节点,轻节点均可以回收,所以不设置条件了
	//IsBackup := this.chain.WitnessChain.FindWitness(Area.Keystore.GetCoinbase().Addr)
	//if !IsBackup {
	//	return
	//}

	total := config.Mining_block_hash_count * 3
	currentWitnessGroup := this.WitnessGroup.PreGroup
	for {
		//没有前置见证人组，则退出循环
		if currentWitnessGroup == nil {
			break
		}

		for _, one := range currentWitnessGroup.Witness {
			if one.Block == nil {
				continue
			}
			total = total - 1
			//找到足够数量之前的见证人组，则退出循环
			if total <= 0 {
				break
			}
		}

		//找到足够数量之前的见证人组，则退出循环
		if total <= 0 {
			break
		}

		currentWitnessGroup = currentWitnessGroup.PreGroup
	}
	if currentWitnessGroup == nil {
		return
	}

	//开始释放引用，回收内存
	currentWitnessGroup.PreGroup = nil
	if currentWitnessGroup.BlockGroup != nil {
		currentWitnessGroup.BlockGroup.PreGroup = nil
	}

	firstWitness := currentWitnessGroup.Witness[0]
	if firstWitness.Block != nil {
		if firstWitness.Block.Witness != nil {
			firstWitness.Block.Witness.PreWitness = nil
		}
		if firstWitness.Block.Group != nil {
			firstWitness.Block.Group.PreGroup = nil
		}
		firstWitness.Block.PreBlock = nil
	}
	firstWitness.PreWitness = nil
	if firstWitness.Group != nil {
		firstWitness.Group.PreGroup = nil
	}

}

/*
回收内存，将前n个见证人之前的见证人链删除。
n不是一个精确的数，大致是config.Mining_block_hash_count参数的3倍
*/
func (this *WitnessChain) GCWitnessV2() {
	_, lastBlock := this.chain.GetLastBlock()
	if lastBlock == nil {
		return
	}
	//由lastBlock向前找300个块
	for i := 0; i < config.Mining_block_hash_count*3; i++ {
		if lastBlock.PreBlock == nil {
			return
		}
		lastBlock = lastBlock.PreBlock
	}

	//通过lastBlock找出Group
	var currentWitnessGroup *WitnessSmallGroup
	if lastBlock.Witness != nil {
		currentWitnessGroup = lastBlock.Witness.Group
	}

	if currentWitnessGroup == nil {
		return
	}

	//开始释放引用，回收内存
	currentWitnessGroup.PreGroup = nil
	if currentWitnessGroup.BlockGroup != nil {
		currentWitnessGroup.BlockGroup.PreGroup = nil
	}

	firstWitness := currentWitnessGroup.Witness[0]
	if firstWitness.Block != nil {
		if firstWitness.Block.Witness != nil {
			firstWitness.Block.Witness.PreWitness = nil
		}
		if firstWitness.Block.Group != nil {
			firstWitness.Block.Group.PreGroup = nil
		}
		firstWitness.Block.PreBlock = nil
	}
	firstWitness.PreWitness = nil
	if firstWitness.Group != nil {
		firstWitness.Group.PreGroup = nil
	}
}

/*
回收内存，将前n个见证人之前的见证人链删除。
n不是一个精确的数，大致是config.Mining_block_hash_count参数的3倍
*/
func (this *WitnessChain) GCWitness() {
	if this.chain.CurrentBlock < config.ChainGCModeHeight {
		this.GCWitnessOld()
	} else {
		this.GCWitnessV2()
	}
}

/*
*
获取当前出块组的最后一个见证人
*/
func (this *WitnessChain) GetCurrGroupLastWitness() *Witness {
	last := this.WitnessGroup.Witness[len(this.WitnessGroup.Witness)-1]

	for {
		if last.WitnessBigGroup != nil {
			break
		}
		last = last.PreWitness
	}

	return last
}

func (this *WitnessChain) GetCurrGroup() *WitnessSmallGroup {
	return this.WitnessGroup
}

type WitnessSort []*Witness

func (e WitnessSort) Len() int {
	return len(e)
}

func (e WitnessSort) Less(i, j int) bool {
	return e[i].VoteNum > e[j].VoteNum
}

func (e WitnessSort) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

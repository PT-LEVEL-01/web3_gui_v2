package mining

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"web3_gui/chain/config"
	lightsnapshot "web3_gui/chain/light/snapshot"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/libp2parea/adapter/engine"
)

// 初始化主链快照
func InitLightChainSnap() {
	// 添加需要快照的对象
	// 主链Chain
	lightsnapshot.Add(GetLongChain())
	// 快照起始高度
	lightsnapshot.StartHeightAt()
}

// 轻节点启动方式加载快照
func StartLightChainSnap() error {
	engine.Log.Info(">>>>>> Light Snapshot[加载] 轻节点方式加载快照")
	BuildSnapshotFirstLightChain()

	// 初始化轻节点快照
	InitLightChainSnap()

	if err := lightsnapshot.Load(); err != nil {
		engine.Log.Info(">>>>>> Light Snapshot[加载] 轻节点加载快照失败  %v", err)
		return err
	}

	return nil
}

// ////////////////////////////// 实现快照接口 ////////////////////////////////
func (this *Chain) LightSnapshotName() string {
	return "light_snapshot_chain"
}

// 序列化见证人链快照碎片
func (this *Chain) LightSnapshotSerialize() ([]byte, error) {
	b, err := this.lightSnapshotSerialize()
	return b, err
}

// 链快照还原Chain
func (this *Chain) LightSnapshotDeSerialize(data []byte) error {
	startAt := config.TimeNow()
	defer engine.Log.Info(">>>>>> Snapshot[反序列化] 主链反序列化耗时(ms):%v,字节:%d", config.TimeNow().Sub(startAt), len(data))

	return this.lightSnapshotDeSerialize(data)
}

// 序列化见证人链快照碎片
func (this *Chain) lightSnapshotSerialize() ([]byte, error) {
	snapshotChain := &SnapshotChain{}
	snapshotWitnessChain := &go_protos.SnapshotWitnessChain{
		WitnessNotGroup:    []*go_protos.SnapshotWitness{},
		AllWitnessNotGroup: make(map[string]*go_protos.SnapshotWitness),
		BigGroupNotGroup:   &go_protos.SnapshotWitnessBigGroup{},
		BigGroup:           []*go_protos.SnapshotWitnessBigGroup{},
		SmallGroup:         []*go_protos.SnapshotWitnessSmallGroup{},
		Group:              []*go_protos.SnapshotGroup{},
		Witness:            []*go_protos.SnapshotWitness{},
		Block:              []*go_protos.SnapshotBlock{},
	}

	// 最后一个有效块
	var lastblock *Block
	// 最后一个有效组
	var lastgroup *Group
	// 最后一个有效小组
	var lastsmallgroup *WitnessSmallGroup
	// 最后一个有效大组
	var lastbiggroup *WitnessBigGroup
	var lastbiggrouptmp *go_protos.SnapshotWitnessBigGroup
	// 最后一个有效见证人
	var lastwitness = this.WitnessChain.WitnessGroup.Witness[0]
	for lastwitness.NextWitness != nil {
		lastwitness = lastwitness.NextWitness
	}

	lastBigGroupPointerAddr := fmt.Sprintf("%p", lastwitness.WitnessBigGroup)

	// 限制快照大小
	// 大约为300(GC回收后大约为300,再乘以2)见证人+保底60组的见证人
	witCount := 0
	limitWitCount := (config.Mining_group_max * config.Witness_backup_group_new) + config.Mining_block_hash_count*3*2

	// 从最后一个见证人,向前查找
	for i := lastwitness; i != nil; i = i.PreWitness {
		if witCount > limitWitCount {
			return nil, errors.New("snapshot size is too big")
		}

		// 见证人
		snapshotWitnessChain.Witness = append(snapshotWitnessChain.Witness, snapshotChain.toSnapshotWitness(i))

		// 块
		if lastblock == nil {
			if i.Block != nil {
				lastblock = i.Block
			}
		}

		// 组
		if lastgroup == nil {
			if i.Group.BlockGroup != nil {
				lastgroup = i.Group.BlockGroup
			}
		}

		// 小组
		if lastsmallgroup == nil {
			if i.Group != nil {
				lastsmallgroup = i.Group
			}
		}

		// 大组
		if i.WitnessBigGroup != nil {
			if i.WitnessBigGroup == lastbiggroup {
				lastbiggrouptmp.Keys = append(lastbiggrouptmp.Keys, snapshotChain.witnessKey(i))
			} else {
				lastbiggrouptmp = snapshotChain.toSnapshotWitnessBigGroupV2(i)
				lastbiggroup = i.WitnessBigGroup
				snapshotWitnessChain.BigGroup = append(snapshotWitnessChain.BigGroup, lastbiggrouptmp)
			}
		}
		witCount++
	}

	// 从最后一个块,向前查找
	for i := lastblock; i != nil; i = i.PreBlock {
		// 统计过的块
		snapshotWitnessChain.Block = append(snapshotWitnessChain.Block, snapshotChain.toSnapshotBlock(i))
	}

	// 从最后一个组,向前查找
	for i := lastgroup; i != nil; i = i.PreGroup {
		snapshotWitnessChain.Group = append(snapshotWitnessChain.Group, snapshotChain.toSnapshotGroup(i))
	}

	// 从最后一个小组,向前查找
	smallgroupindex := 0
	for i := lastsmallgroup; i != nil; i = i.PreGroup {
		sg := snapshotChain.toSnapshotWitnessSmallGroup(i)
		if this.WitnessChain.WitnessGroup == i {
			snapshotWitnessChain.CurrentSmallGroupIndex = int32(smallgroupindex)
		}
		snapshotWitnessChain.SmallGroup = append(snapshotWitnessChain.SmallGroup, sg)
		smallgroupindex++
	}

	snapshotChain.WitnessChain = snapshotWitnessChain

	// 未分配见证人集
	witnessNotGroups := []*go_protos.SnapshotWitness{}
	for i, _ := range this.WitnessChain.witnessNotGroup {
		w := this.WitnessChain.witnessNotGroup[i]
		// 确保最后一个大组与未分配的大组是同一指针
		if lastBigGroupPointerAddr != fmt.Sprintf("%p", w.WitnessBigGroup) {
			return nil, errors.New("snapshot skip")
		}
		witnessNotGroups = append(witnessNotGroups, snapshotChain.toSnapshotWitness(w))
	}

	snapshotChain.WitnessChain.WitnessNotGroup = witnessNotGroups
	//snapshotChain.WitnessChain.BigGroupNotGroup = bigGroupNotGroup

	// 备用见证人
	snapshotWitnessBackup := snapshotChain.toSnapshotWitnessBackup(this.WitnessBackup)

	// 余额管理器
	snapshotBalance := snapshotChain.toSnapshotBalanceManager(this.Balance)

	snapshotChain.No = this.No
	snapshotChain.StartingBlock = this.StartingBlock
	snapshotChain.StartBlockTime = this.StartBlockTime
	snapshotChain.CurrentBlock = this.CurrentBlock
	snapshotChain.PulledStates = this.PulledStates
	snapshotChain.HighestBlock = this.HighestBlock
	//snapshotChain.GroupHeight = this.GroupHeight
	snapshotChain.CurrentBlockHash = this.CurrentBlockHash
	snapshotChain.SyncBlockFinish = this.SyncBlockFinish
	snapshotChain.WitnessBackup = snapshotWitnessBackup
	snapshotChain.WitnessChain = snapshotWitnessChain
	snapshotChain.Balance = snapshotBalance
	snapshotChain.StopSyncBlock = this.StopSyncBlock

	sc := go_protos.SnapshotChain(*snapshotChain)
	b, err := sc.Marshal()
	return b, err
}

// 链快照还原Chain
func (this *Chain) lightSnapshotDeSerialize(data []byte) error {
	// 快照数据
	sc := go_protos.SnapshotChain{}
	err := sc.Unmarshal(data)
	if err != nil {
		return err
	}

	snapshotChain := SnapshotChain(sc)

	// 根据见证人构建链
	// 链接到本地第一个块
	engine.Log.Info(">>>>>> Snapshot[还原] 快照目标 当前高度:%d,同步高度:%d,最高高度:%d", snapshotChain.CurrentBlock, snapshotChain.PulledStates, snapshotChain.HighestBlock)

	/////////////////////////////// 处理基础链表 ///////////////////////////////
	// 当前小组位置
	index := snapshotChain.WitnessChain.CurrentSmallGroupIndex
	// 定义单个链
	// 块
	// 组
	// 大组
	// 小组
	// 见证人
	snapshotBlocks := snapshotChain.WitnessChain.Block
	blocks := make([]*Block, len(snapshotBlocks))
	snapshotGroups := snapshotChain.WitnessChain.Group
	groups := make([]*Group, len(snapshotGroups))
	snapshotBigGroups := snapshotChain.WitnessChain.BigGroup
	bigGroups := make([]*WitnessBigGroup, len(snapshotBigGroups))
	snapshotSmallGroups := snapshotChain.WitnessChain.SmallGroup
	smallGroups := make([]*WitnessSmallGroup, len(snapshotSmallGroups))
	snapshotWitnesses := snapshotChain.WitnessChain.Witness
	witnesses := make([]*Witness, len(snapshotWitnesses))
	// 未分配见证人等
	//snapshotWitnessNotGroup := snapshotChain.WitnessChain.WitnessNotGroup
	//witnessNotGroup := make([]*Witness, len(snapshotWitnessNotGroup))
	witnessNotGroupLength := len(snapshotChain.WitnessChain.WitnessNotGroup)
	witnessNotGroup := make([]*Witness, witnessNotGroupLength)
	allWitnessNotGroup := make(map[string]*Witness)
	//var bigGroupNotGroup *WitnessBigGroup

	// 块链
	for i, _ := range snapshotBlocks {
		block := snapshotChain.toBlock(snapshotBlocks[i])
		// NOTE 稍后处理
		//block.Group
		//block.witness

		if i > 0 {
			blocks[i-1].PreBlock = block
			blocks[i-1].PreBlockID = block.Id
			block.NextBlock = blocks[i-1]
		}

		blocks[i] = block
	}

	// 组链
	for i, _ := range snapshotGroups {
		group := snapshotChain.toGroup(snapshotGroups[i])
		// NOTE 稍后处理
		//group.Blocks

		if i > 0 {
			groups[i-1].PreGroup = group
			group.NextGroup = groups[i-1]
		}

		groups[i] = group
	}

	// 小组链
	for i, _ := range snapshotSmallGroups {
		smallGroup := snapshotChain.toWitnessSmallGroup(snapshotSmallGroups[i])
		// NOTE 稍后处理
		//smallGroup.Witness
		//smallGroup.BlockGroup

		if i > 0 {
			smallGroups[i-1].PreGroup = smallGroup
			smallGroup.NextGroup = smallGroups[i-1]
		}

		smallGroups[i] = smallGroup
	}

	// 大组
	witbgindex := make(map[string]int) // 见证人key：大组数组索引
	for i, _ := range snapshotBigGroups {
		bigGroups[i] = snapshotChain.toWitnessBigGroup(snapshotBigGroups[i])
		for _, k := range snapshotBigGroups[i].Keys {
			witbgindex[k] = i
		}
		for j, _ := range snapshotBigGroups[i].Witnesses {
			bigGroups[i].Witnesses[j].WitnessBigGroup = bigGroups[i]
			allWitnessNotGroup[snapshotBigGroups[i].Witnesses[j].PtrAddr] = bigGroups[i].Witnesses[j]
		}
		for j, _ := range snapshotBigGroups[i].WitnessBackup {
			bigGroups[i].WitnessBackup[j].WitnessBigGroup = bigGroups[i]
			allWitnessNotGroup[snapshotBigGroups[i].WitnessBackup[j].PtrAddr] = bigGroups[i].WitnessBackup[j]
		}
	}

	// 见证人链
	for i, _ := range snapshotWitnesses {
		witness := snapshotChain.toWitness(snapshotWitnesses[i])
		// NOTE 稍后处理
		//witness.Group
		//witness.Block
		//witness.WitnessBigGroup

		// 见证人的大组指针整理
		// 处理内部大组引用指针
		// 同时补充大组中见证人少量信息
		key := snapshotChain.witnessKey(witness)
		bgindex := witbgindex[key]
		witness.WitnessBigGroup = bigGroups[bgindex]
		//if bg, ok := tmpbiggroup[key]; ok {
		//} else {
		//	witnessBigGroup := &WitnessBigGroup{}
		//	for _, v := range snapshotWitnesses[i].WitnessBigGroup.Witnesses {
		//		addr := crypto.AddressCoin(v.Addr)
		//		witnessBigGroup.Witnesses = append(witnessBigGroup.Witnesses, &Witness{
		//			Addr:            &addr,
		//			CreateBlockTime: v.CreateBlockTime,
		//			sleepTime:       v.SleepTime,
		//		})
		//	}

		//	for _, v := range snapshotWitnesses[i].WitnessBigGroup.WitnessBackup {
		//		addr := crypto.AddressCoin(v.Addr)
		//		witnessBigGroup.WitnessBackup = append(witnessBigGroup.WitnessBackup, &Witness{
		//			Addr:            &addr,
		//			CreateBlockTime: v.CreateBlockTime,
		//			sleepTime:       v.SleepTime,
		//		})
		//	}

		//	tmpbiggroup[key] = witnessBigGroup
		//	witness.WitnessBigGroup = witnessBigGroup
		//}

		if i > 0 {
			witnesses[i-1].PreWitness = witness
			witness.NextWitness = witnesses[i-1]
		}

		witnesses[i] = witness
		allWitnessNotGroup[snapshotWitnesses[i].PtrAddr] = witness
	}

	// 所有未分配见证人,备用见证人
	// NOTE：这个大组必定是上一次构建剩下的，应该找上一次构建大组中见证人的指针
	//for i, w := range snapshotChain.WitnessChain.WitnessNotGroup {
	//	// NOTE 处理见证人
	//	if wit, ok := allWitnessNotGroup[w.PtrAddr]; ok {
	//		witnessNotGroup[i] = wit
	//	} else {
	//		witness := snapshotChain.toWitness(snapshotChain.WitnessChain.WitnessNotGroup[i])
	//		witness.WitnessBigGroup = witnesses[0].WitnessBigGroup
	//		witnessNotGroup[i] = witness
	//	}
	//}

	//// 未分配见证人大组
	//// 特殊处理：如果大组的keys数量大于未分配的见证人，这个大组必定是上一次构建剩下的，应该找上一次构建大组的指针
	//isCrossBigGroup := false
	//if len(snapshotChain.WitnessChain.WitnessNotGroup) > 0 {
	//	snapshotbg := snapshotChain.WitnessChain.BigGroupNotGroup
	//	if len(snapshotbg.Keys) > len(snapshotbg.Witnesses) {
	//		isCrossBigGroup = true
	//		bigGroupNotGroup = bigGroups[len(bigGroups)-1]
	//	} else {
	//		if bigGroupNotGroup == nil {
	//			bigGroupNotGroup = &WitnessBigGroup{
	//				Witnesses:     make([]*Witness, 0),
	//				WitnessBackup: make([]*Witness, 0),
	//			}
	//		}
	//		snapshotbg := snapshotChain.WitnessChain.BigGroupNotGroup
	//		for _, w := range snapshotbg.Witnesses {
	//			if v, ok := allWitnessNotGroup[w.PtrAddr]; ok {
	//				// NOTE 处理
	//				// wit.WitnessBigGroup
	//				v.WitnessBigGroup = bigGroupNotGroup
	//				bigGroupNotGroup.Witnesses = append(bigGroupNotGroup.Witnesses, v)
	//			}
	//		}
	//		for _, w := range snapshotbg.WitnessBackup {
	//			if v, ok := allWitnessNotGroup[w.PtrAddr]; ok {
	//				// NOTE 处理
	//				// wit.WitnessBigGroup
	//				v.WitnessBigGroup = bigGroupNotGroup
	//				bigGroupNotGroup.WitnessBackup = append(bigGroupNotGroup.WitnessBackup, v)
	//			}
	//		}
	//	}
	//}

	//// 未分配见证人
	//// 关联特殊处理 "如果大组的keys数量大于未分配的见证人，这个大组必定是上一次构建剩下的，应该找上一次构建大组的指针"
	//if isCrossBigGroup {
	//	for i, _ := range witnessNotGroup {
	//		offset := len(bigGroupNotGroup.Witnesses) - len(witnessNotGroup)
	//		if offset+i >= 0 && offset+i < len(bigGroupNotGroup.Witnesses) {
	//			witnessNotGroup[i] = bigGroupNotGroup.Witnesses[offset+i]
	//		}
	//	}
	//} else {
	//	for i, w := range snapshotWitnessNotGroup {
	//		if v, ok := allWitnessNotGroup[w.PtrAddr]; ok {
	//			v.WitnessBigGroup = bigGroupNotGroup
	//			witnessNotGroup[i] = v
	//		}
	//	}
	//}

	// 未分配见证人
	// 未分配见证人的大组
	// NOTE：这个大组必定是上一次构建剩下的，应该找上一次构建大组的指针
	if witnessNotGroupLength > 0 {
		//offset := len(bigGroups[0].Witnesses) - witnessNotGroupLength
		//for i, _ := range bigGroups[0].Witnesses {
		//	if i == 0 {
		//		continue
		//	}

		//	if i-offset >= 0 && i-offset < witnessNotGroupLength {
		//		if bigGroups[0].Witnesses[i].PreWitness == nil {
		//			bigGroups[0].Witnesses[i].PreWitness = bigGroups[0].Witnesses[i-1]
		//		}
		//		if bigGroups[0].Witnesses[i-1].NextWitness == nil {
		//			bigGroups[0].Witnesses[i-1].NextWitness = bigGroups[0].Witnesses[i]
		//		}
		//		if bigGroups[0].Witnesses[i].WitnessBigGroup == nil {
		//			bigGroups[0].Witnesses[i].WitnessBigGroup = bigGroups[0]
		//		}
		//		witnessNotGroup[i-offset] = bigGroups[0].Witnesses[i]
		//	}
		//}
	}

	/////////////////////////////// 处理链表递归的指针 ///////////////////////////////
	// 处理块链
	for i, _ := range blocks {
		// NOTE 处理
		//block.Group
		//block.witness

		if blocks[i] == nil {
			break
		}

		// block属于哪个group
		isGotGroup := false
		for k, group := range groups {
			if isGotGroup {
				break
			}

			for _, block := range group.Blocks {
				if bytes.Equal(block.Id, blocks[i].Id) {
					blocks[i].Group = groups[k]
					isGotGroup = true
					break
				}
			}
		}
		//if !isGotGroup {
		//	engine.Log.Info(">>>>>> Snapshot[还原.块] 块没有组:%v,%v", hex.EncodeToString(blocks[i].Id), blocks[i].Height)
		//}

		//NOTE 从第i个见证人开始找
		//isGotWitness := false
		//for j := i; j < len(witnesses); j++ {
		//	if witnesses[j].Block == nil {
		//		continue
		//	}
		//	if bytes.Equal(witnesses[j].Block.Id, blocks[i].Id) {
		//		blocks[i].witness = witnesses[j]
		//		isGotWitness = true
		//		break
		//	}
		//}
		//isGotWitness := false
		for j, _ := range witnesses {
			if witnesses[j].Block == nil {
				continue
			}
			if bytes.Equal(witnesses[j].Block.Id, blocks[i].Id) {
				blocks[i].Witness = witnesses[j]
				//isGotWitness = true
				break
			}
		}
		//if !isGotWitness {
		//	engine.Log.Info(">>>>>> Snapshot[还原.块] 块没有见证人:%v,%v", hex.EncodeToString(blocks[i].Id), blocks[i].Height)
		//}
	}

	// 处理见证人链
	for i, _ := range witnesses {
		// NOTE 处理
		//witness.Group
		//witness.Block
		//witness.WitnessBigGroup
		//witness.StopMining

		// 根据组高度还原
		for j, _ := range smallGroups {
			if witnesses[i].Group == nil {
				continue
			}
			if smallGroups[j].Height == witnesses[i].Group.Height {
				witnesses[i].Group = smallGroups[j]
				break
			}
		}

		// 根据块id还原
		for k, _ := range blocks {
			if witnesses[i].Block == nil {
				continue
			}
			if bytes.Equal(witnesses[i].Block.Id, blocks[k].Id) {
				witnesses[i].Block = blocks[k]
				break
			}
		}

		// 见证人与大组通过见证人地址和预计出块时间对应,
		wits := witnesses[i].WitnessBigGroup.Witnesses
		for m, _ := range wits {
			for n, _ := range witnesses {
				if bytes.Equal(*wits[m].Addr, *witnesses[n].Addr) &&
					wits[m].CreateBlockTime == witnesses[n].CreateBlockTime {
					wits[m] = witnesses[n]
					break
				}
			}
		}

		backupwits := witnesses[i].WitnessBigGroup.WitnessBackup
		for m, _ := range backupwits {
			for n, _ := range witnesses {
				if bytes.Equal(*backupwits[m].Addr, *witnesses[n].Addr) &&
					backupwits[m].CreateBlockTime == witnesses[n].CreateBlockTime {
					backupwits[m] = witnesses[n]
					break
				}
			}
		}

		// 停止出块命令
		witnesses[i].StopMining <- true
	}

	// 小组链
	for i, _ := range smallGroups {
		// NOTE 处理
		//smallGroup.Witness
		//smallGroup.BlockGroup

		// 根据见证人组高度
		tmpwitoffset := 1
		shouldwitcount := len(smallGroups[i].Witness)
		for j, _ := range witnesses {
			if tmpwitoffset > shouldwitcount {
				break
			}

			if witnesses[j].Group.Height == smallGroups[i].Height {
				smallGroups[i].Witness[shouldwitcount-tmpwitoffset] = witnesses[j]
				tmpwitoffset++
			}
		}

		// 根据组高度还原
		for k, _ := range groups {
			if smallGroups[i].BlockGroup == nil {
				continue
			}

			if smallGroups[i].BlockGroup.Height == groups[k].Height {
				smallGroups[i].BlockGroup = groups[k]
				break
			}
		}
	}

	// 组链
	for i, group := range groups {
		// NOTE 处理
		//group.Blocks
		for j, gblock := range group.Blocks {
			for k, block := range blocks {
				if bytes.Equal(gblock.Id, block.Id) {
					groups[i].Blocks[j] = blocks[k]
					break
				}
			}
		}
	}

	// 所有未分配见证人,备用见证人
	// NOTE：这个大组必定是上一次构建剩下的，应该找上一次构建大组中见证人的指针
	for i, w := range snapshotChain.WitnessChain.WitnessNotGroup {
		// NOTE 处理见证人
		if wit, ok := allWitnessNotGroup[w.PtrAddr]; ok {
			witnessNotGroup[i] = wit
		} else {
			return errors.New("snapshot internal error")
			//witness := snapshotChain.toWitness(snapshotChain.WitnessChain.WitnessNotGroup[i])
			//witness.WitnessBigGroup = witnesses[0].WitnessBigGroup
			//witnessNotGroup[i] = witness
		}
	}

	// 大组链
	//for i, bigGroup := range bigGroups {
	//	// NOTE 处理
	//	// 内部引用指针
	//	// 根据见证人地址映射
	//	for m, wit := range bigGroup.Witnesses {
	//		for n, wit2 := range witnesses {
	//			if bytes.Equal(*wit.Addr, *wit2.Addr) && wit.CreateBlockTime == wit2.CreateBlockTime {
	//				bigGroups[i].Witnesses[m] = witnesses[n]
	//				break
	//			}
	//		}
	//	}

	//	for m, wit := range bigGroup.WitnessBackup {
	//		for n, wit2 := range witnesses {
	//			if bytes.Equal(*wit.Addr, *wit2.Addr) && wit.CreateBlockTime == wit2.CreateBlockTime {
	//				bigGroups[i].WitnessBackup[m] = witnesses[n]
	//				break
	//			}
	//		}
	//	}
	//}

	/////////////////////////////// 处理Chain指针 ///////////////////////////////
	this.WitnessChain.chain = this
	witnessBackup := snapshotChain.toWitnessBackup(snapshotChain.WitnessBackup)
	witnessBackup.chain = this
	this.WitnessChain.witnessBackup = witnessBackup
	this.WitnessBackup = witnessBackup

	// 余额
	balance := snapshotChain.toBalanceManager(snapshotChain.Balance)
	balance.chain = this
	//balance.txManager = NewTransactionManager(witnessBackup)
	this.Balance.txManager.witnessBackup = witnessBackup
	balance.txManager = this.Balance.txManager
	//balance.txAveGas = snapshotChain.toTxPayAveGas(snapshotChain.Balance.TxAveGas)
	this.Balance = balance

	// 未分配见证人
	//witnessNotGroup := []*Witness{}
	//for _, v := range snapshotChain.WitnessChain.WitnessNotGroup {
	//witnessNotGroup = append(witnessNotGroup, snapshotChain.toWitness4NotGroup(v))
	//}

	this.WitnessChain.witnessNotGroup = witnessNotGroup

	// 当前的小组链
	this.WitnessChain.WitnessGroup = smallGroups[index]

	// 其它
	this.WitnessChain.chain = this
	this.No = snapshotChain.No
	this.StartingBlock = snapshotChain.StartingBlock
	this.StartBlockTime = snapshotChain.StartBlockTime
	this.CurrentBlock = snapshotChain.CurrentBlock
	this.PulledStates = snapshotChain.PulledStates
	this.HighestBlock = snapshotChain.HighestBlock
	//this.GroupHeight = snapshotChain.GroupHeight
	this.CurrentBlockHash = snapshotChain.CurrentBlockHash
	//this.SyncBlockFinish = snapshotChain.SyncBlockFinish
	this.SyncBlockFinish = false // 恢复快照强制为false
	this.StopSyncBlock = snapshotChain.StopSyncBlock
	this.SyncBlockLock = new(sync.RWMutex)
	this.signLock = new(sync.RWMutex)
	this.Temp = nil

	// 补充
	//this.WitnessBackup.syncWitnessVoteFromCache()

	return nil
}

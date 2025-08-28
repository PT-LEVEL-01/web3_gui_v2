package mining

import (
	"bytes"
	"encoding/hex"
	"math"
	"runtime"
	"sync"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/utils"
)

const (
	Mining_Status_Start           = 1 //开始状态，节点启动
	Mining_Status_WaitMulticas    = 2 //
	Mining_Status_WaitImportBlock = 3 //
	Mining_Status_ImportBlock     = 4 //
	// Mining_Status_WaitMulticas = 5 //

)

var MiningStatusLock = new(sync.Mutex)
var MiningStatus = Mining_Status_Start
var BhvoMulticasCache = make(map[string]*BlockHeadVO)

func init() {

}

func SetMiningStatus_() {

}

func AddBlockToCache(bhvo *BlockHeadVO) {
	//先清理缓存
	MiningStatusLock.Lock()
	//pl time
	for k, bhvo := range BhvoMulticasCache {
		if bhvo.BH.Time < config.TimeNow().Unix()-int64(config.Mining_block_time.Seconds())*10 {
			//if time.Unix(bhvo.BH.Time, 0).UnixNano() < config.TimeNow().UnixNano()-config.Mining_block_time.Nanoseconds()*10 {
			delete(BhvoMulticasCache, k)
		}
	}
	//再添加区块
	BhvoMulticasCache[utils.Bytes2string(bhvo.BH.Hash)] = bhvo
	MiningStatusLock.Unlock()

	// engine.Log.Info("导入cache group:%d height:%d ", bhvo.BH.GroupHeight, bhvo.BH.Height)
}

func DelBlockToCache(hash *[]byte) {
	MiningStatusLock.Lock()
	delete(BhvoMulticasCache, utils.Bytes2string(*hash))
	MiningStatusLock.Unlock()
}

/*
*
从缓存中获取区块
*/
func GetBlockByCache(hash *[]byte) *BlockHeadVO {
	MiningStatusLock.Lock()
	bhvo, _ := BhvoMulticasCache[utils.Bytes2string(*hash)]
	MiningStatusLock.Unlock()
	return bhvo
}

/*
*
缓存是否存在
*/
func ExistBlockByCache(hash *[]byte) bool {
	MiningStatusLock.Lock()
	_, ok := BhvoMulticasCache[utils.Bytes2string(*hash)]
	MiningStatusLock.Unlock()
	return ok
}

/*
从缓存中找到对应区块导入
*/
func ImportBlockByCache(hash *[]byte) {
	var bhvo, prebhvo *BlockHeadVO
	// ok := false
	MiningStatusLock.Lock()
	bhvo, _ = BhvoMulticasCache[utils.Bytes2string(*hash)]
	MiningStatusLock.Unlock()
	if bhvo == nil {
		//engine.Log.Info("88888888888888888888888888888888888888888888 %s", hex.EncodeToString(*hash))
		return
	}
	//导入区块，先查找本区块是否也在缓存中

	MiningStatusLock.Lock()
	prebhvo, _ = BhvoMulticasCache[utils.Bytes2string(bhvo.BH.Previousblockhash)]
	MiningStatusLock.Unlock()
	if prebhvo != nil {
		//先导入前置区块
		err := forks.GetLongChain().AddBlockSelf(prebhvo)
		if err == nil {
			//广播区块
			go MulticastBlock(*prebhvo)
			// utils.Go(MulticastBlock(*prebhvo))
			MiningStatusLock.Lock()
			delete(BhvoMulticasCache, utils.Bytes2string(prebhvo.BH.Hash))
			MiningStatusLock.Unlock()
		}
	}
	//再导入本区块
	err := forks.GetLongChain().AddBlockSelf(bhvo)
	if err == nil {
		//广播区块
		go MulticastBlock(*bhvo)
		// utils.Go(MulticastBlock(*bhvo))
		MiningStatusLock.Lock()
		delete(BhvoMulticasCache, utils.Bytes2string(bhvo.BH.Hash))
		MiningStatusLock.Unlock()
	}
}

// /*
// 	开始挖矿
// 	当每个组见证人选出来之后，启动挖矿程序，按顺序定时出块
// */
// func Mining() {
// 	//判断是否同步完成
// 	if GetHighestBlock() <= 0 {
// 		fmt.Println("区块未同步完成，不能挖矿 GetHighestBlock", GetHighestBlock())
// 		return
// 	}
// 	if GetHighestBlock() > GetCurrentBlock() {
// 		fmt.Println("区块未同步完成，不能挖矿 GetCurrentBlock", GetCurrentBlock(), GetHighestBlock())
// 		return
// 	}
// 	if !config.Miner {
// 		fmt.Println("本节点不是旷工节点")
// 		return
// 	}

// 	fmt.Println("启动挖矿程序")

// 	addr := keystore.GetCoinbase()

// 	//用见证人方式出块
// 	fmt.Println("用见证人方式出块")
// 	group := forks.GetLongChain().witnessChain.group
// 	//判断是否已经安排了任务
// 	if group.Task {
// 		// fmt.Println("已经安排了任务，退出")
// 		return
// 	}
// 	group.Task = true

// 	//判断自己出块顺序的时间
// 	for i, one := range forks.GetLongChain().witnessChain.group.Witness {
// 		//自己是见证人才能出块，否则自己出块了，其他节点也不会承认
// 		if bytes.Equal(*one.Addr, addr) {
// 			fmt.Println("自己多少秒钟后出块", config.Mining_block_time*(i+1))
// 			utils.AddTimetask(config.TimeNow().Unix()+int64(config.Mining_block_time*(i+1)),
// 				TaskBuildBlock, Task_class_buildBlock, "")
// 		}

// 	}

// }

// var logblockheight = uint64(152020)

/*
查找未确认的区块
获取其中的交易，用于验证交易
@return    *Block     出块时，应该链接的上一个块
@return    []Block    出块时，应该链接的上一个组的块
*/
func (this *Witness) FindUnconfirmedBlock() (*Block, []Block) {
	//找到上一个块
	var preBlock *Block
	//判断是否是该组第一个块
	isFirst := false
	// engine.Log.Info("FindUnconfirmedBlock SelectionChain")
	group := this.Group.SelectionChain(nil)
	if group == nil {
		isFirst = true
	} else {
		isFirst = false
		//取本组最后一个块
		// fmt.Println("获取本组最后一个块", len(group.Blocks))
		preBlock = group.Blocks[len(group.Blocks)-1]
		// engine.Log.Info("1前置区块 %+v", preBlock)
	}
	// engine.Log.Info("是否是本组第一个块 %v", isFirst)

	//找到上一个组
	preGroup := this.Group
	var preGroupBlock *Group
	var ok bool
	for {
		// if preGroup.Height > logblockheight || preGroup.Height < 145137 {
		// 	engine.Log.Info("-----------寻找上一组 1111 %d", preGroup.Height)
		// }
		ok = false
		preGroup = preGroup.PreGroup
		ok, preGroupBlock = preGroup.CheckBlockGroup(nil)
		if ok {
			// engine.Log.Info("-----------寻找上一组 222222222")
			if isFirst {
				//取本组最后一个块
				// engine.Log.Info("获取上一组最后一个块")
				preBlock = preGroupBlock.Blocks[len(preGroupBlock.Blocks)-1]
				// engine.Log.Info("2前置区块", preBlock)
			}
			break
		}
		// engine.Log.Info("-----------寻找上一组 3333333")
	}

	//查找出未确认的块
	blocks := make([]Block, 0)
	if preGroup.Height != this.Group.Height {
		for _, one := range preGroupBlock.Blocks {
			blocks = append(blocks, *one)
		}
	}
	if group != nil {
		for _, one := range group.Blocks {
			blocks = append(blocks, *one)
		}
	}
	return preBlock, blocks
}

/*
验证未确认的区块
获取其中的交易，用于验证交易
@return    *Block     出块时，应该链接的上一个块
@return    []Block    出块时，应该链接的上一个组的块
*/
func (this *Witness) CheckUnconfirmedBlock(blockhash *[]byte) (*Block, []Block) {

	//找到上一个块
	var preBlock *Block
	//判断是否是该组第一个块
	isFirst := false
	tempBlockHash := blockhash

	// engine.Log.Info("CheckUnconfirmedBlock SelectionChain")
	group := this.Group.SelectionChain(blockhash)

	if group == nil {
		isFirst = true
	} else {
		isFirst = false
		//取本组最后一个块
		// fmt.Println("获取本组最后一个块", len(group.Blocks))
		preBlock = group.Blocks[len(group.Blocks)-1]
		tempBlockHash = &group.Blocks[0].PreBlockID
		// engine.Log.Info("1前置区块 %+v", preBlock)
	}
	// engine.Log.Info("是否是本组第一个块 %v", isFirst)

	//找到上一个组
	preGroup := this.Group
	var preGroupBlock *Group
	var ok bool
	for {
		// if preGroup.Height > logblockheight || preGroup.Height < 145137 {
		// 	engine.Log.Info("-----------寻找上一组 1111 %d", preGroup.Height)
		// }
		ok = false

		ok, preGroupBlock = preGroup.CheckBlockGroup(tempBlockHash)

		if ok {
			// engine.Log.Info("-----------寻找上一组 222222222")
			if isFirst {
				//取本组最后一个块
				// engine.Log.Info("获取上一组最后一个块")
				preBlock = preGroupBlock.Blocks[len(preGroupBlock.Blocks)-1]
				// engine.Log.Info("2前置区块", preBlock)
			}
			break
		} else {
			if preGroupBlock != nil && len(preGroupBlock.Blocks) > 0 {
				tempBlockHash = &preGroupBlock.Blocks[0].Id
			}
		}
		preGroup = preGroup.PreGroup
		// engine.Log.Info("-----------寻找上一组 3333333")
	}

	// for {
	// 	if bytes.Equal(preBlock.Id, bh.Previousblockhash) {
	// 		break
	// 	} else {
	// 		if preBlock.Height <= bh.Height-1 {
	// 			//找不到这个区块了，因为高度都不一样了
	// 			break
	// 		} else {
	// 			preBlock = preBlock.PreBlock
	// 		}
	// 	}
	// }

	preWitness := this.PreWitness
	for {
		if preWitness == nil {
			break
		}
		if preWitness.Block == nil {
			preWitness = preWitness.PreWitness
			continue
		}

		if bytes.Equal(preWitness.Block.Id, *blockhash) {

			preBlock = preWitness.Block
			break
		}
		preWitness = preWitness.PreWitness
		// engine.Log.Info("-----------寻找上一个块 444444444")
	}

	//查找出未确认的块
	blocks := make([]Block, 0)
	if preGroup.Height != this.Group.Height {
		for _, one := range preGroupBlock.Blocks {
			blocks = append(blocks, *one)
		}
	}
	if group != nil {
		for _, one := range group.Blocks {
			blocks = append(blocks, *one)
		}
	}

	// for _, one := range blocks {
	// 	engine.Log.Info("查找出未确认的块 group:%d height:%d hash:%s preHash:%s %+v", one.Group.Height, one.Height, hex.EncodeToString(one.Id), hex.EncodeToString(one.PreBlock.Id), one)
	// }
	return preBlock, blocks
}

// var testBug = false

var testBifurcateHeight = uint64(300)
var witAndRoleRewardSet = sync.Map{} //记录奖励区块奖励;区分见证人和角色奖励;可以记录多个分叉区块的数据

/*
见证人方式出块
出块并广播
@gh    uint64    出块的组高度
@id    []byte    押金id
*/
func (this *Witness) BuildBlock() {
	// var this *Witness
	addrInfo := Area.Keystore.GetCoinbase()

	//自己是见证人才能出块，否则自己出块了，其他节点也不会承认
	if !bytes.Equal(*this.Addr, addrInfo.Addr) {
		return
	}

	//本节点未同步完成，则不应该出块
	if !GetLongChain().SyncBlockFinish {
		return
	}

	//判断出块时间，如果超过了自己的出块时间，则不出块
	//lateTime := int64(time.Since(this.createTime)) - this.CreateBlockTime - this.sleepTime
	//lateTime := int64(config.TimeNow().Sub(this.createTime)) - this.CreateBlockTime - this.sleepTime
	//engine.Log.Info("出块时间晚了:%s 创建时间：%v 预计出块时间：%v 等待时间：%s 当前时间：%s", time.Duration(lateTime), this.createTime.Format("2006-01-02 15:04:05"), time.Unix(this.CreateBlockTime, 0).Format("2006-01-02 15:04:05"), time.Duration(this.sleepTime), config.TimeNow().Format("2006-01-02 15:04:05"))
	//engine.Log.Info("出块时间晚了:%s", time.Duration(lateTime))

	//当前时间-预计出块时间>2倍出块间隔，则认为出块时间太晚了
	lateTime := config.TimeNow().UnixNano() - time.Unix(this.CreateBlockTime, 0).UnixNano()
	if lateTime > int64(2*config.Mining_block_time) {
		engine.Log.Info("出块时间太晚了group:%d", this.Group.Height)
		return
	}

	workModeLockStatic.GetLock()

	start := config.TimeNow()

	//查找出未确认的块
	preBlock, blocks := this.FindUnconfirmedBlock()
	// start2 := config.TimeNow().Sub(start)
	// engine.Log.Info("上一个块高度 %d %d %d", preBlock.Height, blocks[0].Height, blocks[0].witness.Group.Height)
	// engine.Log.Info("获取上一个块高度消耗时间 %s", config.TimeNow().Sub(start))
	engine.Log.Info("=== build block === group:%d height:%d", this.Group.Height, preBlock.Height+1)

	//测试代码，模拟分叉情况
	if false {
		if blocks[0].Height > testBifurcateHeight {
			if len(blocks) > 1 && blocks[len(blocks)-1].Group.Height != this.Group.Height {
				for _, one := range blocks {
					engine.Log.Info("区块高度:%d", one.Height)
				}
				engine.Log.Info("测试生效。前区块高度:%d", preBlock.Height)
				temp := blocks[:len(blocks)-1]
				blocks = temp

				tempBlock := temp[len(temp)-1]
				preBlock = &tempBlock
				testBifurcateHeight = testBifurcateHeight + 100
			}
		}
	}

	//存放交易
	tx := make([]TxItr, 0)
	txids := make([][]byte, 0)

	var reward *Tx_reward
	var tmpVouts []*Vout
	//检查本组是否给上一组见证人奖励
	if this.WitnessBigGroup != preBlock.Witness.WitnessBigGroup {
		// engine.Log.Info("开始构建上一组见证人奖励 %s %d", fmt.Sprintf("%p", preBlock.witness.WitnessBigGroup),
		// preBlock.witness.Group.Height)
		reward, tmpVouts = preBlock.Witness.WitnessBigGroup.CountRewardToWitnessGroup(GetLongChain(), preBlock.Height+1, blocks, preBlock)
		//reward = preBlock.witness.WitnessBigGroup.CountRewardToWitnessGroupByContract(preBlock.Height+1, blocks, preBlock)
		tx = append(tx, reward)
		// engine.Log.Info("reward:%+v", reward)
		txids = append(txids, reward.Hash)
	}

	//云存储代理奖励
	//TODO 需求变更,待修改
	//if 1 == 0 && (preBlock.Height+1)%config.CloudStorage_Reward_Interval == 0 {
	//	cloudStorageProxyPayload := precompiled.BuildCloudStorageProxyRewardInput()
	//	cloudStorageProxyTx := CreateTxCloudStorageProxy(precompiled.CloudStorageProxyContract, preBlock.Height+1, cloudStorageProxyPayload)
	//	tx = append(tx, cloudStorageProxyTx)
	//	txids = append(txids, cloudStorageProxyTx.Hash)
	//}

	//打包撮合交易
	swapTxs, swapIds := oBook.SwapTxPackage()
	if swapTxs != nil && swapIds != nil {
		tx = append(tx, swapTxs...)
		txids = append(txids, swapIds...)
	}

	// start3 := config.TimeNow().Sub(start)
	// engine.Log.Info("394构建区块奖励消耗时间 %s,%s", start2, start3)

	//打包所有交易
	chain := forks.GetLongChain()
	txs, ids := chain.TransactionManager.Package(reward, preBlock.Height+1, blocks, this.CreateBlockTime)
	tx = append(tx, txs...)
	txids = append(txids, ids...)

	// engine.Log.Info("打包消耗时间 %s", config.TimeNow().Sub(start))

	//准备块中的交易
	// fmt.Println("准备块中的交易")
	coinbase := Area.Keystore.GetCoinbase()

	var bh *BlockHead
	now := config.TimeNow().Unix() //time.Now().Unix()
	//now := config.TimeNow().UnixNano()
	//pl time
	//for i := int64(0); i < (config.Mining_block_time*2)-1; i++ {
	for i := int64(0); i < int64(math.Ceil(config.Mining_block_time.Seconds()))*2-1; i++ {
		//开始生成块
		bh = &BlockHead{
			Height:            preBlock.Height + 1, //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
			GroupHeight:       this.Group.Height,   // preGroup.Height + 1,               //矿工组高度
			Previousblockhash: preBlock.Id,         //上一个区块头hash
			NTx:               uint64(len(tx)),     //交易数量
			Tx:                txids,               //本区块包含的交易id
			Time:              now + i,             //unix时间戳
			Witness:           coinbase.Addr,       //此块矿工地址
			ExtSign:           make([][]byte, 0),
		}

		// if !testBug && bh.Height > config.Mining_block_start_height+12 && this.Group.FirstWitness() {
		// 	engine.Log.Error("把区块高度减1")
		// 	testBug = true
		// 	bh.Height = bh.Height - 1
		// }

		bh.BuildMerkleRoot()
		bh.BuildSign(coinbase.Addr, Area.Keystore)
		bh.BuildBlockHash()
		if ok, _ := bh.CheckHashExist(); ok {
			bh = nil
			continue
		} else {
			break
		}
	}
	if bh == nil {
		workModeLockStatic.BackLock()
		engine.Log.Info("Block out failed, all hash have collisions")
		//出块失败，所有hash都有碰撞
		return
	}

	//判断出块时间，如果超过了自己的出块时间，则不出块
	//lateTime = int64(time.Since(this.createTime)) - this.CreateBlockTime - this.sleepTime
	//lateTime = int64(config.TimeNow().Sub(this.createTime)) - this.CreateBlockTime - this.sleepTime

	//当前时间-预计出块时间>2倍出块间隔，则认为出块时间太晚了
	lateTime = config.TimeNow().UnixNano() - time.Unix(this.CreateBlockTime, 0).UnixNano()
	if lateTime > int64(2*config.Mining_block_time) {
		workModeLockStatic.BackLock()
		engine.Log.Info("出块时间太晚了group:%d height:%d", this.Group.Height, bh.Height)
		return
	}

	bhvo := CreateBlockHeadVO(config.StartBlockHash, bh, tx)

	//先保存到数据库再广播，否则其他节点查询不到
	// SaveBlockHead(bhvo)

	bhvo.FromBroadcast = true
	// engine.Log.Info("打包消耗时间 %s", config.TimeNow().Sub(start))

	// workModeLockStatic.BackLock()
	//先自己签名
	ok := workModeLockStatic.CheckBlockSign(bhvo, chain)
	if !ok {
		workModeLockStatic.BackLock()
		engine.Log.Info("自己出块，自己签名失败")
		return
	}

	//先请求其他见证人签名
	//witnessTotal := len(this.WitnessBigGroup.Witnesses)
	err := WitnessSignMulticastBlock(bhvo, this.WitnessBigGroup.Witnesses)
	if err != nil {
		workModeLockStatic.BackLock()
		engine.Log.Warn("先请求其他见证人签名 failed, block is illegal group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		return
	}

	engine.Log.Info("+++ build block Success hash:%s group:%d height:%d preHash:%s %s",
		hex.EncodeToString(bhvo.BH.Hash), bhvo.BH.GroupHeight, bhvo.BH.Height,
		hex.EncodeToString(bhvo.BH.Previousblockhash), config.TimeNow().Sub(start))
	// engine.Log.Info("=== build block Success === Block hash %s", hex.EncodeToString(bhvo.BH.Hash))
	// engine.Log.Info("=== build block Success === pre Block hash %s", hex.EncodeToString(bhvo.BH.Previousblockhash))

	if tmpVouts != nil {
		witAndRoleRewardSet.Store(utils.Bytes2string(bhvo.BH.Hash), tmpVouts)
	}

	//再广播区块
	UniformityMulticastBlock(bhvo)
}

func UniformityMulticastBlock(bhVO *BlockHeadVO) {
	// engine.Log.Info("广播区块 11111111111")
	// AddBlockToCache(bhVO)
	// ImportBlockByCache(&bhVO.BH.Hash)

	//广播区块
	go MulticastBlock(*bhVO)
	// engine.Log.Info("广播区块 22222222222")
	//再导入本区块
	forks.GetLongChain().AddBlockSelf(bhVO)
	// engine.Log.Info("广播区块 3333333333333")
}

/*
获取其他见证人的签名
*/
func WitnessSignMulticastBlock(bhVO *BlockHeadVO, witness []*Witness) error {
	bs, err := bhVO.Proto()
	if err != nil {
		return err
	}
	//签名
	// engine.Log.Info("需要签名，group：%d，height：%d", bhVO.BH.GroupHeight, bhVO.BH.Height)
	err = UniformityBroadcastsSign(bhVO, bs, config.MSGID_multicast_blockhead_sign,
		config.CLASS_uniformity_witness_multicas_blockhead, config.Wallet_sync_block_timeout, witness)
	if err != nil {
		engine.Log.Info("签名失败，group：%d，height：%d,err:%s", bhVO.BH.GroupHeight, bhVO.BH.Height, err.Error())
		return err
	}
	// engine.Log.Info("签名成功，group：%d，height：%d", bhVO.BH.GroupHeight, bhVO.BH.Height)
	bhVO.BH.Hash = nil
	bhVO.BH.BuildSign(bhVO.BH.Witness, Area.Keystore)
	bhVO.BH.BuildBlockHash()
	return nil
}

/*
从白名单连接中，获取其他见证人的签名
*/
func UniformityBroadcastsSign(bhVO *BlockHeadVO, hash *[]byte, msgid uint64,
	waitRequestClass string, timeout int64, witness []*Witness) error {
	whiltlistNodes := Area.NodeManager.GetWhiltListNodes()
	// engine.Log.Info("白名单连接数量:%d", len(whiltlistNodes))
	witnessTotal := len(witness)
	//给已发送的节点放map里，避免重复发送
	allNodes := make(map[string]bool)
	signChan := make(chan *[]byte, len(whiltlistNodes))
	for i, _ := range whiltlistNodes {
		sessionid := whiltlistNodes[i]
		//不要发送给自己
		if bytes.Equal(Area.GetNetId(), sessionid) {
			continue
		}
		//去重
		_, ok := allNodes[utils.Bytes2string(sessionid)]
		if ok {
			continue
		}
		allNodes[utils.Bytes2string(sessionid)] = false
		//区块广播给节点
		utils.Go(func() {
			// engine.Log.Info("请求签名:%s", sessionid.B58String())
			res, err := Area.SendNeighborMsgWaitRequest(msgid, &sessionid, hash, time.Second*time.Duration(timeout))
			//engine.Log.Info("sign sessionId :%s", sessionid.B58String())
			if err != nil {
				// engine.Log.Error("请求签名失败:%s %s", sessionid.B58String(), err.Error())
				res = nil
			} else {
				// engine.Log.Info("请求签名成功:%s", sessionid.B58String())
			}
			select {
			case signChan <- res:
			default:
			}
		}, nil)
	}

	witnessPuks := make(map[string][]byte)
	for _, w := range witness {
		witnessPuks[utils.Bytes2string(*w.Addr)] = w.Puk
	}

	var allSign = make(map[string][]byte)
	for i := 0; i < len(whiltlistNodes); i++ {
		signBs := <-signChan
		if signBs == nil || len(*signBs) == 0 {
			continue
		}
		signProto := new(go_protos.BlockSign)
		err := signProto.Unmarshal(*signBs)
		if err != nil {
			engine.Log.Error("签名解析失败：%s", err.Error())
			continue
		}
		ad := crypto.AddressCoin(signProto.GetWitness())
		//engine.Log.Info("签名见证人地址%s", (&ad).B58String())
		puk, ok := witnessPuks[utils.Bytes2string(ad)]
		if !ok {
			continue
		}

		//验证签名
		if !bhVO.BH.CheckExtSignOne(puk, signProto.GetSign()) {
			engine.Log.Info("签名验证不通过的验证人地址%s", (&ad).B58String())
			continue
		}

		allSign[utils.Bytes2string(ad)] = signProto.Sign
		if len(allSign)+1 >= config.BftMajorityPrinciple(witnessTotal) {
			break
		}
	}

	//根据见证人排序签名
	for _, v := range witness {
		//engine.Log.Info("见证人地址%s", v.Addr.B58String())
		sign, ok := allSign[utils.Bytes2string(*v.Addr)]
		if ok {
			//engine.Log.Info("签名通过见证人地址%s", v.Addr.B58String())
			bhVO.BH.SetExtSign(sign)
		} else {
			engine.Log.Info("未签名见证人地址: %s", v.Addr.B58String())
		}
	}
	engine.Log.Info("签名的总量%d，排序后的量%d,见证人总数%d,应该签名成功数量%d", len(allSign), len(bhVO.BH.ExtSign), witnessTotal, config.BftMajorityPrinciple(witnessTotal)-1)
	if len(bhVO.BH.ExtSign)+1 >= config.BftMajorityPrinciple(witnessTotal) {
		return nil
	}
	return config.ERROR_wait_msg_timeout
}

/*
广播挖到的区块
*/
func MulticastBlock(bhVO BlockHeadVO) {
	// engine.Log.Info("广播区块hash：%d %d %s", bhVO.BH.GroupHeight, bhVO.BH.Height,
	// 	hex.EncodeToString(bhVO.BH.Hash))
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	bs, err := bhVO.Proto() //bhVO.Json()
	if err != nil {
		return
	}
	Area.SendMulticastMsg(config.MSGID_multicast_blockhead, bs)
}

/*
广播挖到的区块，当各个见证人都收到后，导入区块
*/
//func MulticastBlockAndImport(bhVO *BlockHeadVO) error {
//	// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
//	// _, file, line, _ := runtime.Caller(0)
//	// engine.AddRuntime(file, line, goroutineId)
//	// defer engine.DelRuntime(file, line, goroutineId)
//	bs, err := bhVO.Proto() //bhVO.Json()
//	if err != nil {
//		return err
//	}
//
//	head := message_center.NewMessageHead(nil, nil, nil, nil, false, "", "")
//	body := message_center.NewMessageBody(config.MSGID_multicast_witness_blockhead, bs, 0, nil, 0)
//	message := message_center.NewMessage(head, body)
//	message.BuildHash()
//	//先保存这个消息到缓存
//	bodyBs, err := body.Proto()
//	if err != nil {
//		return err
//	}
//	// err = new(sqlite3_db.MessageCache).Add(message.Body.Hash, head.Proto(), bodyBs)
//	mmp := pgo_protos.MessageMulticast{
//		Head: head.Proto(),
//		Body: bodyBs,
//	}
//	err = SetMessageMulticast(message.Body.Hash, &mmp)
//	if err != nil {
//		engine.Log.Error(err.Error())
//		return err
//	}
//	engine.Log.Info("multicast message hash:%s", hex.EncodeToString(message.Body.Hash))
//	return MulticastBlockSync(message)
//}

/*
广播挖到的区块，当各个见证人都收到后，导入区块
*/
func MulticastBlockSync(message *message_center.Message) error {
	whiltlistNodes := Area.NodeManager.GetWhiltListNodes()
	return Area.BroadcastsAll(1, config.MSGID_multicast_witness_blockhead, whiltlistNodes, nil, nil, &message.Body.Hash)
}

/*
发送广播
*/
func UniformityBroadcasts(hash *[]byte, msgid uint64, waitRequestClass string, timeout int64) error {
	// timeout := 4
	// timeoutloopTotal := config.Wallet_multicas_block_time / timeout
	whiltlistNodes := Area.NodeManager.GetWhiltListNodes()
	//给已发送的节点放map里，避免重复发送
	allNodes := make(map[string]bool)

	var timeouterrorlock = new(sync.Mutex)
	var timeouterror error

	//先发送给超级节点
	// superNodes := nodeStore.GetLogicNodes()
	//排除重复的地址
	// superNodes = nodeStore.RemoveDuplicateAddress(superNodes)
	cs := make(chan bool, config.CPUNUM)
	group := new(sync.WaitGroup)
	for i, _ := range whiltlistNodes {
		sessionid := whiltlistNodes[i]
		//不要发送给自己
		if bytes.Equal(Area.GetNetId(), sessionid) {
			continue
		}
		_, ok := allNodes[utils.Bytes2string(sessionid)]
		if ok {
			// engine.Log.Info("repeat node addr: %s", sessionid.B58String())
			continue
		}
		allNodes[utils.Bytes2string(sessionid)] = false
		cs <- false
		group.Add(1)
		//区块广播给节点
		// engine.Log.Info("multcast super node:%s", sessionid.B58String())
		utils.Go(func() {
			success := false
			_, err := Area.SendNeighborMsgWaitRequest(msgid, &sessionid, hash, time.Second*time.Duration(timeout))
			//if err == nil && res != nil {
			//	success = true
			//}

			if err == nil {
				success = true
			} else {
				if err.Error() == config.ERROR_wait_msg_timeout.Error() {
				} else {
					//其他错误不管，当作发送成功
					success = true
				}
			}
			if !success {
				timeouterrorlock.Lock()
				timeouterror = config.ERROR_wait_msg_timeout
				timeouterrorlock.Unlock()
			}
			<-cs
			group.Done()
		}, nil)
	}
	group.Wait()
	// engine.Log.Info("multicast whilt list node time %s", config.TimeNow().Sub(start))

	// engine.Log.Info("multicast proxy node time %s", config.TimeNow().Sub(start))
	return timeouterror
}

/*
发送广播
*/
func UniformityBroadcastsImport(hash *[]byte, msgid uint64, waitRequestClass string, timeout int64) error {
	// timeout := 4
	// timeoutloopTotal := config.Wallet_multicas_block_time / timeout
	whiltlistNodes := Area.NodeManager.GetWhiltListNodes()
	//给已发送的节点放map里，避免重复发送
	allNodes := make(map[string]bool)

	var timeouterrorlock = new(sync.Mutex)
	var timeouterror error

	//先发送给超级节点
	// superNodes := nodeStore.GetLogicNodes()
	//排除重复的地址
	// superNodes = nodeStore.RemoveDuplicateAddress(superNodes)
	cs := make(chan bool, config.CPUNUM)
	group := new(sync.WaitGroup)
	for i, _ := range whiltlistNodes {
		sessionid := whiltlistNodes[i]
		//不要发送给自己
		if bytes.Equal(Area.GetNetId(), sessionid) {
			continue
		}
		_, ok := allNodes[utils.Bytes2string(sessionid)]
		if ok {
			// engine.Log.Info("repeat node addr: %s", sessionid.B58String())
			continue
		}
		allNodes[utils.Bytes2string(sessionid)] = false
		cs <- false
		group.Add(1)
		//区块广播给节点
		// engine.Log.Info("multcast super node:%s", sessionid.B58String())
		utils.Go(func() {
			success := false
			_, err := Area.SendNeighborMsgWaitRequest(msgid, &sessionid, hash, time.Second*time.Duration(timeout))
			//engine.Log.Info("导入返回：%s", string(*res))
			//if err == nil && res != nil {
			//	success = true
			//}
			if err == nil {
				success = true
			} else {
				if err.Error() == config.ERROR_wait_msg_timeout.Error() {
				} else {
					//其他错误不管，当作发送成功
					success = true
				}
			}
			if !success {
				timeouterrorlock.Lock()
				timeouterror = config.ERROR_wait_msg_timeout
				timeouterrorlock.Unlock()
			}
			<-cs
			group.Done()
		}, nil)
	}
	group.Wait()
	// engine.Log.Info("multicast whilt list node time %s", config.TimeNow().Sub(start))

	// engine.Log.Info("multicast proxy node time %s", config.TimeNow().Sub(start))
	return timeouterror
}

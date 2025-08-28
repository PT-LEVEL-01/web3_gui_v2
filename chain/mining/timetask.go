package mining

import (
	"bytes"
	"runtime"
	"time"
	"web3_gui/chain/config"

	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
定时出块
@n    int64    间隔时间
*/
func (this *Witness) SyncBuildBlock(n int64) {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	// engine.Log.Info("定时出块 11111 group:%d futer:%s", this.Group.Height, config.TimeNow().Add(time.Duration(n)))
	this.createTime = config.TimeNow()
	this.sleepTime = n
	timer := time.NewTimer(time.Duration(n))
	select {
	case <-timer.C:
	case <-this.StopMining:
		// engine.Log.Info("停止出块 %d", this.Group.Height)
		timer.Stop()
		return
	}
	this.BuildBlock()
	return
}

/*
保底生成备用见证人组
@n    int64    间隔时间
*/
func (this *Witness) SupplementBuildGroup(n int64) {
	//return
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	// engine.Log.Info("定时出块 11111 group:%d futer:%s", this.Group.Height, config.TimeNow().Add(time.Duration(n)))
	this.createTime = config.TimeNow()
	this.sleepTime = n
	timer := time.NewTimer(time.Duration(n))
	select {
	case <-timer.C:
	case <-this.StopMining:
		// engine.Log.Info("停止出块 %d", this.Group.Height)
		timer.Stop()
		return
	}
	// engine.Log.Info("判断见证人组是否足够")
	//查找后面还有没有自己的定时任务，没有了则创建组
	if this.Group.NextGroup != nil {
		// engine.Log.Info("还有下一组")
		return
	}

	//判断是否是最后一个组的最后一个见证人
	if !bytes.Equal(*this.Group.Witness[len(this.Group.Witness)-1].Addr, *this.Addr) {
		// engine.Log.Info("不是最后一个见证人")
		return
	}

	chain := GetLongChain()
	if !chain.SyncBlockFinish {
		engine.Log.Warn("chain is not SyncBlockFinish")
		return
	}

	workModeLockStatic.GetLock()
	defer workModeLockStatic.BackLock()

	//如果节点导入区块速度慢,就会导致原有的定时任务到达时,认为自己是最后一个组了
	//询问一下其它节点的高度,确认本节点是不是因为导入区块慢造成的(等同于正在同步区块)
	logicNodes := Area.NodeManager.GetLogicNodes()
	_, maxRemoteHeight := FindRemoteCurrentHeightWithPeersAndTimeout(logicNodes, time.Second)
	if int64(maxRemoteHeight)-int64(GetLongChain().CurrentBlock) > config.Mining_group_max {
		engine.Log.Warn("节点导入区块滞后,停止构建新组/定时任务:%d,%d", maxRemoteHeight, GetLongChain().CurrentBlock)
		return
	}
	//chain := GetLongChain()

	//查找出未确认的块
	preBlock, _ := this.FindUnconfirmedBlock()
	// engine.Log.Info("未确认的最后块:%d", preBlock.Height)
	//先统计之前的区块
	preBlock.Witness.Group.BuildGroup(&preBlock.Id)
	chain.CountBlock(preBlock.Witness.Group)

	//engine.Log.Info("后面没有见证者组了，重新构建,当前组：%d", this.Group.Height)
	//先保底构建一个组
	chain.WitnessChain.AdditionalWitnessBackup()
	chain.WitnessChain.BuildWitnessGroupSupplemental()
	chain.WitnessChain.BuildMiningTime()
	return
}

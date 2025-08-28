package engine

import (
	"time"
	"web3_gui/utils"
)

/*
块文件大小计算器
从1KB大小开始下载，时间少于10ms，则大小翻倍到2KB
当大于1MB时，时间少于10ms，则大小增加1MB
*/
type ChunkSizeCalculator struct {
	spendList    []time.Duration //花费时间列表
	networkDelay time.Duration   //网络延迟
	chunkSize    utils.Byte      //当前块大小
}

/*
创建一个块文件大小计算器
*/
func NewChunkSizeCalculator() *ChunkSizeCalculator {
	total := 3
	csc := ChunkSizeCalculator{
		spendList: make([]time.Duration, total),
		chunkSize: utils.KB, //1KB
	}
	return &csc
}

/*
获取块大小
考虑有的网络基本延迟就在400ms，所以传再小的数据包，返回都要400ms
前3个包都传1KB，计算网络平均延迟
当网络延迟大于10秒钟，则不加大包容量
@spend         time.Duration    上一次包传输花费时间
@fullPacket    bool             上一次实际传输内容大小是否等于计算的包大小
*/
func (this *ChunkSizeCalculator) GetChunkSize(spend time.Duration, fullPacket bool) utils.Byte {
	//第一个包传1KB
	if spend == 0 {
		//this.chunkSize = utils.KB
		return this.chunkSize
	}
	//前n个包都传1KB，用于计算平均延迟
	if this.networkDelay == 0 {
		for i, one := range this.spendList {
			if one == 0 {
				this.spendList[i] = spend
			}
		}
		if this.spendList[len(this.spendList)-1] == 0 {
			this.chunkSize = utils.KB
			return this.chunkSize
		}
		//计算平均网络延迟
		var spendTotal time.Duration
		for _, one := range this.spendList {
			spendTotal += one
		}
		this.networkDelay = spendTotal / time.Duration(len(this.spendList))
	}
	//当网络延迟大于1s的时候，网络极差
	if this.networkDelay >= time.Second {
		//如果本次消息花费时间有明显改善，则重新测试网络延迟
		if spend <= this.networkDelay/2 {
			this.reassessment()
			return this.chunkSize
		}
		//若没有明显改善
		this.chunkSize = utils.KB
		return this.chunkSize
	} else if this.networkDelay < time.Second/100 {
		//当网络延迟小于10ms，则网络非常好，可能是内网，包大小指数级增加
		if spend <= time.Second/100 {
			if fullPacket {
				this.chunkSize = this.chunkSize * 2
			}
			return this.chunkSize
		} else {
			this.chunkSize = this.chunkSize / 2
			//当网络变差，包大小只有1KB的时候，重新评估网络
			if this.chunkSize <= utils.KB {
				this.reassessment()
			}
			return this.chunkSize
		}
	} else {
		//当网络延迟大于10ms，小于1s
		if spend > this.networkDelay*3 {
			//网络变差了，重新评估网络
			this.reassessment()
			return this.chunkSize
		} else if spend <= this.networkDelay/2 {
			//网络变好了
			this.reassessment()
			return this.chunkSize
		} else if spend <= this.networkDelay*2 {
			if fullPacket {
				this.chunkSize = this.chunkSize * 2
			}
			return this.chunkSize
		} else {
			this.chunkSize = this.chunkSize / 2
			return this.chunkSize
		}
	}
}

/*
清理缓存，并重新评估网络
*/
func (this *ChunkSizeCalculator) reassessment() {
	for i, _ := range this.spendList {
		this.spendList[i] = 0
	}
	this.networkDelay = 0
	this.chunkSize = utils.KB
}

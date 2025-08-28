package engine

import (
	"context"
	"strconv"
	"sync"
	"time"

	"web3_gui/utils"
)

type SendQueue struct {
	name        string             //
	targetName  string             //
	objctID     uint64             //
	lock        *sync.Mutex        //
	cursorID    uint64             //
	queue       chan *SendPacket   //
	packetMap   *sync.Map          //
	contextRoot context.Context    //
	canceRoot   context.CancelFunc //
}

func NewSendQueue(cache uint64, c context.Context, cance context.CancelFunc, name string) *SendQueue {
	objctID := utils.GetAccNumber()
	// utils.Log.Info().Msgf("发送通道 创建 objctID:%d self:%s", objctID, name)
	sendQueue := &SendQueue{
		name:        name,
		objctID:     objctID,
		lock:        new(sync.Mutex),
		cursorID:    0,
		queue:       make(chan *SendPacket, cache),
		packetMap:   new(sync.Map),
		contextRoot: c,
		canceRoot:   cance,
	}

	return sendQueue
}

/*
添加
*/
func (this *SendQueue) AddAndWaitTimeout(bs *[]byte, timeout time.Duration) error {
	defer utils.PrintPanicStack(nil)

	this.lock.Lock()
	this.cursorID += 1
	id := this.cursorID
	this.lock.Unlock()

	if timeout == 0 {
		timeout = SendTimeOut
	}
	// utils.Log.Info().Msgf("发送通道 等待消息返回 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)

	// start := time.Now()
	// randStr := utils.GetRandomDomain() + utils.GetRandomDomain()

	// 添加消息时, 每500条打印一次发送队列相关内容
	if this.cursorID > 0 && this.cursorID%500 == 0 {
		utils.Log.Warn().Msgf("[发送通道信息] objctID:%d target:%s 总发送数量:%d 剩余发送长度:%d", this.objctID, this.targetName, this.cursorID, len(this.queue))
	}

	sp := &SendPacket{
		ID: id, //
		// timeout:   time.Now().Add(timeout), //超时时间
		bs:        bs,                  //
		resultErr: make(chan error, 1), //
	}
	// utils.Log.Info().Msgf("添加key:%d", id)

	this.packetMap.Store(strconv.Itoa(int(id)), sp)
	select {
	case <-this.contextRoot.Done():
		// utils.Log.Info().Msgf("发送通道 队列销毁1 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		this.packetMap.Delete(strconv.Itoa(int(id)))
		return ERROR_send_cache_close
	case this.queue <- sp:
		// utils.Log.Info().Msgf("发送通道 添加到队列 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		// utils.Log.Info().Msgf("添加到队列:%d", id)
		// utils.Log.Info().Msgf("添加到队列:%s timeout:%d", randStr, timeout/time.Second)
	default:
		// utils.Log.Info().Msgf("删除队列:%d", id)
		this.packetMap.Delete(strconv.Itoa(int(id)))
		return ERROR_send_cache_full
	}

	timer := time.NewTimer(timeout)
	var err error
	select {
	case <-this.contextRoot.Done():
		// utils.Log.Info().Msgf("队列销毁")
		// utils.Log.Info().Msgf("发送通道 队列销毁2 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		err = ERROR_send_cache_close
		timer.Stop()
	case err = <-sp.resultErr:
		// utils.Log.Info().Msgf("发送通道 有返回 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		//返回
		timer.Stop()
		// utils.Log.Info().Msgf("队列返回 %d:%s", id, err)
		// utils.Log.Info().Msgf("返回错误了%d:%s", id, err)
	case <-timer.C:
		//超时
		err = ERROR_send_timeout
		// utils.Log.Info().Msgf("发送通道 返回超时了 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		// utils.Log.Info().Msgf("返回超时了:%d", id)
		// utils.Log.Info().Msgf("%p 队列超时了id:%d 耗时:%s 超时时间:%s", this, id, time.Now().Sub(start), timeout)
	}
	this.packetMap.Delete(strconv.Itoa(int(id)))
	// utils.Log.Info().Msgf("发送通道 执行完 objctID:%d id:%d self:%s target:%s err:%s", this.objctID, id, this.name, this.targetName, err)
	return err
}

func (this *SendQueue) GetQueueChan() <-chan *SendPacket {
	return this.queue
}

/*
设置返回
*/
func (this *SendQueue) SetResult(id uint64, err error) {
	// utils.Log.Info().Msgf("队列设置返回key:%d", id)
	spItr, ok := this.packetMap.Load(strconv.Itoa(int(id)))
	if !ok {
		// utils.Log.Info().Msgf("%p 未找到key:%d", this, id)
		return
	}
	// utils.Log.Info().Msgf("找到key:", strconv.Itoa(int(id)))
	sp := spItr.(*SendPacket)
	select {
	case sp.resultErr <- err:
		// utils.Log.Info().Msgf("返回了", id, err)
	default:
	}
}

/*
设置返回
*/
func (this *SendQueue) Destroy() {
	close(this.queue)
	this.canceRoot()
	// utils.Log.Info().Msgf("发送通道 Destroy objctID:%d self:%s target:%s", this.objctID, this.name, this.targetName)
}

type SendPacket struct {
	ID        uint64     //
	timeout   time.Time  //超时时间
	bs        *[]byte    //
	resultErr chan error //返回错误
}

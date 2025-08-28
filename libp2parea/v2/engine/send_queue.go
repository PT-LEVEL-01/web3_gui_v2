package engine

import (
	"context"
	"github.com/rs/zerolog"
	"math/big"
	"sync"
	"time"
	"web3_gui/utils"
)

type SendQueue struct {
	lock        *sync.Mutex        //
	cursorID    *big.Int           //自增长唯一ID
	queue       chan *SendPacket   //发送队列
	packetMap   *sync.Map          //发送队列中的消息索引。key:string=自增长id;value:*SendPacket=待发送的数据包;
	contextRoot context.Context    //
	canceRoot   context.CancelFunc //
	Log         **zerolog.Logger   //
}

func NewSendQueue(cache uint64, c context.Context, cance context.CancelFunc, log **zerolog.Logger) *SendQueue {
	//objctID := utils.GetAccNumber()
	// Log.Info("发送通道 创建 objctID:%d self:%s", objctID, name)
	sendQueue := &SendQueue{
		lock:        new(sync.Mutex),
		cursorID:    big.NewInt(0),
		queue:       make(chan *SendPacket, cache),
		packetMap:   new(sync.Map),
		contextRoot: c,
		canceRoot:   cance,
		Log:         log,
	}

	return sendQueue
}

/*
添加
*/
func (this *SendQueue) AddAndWaitTimeout(bs *[]byte, timeout time.Duration) utils.ERROR {
	ERR := utils.NewErrorSuccess()
	defer utils.PrintPanicStack(*this.Log)
	this.lock.Lock()
	this.cursorID = new(big.Int).Add(this.cursorID, big.NewInt(1))
	id := this.cursorID.Bytes()
	if timeout == 0 {
		timeout = SendTimeOut
	}
	// Log.Info("发送通道 等待消息返回 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
	sp := &SendPacket{
		ID:        id,                        //
		timeout:   time.Now().Add(timeout),   //
		bs:        bs,                        //
		resultErr: make(chan utils.ERROR, 1), //
	}
	idkey := utils.Bytes2string(id)
	this.packetMap.Store(idkey, sp)
	select {
	case <-this.contextRoot.Done():
		// Log.Info("发送通道 队列销毁1 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		this.packetMap.Delete(idkey)
		ERR = utils.NewErrorBus(ERROR_code_send_cache_close, "")
	default:
		select {
		case <-this.contextRoot.Done():
			// Log.Info("发送通道 队列销毁1 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
			this.packetMap.Delete(idkey)
			ERR = utils.NewErrorBus(ERROR_code_send_cache_close, "")
		case this.queue <- sp:
			// Log.Info("发送通道 添加到队列 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		default:
			// Log.Info("删除队列:%d", id)
			this.packetMap.Delete(idkey)
			ERR = utils.NewErrorBus(ERROR_code_send_cache_full, "")
		}
	}
	this.lock.Unlock()
	if ERR.CheckFail() {
		return ERR
	}
	timer := time.NewTimer(timeout)
	ERR = utils.NewErrorSuccess()
	select {
	case <-this.contextRoot.Done():
		// Log.Info("发送通道 队列销毁2 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		ERR = utils.NewErrorBus(ERROR_code_send_cache_close, "")
		timer.Stop()
	case ERR = <-sp.resultErr:
		// Log.Info("发送通道 有返回 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
		//返回
		timer.Stop()
	case <-timer.C:
		//超时
		ERR = utils.NewErrorBus(ERROR_code_timeout_send, "")
		// Log.Info("发送通道 返回超时了 objctID:%d id:%d self:%s target:%s", this.objctID, id, this.name, this.targetName)
	}
	this.packetMap.Delete(idkey)
	// Log.Info("发送通道 执行完 objctID:%d id:%d self:%s target:%s err:%s", this.objctID, id, this.name, this.targetName, err)
	return ERR
}

func (this *SendQueue) GetQueueChan() <-chan *SendPacket {
	return this.queue
}

/*
设置返回
*/
func (this *SendQueue) SetResult(id []byte, ERR utils.ERROR) {
	// Log.Info("队列设置返回key:%d", id)
	spItr, ok := this.packetMap.Load(utils.Bytes2string(id))
	if !ok {
		// Log.Info("%p 未找到key:%d", this, id)
		return
	}
	// Log.Info("找到key:", strconv.Itoa(int(id)))
	sp := spItr.(*SendPacket)
	select {
	case sp.resultErr <- ERR:
		// Log.Info("返回了", id, err)
	default:
	}
}

/*
设置返回
*/
func (this *SendQueue) Destroy() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.canceRoot()
	close(this.queue)
	//utils.Log.Info().Hex("发送通道销毁", this.cursorID.Bytes()).Send()
}

type SendPacket struct {
	ID        []byte           //
	timeout   time.Time        //超时时间
	bs        *[]byte          //
	resultErr chan utils.ERROR //返回错误
}

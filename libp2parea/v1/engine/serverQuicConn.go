package engine

import (
	"encoding/json"
	"github.com/quic-go/quic-go"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

// 其他计算机对本机的连接
type ServerQuicConn struct {
	sessionBase                      //
	index            uint64          //
	conn             quic.Connection //
	stream           quic.Stream     //
	Ip               string          //
	Connected_time   string          //
	CloseTime        string          //
	outChan          chan *[]byte    //发送队列
	outChanCloseLock *sync.Mutex     //是否关闭发送队列
	outChanIsClose   bool            //是否关闭发送队列。1=关闭
	engine           *Engine         //
	controller       Controller      //
	sendQueue        *SendQueue      //
	isClose          uint32          //
	allowClose       chan bool       //允许调用关闭回调函数，保证先后顺序，close函数最后调用
}

/*
尝试设置为是否关闭
@return    bool    是否设置成功
*/
func (this *ServerQuicConn) SetTryIsClose(isClose bool) bool {
	if isClose {
		return atomic.CompareAndSwapUint32(&this.isClose, 0, 1)
	} else {
		return atomic.CompareAndSwapUint32(&this.isClose, 1, 0)
	}
}

func (this *ServerQuicConn) GetIndex() uint64 {
	return this.index
}

func (this *ServerQuicConn) run() {
	go this.loopSend()
	go this.recv()

}

// 接收客户端消息协程
func (this *ServerQuicConn) recv() {
	defer utils.PrintPanicStack(nil)
	//处理客户端主动断开连接的情况
	var err error
	var handler MsgHandler
	for {
		var packet *Packet
		// if this.isOldVersion {
		// 	packet, err = RecvPackage_old(this.conn)
		// } else {
		packet, err = RecvQuicPackage(this.stream)
		// }
		if err != nil {
			if err.Error() == io.EOF.Error() {
			} else if strings.Contains(err.Error(), "use of closed network connection") {
			} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
			} else if strings.Contains(err.Error(), "connection reset by peer") {
			} else {
				utils.Log.Warn().Msgf("network error ServerConn %s", err.Error())
			}
			break
		} else {
			if packet.MsgID == MSGID_heartbeat {
				// utils.Log.Info().Msgf("收到心跳serverConn")
				//收到心跳并返回
				buff, err := MarshalPacket(MSGID_heartbeat, nil, nil)
				_, err = this.stream.Write(*buff)
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
					} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
					} else if strings.Contains(err.Error(), "An established connection was aborted by the software in your host machine") {
					} else {
						utils.Log.Warn().Msgf("loopSend error:%s", err.Error())
					}
					break
				}
				continue
			}
			packet.Session = this
			// utils.Log.Debug().Msgf("conn recv: %d %s %d\n%s", packet.MsgID, this.Ip, len(packet.Data)+len(packet.Dataplus)+16, hex.EncodeToString(append(packet.Data, packet.Dataplus...)))
			// utils.Log.Debug().Msgf("server conn recv: %d %s <- %s %d", packet.MsgID, this.Ip, this.conn.RemoteAddr(), len(packet.Data)+len(packet.Dataplus)+16)
			handler = this.engine.router.GetHandler(this.areaName, packet.MsgID)
			if handler == nil {
				utils.Log.Warn().Msgf("server The message is not registered, message number: %d", packet.MsgID)
			} else {
				//这里决定了消息是否异步处理
				go this.handlerProcess(handler, packet)
			}
		}
	}
	this.Close()
	//最后一个包接收了之后关闭chan
	//如果有超时包需要等超时了才关闭，目前未做处理
	this.outChanCloseLock.Lock()
	this.outChanIsClose = true
	close(this.outChan)
	this.outChanCloseLock.Unlock()
	// fmt.Println("关闭连接")
}

func (this *ServerQuicConn) loopSend() {
	// timer := time.NewTimer(heartBeat_interval)

	defer this.conn.CloseWithError(0, "")
	// defer timer.Stop()

	var isClose bool
	c := this.sendQueue.GetQueueChan()
	mu := new(sync.Mutex)
	// 开启多个协程，同时从sendQueue中读取消息
	var wg sync.WaitGroup
	for i := 0; i < MaxSendQueueDealCnt; i++ {
		wg.Add(1)
		go func(l *sync.Mutex) {
			defer wg.Done()

			var sp *SendPacket
			for {
				sp, isClose = <-c

				if !isClose {
					// utils.Log.Warn().Msgf("out chan is close")
					return
				}
				l.Lock()
				_, err := this.stream.Write(*sp.bs)
				l.Unlock()
				this.sendQueue.SetResult(sp.ID, err)
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
					} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
					} else {
						utils.Log.Warn().Msgf("loopSend error:%s", err.Error())
					}
					return
				} else {
					// utils.Log.Debug().Msgf("conn send: %d %s %d %d\n%s", msgID, this.conn.RemoteAddr(), len(*buff), n, hex.EncodeToString(*buff))
					// utils.Log.Debug().Msgf("server conn send: %s -> %s %d", this.conn.LocalAddr(), this.conn.RemoteAddr(), len(*sp.bs))
					// utils.Log.Info().Msgf("clent send %s", hex.EncodeToString(*buff))
				}
			}
		}(mu)
	}
	wg.Wait()
}

func (this *ServerQuicConn) handlerProcess(handler MsgHandler, msg *Packet) {
	//消息处理模块报错将不会引起宕机
	defer utils.PrintPanicStack(nil)
	//消息处理前先通过拦截器
	itps := this.engine.interceptor.getInterceptors()
	itpsLen := len(itps)
	for i := 0; i < itpsLen; i++ {
		isIntercept := itps[i].In(this.controller, *msg)
		//
		if isIntercept {
			return
		}
	}
	handler(this.controller, *msg)
	//消息处理后也要通过拦截器
	for i := itpsLen; i > 0; i-- {
		itps[i-1].Out(this.controller, *msg)
	}
}

/*
发送序列化后的数据
*/
func (this *ServerQuicConn) Send(msgID uint64, data, dataplus *[]byte, timeout time.Duration) error {
	var buff *[]byte
	var err error
	buff, err = MarshalPacket(msgID, data, dataplus)
	if err != nil {
		return err
	}
	// utils.Log.Info().Msgf("Send server 111111111")
	return this.sendQueue.AddAndWaitTimeout(buff, timeout)
}

// 给客户端发送数据
func (this *ServerQuicConn) SendJSON(msgID uint64, data interface{}, waite bool) error {
	defer utils.PrintPanicStack(nil)
	// this.packet.IsWait = waite

	f, err := json.Marshal(data)
	if err != nil {
		return err
	}
	buff, err := MarshalPacket(msgID, &f, nil)
	if err != nil {
		return err
	}
	_, err = this.stream.Write(*buff)
	return err
}

// 关闭这个连接
func (this *ServerQuicConn) Close() {
	if !this.SetTryIsClose(true) {
		return
	}
	<-this.allowClose

	closeCallback, exist := this.engine.GetCloseCallback(this.areaName)
	if exist && closeCallback != nil {
		closeCallback(this)
	}
	this.engine.sessionStore.removeSession(this.areaName, this)
	err := this.conn.CloseWithError(0, "")
	if err != nil {
	}
	this.sendQueue.Destroy()
}

func (this *ServerQuicConn) SetName(name string) {
	this.engine.sessionStore.renameSession(this.GetIndex(), name)
	this.name = name
}

// 获取远程ip地址和端口
func (this *ServerQuicConn) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

// 获取节点机器Id
func (this *ServerQuicConn) GetMachineID() string {
	return this.machineID
}

/*
 * GetAreaName 获取区域名称
 *
 * @return	areaName	string	区域名称
 */
func (this *ServerQuicConn) GetAreaName() string {
	return this.areaName
}

/*
 * GetSetGodTime 获取设置超级代理的时间
 *
 * @return	setGodTime	int64	设置超级代理的时间
 */
func (this *ServerQuicConn) GetSetGodTime() int64 {
	return this.setGodTime
}

package engine

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

// 其他计算机对本机的连接
type ServerConn struct {
	sessionBase                   //
	index            uint64       //
	conn             net.Conn     //
	Ip               string       //
	Connected_time   string       //
	CloseTime        string       //
	outChan          chan *[]byte //发送队列
	outChanCloseLock *sync.Mutex  //是否关闭发送队列
	outChanIsClose   bool         //是否关闭发送队列。1=关闭
	engine           *Engine      //
	controller       Controller   //
	sendQueue        *SendQueue   //
	isOldVersion     bool         //是否旧版本
	isClose          uint32       //
	allowClose       chan bool    //允许调用关闭回调函数，保证先后顺序，close函数最后调用
}

/*
尝试设置为是否关闭
@return    bool    是否设置成功
*/
func (this *ServerConn) SetTryIsClose(isClose bool) bool {
	if isClose {
		return atomic.CompareAndSwapUint32(&this.isClose, 0, 1)
		// atomic.StoreUint32(&this.isClose, 1)
	} else {
		return atomic.CompareAndSwapUint32(&this.isClose, 1, 0)
		// atomic.StoreUint32(&this.isClose, 0)
	}
}

func (this *ServerConn) GetIndex() uint64 {
	return this.index
}

func (this *ServerConn) run() {

	// this.packet.Session = this
	go this.loopSend()
	go this.recv()

}

// 接收客户端消息协程
func (this *ServerConn) recv() {
	defer utils.PrintPanicStack(nil)
	//处理客户端主动断开连接的情况
	var err error
	var handler MsgHandler
	for {
		var packet *Packet
		// if this.isOldVersion {
		// 	packet, err = RecvPackage_old(this.conn)
		// } else {
		packet, err = RecvPackage(this.conn)
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
				_, err = this.conn.Write(*buff)
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
			//utils.Log.Debug().Msgf("server conn recv: %d %s <- %s %d", packet.MsgID, this.Ip, this.conn.RemoteAddr(), len(packet.Data)+len(packet.Dataplus)+16)
			handler = this.engine.router.GetHandler(this.areaName, packet.MsgID)
			if handler == nil {
				utils.Log.Warn().Msgf("server The message is not registered, message number：%d", packet.MsgID)
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

func (this *ServerConn) loopSend() {
	// timer := time.NewTimer(heartBeat_interval)

	defer this.conn.Close()
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
				// timer.Reset(heartBeat_interval)
				// select {
				// case sp, isClose = <-c:
				// case <-timer.C:
				// 	//发送心跳
				// 	// buff, err := MarshalPacket(MSGID_heartbeat, nil, nil)
				// 	// _, err = this.conn.Write(*buff)
				// 	// if err != nil {
				// 	// 	if strings.Contains(err.Error(), "use of closed network connection") {
				// 	// 	} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
				// 	// 	} else {
				// 	// 		utils.Log.Warn().Msgf("loopSend error:%s", err.Error())
				// 	// 	}
				// 	// 	return
				// 	// }
				// 	continue
				// }

				if !isClose {
					// utils.Log.Warn().Msgf("out chan is close")
					return
				}
				// err := this.conn.SetWriteDeadline(sp.timeout)
				// if err != nil {
				// 	utils.Log.Warn().Msgf("loopSend error:%s", err.Error())
				// 	return
				// }
				l.Lock()
				_, err := this.conn.Write(*sp.bs)
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
					//utils.Log.Debug().Msgf("server conn send: %s -> %s %d", this.conn.LocalAddr(), this.conn.RemoteAddr(), len(*sp.bs))
					// utils.Log.Info().Msgf("clent send %s", hex.EncodeToString(*buff))
				}
			}
		}(mu)
	}
	wg.Wait()
}

func (this *ServerConn) loopSendOld() {
	var n int
	var err error
	var buff *[]byte
	var isClose = false
	var count = 5
	var total = 0
	for {
		buff, isClose = <-this.outChan
		if !isClose {
			utils.Log.Warn().Msgf("out chan is close")
			break
		}
		this.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
		n, err = this.conn.Write(*buff)
		if err != nil {
			total++
			utils.Log.Warn().Msgf("conn send err: %s", err.Error())
			if total > count {
				this.conn.Close()
			}
		} else {
			total = 0
			// utils.Log.Debug().Msgf("conn send: %d %s %d %d\n%s", msgID, this.conn.RemoteAddr(), len(*buff), n, hex.EncodeToString(*buff))
			// utils.Log.Debug().Msgf("client conn send: %s %d", this.conn.RemoteAddr(), len(*buff))
			// utils.Log.Info().Msgf("clent send %s", hex.EncodeToString(*buff))
		}
		if n < len(*buff) {
			// utils.Log.Warn().Msgf("conn send warn: %d %s length %d %d", msgID, this.conn.RemoteAddr().String(), len(*buff), n)
		}
	}
}

// func (this *ServerConn) Waite(du time.Duration) *Packet {
// 	if this.packet.Wait(du) {
// 		return &this.packet
// 	}
// 	return nil
// }

// func (this *ServerConn) FinishWaite() {
// 	this.packet.FinishWait()
// }

func (this *ServerConn) handlerProcess(handler MsgHandler, msg *Packet) {
	// this.engine.handlerTokenChan <- false
	// goroutineId := GetRandomDomain() + TimeFormatToNanosecondStr()
	// _, file, line, _ := runtime.Caller(0)
	// AddRuntime(file, line, goroutineId)
	// defer DelRuntime(file, line, goroutineId)

	//消息处理模块报错将不会引起宕机
	defer utils.PrintPanicStack(nil)
	// defer func() {
	// 	<-this.engine.handlerTokenChan
	// }()
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
func (this *ServerConn) Send(msgID uint64, data, dataplus *[]byte, timeout time.Duration) error {
	// this.outChanCloseLock.Lock()
	// if this.outChanIsClose {
	// 	this.outChanCloseLock.Unlock()
	// 	return errors.New("send channel is close")
	// }
	// buff, err := MarshalPacket(msgID, data, dataplus)

	var buff *[]byte
	var err error
	// if this.isOldVersion {
	// 	buff = MarshalPacket_old(msgID, data, dataplus)
	// } else {
	buff, err = MarshalPacket(msgID, data, dataplus)
	// }
	if err != nil {
		// this.outChanCloseLock.Unlock()
		return err
	}
	// utils.Log.Info().Msgf("Send server 111111111")
	return this.sendQueue.AddAndWaitTimeout(buff, timeout)
}

// 给客户端发送数据
func (this *ServerConn) SendOld(msgID uint64, data, dataplus *[]byte, waite bool) error {
	this.outChanCloseLock.Lock()
	if this.outChanIsClose {
		this.outChanCloseLock.Unlock()
		return errors.New("send channel is close")
	}
	buff, err := MarshalPacket(msgID, data, dataplus)
	if err != nil {
		this.outChanCloseLock.Unlock()
		return err
	}
	select {
	case this.outChan <- buff:
	default:
		addr := AddressNet([]byte(this.GetName()))
		utils.Log.Warn().Msgf("conn send err chan is full :%s", addr.B58String())
		this.conn.Close()
	}
	this.outChanCloseLock.Unlock()
	return nil
}

//给客户端发送数据
// func (this *ServerConn) Send(msgID uint64, data, dataplus *[]byte, waite bool) (err error) {
// 	defer PrintPanicStack()
// 	// this.packet.IsWait = waite
// 	buff := MarshalPacket(msgID, data, dataplus)
// 	var n int
// 	n, err = this.conn.Write(*buff)
// 	if err != nil {
// 		utils.Log.Warn().Msgf("conn send err: %s", err.Error())
// 	} else {
// 		// utils.Log.Debug().Msgf("conn send: %d %s %d %d\n%s", msgID, this.Ip, len(buff), n, hex.EncodeToString(buff))
// 		// utils.Log.Debug().Msgf("server conn send: %d %s %d", msgID, this.conn.RemoteAddr(), len(*buff))
// 	}
// 	if n < len(*buff) {
// 		utils.Log.Warn().Msgf("conn send warn: %d %s length %d %d", msgID, this.conn.RemoteAddr().String(), len(*buff), n)
// 	}
// 	return
// }

// 给客户端发送数据
func (this *ServerConn) SendJSON(msgID uint64, data interface{}, waite bool) error {
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
	_, err = this.conn.Write(*buff)
	return err
}

// 关闭这个连接
func (this *ServerConn) Close() {
	if !this.SetTryIsClose(true) {
		return
	}
	<-this.allowClose

	closeCallback, exist := this.engine.GetCloseCallback(this.areaName)
	if exist && closeCallback != nil {
		closeCallback(this)
	}
	this.engine.sessionStore.removeSession(this.areaName, this)
	err := this.conn.Close()
	if err != nil {
	}
	this.sendQueue.Destroy()
}

func (this *ServerConn) SetName(name string) {
	this.engine.sessionStore.renameSession(this.GetIndex(), name)
	this.name = name
}

// 获取远程ip地址和端口
func (this *ServerConn) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

// 获取节点机器Id
func (this *ServerConn) GetMachineID() string {
	return this.machineID
}

/*
 * GetAreaName 获取区域名称
 *
 * @return	areaName	string	区域名称
 */
func (this *ServerConn) GetAreaName() string {
	return this.areaName
}

/*
 * GetSetGodTime 获取设置超级代理的时间
 *
 * @return	setGodTime	int64	设置超级代理的时间
 */
func (this *ServerConn) GetSetGodTime() int64 {
	return this.setGodTime
}

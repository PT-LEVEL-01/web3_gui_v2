package engine

import (
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"web3_gui/utils"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

// 本机向其他服务器的连接
type Client struct {
	sessionBase
	index            uint64
	serverName       string
	ip               string
	port             uint32
	conn             net.Conn
	inPack           chan *Packet //接收队列
	outChan          chan *[]byte //发送队列
	outChanCloseLock *sync.Mutex  //是否关闭发送队列
	outChanIsClose   bool         //是否关闭发送队列。1=关闭
	isPowerful       bool         //是否是强连接，强连接有短线重连功能
	engine           *Engine      //
	controller       Controller   //
	sendQueue        *SendQueue   //
	isOldVersion     bool         //是否旧版本
	isClose          uint32       //
	allowClose       chan bool    //允许调用关闭回调函数，保证先后顺序，close函数最后调用
	heartbeat        chan bool    //
	getDataSign      chan bool    // 接收到数据信号，用来优化心跳超时关闭的问题
}

/*
尝试设置为是否关闭
@return    bool    是否设置成功
*/
func (this *Client) SetTryIsClose(isClose bool) bool {
	if isClose {
		return atomic.CompareAndSwapUint32(&this.isClose, 0, 1)
		// atomic.StoreUint32(&this.isClose, 1)
	} else {
		return atomic.CompareAndSwapUint32(&this.isClose, 1, 0)
		// atomic.StoreUint32(&this.isClose, 0)
	}
}

func (this *Client) GetIndex() uint64 {
	return this.index
}

func (this *Client) run() {
	go this.loopSend()
	go this.recv()
	go this.hold()
}

func (this *Client) Connect(ip string, port uint32) (remoteName string, params interface{}, connectKey string, err error) {
	// 如果设置了上帝地址，则忽略其他非上帝地址的连接
	var setGodAddr bool
	if godAddrInfo, exist := this.engine.godInfoes[this.areaName]; exist && godAddrInfo.godIp != "" {
		if godAddrInfo.godIp != ip || godAddrInfo.godPort != uint16(port) {
			// utils.Log.Warn().Msgf("qlw-----不是上帝节点的地址, 不允许连接!!!, ip:%s, port:%d", ip, port)
			return "", nil, "", errors.New("不能连接非上帝节点地址")
		} else {
			utils.Log.Info().Msgf("连上帝节点中！！！！！")
			setGodAddr = true
		}
	}

	utils.Log.Debug().Msgf("start Connecting to:%s:%s selfAreaName:%s", ip, strconv.Itoa(int(port)), hex.EncodeToString([]byte(this.areaName)))
	this.ip = ip
	this.port = port
	this.outChan = make(chan *[]byte, SendQueueCacheNum)
	this.outChanCloseLock = new(sync.Mutex)
	this.outChanIsClose = false

	this.conn, err = net.DialTimeout("tcp", ip+":"+strconv.Itoa(int(port)), time.Second*3)
	if err != nil {
		utils.Log.Error().Msgf("connect error:%s", err.Error())
		return
	}
	// utils.Log.Debug().Msgf("Connecting to 11111:%s:%s", ip, strconv.Itoa(int(port)))
	host, portStr, err := net.SplitHostPort(this.conn.RemoteAddr().String())
	if err != nil {
		utils.Log.Error().Msgf("connect error:%s", err.Error())
		return
	}
	portTemp, err := strconv.Atoi(portStr)
	if err != nil {
		return
	}
	this.ip = host
	this.port = uint32(portTemp)
	// utils.Log.Debug().Msgf("Connecting to 22222:%s:%s", ip, strconv.Itoa(int(port)))
	//权限验证
	auth, exist := this.engine.GetAuth(this.areaName)
	if !exist || auth == nil {
		utils.Log.Error().Msgf("areaName auth don't exist!!!")
		err = errors.New("areaName auth don't exist")
		return
	}
	var setGodTime int64
	remoteName, this.machineID, setGodTime, params, connectKey, err = auth.SendKey(this.conn, this, this.serverName, setGodAddr)
	this.setGodTime = setGodTime
	// utils.Log.Debug().Msgf("Connecting to 33333:%s:%s", ip, strconv.Itoa(int(port)))
	if err != nil {
		utils.Log.Error().Msgf("connect error:%s", err.Error())
		this.conn.Close()
		// this.Close()
		if err.Error() == Error_different_netid.Error() {
			return
		}
		if err.Error() == Error_node_unwanted.Error() {
			return
		}
		return
		//使用旧版本
		// this.conn, err = net.DialTimeout("tcp", ip+":"+strconv.Itoa(int(port)), time.Second*3)
		// if err != nil {
		// 	return
		// }
		// //权限验证
		// remoteName, err = this.engine.auth.SendKeyOld(this.conn, this, this.serverName)
		// if err != nil {
		// 	// this.conn.Close()
		// 	this.Close()
		// 	return
		// }
		// this.isOldVersion = true
	}

	utils.Log.Debug().Msgf("end Connecting to:%s:%s local:%s machineId:%s", ip, strconv.Itoa(int(port)), this.conn.LocalAddr().String(), this.machineID)

	this.controller = &ControllerImpl{
		lock:       new(sync.RWMutex),
		engine:     this.engine,
		attributes: make(map[string]interface{}),
	}
	return
}
func (this *Client) reConnect() {
	for {
		//十秒钟后重新连接
		time.Sleep(time.Second * 10)

		// 如果设置了上帝地址，则不再主动连接其他地址
		// if this.engine.godIp != "" {
		if godAddrInfo, exist := this.engine.godInfoes[this.areaName]; exist && godAddrInfo.godIp != "" {
			if godAddrInfo.godIp != this.ip || godAddrInfo.godPort != uint16(this.port) {
				// utils.Log.Warn().Msgf("qlw-----不是上帝节点的地址, 不允许连接!!!, ip:%s, port:%d", this.ip, this.port)
				continue
			}
		}

		var err error
		this.conn, err = net.Dial("tcp", this.ip+":"+strconv.Itoa(int(this.port)))
		if err != nil {
			continue
		}

		utils.Log.Debug().Msgf("Connecting to %s:%s", this.ip, strconv.Itoa(int(this.port)))

		go this.loopSend()
		go this.recv()
		go this.hold()
		return
	}
}

func (this *Client) recv() {
	defer utils.PrintPanicStack(nil)
	// defer utils.Log.Info().Msgf("recv end")
	var err error
	var handler MsgHandler
	for {
		//		utils.Log.Debug().Msgf("engine client 111111111111")
		var packet *Packet
		// if this.isOldVersion {
		// 	packet, err = RecvPackage_old(this.conn)
		// } else {
		packet, err = RecvPackage(this.conn)
		// }
		if err != nil {
			//			utils.Log.Debug().Msgf("engine client 222222222222")
			// utils.Log.Warn().Msgf("网络错误 Client %s", err.Error())
			if err.Error() == io.EOF.Error() {
			} else if strings.Contains(err.Error(), "use of closed network connection") {
			} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
			} else {
				utils.Log.Warn().Msgf("network error Client:%s", err.Error())
			}
			break
		} else {
			if packet.MsgID == MSGID_heartbeat {
				// utils.Log.Info().Msgf("收到心跳clientConn")
				select {
				case this.heartbeat <- false:
				default:
				}
				//心跳
				continue
			}

			// 如果接收到数据，则发送接收数据信号，进而更新心跳时间
			// 这样可以防止接收数据比较多时，心跳的回应没有及时到达，进而关闭连接的情况
			select {
			case this.getDataSign <- true:
			default:
			}

			packet.Session = this
			// utils.Log.Debug().Msgf("conn recv: %d %s %d\n%s", packet.MsgID, this.conn.RemoteAddr(), len(packet.Data)+len(packet.Dataplus)+16, hex.EncodeToString(append(packet.Data, packet.Dataplus...)))
			//utils.Log.Debug().Msgf("client conn recv: %d %s <- %s %d", packet.MsgID, this.conn.LocalAddr(), this.conn.RemoteAddr(), len(packet.Data)+len(packet.Dataplus)+16)

			handler = this.engine.router.GetHandler(this.areaName, packet.MsgID)
			if handler == nil {
				utils.Log.Warn().Msgf("client The message is not registered, message number:%d", packet.MsgID)
			} else {
				//这里决定了消息是否异步处理
				go this.handlerProcess(handler, packet)
			}

		}
	}

	// close(this.outChan)
	this.outChanCloseLock.Lock()
	this.outChanIsClose = true
	close(this.outChan)
	this.outChanCloseLock.Unlock()
	this.Close()
	if this.isPowerful {
		go this.reConnect()
	}
}

func (this *Client) loopSend() {
	// timer := time.NewTimer(heartBeat_timeout)

	defer this.conn.Close()
	// defer timer.Stop()

	var isClose bool
	mu := new(sync.Mutex)
	c := this.sendQueue.GetQueueChan()
	// 开启多个协程，同时从sendQueue中读取消息
	var wg sync.WaitGroup
	for i := 0; i < MaxSendQueueDealCnt; i++ {
		wg.Add(1)
		go func(mu *sync.Mutex) {
			defer wg.Done()

			var sp *SendPacket
			for {
				sp, isClose = <-c
				// timer.Reset(heartBeat_timeout)
				// select {
				// case sp, isClose = <-c:
				// case <-timer.C:
				// 	//发送心跳
				// 	// buff, err := MarshalPacket(MSGID_heartbeat, nil, nil)
				// 	// _, err = this.conn.Write(*buff)
				// 	// if err != nil {
				// 	// 	if strings.Contains(err.Error(), "use of closed network connection") {
				// 	// 	} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
				// 	// 	} else if strings.Contains(err.Error(), "An established connection was aborted by the software in your host machine") {
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
				mu.Lock()
				// start := time.Now()
				_, err := this.conn.Write(*sp.bs)
				mu.Unlock()
				// utils.Log.Info().Msgf("write 时间 %s", time.Now().Sub(start))
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
					//utils.Log.Debug().Msgf("client conn send: %s -> %s %d", this.conn.LocalAddr(), this.conn.RemoteAddr(), len(*sp.bs))
					// utils.Log.Info().Msgf("clent send %s", hex.EncodeToString(*buff))
				}
			}
		}(mu)
	}
	wg.Wait()
}

func (this *Client) hold() {
	defer this.Close()
	// addr := AddressNet([]byte(this.name))
	// name := addr.B58String()
	timer := time.NewTimer(heartBeat_timeout)
	haveHeart := false
	for {
		// utils.Log.Info().Msgf("发送心跳 11111:%s", name)
		// 心跳间隔
		timer.Reset(heartBeat_interval)
		<-timer.C

		// 查看是否有接收数据的信号，有，则不需要发送心跳
		select {
		case <-this.getDataSign:
			// 接收到数据的信号，表明服务器还存在，进而避免心跳超时直接关闭连接的处理
			haveHeart = true
		default:
			haveHeart = false
		}
		if haveHeart {
			continue
		}

		timer.Reset(heartBeat_timeout)
		//发送心跳
		buff, err := MarshalPacket(MSGID_heartbeat, nil, nil)
		_, err = this.conn.Write(*buff)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
			} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
			} else if strings.Contains(err.Error(), "An established connection was aborted by the software in your host machine") {
			} else {
				utils.Log.Warn().Msgf("loopSend error:%s", err.Error())
			}
			timer.Stop()
			break
		}
		select {
		case <-this.heartbeat:
			// utils.Log.Info().Msgf("发送心跳 22222:%s", name)
			// time.Sleep(heartBeat_interval)
			timer.Stop()
		case <-this.getDataSign:
			// 接收到数据的信号，表明服务器还存在，进而避免心跳超时直接关闭连接的处理
			// time.Sleep(heartBeat_interval)
			timer.Stop()
		case <-timer.C:
			// utils.Log.Info().Msgf("发送心跳 33333:%s", name)
			return
		}
	}
	// utils.Log.Info().Msgf("发送心跳 44444:%s", name)
	// this.Close()
}

func (this *Client) handlerProcess(handler MsgHandler, msg *Packet) {
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
	//	utils.Log.Debug().Msgf("engine client 888888888888")
	handler(this.controller, *msg)
	//	utils.Log.Debug().Msgf("engine client 99999999999999")
	//消息处理后也要通过拦截器
	for i := itpsLen; i > 0; i-- {
		itps[i-1].Out(this.controller, *msg)
	}
}

/*
发送序列化后的数据
*/
func (this *Client) Send(msgID uint64, data, dataplus *[]byte, timeout time.Duration) error {
	var buff *[]byte
	var err error
	buff, err = MarshalPacket(msgID, data, dataplus)
	if err != nil {
		return err
	}
	return this.sendQueue.AddAndWaitTimeout(buff, timeout)
}

// 客户端关闭时,退出recv,send
func (this *Client) Close() {
	if !this.SetTryIsClose(true) {
		return
	}
	<-this.allowClose

	// utils.Log.Info().Msgf("Close session ClientConn")
	closeCallback, exist := this.engine.GetCloseCallback(this.areaName)
	if exist && closeCallback != nil {
		closeCallback(this)
	}
	// utils.Log.Info().Msgf("Close session ClientConn")
	this.engine.sessionStore.removeSession(this.areaName, this)
	// utils.Log.Info().Msgf("Close session ClientConn")
	err := this.conn.Close()
	if err != nil {
	}
	// utils.Log.Info().Msgf("Close session ClientConn")
	this.sendQueue.Destroy()
}

// 获取远程ip地址和端口
func (this *Client) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

// 获取节点机器Id
func (this *Client) GetMachineID() string {
	return this.machineID
}

func (this *Client) SetName(name string) {
	this.engine.sessionStore.renameSession(this.GetIndex(), name)
	this.name = name
}

/*
 * GetAreaName 获取区域名称
 *
 * @return	areaName	string	区域名称
 */
func (this *Client) GetAreaName() string {
	return this.areaName
}

/*
 * GetSetGodTime 获取设置超级代理的时间
 *
 * @return	setGodTime	int64	设置超级代理的时间
 */
func (this *Client) GetSetGodTime() int64 {
	return this.setGodTime
}

func NewClient(areaName []byte, name, ip string, port uint32) *Client {
	client := new(Client)
	client.name = name
	client.inPack = make(chan *Packet, 1000)
	client.areaName = utils.Bytes2string(areaName)
	// client.outData = make(chan *[]byte, 1000)
	client.Connect(ip, port)
	return client
}

/*
随机获取一个域名
*/
func GetRandomDomain() string {
	str := "abcdefghijklmnopqrstuvwxyz"
	// rand.Seed(int64(time.Now().Nanosecond()))
	result := ""
	r := int64(0)
	for i := 0; i < 8; i++ {
		r = GetRandNum(int64(25))
		result = result + str[r:r+1]
	}
	return result
}

/*
获得一个随机数(0 - n]，包含0，不包含n
*/
func GetRandNum(n int64) int64 {
	if n <= 0 {
		return 0
	}
	result, _ := crand.Int(crand.Reader, big.NewInt(int64(n)))
	return result.Int64()
}

func TimeFormatToNanosecondStr() string {
	return time.Now().Format("20060102150405999999999")
}

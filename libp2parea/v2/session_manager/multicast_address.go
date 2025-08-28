package session_manager

import (
	"context"
	"encoding/json"
	ma "github.com/multiformats/go-multiaddr"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
向局域网中广播自己的节点信息
同时接收局域网中其他节点的信息
*/
type MulticastAddressManager struct {
	sessionManager  *SessionManager         //
	nodeManager     *nodeStore.NodeManager  //
	msgChan         chan *AddressInfo       //
	addrInfoMapLock *sync.RWMutex           //
	addrInfoMap     map[string]*AddressTime //
	contextRoot     context.Context         //
	cancelRoot      context.CancelFunc      //
}

func NewMulticastAddressManager(nodeManager *nodeStore.NodeManager, ctx context.Context) *MulticastAddressManager {
	ctxRoot, cancelRoot := context.WithCancel(ctx)
	m := MulticastAddressManager{
		nodeManager:     nodeManager,
		msgChan:         make(chan *AddressInfo, 100),
		addrInfoMapLock: new(sync.RWMutex),
		addrInfoMap:     make(map[string]*AddressTime),
		contextRoot:     ctxRoot,
		cancelRoot:      cancelRoot,
	}
	go m.LoopCleanCache()
	go m.RecvAddress()
	go m.LoopSendAddress()
	return &m
}

/*
定时发送自己节点的地址
*/
func (this *MulticastAddressManager) LoopSendAddress() {
	const interval = time.Minute
	for {
		addrInfo := this.nodeManager.NodeSelf.GetMultiaddrLAN()
		if len(addrInfo) == 0 {
			ips, err := GetLocalPrivateIps()
			if err != nil {
				utils.Log.Error()
				time.Sleep(interval)
				continue
			}
			if len(ips) == 0 {
				time.Sleep(interval)
				continue
			}
			//utils.Log.Info().Uint16("端口", this.nodeManager.NodeSelf.Port).Send()
			mas := BuildAddrs(ips, this.nodeManager.NodeSelf.Port)
			for _, one := range mas {
				addrInfo, ERR := engine.CheckAddr(one)
				if ERR.CheckFail() {
					continue
				}
				this.nodeManager.NodeSelf.AddMultiaddrLAN(addrInfo)
			}
			addrInfo = this.nodeManager.NodeSelf.GetMultiaddrLAN()
			if len(addrInfo) == 0 {
				time.Sleep(interval)
				continue
			}
		}
		//utils.Log.Info().Int("局域网地址数量", len(addrInfo)).Send()
		addrsInfo := AddressInfo{
			Addrs: make([][]byte, 0, len(addrInfo)),
		}
		for _, one := range addrInfo {
			addrsInfo.Addrs = append(addrsInfo.Addrs, one.Multiaddr.Bytes())
		}
		//utils.Log.Info().Interface("局域网广播自己的地址", addrInfo).Send()
		err := SendInterfacesAll(&addrsInfo)
		if err != nil {
			utils.Log.Error().Err(err).Send()
		}
		time.Sleep(interval)
	}
}

/*
接收节点地址
*/
func (this *MulticastAddressManager) RecvAddress() {
	go RecvInterfacesAll(this.msgChan)
	var addrsInfo *AddressInfo
	for {
		select {
		case addrsInfo = <-this.msgChan:
		case <-this.contextRoot.Done():
			return
		}
		//验证并构造地址信息
		addrInfoList := make([]*engine.AddrInfo, 0)
		for _, one := range addrsInfo.Addrs {
			a, err := ma.NewMultiaddrBytes(one)
			if err != nil {
				continue
			}
			aInfoOne, ERR := engine.CheckAddr(a)
			if ERR.CheckFail() {
				continue
			}
			addrInfoList = append(addrInfoList, aInfoOne)
			//utils.Log.Info().Interface("接收到组播地址", *aInfoOne).Send()
		}
		//newAddrs := make([]*engine.AddrInfo, 0)
		this.addrInfoMapLock.Lock()
		//判断地址是否存在
		for _, one := range addrInfoList {

			//utils.Log.Info().Interface("判断地址是否存在", one.Multiaddr.String()).Send()
			_, ok := this.addrInfoMap[one.Multiaddr.String()]
			if ok {
				//utils.Log.Info().Interface("判断地址是否存在", one.Multiaddr.String()).Str("地址已经存在", "").Send()
				continue
			}
			at := AddressTime{
				Addr:       *one,
				CreateTime: time.Now(),
			}
			//添加地址
			this.addrInfoMap[one.Multiaddr.String()] = &at
			go this.sessionManager.connectNet(*one)
		}
		this.addrInfoMapLock.Unlock()
	}
}

/*
循环清理缓存中过期的地址
*/
func (this *MulticastAddressManager) LoopCleanCache() {
	timer := time.NewTimer(time.Second * 10)
	for {
		select {
		case <-timer.C:
		case <-this.contextRoot.Done():
			return
		}
		now := time.Now()
		this.addrInfoMapLock.Lock()
		for k, v := range this.addrInfoMap {
			//判断超时
			if now.Before(v.CreateTime.Add(config.MulticastAddress_cache_timetime)) {
				continue
			}
			delete(this.addrInfoMap, k)
		}
		this.addrInfoMapLock.Unlock()
	}
}

/*
删除
*/
func (this *MulticastAddressManager) DelAddrInfo(addrInfo *engine.AddrInfo) {
	if addrInfo == nil {
		return
	}
	this.addrInfoMapLock.Lock()
	defer this.addrInfoMapLock.Unlock()
	delete(this.addrInfoMap, addrInfo.Multiaddr.String())
}

/*
设置状态为销毁
*/
func (this *MulticastAddressManager) SetDestroy() {
	this.cancelRoot()
}

func BuildAddrs(ips []net.IP, port uint16) []ma.Multiaddr {
	addrs := make([]ma.Multiaddr, 0, len(ips))
	for _, ipOne := range ips {
		a, err := ma.NewMultiaddr("/ip4/" + ipOne.String() + "/tcp/" + strconv.Itoa(int(port)) + "/ws/")
		if err != nil {
			continue
		}
		addrs = append(addrs, a)
	}
	return addrs
}

/*
获取本地局域网ip
*/
func GetLocalPrivateIps() ([]net.IP, error) {
	ips := make([]net.IP, 0)
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, ifOne := range ifs {
		//fmt.Println("网卡列表", ifOne.Index, ifOne.Name, ifOne.Flags.String())
		addrs, err := ifOne.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addrOne := range addrs {
			ip, _, err := net.ParseCIDR(addrOne.String())
			if err != nil {
				return nil, err
			}
			if ip.IsPrivate() {
				ips = append(ips, ip)
			}
		}
	}
	return ips, nil
}

/*
发送给所有网卡
*/
func SendInterfacesAll(addrsInfo *AddressInfo) error {
	ifs, err := net.Interfaces()
	if err != nil {
		return err
	}
	if len(addrsInfo.Addrs) == 0 {
		return nil
	}
	bs, err := json.Marshal(addrsInfo)
	if err != nil {
		return err
	}
	for _, ifOne := range ifs {
		//fmt.Println("网卡列表", ifOne.Index, ifOne.Name, ifOne.Flags.String())
		addrs, err := ifOne.Addrs()
		if err != nil {
			return err
		}
		var ip net.IP
		have := false
		for _, addrOne := range addrs {
			ip, _, err = net.ParseCIDR(addrOne.String())
			if err != nil {
				return err
			}
			if ip.IsPrivate() {
				//fmt.Println("  地址", addrOne.String())
				have = true
				break
			}
		}
		if have {
			err := SendInterfacesOne(ip, &bs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func SendInterfacesOne(localIp net.IP, content *[]byte) error {
	locaAddr, err := net.ResolveUDPAddr("udp", localIp.String()+":")
	if err != nil {
		return err
	}
	//log.Println("发送地址", locaAddr.String())
	//
	conn, err := net.ListenUDP("udp", locaAddr)
	if err != nil {
		log.Println("监听地址出错", err.Error())
		return err
	}
	//log.Println("本地监听地址", conn.LocalAddr())
	defer conn.Close()

	for _, intervalOne := range config.DefaultPortInterval {
		for interval := range intervalOne.Interval {
			remotAddr, err := net.ResolveUDPAddr("udp", config.MulticastAddress+":"+strconv.Itoa(intervalOne.PortBase+interval))
			if err != nil {
				//log.Println("组播地址格式不正确")
				return err
			}
			//log.Println("远端地址", remotAddr.String())
			//发送消息
			_, err = conn.WriteToUDP(*content, remotAddr)
			if err != nil {
				log.Println("发送msg到组播地址出错", err.Error())
				return err
			}
		}
	}
	//utils.Log.Info().Str("本节点发送了自己的信息", "").Send()
	return nil
}

/*
接收组播消息
*/
func RecvInterfacesAll(msgChan chan *AddressInfo) error {
	ifs, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, ifOne := range ifs {
		//fmt.Println("网卡列表", ifOne.Index, ifOne.Name, ifOne.Flags.String())
		addrs, err := ifOne.Addrs()
		if err != nil {
			return err
		}
		var ip net.IP
		have := false
		for _, addrOne := range addrs {
			ip, _, err = net.ParseCIDR(addrOne.String())
			if err != nil {
				return err
			}
			if ip.IsPrivate() {
				//fmt.Println("  地址", addrOne.String())
				have = true
				break
			}
		}
		if have {
			err := RecvInterfacesOne(&ifOne, ip, msgChan)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RecvInterfacesOne(iface *net.Interface, localIp net.IP, msgChan chan *AddressInfo) error {
	for _, intervalOne := range config.DefaultPortInterval {
		for interval := range intervalOne.Interval {
			port := intervalOne.PortBase + interval
			// 绑定IP和端口
			listener, err := net.ListenPacket("udp", localIp.String()+":"+strconv.Itoa(port))
			if err != nil {
				log.Println("错误1", err.Error())
				continue
			}
			// 关闭监听
			defer listener.Close()
			packConnector := ipv4.NewPacketConn(listener)
			// 加入组播
			if err := packConnector.JoinGroup(iface, &net.UDPAddr{IP: net.ParseIP(config.MulticastAddress)}); err != nil {
				log.Println("错误2", err.Error())
				continue
			}
			// 离开组播
			defer packConnector.LeaveGroup(nil, &net.UDPAddr{IP: net.ParseIP(config.MulticastAddress)})
			// 创建一个缓冲区(切片)用于接收数据
			buffer := make([]byte, 1024*2)
			for {
				// 接收数据
				n, _, _, err := packConnector.ReadFrom(buffer)
				if err != nil {
					//log.Println("Error reading:", err)
					utils.Log.Error().Err(err).Send()
					continue
				}
				// 打印接收到的数据和源地址
				data := buffer[:n]
				addrsInfo := new(AddressInfo)
				err = json.Unmarshal(data, addrsInfo)
				if err != nil {
					//log.Println("Error:", err)
					utils.Log.Error().Err(err).Send()
					continue
				}
				select {
				case msgChan <- addrsInfo:
					//utils.Log.Info().Str("接收到内网节点地址广播消息", "").Send()
				default:
				}
				//senderAddr := srcAddr.String()
				// fmt.Printf("Received %d bytes from %s: %s\n", n, senderAddr, hex.Dump(data))
				//log.Printf("Received %d bytes from [%s]:%s\n", n, senderAddr, string(data))
			}
			return nil
		}
	}
	return nil
}

/*
广播消息内容
*/
type AddressInfo struct {
	Addrs [][]byte //地址列表
}

/*
地址刷新时间
*/
type AddressTime struct {
	Addr       engine.AddrInfo //地址
	CreateTime time.Time       //创建时间
}

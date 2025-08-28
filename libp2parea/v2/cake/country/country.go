package country

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

// 这个地址是发送节点的地址（主节点地址），用于节点验证发送节点是否是指定的发送节点
const SendNodeAddr = "FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n"

var areaName string           // 同步数据时, 查找的节点地址
var CacheString string        // 缓存信息
var sendData string           // 发送数据
var addr nodeStore.AddressNet // 同步节点地址
var once sync.Once            // 只初始一次控制器
var areaJsonInfo AreaInfo     // 大区json信息

// 大区信息
type AreaCountry struct {
	Area        *libp2parea.Area // 节点信息
	IsInitNode  bool             // 是否是创世节点标识
	initialized bool             // 是否初始化过
}

// 大区json结构信息
type AreaInfo struct {
	Addresses map[string]string `json:"address"`
}

/*
 * 创建大区
 *
 * @param	area			*Area			所属区域
 * @param	initNode		bool			是不是创世节点标识
 * @param	masterNodeAddr	string			主节点地址, 用于节点验证发送节点是否是指定的发送节点, 如果为空, 则默认地址为FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n
 * @return	country			*AreaCountry	大区信息
 */
func NewAreaCountry(area *libp2parea.Area, initNode bool) *AreaCountry {
	return &AreaCountry{Area: area, IsInitNode: initNode}
}

func (ac *AreaCountry) Start() {
	if ac.initialized {
		return
	}
	ac.initialized = true

	// 初始化start
	ac.Init()

	// 等待节点自治成功
	ac.Area.WaitAutonomyFinish()

	//注册p2p通信模块
	ac.RegisteMSG()
}

/*
 * 启动虚拟节点
 */
func (ac *AreaCountry) startNode() {
	// 等待节点自治成功
	ac.Area.WaitAutonomyFinish()
}

func (ac *AreaCountry) Init() {
	//发送数据到其它节点存储
	area1 := sha256.Sum256([]byte("country"))
	area2 := area1[:]
	areaName = nodeStore.AddressNet(area2).B58String()
	addr = nodeStore.AddressFromB58String(areaName)
}

/*
 * 设置数据
 */
func (ac *AreaCountry) SetData(str string) error {
	if ac.IsInitNode {
		sendData = str
		CacheString = sendData
		//马上执行一次
		err := ac.Send()
		if err != nil {
			return errors.New("exec failed")
		}
		go ac.sendToNode()
		return nil
	}
	return errors.New("the node is not init node,can't set data")
}

// 感知到节点上下线或者到达超时时间，触发
func (ac *AreaCountry) sendToNode() {
	once.Do(func() {
		ticker := time.NewTicker(5 * time.Second)  //刚开始前半小时5s一次同步
		timer := time.NewTimer(1800 * time.Second) //半小时定时
		defer ticker.Stop()
		defer timer.Stop()
		for {
			select {
			case <-ticker.C:
				err := ac.Send()
				if err != nil {
					err := ac.Send()
					if err != nil {
						utils.Log.Info().Msgf("开始同步数据失败:%s", err.Error())
					}
				}
			case <-timer.C:
				ticker = time.NewTicker(10 * time.Second) //半小时后10s一次同步
			}
		}
	})
}

func (ac *AreaCountry) Send() error {
	//utils.Log.Info().Msgf("send data CacheString:", CacheString)
	param := []byte(sendData)
	magneticId, err := ac.Area.SearchNetAddrOneByOneProxy(&addr, nil, ac.Area.GodID, 5)
	if err != nil {
		utils.Log.Info().Msgf("SearchNetAddrOneByOneProxy target:%s err:%s", addr.B58String(), err)
		return err
	}
	//给其它节点发送数据
	for _, node := range magneticId {
		if node.B58String() == addr.B58String() {
			continue
		}
		_, _, _, err = ac.Area.SendP2pMsgProxy(MSGID_P2P_SEND_DATA, &node, nil, nil, &param)
		if err != nil {
			utils.Log.Info().Msgf("err:%s", err)
			return err
		}
	}
	return nil
}

func (ac *AreaCountry) GetNodeData() (string, error) {
	//当前节点的虚拟地址
	res := make(chan string, 5)
	magneticId, err := ac.Area.SearchNetAddrOneByOneProxy(&addr, nil, ac.Area.GodID, 5)
	if err != nil {
		utils.Log.Info().Msgf("SearchNetAddrOneByOneProxy target:%s err:%s", addr.B58String(), err)
		return "", err
	}
	for i, node := range magneticId {
		if node.B58String() == addr.B58String() {
			continue
		}
		go func(nodeAddr *nodeStore.AddressNet) {
			bs, _, _, err := ac.Area.SendP2pMsgProxyWaitRequest(MSGID_P2P_GET_DATA, nodeAddr, nil, nil, nil, time.Second*20)
			if err != nil {
				utils.Log.Info().Msgf("SendP2pMsgProxyWaitRequest err:%s, target:%s", err, nodeAddr.B58String())
				res <- ""
				return
			}
			if bs == nil {
				res <- ""
				return
			}
			var rst string
			json.Unmarshal(*bs, &rst)
			res <- rst
		}(&magneticId[i])
	}
	var s string
	for i := 0; i < 3; i++ {
		if re := <-res; re != "" {
			s = re
			break
		}
	}
	return s, nil
}

/*
 * 获取节点地址列表
 *
 * @return []nodeStore.Addresss
 */
func (ac *AreaCountry) GetAreaSaveNodeIds() ([]nodeStore.AddressNet, error) {
	//当前节点的虚拟地址
	res := make(chan string, 5)
	// 根据地址, 获取存储节点信息
	magneticId, err := ac.Area.SearchNetAddrOneByOneProxy(&addr, nil, ac.Area.GodID, 5)
	if err != nil {
		utils.Log.Info().Msgf("GetAreaSaveNodeIds SearchNetAddrOneByOneProxy target:%s err:%s", addr.B58String(), err)
		return nil, err
	}
	// 依次遍历, 开协程获取数据
	for i, node := range magneticId {
		if node.B58String() == addr.B58String() {
			continue
		}
		go func(nodeAddr *nodeStore.AddressNet) {
			bs, _, _, err := ac.Area.SendP2pMsgProxyWaitRequest(MSGID_P2P_GET_NODE_IDS, nodeAddr, nil, nil, nil, time.Second*20)
			if err != nil {
				utils.Log.Info().Msgf("GetAreaSaveNodeIds SendP2pMsgProxyWaitRequest err:%s, target:%s", err, nodeAddr.B58String())
				res <- ""
				return
			}
			if bs == nil {
				res <- ""
				return
			}
			rst := string(*bs)
			res <- rst
		}(&magneticId[i])
	}
	var s string
	for i := 0; i < 3; i++ {
		if re := <-res; re != "" {
			s = re
			break
		}
	}

	if s == "" {
		utils.Log.Error().Msgf("s 为空!!!!!!")
		return nil, nil
	}

	// 结果按|进行分割
	addrs := strings.Split(s, "|")
	// 依次组装结果
	resAddr := make([]nodeStore.AddressNet, 0)
	for i := range addrs {
		if addrs[i] == "" {
			continue
		}

		resAddr = append(resAddr, nodeStore.AddressFromB58String(addrs[i]))
	}
	return resAddr, nil
}

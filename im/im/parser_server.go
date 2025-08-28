package im

//
//import (
//	"bytes"
//	"github.com/oklog/ulid/v2"
//	"github.com/syndtr/goleveldb/leveldb"
//	"math/big"
//	"strconv"
//	"sync"
//	"time"
//	"web3_gui/config"
//	"web3_gui/im/db"
//	"web3_gui/im/im/v2/imdatachain"
//	"web3_gui/im/model"
//	"web3_gui/libp2parea/v2/engine"
//	"web3_gui/libp2parea/v2/node_store"
//	"web3_gui/utils"
//)
//
///*
//客户端数据链解析器
//*/
//type ProxyParser struct {
//	lock           *sync.RWMutex                        //锁
//	Index          *big.Int                             //必须连续的自增长ID
//	PreHash        []byte                               //
//	addrSelf       nodeStore.AddressNet                 //本节点地址
//	uploadChan     chan imdatachain.DataChainProxyItr   //
//	downloadChan   chan []imdatachain.DataChainProxyItr //
//	indexParseLock *sync.RWMutex                        //锁
//	IndexParse     *big.Int                             //本地解析到的数据链高度
//	parseSignal    chan imdatachain.DataChainProxyItr   //新消息需要解析的信号
//}
//
//func NewProxyParser(addrSelf nodeStore.AddressNet, dataChainChan chan imdatachain.DataChainProxyItr,
//	downloadChan chan []imdatachain.DataChainProxyItr) (*ProxyParser, utils.ERROR) {
//	cp := &ProxyParser{
//		lock:           new(sync.RWMutex),
//		addrSelf:       addrSelf,
//		uploadChan:     dataChainChan,
//		downloadChan:   downloadChan,
//		indexParseLock: new(sync.RWMutex),
//		parseSignal:    make(chan imdatachain.DataChainProxyItr, 100),
//	}
//	go cp.LoopDownloadChan()
//	ERR := cp.LoadDBShot()
//	go cp.loopParserClientDatachain()
//	return cp, ERR
//}
//
///*
//加载本地数据库中好友列表和聊天记录
//*/
//func (this *ProxyParser) LoadDBShot() utils.ERROR {
//	indexBs, endItr, ERR := db.ImProxyClient_LoadShot(&this.addrSelf)
//	if !ERR.CheckSuccess() {
//		utils.Log.Info().Msgf("加载快照高度错误:%s", ERR.String())
//		return ERR
//	}
//	//新节点，初始化数据链
//	this.Index = big.NewInt(0)
//	this.IndexParse = big.NewInt(0)
//	if endItr == nil {
//		this.InitDataChain()
//	} else {
//		this.PreHash = endItr.GetHash()
//		endIndex := endItr.GetIndex()
//		this.Index = &endIndex
//		this.IndexParse.SetBytes(indexBs)
//	}
//	return utils.NewErrorSuccess()
//}
//
///*
//初始化数据链
//*/
//func (this *ProxyParser) InitDataChain() utils.ERROR {
//	//dataChainInit := imdatachain.NewFirendListInit(Node.GetNetId(), nil)
//	//ERR := this.SaveDataChain(dataChainInit)
//	return ERR
//}

package im

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/oklog/ulid/v2"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"strconv"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/im/subscription"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
客户端数据链解析器
*/
type ClientParser struct {
	lock               *sync.RWMutex                        //锁
	PreHash            []byte                               //
	addrSelf           nodeStore.AddressNet                 //本节点地址
	uploadChan         chan imdatachain.DataChainProxyItr   //
	downloadChan       chan []imdatachain.DataChainProxyItr //从代理节点同步下来的已上链消息
	downloadNoLinkChan chan []imdatachain.DataChainProxyItr //从代理节点同步下来的未上链消息
	indexParseLock     *sync.RWMutex                        //锁
	IndexKnit          *big.Int                             //必须连续的自增长ID
	IndexParse         *big.Int                             //本地解析到的数据链高度
	parseSignal        chan imdatachain.DataChainProxyItr   //新消息需要解析的信号
	groupOperationChan chan imdatachain.DataChainProxyItr   //群操作管道
	cancel             context.CancelFunc                   //
	ctx                context.Context                      //
	syncSendTextSignal chan *SendTextTask                   //异步发送文本消息信号
	syncSendFileSignal chan bool                            //异步发送文件信号
}

func NewClientParser(addrSelf nodeStore.AddressNet, dataChainChan chan imdatachain.DataChainProxyItr, downloadChan,
	downloadNoLinkChan chan []imdatachain.DataChainProxyItr, groupOperationChan chan imdatachain.DataChainProxyItr) (*ClientParser, utils.ERROR) {
	cp := &ClientParser{
		lock:               new(sync.RWMutex),
		addrSelf:           addrSelf,
		uploadChan:         dataChainChan,
		downloadChan:       downloadChan,
		downloadNoLinkChan: downloadNoLinkChan,
		indexParseLock:     new(sync.RWMutex),
		parseSignal:        make(chan imdatachain.DataChainProxyItr, 100),
		groupOperationChan: groupOperationChan, //群操作管道
		syncSendTextSignal: make(chan *SendTextTask, 100),
		syncSendFileSignal: make(chan bool, 1),
	}
	cp.ctx, cp.cancel = context.WithCancel(context.Background())
	go cp.LoopDownloadChan()
	go cp.LoopDownloadChanNoLink()
	ERR := cp.LoadDBShot()
	go cp.loopParserClientDatachain()
	go cp.loopSendFile()
	return cp, ERR
}

/*
获取解析的高度
*/
func (this *ClientParser) GetParseIndex() *big.Int {
	this.indexParseLock.RLock()
	defer this.indexParseLock.RUnlock()
	return new(big.Int).SetBytes(this.IndexParse.Bytes())
}

/*
加载本地数据库中好友列表和聊天记录
*/
func (this *ClientParser) LoadDBShot() utils.ERROR {
	indexBs, endItr, ERR := db.ImProxyClient_LoadShot(*Node.GetNetId(), this.addrSelf)
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("加载快照高度错误:%s", ERR.String())
		return ERR
	}
	//新节点，初始化数据链
	this.IndexKnit = big.NewInt(0)
	this.IndexParse = big.NewInt(0)
	if endItr == nil {
		this.InitDataChain()
	} else {
		this.PreHash = endItr.GetHash()
		endIndex := endItr.GetIndex()
		this.IndexKnit = &endIndex
		this.IndexParse.SetBytes(indexBs)
	}
	return utils.NewErrorSuccess()
}

/*
初始化数据链
*/
func (this *ClientParser) InitDataChain() utils.ERROR {
	dataChainInit := imdatachain.NewFirendListInit(*Node.GetNetId(), nodeStore.AddressNet{})
	ERR := this.SaveDataChain(dataChainInit)
	return ERR
}

/*
保存一条记录，添加到数据链中，并上传到代理节点或直接发送给对方
*/
func (this *ClientParser) SaveDataChain(itr imdatachain.DataChainProxyItr) utils.ERROR {
	//dhPuk := Node.Keystore.GetDHKeyPair().KeyPair.GetPublicKey()
	//dhPrk := Node.Keystore.GetDHKeyPair().KeyPair.GetPrivateKey()
	//utils.Log.Info().Msgf("打印自己的公钥:%s 私钥:%s", hex.EncodeToString(dhPuk[:]), hex.EncodeToString(dhPrk[:]))
	//utils.Log.Info().Msgf("保存数据链:%d", itr.GetCmd())
	//utils.Log.Info().Msgf("保存数据链:%d %+v", itr.GetCmd(), itr)
	this.lock.Lock()
	defer this.lock.Unlock()
	//utils.Log.Info().Msgf("保存数据链:%d", itr.GetCmd())
	//查询这条消息是否重复
	if itr.GetID() != nil && len(itr.GetID()) != 0 {
		//utils.Log.Info().Msgf("保存数据链")
		proxyItr, ERR := db.ImProxyClient_FindDataChainByID(*Node.GetNetId(), this.addrSelf, itr.GetID())
		if ERR.CheckFail() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		//utils.Log.Info().Msgf("保存数据链")
		//已经存在
		if proxyItr != nil {
			return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_exist, "")
		}
	}
	//查询这条消息是否重复
	if itr.GetSendID() != nil && len(itr.GetSendID()) != 0 {
		//utils.Log.Info().Msgf("保存数据链")
		ok, ERR := db.ImProxyClient_FindDataChainBySendID(*Node.GetNetId(), this.addrSelf, itr.GetSendID())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		//utils.Log.Info().Msgf("保存数据链")
		//已经存在
		if ok {
			return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_exist, "")
		}
	}
	//utils.Log.Info().Msgf("保存数据链:%d %+v", itr.GetCmd(), itr)
	proxyBase := itr.GetBase()
	addrFrom := itr.GetAddrFrom()
	addrTo := itr.GetAddrTo()
	//utils.Log.Info().Msgf("发送者与接收者:%d %+v %+v", itr.GetCmd(), addrFrom, addrTo)

	//发送者是自己，则去数据库查询发送索引，索引为连续增长的整数值。
	if bytes.Equal(itr.GetAddrFrom().GetAddr(), this.addrSelf.GetAddr()) {
		if len(addrTo.GetAddr()) > 0 {
			index, ERR := db.ImProxyClient_FindSendIndex(*config.DBKEY_improxy_user_datachain_sendIndex_knit, *Node.GetNetId(), addrFrom, addrTo)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("错误:%s", ERR.String())
				return ERR
			}
			//utils.Log.Info().Msgf("查询到用户的sendindex为:%s %s %+v", addrFrom.B58String(), addrTo.B58String(), index.Bytes())
			index = new(big.Int).Add(index, big.NewInt(1))
			proxyBase.SendIndex = index
		}
	} else {
		//发送者不是自己，又没有sendIndex的，不合法
		if proxyBase.SendIndex == nil {
			return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_params_fail, "sendIndex is 0")
		}
		//自己接收的消息，判断发送者的sendIndex是否连续，不连续消息则有遗漏
		index, ERR := db.ImProxyClient_FindSendIndex(*config.DBKEY_improxy_user_datachain_sendIndex_knit, *Node.GetNetId(), addrFrom, addrTo)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		if index.Cmp(big.NewInt(0)) == 0 {
			//当对方sendIndex无记录时，对方可以从任意index开始
		} else {
			//有记录，则必须连续
			index = new(big.Int).Add(index, big.NewInt(1))
			if !bytes.Equal(index.Bytes(), proxyBase.SendIndex.Bytes()) {
				utils.Log.Info().Msgf("发送索引不连续:%+v %+v", index.Bytes(), proxyBase.SendIndex.Bytes())
				return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_sendIndex_discontinuity, "")
			}
		}
	}
	//utils.Log.Info().Msgf("保存数据链:%d %+v", itr.GetCmd(), itr)
	itr.SetStatus(config.IMPROXY_datachain_status_notSend)
	//utils.Log.Info().Msgf("保存数据链")
	newIndex := new(big.Int).Add(this.IndexKnit, big.NewInt(1))
	itr.SetIndex(*newIndex)
	itr.SetPreHash(this.PreHash)
	//utils.Log.Info().Msgf("保存数据链:%d %+v", itr.GetCmd(), itr)
	//对方发的消息
	//if itr.GetCmd() == config.IMPROXY_Command_server_msglog_add {
	//}
	//自己发出的消息，需要加密
	if itr.GetClientItr() != nil && bytes.Equal(itr.GetAddrFrom().GetAddr(), this.addrSelf.GetAddr()) {
		var dhPuk *keystore.Key
		var ERR utils.ERROR
		dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
		if ERR.CheckFail() {
			return ERR
		}
		dhPrk := dhKey.GetPrivateKey()
		//与好友的公钥生成协商密钥
		if len(addrTo.GetAddr()) > 0 {
			dhPuk, ERR = db.ImProxyClient_FindUserDhPuk(*Node.GetNetId(), addrTo)
			//friend, ERR := db.ImProxyClient_FindUserinfo(addrTo)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("错误:%s", ERR.String())
				return ERR
			}
		} else {
			//与自己的公钥生成协商密钥
			pukKey := dhKey.GetPublicKey()
			dhPuk = &pukKey
		}
		//utils.Log.Info().Msgf("加密用公私密钥:%+v %+v", dhPrk, dhPuk)
		//生成共享密钥sharekey
		sharekey, err := keystore.KeyExchange(keystore.NewDHPair(dhPrk, *dhPuk))
		if err != nil {
			utils.Log.Error().Msgf("错误:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		//开始加密
		ERR = itr.EncryptContent(sharekey[:])
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		//把自己的公钥发送给对方
		pukSelf := dhKey.GetPublicKey()
		pukInfo, ERR := config.BuildDhPukInfoV1(pukSelf[:])
		if ERR.CheckFail() {
			return ERR
		}
		itr.GetBase().DhPuk = pukInfo
	}
	//utils.Log.Info().Msgf("保存数据链:%d %+v", itr.GetCmd(), itr)
	itr.BuildHash()
	//utils.Log.Info().Msgf("保存数据链:%+v", itr)

	batch := new(leveldb.Batch)
	//把数据链保存到数据库
	ERR := db.ImProxyClient_SaveDataChainMore(*Node.GetNetId(), batch, itr)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	//保存发送者的sendIndex
	if proxyBase.SendIndex != nil && (len(addrTo.GetAddr()) > 0) {
		//utils.Log.Info().Msgf("保存发送者的sendindex:%s %s %+v", addrFrom.B58String(), addrTo.B58String(), proxyBase.SendIndex.Bytes())
		ERR = db.ImProxyClient_SaveSendIndex(*config.DBKEY_improxy_user_datachain_sendIndex_knit, *Node.GetNetId(),
			addrFrom, addrTo, proxyBase.SendIndex.Bytes(), batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
	}
	err := db.SaveBatch(batch)
	if err != nil {
		utils.Log.Error().Msgf("错误:%s", err.Error())
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("保存数据链")
	this.IndexKnit = newIndex
	this.PreHash = itr.GetHash()
	select {
	case this.parseSignal <- itr:
	default:
	}
	//utils.Log.Info().Msgf("保存数据链")
	utils.Log.Info().Msgf("保存并解析自己的数据链:%s %d", newIndex.String(), itr.GetCmd())
	//utils.Log.Info().Msgf("保存并解析自己的数据链:%s %d %+v", newIndex.String(), itr.GetCmd(), itr)
	//utils.Log.Info().Msgf("保存并解析自己的数据链:%s %d %s", newIndex.String(), itr.GetCmd(), hex.EncodeToString(itr.GetID()))
	return utils.NewErrorSuccess()
}

/*
发送文本消息
*/
func (this *ClientParser) SendText(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	//utils.Log.Info().Str("开始发送文本消息", "").Send()
	task := NewSendTextTask(proxyItr)
	//utils.Log.Info().Int("开始发送文本消息", 0).Send()
	this.syncSendTextSignal <- task
	//utils.Log.Info().Str("开始发送文本消息", "").Send()
	ERR := <-task.errChan
	//utils.Log.Info().Str("开始发送文本消息", "").Send()
	return ERR
}

/*
发送文件
*/
func (this *ClientParser) SendFile() {
	this.syncSendFileSignal <- false
}

/*
循环发送文件
*/
func (this *ClientParser) loopSendFile() {
	//utils.Log.Info().Str("循环发送消息", "").Send()
	var ERR utils.ERROR
	for {
		if ERR.CheckFail() {
			//utils.Log.Info().Str("循环发送消息 等待50秒", ERR.String()).Send()
			time.Sleep(time.Second * 50)
		}

		var sendFileInfo *imdatachain.SendFileInfo
		var sendTextTask *SendTextTask
		for {
			//utils.Log.Info().Str("循环发送消息", "").Send()
			sendTextTask = nil
			//优先检查是否已经退出
			select {
			case <-this.ctx.Done():
				return
			default:
			}
			//utils.Log.Info().Str("循环发送消息", "").Send()
			//优先发送文本消息
			select {
			case sendTextTask = <-this.syncSendTextSignal:
				utils.Log.Info().Str("循环发送消息 有文本消息", "").Send()
			default:
			}
			if sendTextTask == nil && sendFileInfo != nil {
				utils.Log.Info().Str("循环发送文件消息", "").Send()
				//开始发送文件
				var datachainFile *imdatachain.DatachainFile
				datachainFile, ERR = sendFileInfo.BuildDatachainFile(*Node.GetNetId())
				if ERR.CheckFail() {
					//如果是系统错误，默认为文件无法读取，删除文件发送任务
					if ERR.Code == config.ERROR_CODE_system_error_self {
						ERR = db.ImProxyClient_DelSendFileList(*Node.GetNetId(), sendFileInfo)
						if ERR.CheckFail() {
							utils.Log.Error().Str("ERR", ERR.String()).Send()
							break
						}
						sendFileInfo = nil
						continue
					}
					utils.Log.Error().Str("ERR", ERR.String()).Send()
					break
				}

				//发送
				ERR = this.SaveDataChain(datachainFile.GetProxyItr())
				if !ERR.CheckSuccess() {
					break
				} else {
					sendFileInfo.BlockIndex++
					if sendFileInfo.BlockTotal == sendFileInfo.BlockIndex {
						//传完后删除
						ERR = db.ImProxyClient_DelSendFileList(*Node.GetNetId(), sendFileInfo)
						if ERR.CheckFail() {
							utils.Log.Error().Str("ERR", ERR.String()).Send()
							break
						}
						sendFileInfo = nil
					} else {
						//跟新上传进度
						ERR = db.ImProxyClient_SaveSendFileList(*Node.GetNetId(), sendFileInfo)
						if ERR.CheckFail() {
							utils.Log.Error().Str("ERR", ERR.String()).Send()
							break
						}
					}
				}
				continue
			}

			if sendTextTask == nil && sendFileInfo == nil {
				//utils.Log.Info().Hex("循环发送消息", this.GroupID).Send()
				select {
				case <-this.syncSendFileSignal:
					//utils.Log.Info().Str("循环发送消息", "").Send()
					//查询待发送的离线文件
					sendFileInfo, ERR = db.ImProxyClient_FindSendFileList(*Node.GetNetId(), *Node.GetNetId())
					if ERR.CheckFail() {
						utils.Log.Error().Str("ERR", ERR.String()).Send()
						break
					}
				case sendTextTask = <-this.syncSendTextSignal:
					utils.Log.Info().Str("循环发送消息 有文本消息", "").Send()
				case <-this.ctx.Done():
					//utils.Log.Info().Str("循环发送消息", "").Send()
					return
				}
				//utils.Log.Info().Hex("循环发送消息", this.GroupID).Send()
			}

			if sendTextTask != nil {
				//utils.Log.Info().Str("循环发送文本消息", "").Send()
				//设置sendIndex
				//sendTextTask.proxyItr.GetBase().SendIndex = new(big.Int).Add(sendIndex, big.NewInt(1))
				//发送
				ERR = this.SaveDataChain(sendTextTask.proxyItr)
				//ERR = SendGroupDataChain(this.addrKnitProxys, sendTextTask.proxyItr)
				if ERR.CheckFail() {
					utils.Log.Error().Msgf("循环发送消息 错误:%s", ERR.String())
					break
				}
				//utils.Log.Info().Str("循环发送消息", "").Send()
				sendTextTask.errChan <- ERR
				continue
			}
		}
	}
}

//
///*
//循环发送文件
//*/
//func (this *ClientParser) loopSendFileOld() {
//	utils.Log.Info().Str("循环发送消息", "").Send()
//	var ERR utils.ERROR
//	var sendFileInfo *imdatachain.SendFileInfo
//	for {
//		if ERR.CheckFail() {
//			utils.Log.Info().Str("循环发送消息 等待50秒", ERR.String()).Send()
//			time.Sleep(time.Second * 50)
//		}
//		//获取自己的发送index
//		//sendIndex := this.GetSendIndex()
//		utils.Log.Info().Str("循环发送消息", "").Send()
//		//优先检查是否已经退出
//		select {
//		case <-this.ctx.Done():
//			return
//		default:
//		}
//		utils.Log.Info().Str("循环发送消息", "").Send()
//
//		if sendFileInfo == nil {
//			select {
//			case <-this.syncSendFileSignal:
//				utils.Log.Info().Str("循环发送消息", "").Send()
//				//查询待发送的离线文件
//				sendFileInfo, ERR = db.ImProxyClient_FindSendFileList(Node.GetNetId(), Node.GetNetId())
//				if ERR.CheckFail() {
//					utils.Log.Error().Str("ERR", ERR.String()).Send()
//					continue
//				}
//			case <-this.ctx.Done():
//				utils.Log.Info().Str("循环发送消息", "").Send()
//				return
//			}
//		}
//
//		utils.Log.Info().Msgf("文件拆分个数:%+v", sendFileInfo)
//		//开始发送文件
//		var datachainFile *imdatachain.DatachainFile
//		datachainFile, ERR = sendFileInfo.BuildDatachainFile(Node.GetNetId())
//		if ERR.CheckFail() {
//			utils.Log.Error().Str("ERR", ERR.String()).Send()
//			continue
//		}
//
//		ERR = this.SaveDataChain(datachainFile.GetProxyItr())
//		if !ERR.CheckSuccess() {
//			utils.Log.Error().Str("ERR", ERR.String()).Send()
//			continue
//		} else {
//			sendFileInfo.BlockIndex++
//			if sendFileInfo.BlockTotal == sendFileInfo.BlockIndex {
//				//传完后删除
//				ERR = db.ImProxyClient_DelSendFileList(Node.GetNetId(), sendFileInfo)
//				if ERR.CheckFail() {
//					utils.Log.Error().Str("ERR", ERR.String()).Send()
//					continue
//				}
//				sendFileInfo = nil
//			} else {
//				//跟新上传进度
//				ERR = db.ImProxyClient_SaveSendFileList(Node.GetNetId(), sendFileInfo)
//				if ERR.CheckFail() {
//					utils.Log.Error().Str("ERR", ERR.String()).Send()
//					continue
//				}
//			}
//		}
//	}
//}

/*
循环读取下载通道中的新消息
*/
func (this *ClientParser) LoopDownloadChan() {
	for itrs := range this.downloadChan {
		if itrs == nil || len(itrs) == 0 {
			continue
		}
		for _, one := range itrs {
			//utils.Log.Info().Msgf("保存到这里")
			//有index则是多端同步下来的链上消息
			ERR := db.ImProxyClient_SaveDataChainMore(*Node.GetNetId(), nil, one)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("保存错误:%s", ERR.String())
			}
		}
	}
}

/*
循环读取下载通道中的离线消息
*/
func (this *ClientParser) LoopDownloadChanNoLink() {
	for itrs := range this.downloadNoLinkChan {
		if itrs == nil || len(itrs) == 0 {
			continue
		}
		for _, one := range itrs {
			//utils.Log.Info().Msgf("新消息:%+v", one)
			//index := one.GetIndex()
			//utils.Log.Info().Msgf("保存到这里")

			newItr := imdatachain.NewDataChainProxyMsgLog(one)
			//utils.Log.Info().Msgf("转换后的消息:%+v", newItr)
			////解密消息
			//ERR = itr.DecryptContent(sharekey[:])
			ERR := DecryptContent(newItr)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("解密错误:%s", ERR.String())
				//return ERR
				continue
			}

			ERR = this.SaveDataChain(newItr)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("保存错误:%s", ERR.String())
			}
		}
	}
}

/*
循环等待新消息信号
新消息来了，开始解析消息内容
自己发送的消息，放入发送通道
*/
func (this *ClientParser) loopParserClientDatachain() {
	for proxyItr := range this.parseSignal {
		//utils.Log.Info().Msgf("收到上传和发送消息")
		for {
			//utils.Log.Info().Msgf("收到上传和发送消息")
			this.lock.Lock()
			//utils.Log.Info().Msgf("收到上传和发送消息")
			lastIndex := this.IndexKnit
			this.lock.Unlock()
			//utils.Log.Info().Msgf("收到上传和发送消息")
			if this.IndexParse.Cmp(lastIndex) >= 0 {
				utils.Log.Info().Msgf("收到上传和发送消息")
				break
			}
			//utils.Log.Info().Msgf("收到上传和发送消息")
			//开始解析
			parseIndexBs := new(big.Int).Add(this.IndexParse, big.NewInt(1))
			//utils.Log.Info().Msgf("开始解析数据链:%+v", parseIndexBs.Bytes())
			itr, ERR := db.ImProxyClient_FindDataChainByIndex(*Node.GetNetId(), this.addrSelf, parseIndexBs.Bytes())
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询数据链出错:%s", ERR.String())
				break
			}
			//index := itr.GetIndex()
			//utils.Log.Info().Msgf("开始解析数据链:%s %d", index.String(), itr.GetCmd())
			ERR = this.parseClientDatachain(itr)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("解析数据链出错:%s", ERR.String())
				break
			}
			this.IndexParse = parseIndexBs
			//utils.Log.Info().Msgf("开始解析数据链")
		}
		//给上传消息提供一个信号
		select {
		case this.uploadChan <- proxyItr:
		default:
		}
	}
}

/*
解析一条数据链消息
*/
func (this *ClientParser) parseClientDatachain(itr imdatachain.DataChainProxyItr) utils.ERROR {
	//utils.Log.Info().Msgf("解析一条数据链消息:%d", itr.GetCmd())
	var mcVO *model.MessageContentVO
	var ERR utils.ERROR
	if itr.GetBase().Content != nil && len(itr.GetBase().Content) > 0 {
		var dhPuk *keystore.Key
		dhKey, ERR := Node.Keystore.GetDhAddrKeyPair(config.Wallet_keystore_default_pwd)
		if ERR.CheckFail() {
			return ERR
		}
		dhPrk := dhKey.GetPrivateKey()
		if bytes.Equal(itr.GetAddrFrom().GetAddr(), this.addrSelf.GetAddr()) {
			//自己的设置消息，没有目标地址，使用自己的公钥加密
			if len(itr.GetAddrTo().GetAddr()) == 0 {
				pukKey := dhKey.GetPublicKey()
				dhPuk = &pukKey
			} else {
				//自己发出的消息
				//查询对方的公钥
				dhPuk, ERR = db.ImProxyClient_FindUserDhPuk(*Node.GetNetId(), itr.GetAddrTo())
				if !ERR.CheckSuccess() {
					utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
					return ERR
				}
			}
		} else {
			//好友发送给自己的消息
			if itr.GetBase().DhPuk != nil && len(itr.GetBase().DhPuk) != 0 {
				//对方发来的消息，优先使用数据链中发来的公钥
				//utils.Log.Info().Msgf("解析公钥信息")
				//解析公钥信息
				dhPuk, ERR = config.ParseDhPukInfoV1(itr.GetBase().DhPuk)
				if !ERR.CheckSuccess() {
					utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
					return ERR
				}
			} else {
				//utils.Log.Info().Msgf("解析公钥信息")
				dhPuk, ERR = db.ImProxyClient_FindUserDhPuk(*Node.GetNetId(), itr.GetAddrFrom())
				if !ERR.CheckSuccess() {
					utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
					return ERR
				}
				//utils.Log.Info().Msgf("查询好友公钥信息:%s %s", itr.GetAddrFrom().B58String(), hex.EncodeToString(dhPuk[:]))
			}
		}

		//生成共享密钥sharekey
		sharekey, err := keystore.KeyExchange(keystore.NewDHPair(dhPrk, *dhPuk))
		if err != nil {
			ERR = utils.NewErrorSysSelf(err)
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		//协商密钥解密内容
		ERR = itr.DecryptContent(sharekey[:])
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		clientItr := itr.GetClientItr()
		clientItr.SetProxyItr(itr)
		ok := clientItr.CheckCmd()
		if !ok {
			return utils.NewErrorBus(config.ERROR_CODE_IM_datachain_cmd_fail, "")
		}
	}

	//utils.Log.Info().Msgf("解析一条数据链消息:%d", itr.GetCmd())
	ERR = utils.NewErrorSuccess()
	batch := new(leveldb.Batch)
	switch itr.GetCmd() {
	case config.IMPROXY_Command_server_init:
	case config.IMPROXY_Command_server_forward:
		//utils.Log.Info().Msgf("保存一个消息到未发送列表:%+v", itr.GetID())
		clientItr := itr.GetClientItr()
		//utils.Log.Info().Msgf("解析一条数据链消息:%d %d", itr.GetCmd(), clientItr.GetClientCmd())
		//需要发送的消息，先保存到未发送列表
		ERR = db.ImProxyClient_SaveDataChainSendFail(*Node.GetNetId(), itr, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
		mcVO, ERR = parseDataChainMessageType(clientItr, batch)
	case config.IMPROXY_Command_server_msglog_add:
		clientItr := itr.GetClientItr()
		mcVO, ERR = parseDataChainMessageType(clientItr, batch)
	case config.IMPROXY_Command_server_group_create:
		this.groupOperationChan <- itr
	case config.IMPROXY_Command_server_group_members:
		this.groupOperationChan <- itr
	case config.IMPROXY_Command_server_group_dissolve:
		this.groupOperationChan <- itr
	case config.IMPROXY_Command_server_setup:
		clientItr := itr.GetClientItr()
		switch clientItr.GetClientCmd() {
		case config.IMPROXY_Command_client_remarksname: //备注昵称
			mcVO, ERR = remarksName(clientItr, batch)
		default:
			utils.Log.Error().Msgf("未找到数据链中Client命令:%d", clientItr.GetClientCmd())
			ERR = utils.NewErrorBus(config.ERROR_CODE_IM_datachain_cmd_exist, "未找到数据链中Client命令:"+strconv.Itoa(clientItr.GetClientCmd()))
		}
	default:
		utils.Log.Error().Msgf("未找到数据链中Proxy命令:%d", itr.GetCmd())
		ERR = utils.NewErrorBus(config.ERROR_CODE_IM_datachain_cmd_exist, "未找到数据链中Proxy命令:"+strconv.Itoa(int(itr.GetCmd())))
	}
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	if mcVO != nil {
		//utils.Log.Info().Msgf("消息内容长度:%d", len(mcVO.Content))
	}
	proxyClientAddr := itr.GetProxyClientAddr()
	index := itr.GetIndex()
	ERR = db.ImProxyClient_SaveShot(*Node.GetNetId(), proxyClientAddr, index.Bytes(), batch)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	if mcVO != nil {
		//utils.Log.Info().Msgf("消息内容长度:%d", len(mcVO.Content))
	}
	//保存发送者的sendIndex
	proxyBase := itr.GetBase()
	if proxyBase.SendIndex != nil && len(itr.GetAddrTo().GetAddr()) > 0 {
		addrFrom := itr.GetAddrFrom()
		addrTo := itr.GetAddrTo()
		ERR = db.ImProxyClient_SaveSendIndex(*config.DBKEY_improxy_user_datachain_sendIndex_parse, *Node.GetNetId(),
			addrFrom, addrTo, proxyBase.SendIndex.Bytes(), batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return ERR
		}
	}
	if mcVO != nil {
		//utils.Log.Info().Msgf("消息内容长度:%d", len(mcVO.Content))
	}
	err := db.SaveBatch(batch)
	if err != nil {
		ERR = utils.NewErrorSysSelf(err)
		utils.Log.Error().Msgf("错误:%s", ERR.String())
		return ERR
	}
	utils.Log.Info().Msgf("解析一条数据链消息:%d", itr.GetCmd())
	utils.Log.Info().Msgf("打印消息:%+v", mcVO)
	if mcVO != nil {
		subscription.AddSubscriptionMsg(mcVO)
	}
	utils.Log.Info().Msgf("解析一条数据链消息:%s", index.String())
	ERR = utils.NewErrorSuccess()
	return ERR
}

/*
解析消息类型
*/
func parseDataChainMessageType(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	var mcVO *model.MessageContentVO
	var ERR utils.ERROR
	switch clientItr.GetClientCmd() {
	case config.IMPROXY_Command_client_addFriend: //添加好友
		mcVO, ERR = addFriend(clientItr, batch)
	case config.IMPROXY_Command_client_agreeFriend: //同意好友申请
		mcVO, ERR = agreeFriend(clientItr, batch)
	case config.IMPROXY_Command_client_del: //删除好友
		mcVO, ERR = delFriend(clientItr, batch)
	case config.IMPROXY_Command_client_sendText: //发送文本消息
		mcVO, ERR = sendText(clientItr, batch)
	case config.IMPROXY_Command_client_group_invitation: //邀请入群
		mcVO, ERR = invitationGroupMember(clientItr, batch)
	case config.IMPROXY_Command_client_group_accept: //接受好友邀请入群
		mcVO, ERR = acceptGroupMember(clientItr, batch)
	case config.IMPROXY_Command_client_group_addMember: //管理员同意添加用户入群
		mcVO, ERR = addGroupMember(clientItr, batch)
	case config.IMPROXY_Command_client_file: //发送文件
		//utils.Log.Info().Msgf("解析一条数据链消息:%d", itr.GetCmd())
		mcVO, ERR = sendFile(clientItr, batch)
	case config.IMPROXY_Command_client_voice: //发送语音消息
		mcVO, ERR = sendVoice(clientItr, batch)
	default:
		utils.Log.Error().Msgf("未找到数据链中Client命令:%d", clientItr.GetClientCmd())
		ERR = utils.NewErrorBus(config.ERROR_CODE_IM_datachain_cmd_exist, "未找到数据链中Client命令:"+strconv.Itoa(clientItr.GetClientCmd()))
	}
	if mcVO != nil {
		//utils.Log.Info().Msgf("消息内容长度:%d", len(mcVO.Content))
	}
	return mcVO, ERR
}

/*
申请添加好友
*/
func addFriend(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("申请添加好友")
	clientBase := clientItr.(*imdatachain.DatachainAddrFriend)
	if bytes.Equal(Node.GetNetId().GetAddr(), clientBase.AddrFrom.GetAddr()) {
		utils.Log.Info().Msgf("申请添加好友")
		userinfo, ERR := db.ImProxyClient_FindUserinfo(*Node.GetNetId(), clientBase.AddrTo)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("申请添加好友:%s", ERR.String())
			return nil, ERR
		}
		if userinfo == nil {
			userinfo = model.NewUserInfo(clientBase.AddrTo)
		}
		userinfo.Token = clientItr.GetProxyItr().GetID()
		userinfo.Time = clientItr.GetProxyItr().GetSendTime()
		//自己是发送者
		ERR = db.SaveUserListLocalApply(*Node.GetNetId(), userinfo, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("申请添加好友:%s", ERR.String())
			return nil, ERR
		}
		return nil, utils.NewErrorSuccess()
	} else {
		utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
		//自己是接收者
		//先判断好友列表里面有没有
		userList, ERR := db.FindUserListByAddr(*Node.GetNetId(), clientBase.AddrFrom)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询自己好友列表 错误:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
		//已经在好友列表中了
		if userList != nil {
			utils.Log.Info().Msgf("已经在好友列表中了")
			return nil, utils.NewErrorSuccess()
		}
		//判断近期是否有重复邀请
		userinfo, ERR := db.FindUserListApplyByAddr(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_addr,
			*Node.GetNetId(), clientBase.AddrFrom, nil)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		//判断短时间是否有重复邀请
		now := time.Now().Unix()
		if userinfo != nil && userinfo.Status != 1 && userinfo.Time < now && now-userinfo.Time < 60*60*24 {
			//有重复邀请
			return nil, utils.NewErrorSuccess()
		}

		//utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
		userinfo, ERR = db.ImProxyClient_FindUserinfo(*Node.GetNetId(), clientBase.AddrFrom)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
		if userinfo == nil {
			userinfo = model.NewUserInfo(clientBase.AddrFrom)
		}
		//解析公钥
		dhPuk, ERR := config.ParseDhPukInfoV1(clientBase.GetProxyItr().GetBase().DhPuk)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		userinfo.Time = clientBase.Time
		userinfo.GroupDHPuk = dhPuk[:]
		userinfo.Token = clientItr.GetProxyItr().GetID()
		userinfo.Time = clientItr.GetProxyItr().GetRecvTime()
		userinfo.Nickname = clientBase.Nickname
		utils.Log.Info().Msgf("保存token:%s", hex.EncodeToString(userinfo.Token))

		//添加到待确认列表
		ERR = db.SaveUserListRemoteApply(*Node.GetNetId(), userinfo, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("保存到用户列表 错误:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
		//给前端发送一个通知
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
		//AddSubscriptionMsg(&msgInfo)
		return &msgInfo, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("申请添加好友完成")
}

/*
同意好友申请
*/
func agreeFriend(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("同意好友申请")
	clientBase := clientItr.(*imdatachain.DatachainAgreeFriend)
	if bytes.Equal(Node.GetNetId().GetAddr(), clientBase.AddrFrom.GetAddr()) {
		utils.Log.Info().Msgf("同意好友申请")
		//查询列表中的好友
		user, ERR := db.FindUserListApplyByToken(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_index,
			*Node.GetNetId(), clientBase.TokenLocal)
		//user, ERR := db.FindUserListApplyByIndex(config.DBKEY_apply_remote_userlist, Node.GetNetId(), uint64(clientBase.CreateTime))
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
			return nil, ERR
		}
		//用户不在本列表中
		if user == nil {
			//return utils.NewErrorBus(config.ERROR_CODE_IM_user_not_exist, "")
			user = model.NewUserInfo(clientBase.AddrTo)
		}
		//修改申请
		user.Status = 1
		ERR = db.SaveUserListRemoteApply(*Node.GetNetId(), user, batch)
		//ERR = db.SaveUserList(config.DBKEY_apply_remote_userlist, Node.GetNetId(), user, batch)
		//ERR := db.UpdateUserList(*user, config.DBKEY_apply_remote_userlist)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("错误:%s", ERR.String())
			return nil, ERR
		}
		//保存到好友列表
		ERR = db.SaveUserList(*Node.GetNetId(), user, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("保存到列表 错误:%s", ERR.String())
			return nil, ERR
		}

		//把自己和对方的sendindex都重置
		//对方的
		//ERR = db.ImProxyClient_SaveSendIndex(config.DBKEY_improxy_user_datachain_sendIndex_parse, Node.GetNetId(),
		//	clientBase.AddrTo, clientBase.AddrFrom, nil, batch)
		//if !ERR.CheckSuccess() {
		//	return nil, ERR
		//}
		////自己的
		//ERR = db.ImProxyClient_SaveSendIndex(config.DBKEY_improxy_user_datachain_sendIndex_parse, Node.GetNetId(),
		//	clientBase.AddrFrom, clientBase.AddrTo, nil, batch)
		//if !ERR.CheckSuccess() {
		//	return nil, ERR
		//}

		//ERR = db.AddUserList(user, config.DBKEY_friend_userlist)
		//return nil, ERR
		//给前端发送一个通知
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
		//AddSubscriptionMsg(&msgInfo)
		return &msgInfo, utils.NewErrorSuccess()
	} else {
		utils.Log.Info().Msgf("同意好友申请")
		//自己是接收者
		//判断申请列表里面有没有这个令牌
		userInfo, ERR := db.FindUserListApplyByToken(*config.DBKEY_apply_local_userlist, *config.DBKEY_apply_local_userlist_index,
			*Node.GetNetId(), clientBase.TokenRemote)
		//userInfo, ERR := db.FindUserListByAddr(config.DBKEY_apply_local_userlist, Node.GetNetId(), clientBase.AddrFrom)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
			return nil, ERR
		}
		if userInfo == nil {
			//申请列表里面没有，则判断好友列表里面有没有
			userInfo, ERR = db.FindUserListByAddr(*Node.GetNetId(), clientBase.AddrFrom)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
				return nil, ERR
			}
			//好友列表中也没有
			if userInfo == nil {
				return nil, utils.NewErrorBus(config.ERROR_CODE_IM_invalid_Agree_Add_Friend, "")
			}
		}
		//解析公钥信息
		dhPuk, ERR := config.ParseDhPukInfoV1(clientBase.DhPukInfo)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
			return nil, ERR
		}
		userInfo.GroupDHPuk = dhPuk[:]
		ERR = db.SaveUserList(*Node.GetNetId(), userInfo, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("保存到列表 错误:%s", ERR.String())
			return nil, ERR
		}

		//把自己和对方的sendindex都重置
		//对方的
		//ERR = db.ImProxyClient_SaveSendIndex(config.DBKEY_improxy_user_datachain_sendIndex_parse, Node.GetNetId(),
		//	clientBase.AddrFrom, clientBase.AddrTo, nil, batch)
		//if !ERR.CheckSuccess() {
		//	return nil, ERR
		//}
		//自己的
		//ERR = db.ImProxyClient_SaveSendIndex(config.DBKEY_improxy_user_datachain_sendIndex_parse, Node.GetNetId(),
		//	clientBase.AddrTo, clientBase.AddrFrom, nil, batch)
		//if !ERR.CheckSuccess() {
		//	return nil, ERR
		//}

		//给前端发送一个通知
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
		//AddSubscriptionMsg(&msgInfo)
		return &msgInfo, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("同意好友申请完成")
	//return utils.NewErrorSuccess()
}

/*
删除好友
*/
func delFriend(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("删除好友")
	clientBase := clientItr.(*imdatachain.DatachainDelFriend)
	friendAddr := clientBase.AddrFrom
	if bytes.Equal(Node.GetNetId().GetAddr(), clientBase.AddrFrom.GetAddr()) {
		friendAddr = clientBase.AddrTo
	}
	utils.Log.Info().Msgf("删除好友")
	ERR := db.DelUserListByAddr(*Node.GetNetId(), friendAddr, batch)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("删除好友 错误:%s", ERR.String())
		return nil, ERR
	}
	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
	//给前端发送一个通知
	//AddSubscriptionMsg(&msgInfo)
	return &msgInfo, utils.NewErrorSuccess()
}

/*
发送文本消息
*/
func sendText(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	//utils.Log.Info().Msgf("解析并保存文本消息记录:%+v", clientItr)
	clientBase := clientItr.(*imdatachain.DataChainSendText)
	if bytes.Equal(Node.GetNetId().GetAddr(), clientBase.AddrFrom.GetAddr()) {
		//utils.Log.Info().Msgf("解析并保存文本消息记录")

		//id := ulid.Make()
		//utils.Log.Info().Msgf("给好友发送消息：%s", content)
		messageOne := &model.MessageContent{
			Type:       config.MSG_type_text,            //消息类型
			FromIsSelf: true,                            //是否自己发出的
			From:       clientBase.AddrFrom,             //发送者
			To:         clientBase.AddrTo,               //接收者
			Content:    clientBase.Content,              //消息内容
			Time:       time.Now().Unix(),               //时间
			SendID:     clientItr.GetProxyItr().GetID(), //
			QuoteID:    clientBase.QuoteID,              //
			State:      config.MSG_GUI_state_not_send,   //
			//PullAndPushID: pullAndPushID,                 //
		}
		//utils.Log.Info().Msgf("发送消息：%+v", messageOne)
		//保存未发送状态的消息
		messageOne, ERR := db.AddMessageHistoryV2(*Node.GetNetId(), messageOne, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("解析并保存文本消息记录")
		msgVO := messageOne.ConverVO()
		msgVO.Subscription = config.SUBSCRIPTION_type_msg
		msgVO.State = config.MSG_GUI_state_not_send
		//AddSubscriptionMsg(msgVO)
		return msgVO, utils.NewErrorSuccess()
	} else {
		//utils.Log.Info().Msgf("解析并保存文本消息记录")
		//自己是接收者
		//查看发送者是否在好友列表中
		userinfo, ERR := db.FindUserListByAddr(*Node.GetNetId(), clientBase.AddrFrom)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("查看发送者是否在好友列表中 错误:%s", ERR.String())
			return nil, ERR
		}
		//不在好友列表中，则添加到好友申请列表中
		if userinfo == nil {
			//判断近期是否有重复邀请
			userinfo, ERR := db.FindUserListApplyByAddr(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_addr,
				*Node.GetNetId(), clientBase.AddrFrom, nil)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
				return nil, ERR
			}
			//判断短时间是否有重复邀请
			now := time.Now().Unix()
			if userinfo != nil && userinfo.Status != 1 && userinfo.Time < now && now-userinfo.Time < 60*60*24 {
				//有重复邀请
				return nil, utils.NewErrorSuccess()
			}

			//utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
			userinfo, ERR = db.ImProxyClient_FindUserinfo(*Node.GetNetId(), clientBase.AddrFrom)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
				return nil, ERR
			}
			//utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
			if userinfo == nil {
				userinfo = model.NewUserInfo(clientBase.AddrFrom)
			}

			//解析公钥
			dhPuk, ERR := config.ParseDhPukInfoV1(clientBase.GetProxyItr().GetBase().DhPuk)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
				return nil, ERR
			}
			userinfo.GroupDHPuk = dhPuk[:]
			userinfo.Token = clientItr.GetProxyItr().GetID()
			userinfo.Time = clientItr.GetProxyItr().GetRecvTime()
			//userinfo.Nickname = clientBase.Nickname

			//添加到好友申请列表
			ERR = db.SaveUserListRemoteApply(*Node.GetNetId(), userinfo, batch)
			//ERR = db.SaveUserList(config.DBKEY_apply_remote_userlist, Node.GetNetId(), userinfo, batch)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("添加到好友申请列表 错误:%s", ERR.String())
				return nil, ERR
			}
			//给前端发送一个通知
			msgInfo := &model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
			return msgInfo, utils.NewErrorSuccess()
		}
		//保存聊天记录
		messageOne := &model.MessageContent{
			Type:       config.MSG_type_text,            //消息类型
			FromIsSelf: false,                           //是否自己发出的
			From:       clientBase.AddrFrom,             //发送者
			To:         clientBase.AddrTo,               //接收者
			Content:    clientBase.Content,              //消息内容
			Time:       time.Now().Unix(),               //时间
			SendID:     clientItr.GetProxyItr().GetID(), //消息唯一ID
			RecvID:     ulid.Make().Bytes(),             //
			QuoteID:    clientBase.QuoteID,              //
			State:      config.MSG_GUI_state_success,    //
		}
		messageOne, ERR = db.AddMessageHistoryV2(*Node.GetNetId(), messageOne, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("添加消息记录 错误:%s", ERR.String())
			return nil, ERR
		}
		//给前端发送一个通知
		msgInfo := messageOne.ConverVO()
		msgInfo.Subscription = config.SUBSCRIPTION_type_msg
		//AddSubscriptionMsg(msgInfo)
		//utils.Log.Info().Msgf("解析并保存文本消息记录完成")
		return msgInfo, utils.NewErrorSuccess()
	}
	//return utils.NewErrorSuccess()
}

/*
邀请好友入群
*/
func invitationGroupMember(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	//proxyItr := clientItr.GetProxyItr()
	//utils.Log.Info().Msgf("邀请好友入群:%+v", clientItr)
	//utils.Log.Info().Msgf("打印各个地址: self:%s pitrFrom:%s pitrTo:%s citrFrom:%s citrTo:%s", Node.GetNetId().B58String(),
	//	clientItr.GetProxyItr().GetAddrFrom().B58String(), clientItr.GetProxyItr().GetAddrTo().B58String(),
	//	clientItr.GetAddrFrom().B58String(), clientItr.GetAddrTo().B58String())
	groupInvitation := clientItr.(*imdatachain.DatachainGroupInvitation)
	if bytes.Equal(Node.GetNetId().GetAddr(), clientItr.GetAddrFrom().GetAddr()) {
		token := ulid.Make().Bytes()
		//utils.Log.Info().Msgf("保存邀请到列表")
		userinfo := &model.UserInfo{
			Addr:      clientItr.GetAddrTo(),
			IsGroup:   true,
			GroupId:   groupInvitation.GroupID,
			Proxy:     new(sync.Map),
			AddrAdmin: groupInvitation.AdminAddr,
			Time:      clientItr.GetProxyItr().GetSendTime(),
		}
		userinfo.Token = token
		//自己是发送者
		ERR := db.SaveUserListLocalApply(*Node.GetNetId(), userinfo, batch)
		//ERR := db.SaveUserListForGroup(config.DBKEY_apply_local_userlist, Node.GetNetId(), userinfo, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("申请添加好友:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("给前端发送一个通知")
		//给前端发送一个通知
		//msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
		//AddSubscriptionMsg(&msgInfo)
		return nil, ERR
	} else {
		//判断近期是否有重复邀请
		userinfo, ERR := db.FindUserListApplyByAddr(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_addr,
			*Node.GetNetId(), clientItr.GetAddrFrom(), groupInvitation.GroupID)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		//判断短时间是否有重复邀请
		now := time.Now().Unix()
		if userinfo != nil && userinfo.Status != 1 && userinfo.Time < now && now-userinfo.Time < 60*60*24 {
			//有重复邀请
			return nil, utils.NewErrorSuccess()
		}

		//utils.Log.Info().Msgf("保存邀请到列表")
		userinfo, ERR = db.ImProxyClient_FindUserinfo(*Node.GetNetId(), clientItr.GetAddrFrom())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		if userinfo == nil {
			userinfo = model.NewUserInfo(clientItr.GetAddrFrom())
		}
		userinfo.IsGroup = true
		userinfo.GroupId = groupInvitation.GroupID
		userinfo.AddrAdmin = groupInvitation.AdminAddr
		userinfo.Token = clientItr.GetProxyItr().GetID()
		userinfo.Time = clientItr.GetProxyItr().GetRecvTime()
		userinfo.Nickname = groupInvitation.Nickname
		userinfo.RemarksName = groupInvitation.GroupNickname

		//添加到待确认列表
		ERR = db.SaveUserListRemoteApply(*Node.GetNetId(), userinfo, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("保存到用户列表 错误:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("给前端发送一个通知")
		//给前端发送一个通知
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
		//AddSubscriptionMsg(&msgInfo)
		return &msgInfo, utils.NewErrorSuccess()
	}
	//return utils.NewErrorSuccess()
}

/*
接受好友邀请入群
*/
func acceptGroupMember(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	//utils.Log.Info().Msgf("接受好友邀请入群:%+v", clientItr)
	//proxyItr := clientItr.GetProxyItr()
	groupAccept := clientItr.(*imdatachain.DatachainGroupAccept)
	if bytes.Equal(Node.GetNetId().GetAddr(), clientItr.GetAddrFrom().GetAddr()) {
		//utils.Log.Info().Msgf("同意好友申请")
		//查询列表中的好友
		user, ERR := db.FindUserListApplyByToken(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_index,
			*Node.GetNetId(), groupAccept.Token)
		//users, ERR := db.FindUserListAll(config.DBKEY_apply_remote_userlist, Node.GetNetId())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询列表 错误:%s", ERR.String())
			return nil, ERR
		}
		if user == nil {
			user = model.NewUserInfo(clientItr.GetAddrTo())
		}
		user.Status = 1
		ERR = db.SaveUserListRemoteApply(*Node.GetNetId(), user, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("错误:%s", ERR.String())
			return nil, ERR
		}
		//给前端发送一个通知
		//msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
		//AddSubscriptionMsg(&msgInfo)
		return nil, ERR
	} else {
		//utils.Log.Info().Msgf("-----同意好友申请:%+v", groupAccept)
		//检查群管理员是否是自己
		user, ERR := db.ImProxyClient_FindGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(), groupAccept.GroupID, *Node.GetNetId())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		if user == nil || !user.Admin {
			//加入群聊，发给我，但自己不是管理员
			//utils.Log.Info().Msgf("自己不是管理员")
			return nil, utils.NewErrorSuccess()
		}

		//判断近期是否有重复邀请
		userinfo, ERR := db.FindUserListApplyByAddr(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_addr,
			*Node.GetNetId(), clientItr.GetAddrFrom(), groupAccept.GroupID)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		//判断短时间是否有重复邀请
		now := time.Now().Unix()
		if userinfo != nil && userinfo.Status != 1 && userinfo.Time < now && now-userinfo.Time < 60*60*24 {
			//有重复邀请
			return nil, utils.NewErrorSuccess()
		}

		//群管理员收到带有签名的群成员入群申请
		//utils.Log.Info().Msgf("好友入群申请")
		//判断申请列表里面有没有这个好友

		userinfo, ERR = db.ImProxyClient_FindUserinfo(*Node.GetNetId(), clientItr.GetAddrFrom())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		if userinfo == nil {
			userinfo = model.NewUserInfo(clientItr.GetAddrFrom())
		}
		userinfo.IsGroup = true
		userinfo.GroupId = groupAccept.GroupID
		userinfo.AddrAdmin = groupAccept.AddrTo
		userinfo.GroupAcceptTime = groupAccept.AcceptTime
		userinfo.GroupDHPuk = groupAccept.DHPuk
		userinfo.GroupSign = groupAccept.Sign
		userinfo.GroupSignPuk = groupAccept.SignPuk
		userinfo.Token = clientItr.GetProxyItr().GetID()
		userinfo.Time = groupAccept.GetProxyItr().GetRecvTime()
		userinfo.Nickname = groupAccept.Nickname
		userinfo.RemarksName = groupAccept.GroupNickname
		ERR = db.SaveUserListRemoteApply(*Node.GetNetId(), userinfo, batch)
		//ERR = db.SaveUserListForGroup(config.DBKEY_apply_remote_userlist, Node.GetNetId(), userinfo, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("保存到列表 错误:%s", ERR.String())
			return nil, ERR
		}
		//给前端发送一个通知
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
		//AddSubscriptionMsg(&msgInfo)
		return &msgInfo, utils.NewErrorSuccess()
	}
}

/*
管理员同意用户入群
*/
func addGroupMember(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	//utils.Log.Info().Msgf("接受好友邀请入群:%+v", clientItr)
	//proxyItr := clientItr.GetProxyItr()
	groupAddMember := clientItr.(*imdatachain.DatachainGroupAddMember)
	if bytes.Equal(Node.GetNetId().GetAddr(), clientItr.GetAddrFrom().GetAddr()) {
		//utils.Log.Info().Msgf("同意好友申请")
		//检查群管理员是否是自己
		user, ERR := db.ImProxyClient_FindGroupMember(*config.DBKEY_improxy_group_members_parser, *Node.GetNetId(),
			groupAddMember.GroupID, *Node.GetNetId())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		if user == nil || !user.Admin {
			//加入群聊，发给我，但自己不是管理员
			//utils.Log.Info().Msgf("自己不是管理员")
			return nil, utils.NewErrorSuccess()
		}
		//群管理员收到带有签名的群成员入群申请
		//utils.Log.Info().Msgf("好友入群申请")
		//判断申请列表里面有没有这个好友
		userinfo, ERR := db.FindUserListApplyByToken(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_index,
			*Node.GetNetId(), groupAddMember.Token)
		//userinfo, ERR := db.ImProxyClient_FindUserinfo(Node.GetNetId(), clientItr.GetAddrFrom())
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
			return nil, ERR
		}
		if userinfo == nil {
			//userinfo = model.NewUserInfo(clientItr.GetAddrFrom())
			return nil, utils.NewErrorSuccess()
		}
		//修改状态
		userinfo.Status = 1
		ERR = db.SaveUserListRemoteApply(*Node.GetNetId(), userinfo, batch)
		//ERR = db.SaveUserListForGroup(config.DBKEY_apply_remote_userlist, Node.GetNetId(), userinfo, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("保存到列表 错误:%s", ERR.String())
			return nil, ERR
		}
		//给前端发送一个通知
		msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
		//AddSubscriptionMsg(&msgInfo)
		return &msgInfo, utils.NewErrorSuccess()
	} else {
		//utils.Log.Info().Msgf("-----同意好友申请:%+v", groupAccept)
		return nil, utils.NewErrorSuccess()
	}
}

/*
发送文件
*/
func sendFile(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	dataChainFile := clientItr.(*imdatachain.DatachainFile)
	//utils.Log.Info().Msgf("解析并保存文件:%d", dataChainFile.BlockIndex)
	//utils.Log.Info().Msgf("解析并保存文件:%+v %+v", dataChainFile.Hash, dataChainFile.SendTime)
	//index := clientItr.GetProxyItr().GetIndex()
	addrRemote := clientItr.GetAddrFrom()
	fromIsSelf := false
	if bytes.Equal(Node.GetNetId().GetAddr(), dataChainFile.GetAddrFrom().GetAddr()) {
		fromIsSelf = true
		addrRemote = clientItr.GetAddrTo()
	}

	//utils.Log.Info().Msgf("创建群或者修改群信息")

	//是一个新文件，则保存新记录
	if dataChainFile.BlockIndex == 0 {
		//utils.Log.Info().Msgf("解析并保存文件")
		var msgContent *model.MessageContent
		//utils.Log.Info().Str("mimeType", dataChainFile.MimeType).Send()
		//是base64编码图片
		if len(dataChainFile.MimeType) >= 5 && dataChainFile.MimeType[:5] == "image" && dataChainFile.Size <= uint64(config.FILE_image_size_max) {
			//utils.Log.Info().Str("mimeType", dataChainFile.MimeType).Send()
			msgContent = model.NewMsgContentImgBase64(fromIsSelf, dataChainFile.GetAddrFrom(), dataChainFile.GetAddrTo(),
				time.Now().Unix(), dataChainFile.SendTime, dataChainFile.GetProxyItr().GetID(), dataChainFile.Name,
				dataChainFile.MimeType, dataChainFile.Size, dataChainFile.Hash, dataChainFile.BlockTotal,
				dataChainFile.BlockIndex, [][]byte{dataChainFile.GetProxyItr().GetID()})
		} else {
			//utils.Log.Info().Str("mimeType", dataChainFile.MimeType).Send()
			msgContent = model.NewMsgContentFile(fromIsSelf, dataChainFile.GetAddrFrom(), dataChainFile.GetAddrTo(),
				time.Now().Unix(), dataChainFile.SendTime, dataChainFile.GetProxyItr().GetID(), dataChainFile.Name,
				dataChainFile.MimeType, dataChainFile.Size, dataChainFile.Hash, dataChainFile.BlockTotal,
				dataChainFile.BlockIndex, [][]byte{dataChainFile.GetProxyItr().GetID()})
		}
		msgContent.State = config.MSG_GUI_state_not_send
		//utils.Log.Info().Msgf("发送消息：%+v", messageOne)

		//如果是自己发送的文件，进度显示网络发送的进度。保存传送进度到数据库
		if !msgContent.FromIsSelf {
			msgContent.TransProgress = int(float64(msgContent.FileBlockIndex+1) / float64(msgContent.FileBlockTotal) * 100)
		}

		//保存未发送状态的消息
		messageOne, ERR := db.AddMessageHistoryV2(*Node.GetNetId(), msgContent, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("解析并保存文件id:%+v %+v %+v", dataChainFile.GetProxyItr().GetID(), messageOne.FileContent, dataChainFile.Hash)
		//utils.Log.Info().Msgf("解析并保存文件")
		//判断图片文件是否已经传完
		//if messageOne.FileType == config.FILE_type_image_base64 && messageOne.FileBlockTotal == messageOne.FileBlockIndex+1 {
		//图片传完的，拼接起来
		ERR = JoinImageBase64(messageOne)
		if ERR.CheckFail() {
			return nil, ERR
		}
		//}
		//utils.Log.Info().Msgf("解析并保存文件")
		msgVO := messageOne.ConverVO()
		msgVO.Subscription = config.SUBSCRIPTION_type_msg
		msgVO.State = config.MSG_GUI_state_not_send
		return msgVO, utils.NewErrorSuccess()
	}
	//utils.Log.Info().Msgf("解析并保存文件")
	//是文件续传
	key := config.DBKEY_improxy_user_message_history_fileHash_recv
	if fromIsSelf {
		key = config.DBKEY_improxy_user_message_history_fileHash_send
	}
	//utils.Log.Info().Msgf("保存记录:%d %d %+v", dataChainFile.BlockIndex, dataChainFile.BlockTotal, dataChainFile.Hash)
	msgContentOld, ERR := db.FindMessageHistoryByFileHash(*key, *Node.GetNetId(), addrRemote, dataChainFile.Hash, dataChainFile.SendTime)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
		return nil, ERR
	}
	msgContentOld.FileContent = append(msgContentOld.FileContent, dataChainFile.GetProxyItr().GetID())
	msgContentOld.FileBlockIndex = dataChainFile.BlockIndex
	//保存传送进度到数据库
	if !msgContentOld.FromIsSelf {
		msgContentOld.TransProgress = int(float64(msgContentOld.FileBlockIndex+1) / float64(msgContentOld.FileBlockTotal) * 100)
	}
	ERR = db.UpdateMessageHistoryByFileHash(*Node.GetNetId(), msgContentOld, batch)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
		return nil, ERR
	}
	//判断图片文件是否已经传完
	//if msgContentOld.FileType == config.FILE_type_image_base64 && msgContentOld.FileBlockTotal == msgContentOld.FileBlockIndex+1 {
	//图片传完的，拼接起来
	ERR = JoinImageBase64(msgContentOld)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//}
	//utils.Log.Info().Msgf("解析并保存文件")
	msgInfo := msgContentOld.ConverVO()
	msgInfo.Subscription = config.SUBSCRIPTION_type_msg
	//如果是自己发送的文件，进度显示网络发送的进度
	if msgContentOld.FromIsSelf && msgContentOld.FileType == config.FILE_type_file {
		msgInfo = nil
	}
	//给前端发送一个通知
	//msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_msg}
	//AddSubscriptionMsg(&msgInfo)
	return msgInfo, utils.NewErrorSuccess()
}

/*
发送语音消息
*/
func sendVoice(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	//utils.Log.Info().Msgf("解析并保存语音消息记录:%+v", clientItr)
	clientBase := clientItr.(*imdatachain.DatachainVoice)

	messageOne := &model.MessageContent{
		Type:       config.MSG_type_voice, //消息类型
		FromIsSelf: true,                  //是否自己发出的
		From:       clientBase.AddrFrom,   //发送者
		To:         clientBase.AddrTo,     //接收者
		//Content:    clientBase.Content,              //消息内容
		Time:   time.Now().Unix(),               //时间
		SendID: clientItr.GetProxyItr().GetID(), //
		//QuoteID:    clientBase.QuoteID,              //
		State:         config.MSG_GUI_state_not_send, //
		PullAndPushID: uint64(clientBase.Second),     //
		FileMimeType:  clientBase.MimeType,
		FileName:      clientBase.Name,
		FileSendTime:  clientBase.Second,
	}
	if clientBase.BlockCoding != "" {
		messageOne.Content = []byte(clientBase.BlockCoding)
	}

	if bytes.Equal(Node.GetNetId().GetAddr(), clientBase.AddrFrom.GetAddr()) {
		//utils.Log.Info().Msgf("解析并保存文本消息记录")
		//id := ulid.Make()
		//utils.Log.Info().Msgf("给好友发送消息：%s", content)

		//utils.Log.Info().Msgf("发送消息：%+v", messageOne)
		//保存未发送状态的消息
		messageOne, ERR := db.AddMessageHistoryV2(*Node.GetNetId(), messageOne, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("解析并保存文本消息记录 错误:%s", ERR.String())
			return nil, ERR
		}
		//utils.Log.Info().Msgf("解析并保存文本消息记录")
		msgVO := messageOne.ConverVO()
		msgVO.Subscription = config.SUBSCRIPTION_type_msg
		msgVO.State = config.MSG_GUI_state_not_send
		//AddSubscriptionMsg(msgVO)
		return msgVO, utils.NewErrorSuccess()
	} else {
		//utils.Log.Info().Msgf("解析并保存文本消息记录")
		//自己是接收者
		//查看发送者是否在好友列表中
		userinfo, ERR := db.FindUserListByAddr(*Node.GetNetId(), clientBase.AddrFrom)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("查看发送者是否在好友列表中 错误:%s", ERR.String())
			return nil, ERR
		}
		//不在好友列表中，则添加到好友申请列表中
		if userinfo == nil {
			//判断近期是否有重复邀请
			userinfo, ERR := db.FindUserListApplyByAddr(*config.DBKEY_apply_remote_userlist, *config.DBKEY_apply_remote_userlist_addr,
				*Node.GetNetId(), clientBase.AddrFrom, nil)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
				return nil, ERR
			}
			//判断短时间是否有重复邀请
			now := time.Now().Unix()
			if userinfo != nil && userinfo.Status != 1 && userinfo.Time < now && now-userinfo.Time < 60*60*24 {
				//有重复邀请
				return nil, utils.NewErrorSuccess()
			}

			//utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
			userinfo, ERR = db.ImProxyClient_FindUserinfo(*Node.GetNetId(), clientBase.AddrFrom)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
				return nil, ERR
			}
			//utils.Log.Info().Msgf("有新好友申请:%s", clientBase.AddrFrom.B58String())
			if userinfo == nil {
				userinfo = model.NewUserInfo(clientBase.AddrFrom)
			}

			//解析公钥
			dhPuk, ERR := config.ParseDhPukInfoV1(clientBase.GetProxyItr().GetBase().DhPuk)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询用户信息 错误:%s", ERR.String())
				return nil, ERR
			}
			userinfo.GroupDHPuk = dhPuk[:]
			userinfo.Token = clientItr.GetProxyItr().GetID()
			userinfo.Time = clientItr.GetProxyItr().GetRecvTime()
			//userinfo.Nickname = clientBase.Nickname

			//添加到好友申请列表
			ERR = db.SaveUserListRemoteApply(*Node.GetNetId(), userinfo, batch)
			//ERR = db.SaveUserList(config.DBKEY_apply_remote_userlist, Node.GetNetId(), userinfo, batch)
			if !ERR.CheckSuccess() {
				utils.Log.Info().Msgf("添加到好友申请列表 错误:%s", ERR.String())
				return nil, ERR
			}
			//给前端发送一个通知
			msgInfo := &model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_addFriend}
			return msgInfo, utils.NewErrorSuccess()
		}
		messageOne.FromIsSelf = false
		messageOne.RecvID = ulid.Make().Bytes()
		messageOne.State = config.MSG_GUI_state_success
		messageOne, ERR = db.AddMessageHistoryV2(*Node.GetNetId(), messageOne, batch)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("添加消息记录 错误:%s", ERR.String())
			return nil, ERR
		}
		//给前端发送一个通知
		msgInfo := messageOne.ConverVO()
		msgInfo.Subscription = config.SUBSCRIPTION_type_msg
		//AddSubscriptionMsg(msgInfo)
		//utils.Log.Info().Msgf("解析并保存文本消息记录完成")
		return msgInfo, utils.NewErrorSuccess()
	}
	//return utils.NewErrorSuccess()
}

/*
修改昵称
*/
func remarksName(clientItr imdatachain.DataChainClientItr, batch *leveldb.Batch) (*model.MessageContentVO, utils.ERROR) {
	utils.Log.Info().Msgf("修改昵称")
	clientRemarksname := clientItr.(*imdatachain.ClientRemarksname)
	userinfo, ERR := db.FindUserListByAddr(*Node.GetNetId(), clientRemarksname.Addr)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改昵称 错误:%s", ERR.String())
		return nil, ERR
	}
	userinfo.RemarksName = clientRemarksname.Remarksname
	ERR = db.SaveUserList(*Node.GetNetId(), userinfo, batch)
	if !ERR.CheckSuccess() {
		utils.Log.Error().Msgf("修改昵称 错误:%s", ERR.String())
		return nil, ERR
	}
	//给前端发送一个通知
	msgInfo := model.MessageContentVO{Subscription: config.SUBSCRIPTION_type_agreeFriend}
	return &msgInfo, utils.NewErrorSuccess()

}

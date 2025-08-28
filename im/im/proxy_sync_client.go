package im

import (
	"context"
	"math/big"
	"sync"
	"time"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/im/im/imdatachain"
	"web3_gui/im/model"
	"web3_gui/im/subscription"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type SyncClient struct {
	lock               *sync.RWMutex                        //
	ServerAddr         nodeStore.AddressNet                 //代理节点地址
	ClientAddr         nodeStore.AddressNet                 //本节点地址
	remoteIndex        *big.Int                             //远端已经上传成功的index
	localIndex         *big.Int                             //本地最新index
	Status             int                                  //状态
	uploadCache        chan imdatachain.DataChainProxyItr   //上传消息缓冲池
	downloadChan       chan []imdatachain.DataChainProxyItr //从代理节点同步下来的已上链消息
	downloadNoLinkChan chan []imdatachain.DataChainProxyItr //从代理节点同步下来的未上链消息
	ticker             *time.Ticker                         //
	ctx                context.Context                      //
	cancel             context.CancelFunc                   //
	parseClient        *ClientParser                        //
}

func NewSyncClient(addrServer, addrClient nodeStore.AddressNet, downloadChan, downloadNoLinkChan chan []imdatachain.DataChainProxyItr,
	parseClient *ClientParser) *SyncClient {
	sc := SyncClient{
		lock:               new(sync.RWMutex),
		ServerAddr:         addrServer,
		ClientAddr:         addrClient,
		remoteIndex:        new(big.Int),
		localIndex:         new(big.Int),
		uploadCache:        make(chan imdatachain.DataChainProxyItr, 100),
		downloadChan:       downloadChan,
		downloadNoLinkChan: downloadNoLinkChan,
		ticker:             time.NewTicker(config.IMPROXY_client_download_interval),
		parseClient:        parseClient,
	}
	sc.ctx, sc.cancel = context.WithCancel(context.Background())
	go sc.DownloadDataChain()
	go sc.loopUploadDataChain()
	return &sc
}

/*
获取远端已经同步的index
*/
func (this *SyncClient) GetRemoteIndex() {
	count := 0
	bTimer := utils.NewBackoffTimerChan(config.SyncNewBackoffTimer...)
	for {
		//utils.Log.Info().Msgf("获取远端已经同步的index")
		interval := bTimer.Wait(this.ctx)
		if interval == 0 && count != 0 {
			//销毁了
			utils.Log.Info().Msgf("销毁了")
			break
		}
		count++
		index, ERR := GetProxyDataChainIndex(this.ServerAddr)
		if !ERR.CheckSuccess() {
			utils.Log.Info().Msgf("获取远端已经同步的index 错误:%s", ERR.String())
			continue
		}
		this.remoteIndex = index
		this.lock.Lock()
		this.Status++
		this.lock.Unlock()
		break
	}
	return
}

/*
同步上传数据链记录
*/
func (this *SyncClient) loopUploadDataChain() {
	//limit := uint64(100)

	//utils.Log.Info().Msgf("为什么是0:%+v %+v", limitBig, big.NewInt(100))
	for {
		select {
		case <-this.uploadCache:
		case <-this.ctx.Done():
			//销毁了
			utils.Log.Info().Msgf("销毁了")
			return
		}
		//utils.Log.Info().Msgf("同步上传数据链记录")
		//查询远端数据链高度
		this.GetRemoteIndex()
		//查询本地数据链索引高度
		itr, ERR := db.ImProxyClient_FindDataChainLast(*Node.GetNetId(), this.ClientAddr)
		//_, endIndexBs, ERR := db.ImProxyClient_LoadShot(&this.ServerAddr)
		if !ERR.CheckSuccess() {
			utils.Log.Error().Msgf("查询数据库错误:%s", ERR.String())
			continue
		}
		//utils.Log.Info().Msgf("同步上传数据链记录")
		//对比高度
		endIndex := itr.GetIndex()
		if endIndex.Cmp(this.remoteIndex) != 1 {
			continue
		}
		//utils.Log.Info().Msgf("同步上传数据链记录")
		//高度太低，需要上传，计算上传轮次
		interval := new(big.Int).Sub(&endIndex, this.remoteIndex)
		limitBig := big.NewInt(config.IMPROXY_sync_total_once)
		utils.Log.Info().Msgf("计算上传轮次:%+v %+v %+v", interval.Bytes(), limitBig.Bytes(), big.NewInt(100).Bytes())
		d, m := new(big.Int).DivMod(interval, limitBig, limitBig)
		startIndexBs := this.remoteIndex.Bytes()
		count := d.Uint64()
		if m.Uint64() != 0 {
			count++
		}
		for i := uint64(0); i < count; i++ {
			//utils.Log.Info().Msgf("同步上传数据链记录")
			itrs, ERR := db.ImProxyClient_FindDataChainRange(*Node.GetNetId(), this.ClientAddr, startIndexBs, config.IMPROXY_sync_total_once)
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("查询错误:%s", ERR.String())
				break
			}
			//utils.Log.Info().Msgf("同步上传数据链记录")
			//开始上传数据链
			ERR = UploadDataChain(this.ServerAddr, itrs)
			//不论成功，都要给一个结果
			for _, one := range itrs {
				//utils.Log.Info().Msgf("上传的数据链消息:%+v", one)
				engine.ResponseItrKey(config.FLOOD_key_addFriend, one.GetID(), ERR)
				engine.ResponseItrKey(config.FLOOD_key_agreeFriend, one.GetID(), ERR)
				//flood.ResponseItrKey(config.FLOOD_key_addFriend, one.GetID(), ERR)
				//flood.ResponseItrKey(config.FLOOD_key_agreeFriend, one.GetID(), ERR)

			}

			//修改状态，并推送给前端
			for _, one := range itrs {
				//修改状态并通知前端
				if one.GetCmd() == config.IMPROXY_Command_server_forward {
					//顺便解密，方便后面推送
					ERR := DecryptContent(one)
					if ERR.CheckFail() {
						utils.Log.Info().Msgf("解密失败:%s", ERR.String())
						continue
					}
					clientOne := one.GetClientItr()
					if clientOne.GetClientCmd() == config.IMPROXY_Command_client_sendText {
						//修改状态为失败
						clientBase := clientOne.(*imdatachain.DataChainSendText)
						messageOne := &model.MessageContent{
							Type:       config.MSG_type_text,         //消息类型
							FromIsSelf: true,                         //是否自己发出的
							From:       clientBase.AddrFrom,          //发送者
							To:         clientBase.AddrTo,            //接收者
							Content:    clientBase.Content,           //消息内容
							Time:       time.Now().Unix(),            //时间
							SendID:     one.GetID(),                  //
							QuoteID:    clientBase.QuoteID,           //
							State:      config.MSG_GUI_state_success, //
							//PullAndPushID: pullAndPushID,                 //
						}
						if !ERR.CheckSuccess() {
							messageOne.State = config.MSG_GUI_state_fail
						} else {
							//发送成功
							messageOne.State = config.MSG_GUI_state_success
						}
						//更新状态，这里的ERR不能影响外面的ERR，外面的ERR后面还需要继续使用
						ERR := db.UpdateSendMessageStateV2(one.GetAddrFrom(), one.GetAddrTo(), one.GetID(), messageOne.State)
						if !ERR.CheckSuccess() {
							utils.Log.Error().Msgf("修改消息状态 错误:%s", ERR.String())
							continue
						}
						//给前端一个通知
						msgVO := messageOne.ConverVO()
						msgVO.Subscription = config.SUBSCRIPTION_type_msg
						msgVO.State = messageOne.State
						subscription.AddSubscriptionMsg(msgVO)
					}
				}
			}

			//发送失败
			if !ERR.CheckSuccess() {
				utils.Log.Error().Msgf("发送消息 错误:%s", ERR.String())
				break
			}

			//发送成功
			for _, one := range itrs {
				//utils.Log.Info().Msgf("发送成功，删除未发送消息:%+v", one.GetID())
				//删除发送失败列表中的记录
				ERR = db.ImProxyClient_DelDataChainSendFail(*Node.GetNetId(), [][]byte{one.GetID()})
				if !ERR.CheckSuccess() {
					utils.Log.Error().Msgf("修改消息状态 错误:%s", ERR.String())
					continue
				}
			}

			//utils.Log.Info().Msgf("同步上传数据链记录")
			//
			index := itrs[len(itrs)-1].GetIndex()
			startIndexBs = index.Bytes()
			this.lock.Lock()
			this.remoteIndex = &index
			this.lock.Unlock()
		}
	}
}

/*
上传一条数据链记录
*/
func (this *SyncClient) UploadDataChain(itr imdatachain.DataChainProxyItr) {
	select {
	case this.uploadCache <- itr:
	default:
	}
	if itr == nil {
		return
	}
	index := itr.GetIndex()
	this.lock.Lock()
	if index.Cmp(this.localIndex) == 1 {
		this.localIndex = &index
	}
	this.lock.Unlock()
}

/*
下载数据链记录
*/
func (this *SyncClient) DownloadDataChain() {
	for {
		utils.Log.Info().Msgf("开始下载离线消息")
		//查询本地已经解析到的数据链高度
		index := this.parseClient.GetParseIndex()

		//this.lock.RLock()
		//index := this.localIndex
		//this.lock.RUnlock()
		index = new(big.Int).Add(index, big.NewInt(1))
		itrs, itrs2, ERR := DownloadDataChain(this.ServerAddr, index.Bytes())
		if ERR.CheckSuccess() {
			utils.Log.Info().Msgf("下载的离线消息:%+v %d %d", index.Bytes(), len(itrs), len(itrs2))
			this.downloadChan <- itrs
			this.downloadNoLinkChan <- itrs2
		} else {
			utils.Log.Error().Msgf("下载离线消息 错误:%s", ERR.String())
		}
		select {
		case <-this.ticker.C:
		case <-this.ctx.Done():
			this.ticker.Stop()
			//销毁了
			utils.Log.Info().Msgf("销毁了")
			break
		}
	}
}

/*
销毁
*/
func (this *SyncClient) Destroy() {
	this.cancel()
}

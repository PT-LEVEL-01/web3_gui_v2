package rpc_server

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/big"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/event"
	"web3_gui/chain/event/types"
	"web3_gui/chain/mining"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
)

type SubscriberService struct {
	chain *mining.Chain
	go_protos.UnimplementedSubscriberServer
	eventSub *event.ContractEventSub
	ctx      context.Context
}

func NewSubscriberService(ctx context.Context, chain *mining.Chain) *SubscriberService {
	return &SubscriberService{
		ctx:                           ctx,
		chain:                         chain,
		UnimplementedSubscriberServer: go_protos.UnimplementedSubscriberServer{},
		eventSub:                      event.NewSubscriber(chain.Balance.GetEventBus()),
	}
}
func (s *SubscriberService) EventSub(req *go_protos.SubscriberReq, server go_protos.Subscriber_EventSubServer) error {
	//解析参数
	payload := &go_protos.SubscribeContractEventPayload{}
	err := proto.Unmarshal(req.Payload, payload)
	if err != nil {
		return err
	}
	startBlock := payload.StartBlock
	endBlock := payload.EndBlock
	topic := payload.Topic
	contractAddr := payload.ContractAddress
	if err = s.checkSubscribeContractEventPayload(startBlock, endBlock, contractAddr); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	engine.Log.Info("收到合约事件订阅请求:startBlock-%d,endBlock-%d,contractAddr-%s,topic-%s", startBlock, endBlock, contractAddr, topic)

	return s.sendContractEvent(server, startBlock, endBlock, contractAddr, topic)

}
func (s *SubscriberService) checkSubscribeContractEventPayload(startBlock, endBlock int64, contractAddr string) error {
	if startBlock < -1 || endBlock < -1 || (endBlock != -1 && startBlock > endBlock) {
		return errors.New("无效的startBlock或者endBlock")
	}
	if contractAddr == "" {
		return errors.New("合约地址无效")
	}
	addr := crypto.AddressFromB58String(contractAddr)
	if !crypto.ValidAddr(config.AddrPre, addr) {
		return errors.New("合约地址无效")
	}
	return nil
}

func (s *SubscriberService) sendContractEvent(server go_protos.Subscriber_EventSubServer, startBlock, endBlock int64, contractAddr, topic string) error {
	//server.Send()

	var (
		alreadySendHistoryHeight int64
		err                      error
	)
	if startBlock == -1 && endBlock == 0 {
		return status.Error(codes.OK, "ok")
	}
	if (startBlock == -1 && endBlock == -1) || (startBlock == 0 || endBlock == 0) {
		//发送新的合约事件
		return s.sendNewContractEvent(server, startBlock, endBlock, contractAddr, topic, -1)
	}
	if startBlock != -1 {
		if alreadySendHistoryHeight, err = s.sendHistoryContractEvent(server, startBlock, endBlock, contractAddr, topic); err != nil {
			return err
		}
	}
	if startBlock == -1 {
		alreadySendHistoryHeight = -1
	}
	if alreadySendHistoryHeight == 0 {
		return status.Error(codes.OK, "ok")
	}

	return s.sendNewContractEvent(server, startBlock, endBlock, contractAddr, topic, alreadySendHistoryHeight)

}

// 一级操作
func (s *SubscriberService) sendHistoryContractEvent(server go_protos.Subscriber_EventSubServer, startBlock, endBlock int64, contractAddr, topic string) (int64, error) {

	if startBlock < 0 {
		startBlock = 0
	}
	_, err := s.doSendHistoryContractEvent(server, startBlock, endBlock, contractAddr, topic)
	if err != nil {
		engine.Log.Error("sendHistoryContractEvent失败:%s", err.Error())
		return -1, err
	}
	return 0, nil
}

// 一级操作
func (s *SubscriberService) sendNewContractEvent(server go_protos.Subscriber_EventSubServer, startBlock, endBlock int64, contractAddr, topic string, alreadySendHistoryHeight int64) error {
	var (
		err          error
		historyBlock int64
	)

	eventCh := make(chan types.ContractEvent)
	sub := s.eventSub.SubscribeContractEvent(eventCh)
	defer sub.Unsubscribe()
	for {
		select {
		case e := <-eventCh:
			engine.Log.Info("订阅者接收到事件", config.TimeNow())
			contractEventInfoList := e.ContractEventInfoList.ContractEvents
			blockHeight := contractEventInfoList[0].BlockHeight
			if endBlock > 0 && blockHeight > endBlock {
				return status.Error(codes.OK, "ok")
			}
			if alreadySendHistoryHeight != -1 && blockHeight > alreadySendHistoryHeight {
				//发送历史
				historyBlock, err = s.doSendHistoryContractEvent(server, alreadySendHistoryHeight+1, blockHeight, contractAddr, topic)
				if err != nil {
					engine.Log.Error("发送历史合约事件失败:%s", err.Error())
				}
				if endBlock > 0 && historyBlock >= endBlock {
					return status.Error(codes.OK, "OK")
				}
				alreadySendHistoryHeight = -1
				continue
			}
			if err = s.doSend(server, contractEventInfoList, contractAddr, topic); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			if endBlock > 0 && blockHeight >= endBlock {
				return status.Error(codes.OK, "ok")
			}
		case <-server.Context().Done():
			return nil
		case <-s.ctx.Done():
			return nil
		}
	}
	return nil
}

// 获取从开始到结束区块所有的历史事件记录，然后发送
func (s *SubscriberService) doSendHistoryContractEvent(server go_protos.Subscriber_EventSubServer, startBlock, endBlock int64, contractAddr, topic string) (int64, error) {
	//从开始区块到结束区块的日子获取
	i := startBlock
	for {
		select {
		case <-s.ctx.Done():
			return -1, status.Error(codes.Internal, "")
		default:
			//速率限制
			//if err:=s.getRateLimitToken();err!=nil{
			//
			//}
			if endBlock > 0 && i > endBlock {
				return i - 1, nil
			}
			if err := s.sendSubscribeContractEventByBlock(server, i, contractAddr, topic); err != nil {
				return -1, status.Error(codes.Internal, err.Error())
			}
			i++
		}
	}
}

// 发送某个区块的合约事件
func (s *SubscriberService) sendSubscribeContractEventByBlock(server go_protos.Subscriber_EventSubServer, block int64, contractAddr, topic string) error {
	//获取该区块的消息
	var err error
	key := append([]byte(config.DBKEY_BLOCK_EVENT), big.NewInt(block).Bytes()...)
	fvs, err := db.LevelDB.GetDB().HGetAll(key)
	if err != nil {
		engine.Log.Error("获取区块的合约事件日志失败%s", err.Error())
		return err
	}
	contractEvents := []*go_protos.ContractEventInfo{}
	for _, fv := range fvs {
		//解析到结构体
		eventList := &go_protos.ContractEventInfoList{}
		err = proto.Unmarshal(fv.Value, eventList)
		if err != nil {
			engine.Log.Error("从db中解析事件结构失败%s", err.Error())
			return err
		}
		contractEvents = append(contractEvents, eventList.ContractEvents...)
	}
	return s.doSend(server, contractEvents, contractAddr, topic)
}

// 执行发送
func (s *SubscriberService) doSend(server go_protos.Subscriber_EventSubServer, contractEvents []*go_protos.ContractEventInfo, contractAddr, topic string) error {
	var (
		err error
	)

	needSend := []*go_protos.ContractEventInfo{}
	for _, eventInfo := range contractEvents {
		if eventInfo.ContractAddress != contractAddr || (topic != "" && eventInfo.Topic != topic) {
			continue
		}
		needSend = append(needSend, eventInfo)
	}
	if len(needSend) == 0 {
		return nil
	}
	eventBytes, err := proto.Marshal(&go_protos.ContractEventInfoList{ContractEvents: needSend})
	if err != nil {
		return err
	}
	result := &go_protos.SubscriberResp{Data: eventBytes}
	if err = server.Send(result); err != nil {
		return errors.New("发送合约事件数据失败,错误信息：" + err.Error())
	}
	return nil
}

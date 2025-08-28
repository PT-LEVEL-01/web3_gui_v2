package country

import (
	"encoding/json"

	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/utils"
)

func (ac *AreaCountry) RegisteMSG() {
	if ac == nil || ac.Area == nil {
		return
	}

	ac.Area.Register_p2p(MSGID_P2P_GET_DATA, ac.GetDefineData)
	ac.Area.Register_p2p(MSGID_P2P_GET_DATA_BACK, ac.GetDefineDataBack)

	ac.Area.Register_p2p(MSGID_P2P_SEND_DATA, ac.SaveData)
	ac.Area.Register_p2p(MSGID_P2P_SEND_DATA_BACK, ac.SaveDataBack)

	ac.Area.Register_p2p(MSGID_P2P_GET_NODE_IDS, ac.GetAreaNodeIds)
	ac.Area.Register_p2p(MSGID_P2P_GET_NODE_IDS_RECV, ac.GetAreaNodeIdsRecv)
}

// 获取其他节点内存数据
func (ac *AreaCountry) GetDefineData(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if ac == nil || ac.Area == nil {
		return
	}

	// var msgHash string
	// if message != nil && message.Body != nil {
	// 	msgHash = hex.EncodeToString(message.Body.Hash)
	// }
	// utils.Log.Info().Msgf("GetDefineData msgHash:%s", msgHash)
	if CacheString == "" {
		ac.Area.SendP2pReplyMsg(message, MSGID_P2P_GET_DATA_BACK, nil)
		return
	}
	rst, _ := json.Marshal(CacheString)
	//utils.Log.Info().Msgf("GetDefineData result:%s", string(rst))
	ac.Area.SendP2pReplyMsg(message, MSGID_P2P_GET_DATA_BACK, &rst)
}

func (ac *AreaCountry) GetDefineDataBack(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// msgHash := hex.EncodeToString(message.Body.Hash)
	// utils.Log.Info().Msgf("GetDefineDataBack msgHash:%s", msgHash)
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

// 保存其它节点发来的数据
func (ac *AreaCountry) SaveData(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if ac == nil || ac.Area == nil {
		return
	}

	utils.Log.Info().Msgf("[%s] 获取发送节点的地址: %s", ac.Area.GetNetId().B58String(), message.Head.Sender.B58String())
	if message.Head.Sender.B58String() != SendNodeAddr {
		rst := []byte("send failed;send node auth failed")
		ac.Area.SendP2pReplyMsg(message, MSGID_P2P_SEND_DATA_BACK, &rst)
		return
	}
	if message.Body.Content == nil {
		CacheString = ""
		if len(areaJsonInfo.Addresses) > 0 {
			areaJsonInfo.Addresses = make(map[string]string)
		}
	} else {
		newCacheString := string(*message.Body.Content)
		// 值不同时再做更新
		if newCacheString != CacheString {
			CacheString = newCacheString
			// 解析大区信息
			err := json.Unmarshal([]byte(CacheString), &areaJsonInfo)
			if err != nil {
				utils.Log.Error().Msgf("json.Unmarshal err:%s", err)
			}
		}
	}
	//同步自身节点的数据到其它节点
	ac.Area.SendP2pReplyMsg(message, MSGID_P2P_SEND_DATA_BACK, nil)
}

func (ac *AreaCountry) SaveDataBack(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

// 获取大区保存节点信息
func (ac *AreaCountry) GetAreaNodeIds(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if ac == nil || ac.Area == nil {
		ac.Area.SendP2pReplyMsg(message, MSGID_P2P_GET_NODE_IDS_RECV, nil)
		return
	}

	// 组装节点地址列表
	var res string
	for _, v := range areaJsonInfo.Addresses {
		if v == "" {
			continue
		}

		if res != "" {
			// 节点地址之间用|隔离开
			res += "|"
		}
		res += v
	}

	//同步自身节点的数据到其它节点
	content := []byte(res)
	ac.Area.SendP2pReplyMsg(message, MSGID_P2P_GET_NODE_IDS_RECV, &content)
}

// 获取大区保存节点信息返回
func (ac *AreaCountry) GetAreaNodeIdsRecv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

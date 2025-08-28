package light

import (
	"bytes"
	"encoding/json"
	"errors"
	"web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain/rpc"
	utils2 "web3_gui/chain/utils"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
)

func RegisterDomainMsg() {
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_FIND_NAME, FindName)                    //查找域名
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1, CreateOfflineTxV1)   //查找域名
	Area.Register_neighbor(config.MSGID_P2P_NODE_HANDLEMULTIACCOUNTS, handleMultiAccounts) //查找域名
}

/*
查找域名
*/
func FindName(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_NAME_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_FIND_NAME_REV, pkg(rpc.SystemError, err.Error()))
}

/*
构建离线交易V1
*/
func CreateOfflineTxV1(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	// 获取密码
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(model.NoField, "pwd"))
		return
	}
	pwd := pwdItr.(string)

	// 获取nonce
	nonceItr, ok := rj.Get("nonce")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(model.NoField, "nonce"))
		return
	}
	nonce := uint64(nonceItr.(float64))
	if nonce < 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(5010, "nonce"))
		return
	}

	// 获取currentHeight
	currentHeightItr, ok := rj.Get("currentHeight")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(model.NoField, "currentHeight"))
		return
	}
	currentHeight := uint64(currentHeightItr.(float64))
	if currentHeight <= 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(5010, "currentHeight"))
		return
	}

	// 获取冻结高度
	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}

	// 获取domainType
	domainType := uint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = uint64(domainTypeItr.(float64))
	}

	// 获取keyStorePath 钱包路径
	keyStorePath := ""
	keyStorePathItr, ok := rj.Get("key_store_path")
	if ok {
		keyStorePath = keyStorePathItr.(string)
	}

	// 获取tag
	tagItr, ok := rj.Get("tag")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(model.NoField, "tag"))
		return
	}
	tag := tagItr.(string)

	// 获取jsonData
	jsonDataItr, ok := rj.Get("jsonData")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(model.NoField, "jsonData"))
		return
	}
	jsonDataItrs := jsonDataItr.(string)

	result := mining.BuildOfflineTx(keyStorePath, pwd, nonce, currentHeight, frozenHeight, domainType, domain, tag, jsonDataItrs)
	dataInfo := utils2.ParseDataInfo(result)
	if dataInfo.Code != 200 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(model.Nomarl, errors.New(dataInfo.Data.(string))))
		return
	}

	info := dataInfo.Data.(map[string]interface{})
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATEOFFLINETXV1_REV, pkg(model.Success, info))
}

// 查询多地址账户信息
func handleMultiAccounts(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_P2P_NODE_HANDLEMULTIACCOUNTS_rev, pkg(rpc.SystemError, err.Error()))
		return
	}
	addressesP, ok := rj.Get("addresses")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_P2P_NODE_HANDLEMULTIACCOUNTS_rev, pkg(model.NoField, "addresses"))
		return
	}

	bs, err := json.Marshal(addressesP)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_P2P_NODE_HANDLEMULTIACCOUNTS_rev, pkg(model.TypeWrong, "addresses"))
		return
	}
	addresses := make([]string, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	err = decoder.Decode(&addresses)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_P2P_NODE_HANDLEMULTIACCOUNTS_rev, pkg(model.TypeWrong, "addresses"))
		return
	}

	if len(addresses) <= 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_P2P_NODE_HANDLEMULTIACCOUNTS_rev, pkg(model.NoField, "addresses"))
		return
	}

	vos := make([]rpc.AccountVO, 0)
	for i, val := range addresses {
		addr := crypto.AddressFromB58String(val)
		var ba, fba, baLockup uint64
		ba, fba, baLockup = mining.GetBalanceForAddrSelf(addr)

		vo := rpc.AccountVO{
			Index:       i,
			AddrCoin:    val,
			Type:        mining.GetAddrState(addr),
			Value:       ba,       //可用余额
			ValueFrozen: fba,      //冻结余额
			ValueLockup: baLockup, //
		}
		vos = append(vos, vo)
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_P2P_NODE_HANDLEMULTIACCOUNTS_rev, pkg(model.Success, vos))
}

package libp2parea

import (
	"time"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

/*
 * 获取指定节点的机器id
 *
 * @return	machineID	string		目标的机器id
 */
func (this *Area) GetNodeMachineID(recvid *nodeStore.AddressNet) string {
	bs, ok, isSendSelf, err := this.MessageCenter.SendP2pMsgWaitRequest(config.MSGID_search_node, recvid, nil, time.Second*10)
	if err != nil {
		return ""
	}
	if isSendSelf || !ok {
		return ""
	}

	// bs := flood.WaitRequest(config.CLASS_get_MachineID, hex.EncodeToString(message.Body.Hash), 0)
	// bs, _ := flood.WaitRequest(config.CLASS_get_MachineID, utils.Bytes2string(message.Body.Hash), 0)
	if bs == nil {
		return ""
	}
	return utils.Bytes2string(*bs)
}

package im

import (
	"web3_gui/im/im/imdatachain"
	"web3_gui/utils"
)

type GroupSyncMinorManager struct {
}

func NewGroupSyncMinorManager() (*GroupSyncMinorManager, utils.ERROR) {
	gsmm := GroupSyncMinorManager{}
	ERR := gsmm.LoadDB()
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	return &gsmm, utils.NewErrorSuccess()
}

func (this *GroupSyncMinorManager) LoadDB() utils.ERROR {
	return utils.NewErrorSuccess()
}

func (this *GroupSyncMinorManager) ParseGroupDataChain(proxyItr imdatachain.DataChainProxyItr) utils.ERROR {
	return utils.NewErrorSuccess()
}

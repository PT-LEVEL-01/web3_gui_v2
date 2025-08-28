package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"strconv"
	"sync"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/utils"
)

var factoryMap *sync.Map = new(sync.Map) //key:uint64=;value:=;

type factoryFun func(bs []byte) (DataChainProxyItr, error)

func RegisterLog(id uint64, f factoryFun) {
	_, ok := factoryMap.LoadOrStore(id, f)
	if ok {
		utils.Log.Error().Msgf("重复注册的数据链:%d", id)
	}
}

func ParseDataChain(bs []byte) (DataChainProxyItr, utils.ERROR) {
	base := go_protos.ImDataChainProxyBase{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	itr, ok := factoryMap.Load(base.Base.Command)
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_CODE_Not_present, "ParseDataChain not fond command:"+strconv.Itoa(int(base.Base.Command)))
	}
	f := itr.(factoryFun)
	dataChainItr, err := f(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return dataChainItr, utils.NewErrorSuccess()
}

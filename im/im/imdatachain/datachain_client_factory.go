package imdatachain

import (
	"github.com/gogo/protobuf/proto"
	"strconv"
	"sync"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/utils"
)

var factoryClientMap *sync.Map = new(sync.Map) //key:uint64=;value:=;

type factoryClientFun func(bs []byte) (DataChainClientItr, error)

func RegisterCmdClient(id uint64, f factoryClientFun) {
	_, ok := factoryClientMap.LoadOrStore(id, f)
	if ok {
		utils.Log.Error().Msgf("重复注册的数据链:%d", id)
	}
}

func ParseDataChainClient(bs []byte) (DataChainClientItr, utils.ERROR) {
	base := go_protos.ImDataChainClientBase{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	itr, ok := factoryClientMap.Load(base.Base.Command)
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_CODE_Not_present, "ParseDataChainClient not fond command:"+strconv.Itoa(int(base.Base.Command)))
	}
	f := itr.(factoryClientFun)
	dataChainItr, err := f(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return dataChainItr, utils.NewErrorSuccess()
}

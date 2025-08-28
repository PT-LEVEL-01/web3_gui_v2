package object_beans

import (
	"github.com/gogo/protobuf/proto"
	"strconv"
	"sync"
	"web3_gui/config"
	"web3_gui/im/protos/go_protos"
	"web3_gui/utils"
)

var factoryMap *sync.Map = new(sync.Map) //key:uint64=;value:=;

type factoryFun func(bs []byte) (ObjectItr, error)

func RegisterObjectClass(class uint64, f factoryFun) {
	_, ok := factoryMap.LoadOrStore(class, f)
	if ok {
		utils.Log.Error().Msgf("重复注册的数据链:%d", class)
	}
}

func ParseObject(bs []byte) (ObjectItr, utils.ERROR) {
	base := go_protos.ImDataChainClientBase{}
	err := proto.Unmarshal(bs, &base)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	itr, ok := factoryMap.Load(base.Base.Command)
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_CODE_Not_present, "ParseDataChainClient not fond command:"+strconv.Itoa(int(base.Base.Command)))
	}
	f := itr.(factoryFun)
	dataChainItr, err := f(bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return dataChainItr, utils.NewErrorSuccess()
}

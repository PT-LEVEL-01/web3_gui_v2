package build

import "web3_gui/chain/mining"

const (
	Event_type_pay        = 1 //转账
	Event_type_deposit_in = 2 //见证人质押
)

type EventHandler func() mining.TxItr

type Event struct {
	StartHeight  uint64       //开始区块高度
	EndHeight    uint64       //结束区块高度
	EventType    int          //事件类型
	EventHandler EventHandler //
}

/*
构建指定高度的事件
@et        int       事件类型
@height    uint64    区块高度
*/
func CreateEventBlockHeight(et int, height uint64, handler EventHandler) Event {
	e := Event{
		StartHeight:  height,
		EndHeight:    height,
		EventType:    et,
		EventHandler: handler,
	}
	return e
}

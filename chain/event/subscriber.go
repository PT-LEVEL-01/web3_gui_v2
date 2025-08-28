package event

import (
	"web3_gui/chain/event/types"
	"web3_gui/chain/protos/go_protos"
)

type Subscriber interface {
	OnListen(*Event)
	OnExit()
}
type ContractEventSub struct {
	contractEventFeed Feed
}

func (c *ContractEventSub) OnListen(e *Event) {
	if contractEventInfoList, ok := e.Payload.(*go_protos.ContractEventInfoList); ok {
		go c.contractEventFeed.Send(types.ContractEvent{ContractEventInfoList: contractEventInfoList})
	}
}
func (c *ContractEventSub) OnExit() {

}
func NewSubscriber(bus EventBus) *ContractEventSub {
	sub := &ContractEventSub{}
	bus.Subscriber(ContractEventTopic, sub)
	return sub
}
func (c *ContractEventSub) SubscribeContractEvent(ch chan<- types.ContractEvent) Subscription {
	return c.contractEventFeed.Subscribe(ch)
}

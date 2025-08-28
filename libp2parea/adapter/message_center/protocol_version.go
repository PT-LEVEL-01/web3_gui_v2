package message_center

import "web3_gui/libp2parea/adapter/engine"

type MsgHandler func(c engine.Controller, msg engine.Packet, message *Message)

package engine

import (
	"strconv"
	"testing"
	"time"
	"web3_gui/utils"
)

func TestEngine(t *testing.T) {
	//example_engine_oneport()
	example_engine_client()
	//time.Sleep(time.Second * 50)
	//example_engine_tcp()
	//utils.Log.Info().Str("tcp服务器执行完成", "").Send()
	//example_engine_quic()
	//utils.Log.Info().Str("quic服务器执行完成", "").Send()
}

var port_server = uint16(8080)
var port_client = uint16(8081)

func TestEngineRegRpc(t *testing.T) {
	engine := NewEngine("test_server")
	engine.RegisterMsg(msgid_getinfo, GetInfo)
	ERR := engine.RegisterRPC(1, rpc_getinfo_method, RPC_getinfo, "")
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		panic(ERR.String())
	}
}

func TestEngineOnePort(t *testing.T) {
	engine := NewEngine("test_server")
	engine.RegisterMsg(msgid_getinfo, GetInfo)
	engine.RegisterRPC(1, rpc_getinfo_method, RPC_getinfo, "")
	ERR := engine.ListenOnePort(port_server, true)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
}

/*
客户端使用
*/
func example_engine_client() {
	engine := NewEngine("test_client")
	//ERR := engine.ListenOnePort(port_client, true)
	//if ERR.CheckFail() {
	//	utils.Log.Error().Str("ERR", ERR.String()).Send()
	//	return
	//}
	engine.RegisterMsg(msgid_getinfo_recv, GetInfo_recv)
	s, ERR := engine.DialTcpAddr("127.0.0.1", port_server)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	bs := []byte("hello server")
	_, ERR = s.Send(msgid_getinfo, &bs, time.Second*10)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}

	//开始rpc请求
	postResult, err := Post("127.0.0.1:"+strconv.Itoa(int(port_server)), "", "", rpc_getinfo_method, map[string]interface{}{"hello": "server"})
	if ERR.CheckFail() {
		utils.Log.Error().Err(err).Send()
		return
	}
	utils.Log.Info().Interface("postResult", postResult).Send()
}

var msgid_getinfo = uint64(11)
var msgid_getinfo_recv = uint64(12)

func example_engine_tcp() {
	port := uint16(8888)
	engine := NewEngine("test_server")
	engine.RegisterMsg(msgid_getinfo, GetInfo)
	ERR := engine.ListenTCP("", port, true)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	time.Sleep(2 * time.Second)
	engineClient := NewEngine("test_client")
	engineClient.RegisterMsg(msgid_getinfo_recv, GetInfo_recv)
	s, ERR := engineClient.DialTcpAddr("127.0.0.1", port)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	bs := []byte("hello server")
	_, ERR = s.Send(msgid_getinfo, &bs, time.Second)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	time.Sleep(2 * time.Second)
	engine.Destroy()
	engineClient.Destroy()
}

func example_engine_quic() {

	port := uint16(8888)
	engine := NewEngine("test_server")
	engine.RegisterMsg(msgid_getinfo, GetInfo)
	ERR := engine.ListenQuic("", port, true)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	time.Sleep(2 * time.Second)
	engineClient := NewEngine("test_client")
	engineClient.RegisterMsg(msgid_getinfo_recv, GetInfo_recv)
	s, ERR := engineClient.DialQuicAddr("127.0.0.1", port)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	bs := []byte("hello server")
	_, ERR = s.Send(msgid_getinfo, &bs, time.Second)
	if ERR.CheckFail() {
		utils.Log.Error().Str("error", ERR.String()).Send()
		return
	}
	time.Sleep(2 * time.Second)
	engine.Destroy()
	engineClient.Destroy()
}

func GetInfo(msg *Packet) {
	utils.Log.Info().Str("收到消息getinfo", string(msg.Data)).Send()
	bs := []byte("hello client")
	//回复消息
	_, ERR := msg.Session.Send(msgid_getinfo_recv, &bs, time.Second)
	if ERR.CheckFail() {
		utils.Log.Error().Str("error", ERR.String()).Send()
		return
	}
}
func GetInfo_recv(msg *Packet) {
	utils.Log.Info().Str("收到getinfo消息返回", string(msg.Data)).Send()
}

const rpc_getinfo_method = "getinfo"

func RPC_getinfo(params *map[string]interface{}) PostResult {
	//utils.Log.Info().Interface("params", params).Send()
	return PostResult{
		Code: utils.NewErrorSuccess().Code,
	}
}

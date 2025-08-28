package main

import (
	"fmt"
	"github.com/miekg/dns"
	"runtime"
	"sync"
	"time"
	"web3_gui/proxy/godivert"
	"web3_gui/proxy/godivert/header"
	"web3_gui/utils"
)

func main() {
	utils.LogBuildDefaultFile("log.txt")
	Example()
}

var names = new(sync.Map)   //保存需要解析的域名
var nameIps = new(sync.Map) //

func Example() {
	//设置要监听的域名
	names.Store("baidu.com", "")

	winDivert, err := godivert.NewWinDivertHandle("true")
	if err != nil {
		panic(err)
	}
	defer winDivert.Close()

	packetChan, err := winDivert.Packets()
	if err != nil {
		panic(err)
	}

	for range runtime.NumCPU() {
		go checkPacket(winDivert, packetChan)
	}
	utils.Log.Info().Str("start", "").Send()
	time.Sleep(10 * time.Minute)
}

//var counter = atomic.Uint64{}

func checkPacket(wd *godivert.WinDivertHandle, packetChan <-chan *godivert.Packet) {
	for packet := range packetChan {
		//newCounter := counter.Add(1)
		//utils.Log.Info().Uint64("start", newCounter).Send()
		countPacket(packet)
		//if packet.DstIP().Equal(cloudflareDNS4) {
		//	continue
		//}
		//_, err := wd.Send(packet)
		_, err := packet.Send(wd)
		if err != nil {
			fmt.Println("send error:", err.Error())
		}
		//fmt.Println("send size:", n)
		//utils.Log.Info().Uint64("end", newCounter).Send()
	}
	//utils.Log.Error().Str("ERROR", "结束了").Send()
}

func countPacket(packet *godivert.Packet) {
	//if packet.DstIP().Equal(net.ParseIP("114.114.114.114")) {
	//	utils.Log.Info().Hex("dns", packet.Raw).Send()
	//	utils.Log.Info().Hex("dns", packet.Raw[28:]).Send()
	//}

	//解析DNS请求。前28字节是IP头部和UDP头部信息，去掉后剩下DNS报文头部
	msg := dns.Msg{}
	if err := msg.Unpack(packet.Raw[28:]); err != nil {
		//utils.Log.Error().Err(err).Send()
		return
	}

	// 打印DNS请求的一些基本信息
	//fmt.Println("ID:", msg.Id)
	//fmt.Println("Recursion Desired:", msg.RecursionDesired)
	//fmt.Println("Question:", msg.Question)
	//utils.Log.Info().Uint16("ID", msg.Id).Bool("Recursion Desired", msg.RecursionDesired).Interface("Question", msg.Question).Send()

	// 如果请求包含问题查询部分，打印问题中的域名和类型
	for _, one := range msg.Question {
		//q := msg.Question[0]
		//fmt.Printf("Domain: %s, Type: %s\n", q.Name, dns.Type(q.Qtype).String())
		utils.Log.Info().Str("域名", one.Name).Str("类型", dns.Type(one.Qtype).String()).Send()

	}

	for _, one := range msg.Answer {
		utils.Log.Info().Str("answer", one.String()).Send()
	}

	switch packet.NextHeaderType() {
	case header.ICMPv4:
	case header.ICMPv6:
	case header.TCP:
		//utils.Log.Info().Hex("TCP协议", packet.Raw).Send()
	case header.UDP:
		//utils.Log.Info().Hex("UDP协议", packet.Raw).Send()
	default:
		utils.Log.Error().Hex("未知协议", packet.Raw).Send()
	}
}

/*
dns解析问答
*/
type DNSAnswer struct {
	QuestionID int
	DNSName    string
	IPAddress  []string
}

func NewDNSAnswer(id int) *DNSAnswer {
	return &DNSAnswer{
		QuestionID: id,
	}
}

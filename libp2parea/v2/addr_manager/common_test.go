package addr_manager

import (
	ma "github.com/multiformats/go-multiaddr"
	"testing"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func TestCheckAddr(t *testing.T) {
	testClass := []struct {
		input  string
		expect bool
	}{
		{input: "/ip4/127.0.0.1", expect: false},
		{input: "/ip4/127.0.0.1/tcp/1234", expect: true},
		{input: "/ip4/127.0.0.1/tcp/1234/http/", expect: true},
		{input: "/ip4/127.0.0.1/tcp/1234/ws", expect: true},
		{input: "/ip4/127.0.0.1/udp/1234/quic", expect: true},
		{input: "/ip4/127.0.0.1/tcp/1234/quic", expect: false},
		{input: "/ip6/127.0.0.1/tcp/1234", expect: true},
		{input: "/dns/example.com/tcp/1234", expect: true},
		{input: "/dns/127.0.0.1/tcp/1234", expect: true},
	}

	for _, one := range testClass {
		a, err := ma.NewMultiaddr(one.input)
		if err != nil {
			panic(err)
		}
		_, ERR := engine.CheckAddr(a)
		ok := ERR.CheckSuccess()
		if ok != one.expect {
			if ERR.CheckFail() {
				utils.Log.Info().Str("ERR", ERR.String()).Send()
			}
			t.Errorf("CheckAddr(%q) = %v, expected %v", one.input, ok, one.expect)
		}
	}

	a, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/1234/http/")
	b, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/1235")
	x := ma.Join(a, b)
	utils.Log.Info().Str("合并地址", x.String()).Send()

	addrs := ma.Split(x)
	for _, one := range addrs {
		utils.Log.Info().Str("解析合并地址", one.String()).Send()
	}

}

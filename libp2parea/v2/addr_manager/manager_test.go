package addr_manager

import (
	ma "github.com/multiformats/go-multiaddr"
	"testing"
	"web3_gui/utils"
)

func TestManager(t *testing.T) {
	addr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/65535/")
	if err != nil {
		panic(err)
	}

	manager := NewAddrManager()
	manager.AddSuperPeerAddr(addr)
	ips, dns := manager.LoadAddrForAll()
	utils.Log.Info().Interface("ips", ips).Interface("dns", dns).Send()
}

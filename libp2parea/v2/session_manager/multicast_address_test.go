package session_manager

import (
	"fmt"
	"net"
	"testing"
)

func TestBuildAddrs(t *testing.T) {
	ips := []net.IP{net.IP{192, 168, 0, 1}, net.IP{10, 0, 0, 1}}
	mas := BuildAddrs(ips, 1901)
	for _, one := range mas {
		fmt.Println("地址列表", one.String())
	}
}

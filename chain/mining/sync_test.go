package mining

import (
	"fmt"
	"testing"
	"web3_gui/libp2parea/adapter/nodeStore"
)

func TestPeerBlockInfoDESC(t *testing.T) {
	peers := make([]*PeerBlockInfo, 0)
	addrOne := nodeStore.AddressNet([]byte("123"))
	one1 := PeerBlockInfo{&addrOne, 1}
	peers = append(peers, &one1)
	one2 := PeerBlockInfo{&addrOne, 5}
	peers = append(peers, &one2)
	one3 := PeerBlockInfo{&addrOne, 3}
	peers = append(peers, &one3)
	pbi := NewPeerBlockInfoDESC(peers)
	for _, one := range pbi.Sort() {
		fmt.Println("列表", one)
	}
}

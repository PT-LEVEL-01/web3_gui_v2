package nodeStore

import (
	"testing"
	"web3_gui/utils"
)

func TestLogicNumBuider(t *testing.T) {
	//生成
	//for i := 0; i < 10; i++ {
	//	domain := utils.GetRandomDomain()
	//	utils.Log.Info().Str("domain", domain).Send()
	//}
	domain := utils.GetRandomDomain()
	hashBs := utils.Hash_SHA3_256([]byte(domain))

	buider := NewLogicNumBuider(hashBs, NodeIdLevel)
	buider.GetNodeNetworkNum()
	netIDs := buider.GetNodeNetworkNum()
	for _, netID := range netIDs {
		utils.Log.Info().Msgf("%b", netID)
		//utils.Log.Info().RawCBOR("id", *netID).Send()
		//utils.Log.Info().Hex("id", *netID).Send()
	}
}

package unit

import (
	"bytes"
	"math/big"

	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

/*
多次查询逻辑节点并对比结果是否正确
*/
func LoopSendSearch(areas []*libp2parea.Area) {

	netid := areas[0].GetNetId()
	logicIds := GetLogicNetID(&netid)
	for i, logicOne := range logicIds {
		wantId := globalSearchNode(areas, *logicOne)
		for _, one := range areas {
			utils.Log.Info().Msgf("节点:%s 查找:%s want:%s", one.GetNetId().B58String(), logicOne.B58String(), wantId.B58String())
			magneticId, err := one.SearchNetAddr(logicOne)
			if err != nil {
				utils.Log.Error().Msgf("发送SearchNetAddr消息失败:%s", err.Error())
				continue
			}

			if !bytes.Equal(wantId, *magneticId) {
				utils.Log.Error().Msgf("磁力节点不相等 %d:%s from:%s", i, magneticId.B58String(), one.GetNetId().B58String())
			} else {
				// utils.Log.Info().Msgf("磁力节点 %d:%s", i, magnetic.B58String())
			}

		}
	}

}

/*
得到保存数据的逻辑节点
@idStr  id十六进制字符串
@return 4分之一节点
*/
func GetLogicNetID(id *nodeStore.AddressNet) (logicIds []*nodeStore.AddressNet) {
	bas := nodeStore.BuildArithmeticSequence(100)
	logicIds = make([]*nodeStore.AddressNet, 0, len(bas))
	idInt := new(big.Int).SetBytes(*id)
	for _, one := range bas {
		bs := new(big.Int).Xor(idInt, one).Bytes()
		newbs := utils.FullHighPositionZero(&bs, 32)
		mh := nodeStore.AddressNet(*newbs)
		logicIds = append(logicIds, &mh)
	}
	return
}

func globalSearchNode(areas []*libp2parea.Area, find nodeStore.AddressNet) nodeStore.AddressNet {
	// findNetid := nodeStore.AddressFromB58String("ED7buQ4uJy5DSsYNUKmzxgmUVTDQKRxsbKPwtwgBre3W")

	kad := nodeStore.NewKademlia(0)

	// netBigIds := make([]*big.Int, 0)
	for _, one := range areas {
		kad.Add(new(big.Int).SetBytes(one.GetNetId()))
		// netBigIds = append(netBigIds, new(big.Int).SetBytes(one.GetNetId()))
	}

	dstId := kad.Get(new(big.Int).SetBytes(find))
	dstIdBs := dstId[0].Bytes()
	dstIdBsP := utils.FullHighPositionZero(&dstIdBs, 32)

	return nodeStore.AddressNet(*dstIdBsP)
}

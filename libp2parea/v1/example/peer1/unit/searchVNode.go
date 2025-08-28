package unit

import (
	"bytes"
	"math/big"

	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

/*
多次查询逻辑虚拟节点并对比结果是否正确
*/
func LoopSendSearchVnode(areas []*libp2parea.Area) {
	//第一步：先得到标准的虚拟节点，应该有的逻辑节点，并打印出来。
	//第二部：对比实际虚拟节点和标准的虚拟节点的逻辑是否一致。

	netid := areas[0].Vm.GetVnodeDiscover().Vnode.Vid
	logicIds := GetLogicVnodeNetID(&netid)
	for i, logicOne := range logicIds {
		wantId := globalSearchVNode(areas, *logicOne)
		for _, one := range areas {
			utils.Log.Info().Msgf("虚拟节点:%s 查找:%s want:%s", one.GetNetId().B58String(), logicOne.B58String(), wantId.B58String())
			magneticId, err := one.SearchVnodeId(logicOne)
			if err != nil {
				utils.Log.Error().Msgf("发送SearchVnodeNetAddr消息失败:%s", err.Error())
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
func GetLogicVnodeNetID(id *virtual_node.AddressNetExtend) (logicIds []*virtual_node.AddressNetExtend) {
	bas := nodeStore.BuildArithmeticSequence(100)
	logicIds = make([]*virtual_node.AddressNetExtend, 0, len(bas))
	idInt := new(big.Int).SetBytes(*id)
	for _, one := range bas {
		bs := new(big.Int).Xor(idInt, one).Bytes()
		newbs := utils.FullHighPositionZero(&bs, 32)
		mh := virtual_node.AddressNetExtend(*newbs)
		logicIds = append(logicIds, &mh)
	}
	return
}

func globalSearchVNode(areas []*libp2parea.Area, find virtual_node.AddressNetExtend) virtual_node.AddressNetExtend {
	kad := nodeStore.NewKademlia(0)
	for _, one := range areas {
		vnodeinfos := one.Vm.GetVnodeSelf()
		for _, vnodeAddrOne := range vnodeinfos {
			//虚拟节点中index为0的是发现节点，SearchVnode操作时，排除这些节点
			if vnodeAddrOne.Index == 0 {
				continue
			}
			kad.Add(new(big.Int).SetBytes(vnodeAddrOne.Vid))
		}
	}
	dstId := kad.Get(new(big.Int).SetBytes(find))
	dstIdBs := dstId[0].Bytes()
	dstIdBsP := utils.FullHighPositionZero(&dstIdBs, 32)

	return virtual_node.AddressNetExtend(*dstIdBsP)
}

package nodeStore

import (
	"math/big"
	"web3_gui/utils"
)

const (
	Str_zaro      = "0000000000000000000000000000000000000000000000000000000000000000" //字符串0
	Str_maxNumber = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" //256位的最大数十六进制表示id

	NodeIdLevel uint = 256 //节点id长度

	Node_min = 0 //每个节点的最少连接数量

	Other_area_node_conn_size = 2 //其他逻辑域中，每个域中保存连接的节点数量

	Node_type_all        NodeClass = 0 //包含所有类型
	Node_type_logic      NodeClass = 1 //自己需要的逻辑节点
	Node_type_client     NodeClass = 2 //保存其他逻辑节点连接到自己的节点，都是超级节点
	Node_type_proxy      NodeClass = 3 //被代理的节点
	Node_type_other      NodeClass = 4 //每个节点有最少连接数量
	Node_type_white_list NodeClass = 5 //连接白名单
	Node_type_oneByone   NodeClass = 6 //onebyone规则连接类型

)

var (
	Number_interval   []*big.Int = BuildArithmeticSequence(16) //相隔距离16分之一
	Number_interval06 []*big.Int = BuildArithmeticSequence(6)  //相隔距离6分之一
	Number_quarter    []*big.Int = BuildArithmeticSequence(4)  //最大id的四分之一
)

type NodeClass int

/*
构建等差数列
@num    int    几分之一 值为16，则为十六分之一
*/
func BuildArithmeticSequence(num int) []*big.Int {
	number_interval := make([]*big.Int, 0, num-1)
	Number_max, ok := new(big.Int).SetString(Str_maxNumber, 16)
	if !ok {
		panic("id string format error")
	}
	one_sixteenth := new(big.Int).Div(Number_max, big.NewInt(int64(num)))
	for i := 1; i < num; i++ {
		number_interval = append(number_interval, new(big.Int).Mul(one_sixteenth, big.NewInt(int64(i))))
	}
	return number_interval
}

/*
	得到保存数据的逻辑节点
	@idStr  id十六进制字符串
	@return 16分之一节点
*/
// func GetLogicIds(id *utils.Multihash) (logicIds []*utils.Multihash) {

// 	logicIds = make([]*utils.Multihash, 0)
// 	idInt := new(big.Int).SetBytes(id.Data())
// 	for _, one := range Number_interval {
// 		bs := new(big.Int).Xor(idInt, one).Bytes()
// 		mhbs, _ := utils.Encode(bs, config.HashCode)
// 		mh := utils.Multihash(mhbs)
// 		logicIds = append(logicIds, &mh)
// 	}

// 	return
// }

/*
	得到保存数据的逻辑节点
	@idStr  id十六进制字符串
	@return 4分之一节点
*/
// func GetQuarterLogicAddrNetByAddrCoin(id *crypto.AddressCoin) (logicIds []*AddressNet) {

// 	logicIds = make([]*AddressNet, 0)
// 	logicIds = append(logicIds, id)
// 	idInt := new(big.Int).SetBytes(*id)
// 	for _, one := range Number_quarter {
// 		bs := new(big.Int).Xor(idInt, one).Bytes()
// 		mh := AddressNet(bs)
// 		logicIds = append(logicIds, &mh)
// 	}
// 	return
// }

/*
得到保存数据的逻辑节点
@idStr  id十六进制字符串
@return 4分之一节点
*/
func GetQuarterLogicAddrNetByAddrNet(id *AddressNet) ([]*AddressNet, utils.ERROR) {
	addrPre := id.GetPre()
	logicIds := make([]*AddressNet, 0, 4)
	logicIds = append(logicIds, id)
	idInt := new(big.Int).SetBytes(id.Data())
	for _, one := range Number_quarter {
		bs := new(big.Int).Xor(idInt, one).Bytes()
		newbs := utils.FullHighPositionZero(&bs, 32)
		mh, ERR := BuildAddrByData(addrPre, *newbs)
		if ERR.CheckFail() {
			return nil, ERR
		}
		//mh := AddressNet(*newbs)
		logicIds = append(logicIds, &mh)
	}
	return logicIds, utils.NewErrorSuccess()
}

/*
获得指定节点的100个磁力节点
*/
func GetMagneticID100(id *AddressNet) ([]*AddressNet, utils.ERROR) {
	addrPre := id.GetPre()
	bas := BuildArithmeticSequence(100)
	logicIds := make([]*AddressNet, 0, len(bas))
	idInt := new(big.Int).SetBytes(id.Data())
	for _, one := range bas {
		bs := new(big.Int).Xor(idInt, one).Bytes()
		newbs := utils.FullHighPositionZero(&bs, 32)
		mh, ERR := BuildAddrByData(addrPre, *newbs)
		if ERR.CheckFail() {
			return nil, ERR
		}
		logicIds = append(logicIds, &mh)
	}
	return logicIds, utils.NewErrorSuccess()
}

/*
获得指定节点的16个磁力节点
*/
func GetMagneticID16(addrPre string) ([]*AddressNet, utils.ERROR) {
	//addrPre := id.GetPre()
	logicIds := make([]*AddressNet, 0, 16)
	for i := 0; i < 16; i++ {
		addrBs := big.NewInt(int64(i)).Bytes()
		newbs := utils.FullHighPositionZero(&addrBs, 1)
		//mh := AddressNet(*newbs)
		mh, ERR := BuildAddrByData(addrPre, *newbs)
		if ERR.CheckFail() {
			return nil, ERR
		}
		logicIds = append(logicIds, &mh)
	}
	return logicIds, utils.NewErrorSuccess()
}

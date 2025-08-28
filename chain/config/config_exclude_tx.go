package config

import (
	"encoding/hex"
	"sync"
	// "web3_gui/utils"
	// "web3_gui/keystore/adapter/crypto"
	// "web3_gui/utils"
)

// const BlockRewardHeightOffset = 1600000 //区块奖励有160万高度的偏移量

const Mining_block_start_height_jump = 0 //跳过区块高度不验证签名和交易锁定上链高度//631751//632060/634818/661575/700000//719745//814565/940000/981720/
// const WitnessOrderCorrectStart = 0       //599802/0
// const WitnessOrderCorrectEnd = 0         //719745/981720
const RandomHashHeightMin = 0 //用一个无限高度控制随机数//
const NextHashHeightMax = 0   //
const ChainGCModeHeight = 0   //根据高度使用不同的链回收机制

// const CheckAddBlacklistChangeHeight = 100000 //此高度之前，条件判断错误。此高度之后纠正之

/*
1169333    5c91873af3f9b6d9b18c649d2c0a4c49b2540a31fee26546e259bc5fce88c2e9  +

1169334    ef57d947d2f7a10d5ade6024e4bef1c74743f1d79807914ee9e268e6074d71c6  +
1169335    77ce8deff0c43893e7fba2312d934f32a715063aa7ade53ae539bdcfa012b8a1  -
1169335    b89ace0f46b39cbddec13b6cd3bc3cba99cff244477e207928d0cd7d625d9907  +

1169336    e00171f80fb6dddaeb1aec410169c2348e0514a373f03e189380393cf244da36  -
1169336    9910e28c5868fbfa56e4cd052c01725b2b28f7462827c2e0eccba6bad1506fed  +
1169337    09f5d7c1211d190a7d93afd734412277b48fdc8ef2179552e0ca2ea81c2b44a4  +

1169338    2727a5e5361425d58a59275a1346bab2470310b387a45c3784dcdcae79b86bbd  +
*/
// const FixBuildGroupBUGHeightMax = 0 //1367480

var RandomHashFixed = []byte{}

const RandomHashFixedStr = "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"

var SpecialAddrs = []byte{}

var SpecialBlockHash = []byte{}

// 区块回滚高度
var CutBlockHeight = uint64(0)
var CutBlockHash = []byte{}

var NextHash = new(sync.Map)

var BlockNextHash = new(sync.Map) //临时修改区块的nexthash

var RandomMap = new(sync.Map) //make(map[uint64]*[]byte)

var Exclude_Tx = []ExcludeTx{
	// ExcludeTx{781928, "01000000000000000353039004084e7a5111fccf02464e67b0397e9d09893fbc64545fb778c1111a", nil},
	// ExcludeTx{781949, "0100000000000000c2e0efd51a46e419a768dc3959442365db60e1f43b3d90887711740ac75d6966", nil},
	// ExcludeTx{786577, "0400000000000000d59e4560f40a275c345f34e8673a0d092e84e9732d3771ad81aa03cbb9608c7f", nil},
	// ExcludeTx{786721, "04000000000000006987aa770b5dff96258a8fc024fc9b83051613cbf4f647b3a1c439f6f44a7633", nil},
	// ExcludeTx{786747, "040000000000000058badafa534e3a3a52bdaffab50bafe8e0180d80d76abb51b1a9bb5443feac26", nil},
	// ExcludeTx{790151, "0400000000000000f57c4145566e51fd99ba3b7bc25e4ee989c9eb7f9648e118c976a9ff525635bc", nil},
	// ExcludeTx{792068, "0400000000000000f448e3bc213b15eafb2809a1b8d50295dcec0d552cbe0bbb98736750f5dfb319", nil},
	// ExcludeTx{802451, "0400000000000000b75ee1f88f729e64e8efd1f7393ac59f46d8c17a8f16801c4fed67810234bb23", nil},
	// ExcludeTx{810149, "07000000000000002df50e6c9576b14362b92fdf444494f3096f17ec9129bd6b39459b587f598da8", nil},
	// ExcludeTx{810154, "07000000000000006f5a027990165260dc653c66ef0cf6d3160016af9b8aca3f2ad2ad9a1025f469", nil},
	// ExcludeTx{815935, "040000000000000095eaa3b45aeb13d1f4d76db1b2277ee70225491b0d9bbe4fc4511b63a9675167", nil},
}

var BlockHashs = make([][]byte, 0)

// key:某段连续区块的首块hash,value:连续的区块hash数组
var BlockHashsMap = make(map[string][][]byte, 0)

/*
1.31个以内，按现有见证人数量平均分配（假如只有5个见证人，则5个见证人平均分）。
2.31-99个，均分部分按现有见证人数量均分。
3.大于99个见证人，均分部分给前99个见证人均分。排名99之后的见证人没有奖励。
*/
const Mining_witness_average_height = 0 //超过这一高度，区块奖励按新的规则算//624827
const Reward_witness_height = 0         //新版奖励起始高度//745333
/*
上一版本的奖励还有一处bug，就是保存的见证人排序是按投票数量排序，
然而寻找出块的时候，prehash导致不连续，导致寻找的出块见证人数量比预计的少，奖励也变少。
新版将解决这个问题。
*/
const Reward_witness_height_new = 0 //新版奖励起始高度//800000

var DBUG_import_height_max = uint64(0) //dbug模式，只加载到n块高度

const Store_name_new = "storereward"
const Store_name_new_height = 0

func init() {
	// CutBlockHeight = 9574
	// blockhashStr := "b7badbf0fcef73365aa2a03e19f72ca530288054374fadf63be89b6be9375ddb"
	// CutBlockHash, _ = hex.DecodeString(blockhashStr)

	BuildRandom()
	BuildNextHash()
	CustomOrderBlockHash()

	for i, one := range Exclude_Tx {
		bs, err := hex.DecodeString(one.TxStr)
		if err != nil {
			panic("交易hash不规范:" + err.Error())
		}
		Exclude_Tx[i].TxByte = bs
		Exclude_Tx[i].TxStr = ""
		// one.TxByte = bs
		// one.TxStr = ""
	}
	// engine.Log.Info("打印需要排除的交易hash %+v", Exclude_Tx)
}

type ExcludeTx struct {
	Height uint64
	TxStr  string
	TxByte []byte
}

func BuildRandom() {
	// addrs := "MMS792Dg5nQQvpYoajJtcrsCR4vFG4xeyU9z4"
	// SpecialAddrs = crypto.AddressFromB58String(addrs)

	// blockhashStr := "672fbd60d432c899f41d32e47ce8965125a378f79c27a7330ee97d5aae34c184"
	// SpecialBlockHash, _ = hex.DecodeString(blockhashStr)

	// RandomHashFixed, _ = hex.DecodeString(RandomHashFixedStr) //599803

	// blockhash, _ := hex.DecodeString("327ed7c01ac82c31f3454f4a652921e4c14045c1220b566d08e8241358b93a03") //599803
	// random, _ := hex.DecodeString("536c6049a7310afb957553ef9643a6efff586951cddb9832a556fbca912ced34")
	// RandomMap.Store(utils.Bytes2string(blockhash), &random)

	// blockhash, _ = hex.DecodeString("c7e52c1922ff9e7d9de87eca78e25ccf378c281b6340f982f5837c72fafad618") //630456
	// random, _ = hex.DecodeString("8b4a13c54087342bd20979ef8e0a10045897cf5a0473b2a95f192f6fb004b031")
	// RandomMap.Store(utils.Bytes2string(blockhash), &random)

	// blockhash, _ = hex.DecodeString("a65a7bf2f888876c10b2d3ec7827d59cc8f4f22cd7538d02cde186724bdfa036") //599800
	// random, _ = hex.DecodeString("903d5e2e20aba4464bbfc929419d7eb5b63abb8227ac56b1feec63b0c15fc0fd")
	// RandomMap.Store(utils.Bytes2string(blockhash), &random)

}

/*
修改nexthash
*/
func BuildNextHash() {
	// blockHash, _ := hex.DecodeString("5e8515f972178a7d4fb31296790c95748e9ad357b20f70d337467bd4b459c97b")
	// nextHash, _ := hex.DecodeString("8d76763b970127d84cbe097c78575b9bd1bcabda6d3bf99416a0a1fa3040cd57")
	// NextHash.Store(utils.Bytes2string(blockHash), &nextHash)

	// blockhash, _ = hex.DecodeString("a65a7bf2f888876c10b2d3ec7827d59cc8f4f22cd7538d02cde186724bdfa036")
	// random, _ = hex.DecodeString("2d53b5e4e21ba7ff9508418e450dc17aeafc0c4e007316a45e19199c92d34b85")
	// NextHash.Store(utils.Bytes2string(blockhash), &random)
}

/*
按照指定的顺序加载区块
一个测试方法，加载到这里的第一个区块后，就按照这里的顺序加载指定的区块
*/
func CustomOrderBlockHash() {
	if true {
		//添加一段顺序加载区块
		addCustomOrderBlockHash(
		//"9c0a8c3f3fb2bd0347707473ab1bd82730774977f95c7225d67f4fe3325f8aa7",
		//"74f374a90889d25591849b2b49f1b0c75a9756aadc260875b7d9261cbf75ddba",
		//"26a18b388bd48d7c33b156a2ffb3855836855eb6c24634551620bb34fd8f946a",
		//"f63b1e86f1758f7178e998ae810f93339b9c5c3469085b9b7603a6999177b473",
		//"d5a1515d560d961efee122c8faaae4852068a24751bc6145fcf0847001003752",
		//"7d51c5eaa1cbe441df89f5fc3a25f36f9e3c2a02fb0d41da13900cfd662b52bf",
		//"1eb1634af6be8b17ea3aacfc27e8552c85638cd7b2d032db6dde66f8c9551fd3",
		//"c09339d8de81265cd398ce12432c8a4e8050f1e080c7aae24ed8682731f0fde5",
		//"2ef84fd626aa34fae7c066014c873907d5e58a420ef16b5709e8d9341e80c876",
		//"44eee9d181e0d6791ef21406b27e8e65bd504e5ca851ab268454d7ff861b9b91",
		//"093ec9d23645f633198e0fb0899fb1d7f1d252412658997b0b402827a9ef0b13",
		//"ccd9f1ec2dd25f02d6a47f5e543f5880a72a353ec58686e655cb0c32972e9f48",
		)
	}
}

// 添加一段区块加载顺序
// 首块哈希为key,连续哈希为value的Map表
func addCustomOrderBlockHash(hashs ...string) {
	if len(hashs) < 2 {
		return
	}

	blockHashs := make([][]byte, 0)
	firstHash := ""
	for i, hashStr := range hashs {
		if i == 0 {
			firstHash = hashStr
		}
		hash, _ := hex.DecodeString(hashStr)
		blockHashs = append(blockHashs, hash)
	}
	BlockHashsMap[firstHash] = blockHashs
}
func UpdateSoreName(height uint64) {
	// if height == Store_name_new_height {
	// 	Name_store = Store_name_new
	// }
}

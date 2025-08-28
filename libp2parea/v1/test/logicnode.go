package main

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

func main() {
	BuildIds()

	// distance()

	// ExampleBuildLogicID()

}

/*
测试查询逻辑节点
*/
func BuildIds() {

	fmt.Println("---------- BuildIds ----------")

	ids := []string{
		"GJMNovKvmN1VeYHCjP2b27NkXxkRNDdVKrziTzy4FPyM",
		"FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n",
		"4MTCfHur4XPmRhpUByGC7Z9ve3pbezDyiYuYM6cA5Lmj",
		"6Wrco6JKC5c7pg8qp5Z5mNnayHqdW1eyLExHboihCY4c",
		"5o1oZJwByLAgT36BXXbBRDJ8UDhUUY4JLJbjR3aG6ZEp",
		"qUX2PsQATzSaofU6Q7sGmipu1g2V1H2KwgdzcqZsbN7",
		"GQHsVDddC44AM6z4LmS6f1brkJn1cNYhrms75ZEwchAG",
	}
	// 本节点为 Bsyuy8Cpg5VWi69axQKaU6pLbHkWHffCDjcQEFJC1qEr
	// --逻辑节点 D79GZyCBcNKyvc3pHco6nvnzJuTGiaYDLqFjkLG1fkcR
	// --逻辑节点 2w5QBfujmLTAvesJRyRpxZFj4D4PJTEbhDVQJt1kbDmk

	// ids := []string{

	// 	"HCzFqZSNAZysyBazDQRtpFzy6eq7JigKVbQSqp3Zw1m",

	// 	"BL9ZL1qX8vGRNUaZEq1E4svhjX1fiFm8iYJrMLHV7NTG",
	// 	"8VmyBM5XmJRzPmQu7TC82TxdkX5BNc87sMaSfyft9fXW",
	// 	"9eQ642R1zttsGw4ikdn5vEtG2f18GGiV2NBXL9gsszmM",
	// 	"CsZqHcCVTukktiBDv9Cc1ELrhN7KVKwGnwzPNG2a9obV",

	// 	"3yk5sZhd7o7PGT6i5muUXSCHDHxjoDAWuaMbqX1K1NXT",
	// 	"EuNqmo354mxgizaZUgNkd59M2ho8nPWMs4QEN13W37Xo",
	// 	"GKr849QTWzmhfMkpT4iatrvQNmREXBZrAkwHNmCDZx5m",
	// 	"AxwnZMGyNsLSLHRB4nN2AvuFodtLZwYkCdC5L49beyEE",
	// }

	for n := 0; n < len(ids); n++ {
		fmt.Println("本节点为", ids[n])
		index := n

		idMH := nodeStore.AddressFromB58String(ids[index])
		idsm := nodeStore.NewIds(idMH, 256)
		for i, one := range ids {
			if i == index {
				continue
			}

			idMH := nodeStore.AddressFromB58String(one)
			idsm.AddId(idMH)
			//		ok, remove := idsm.AddId(&idMH)
			//		if ok {
			//			fmt.Println(one, remove)
			//		}
		}

		is := idsm.GetIds()
		for _, one := range is {

			idOne := nodeStore.AddressNet(one)

			fmt.Println("--逻辑节点", idOne.B58String())
		}

	}

}

/*
计算节点距离
*/
func distance() {

	fmt.Println("---------- distance ----------")

	ids := []string{
		"W1aLWC4unTJZhSFc4VNLFsazAJ1PyTocV7agmteQDL3J3N",
		"W1gfVGa52yUJ4Gws4TiA9YbwGP8qCGgaYeeT8APjSiNk6U",
		"W1j9RJ1xYHaoAuRk2HGBrVA82njoxFAoctYKQMH43k8hXu",
		"W1atFt7bJ5Ubk4MXuV5GfsEYE7srWXR51exDgUEJcVr5fZ",
		"W1n9XtbLAjRsh9sr2kbwfkfy3VGenyhazbHJwrEYsnDZ8M",
	}

	index := 4

	kl := nodeStore.NewKademlia(0)
	for i, one := range ids {
		if i == index {
			continue
		}

		idMH, _ := utils.FromB58String(one)
		kl.Add(new(big.Int).SetBytes(idMH.Data()))

	}

	idMH, _ := utils.FromB58String(ids[index])
	is := kl.Get(new(big.Int).SetBytes(idMH.Data()))
	src := new(big.Int).SetBytes(idMH.Data())

	//	is := idsm.GetIds()
	for _, one := range is {
		tag := new(big.Int).SetBytes(one.Bytes())
		juli := tag.Xor(tag, src)

		bs, err := utils.Encode(one.Bytes(), utils.SHA3_256)
		if err != nil {
			fmt.Println("编码失败")
			continue
		}
		mh := utils.Multihash(bs)

		fmt.Println("排序结果", mh.B58String(), "距离", hex.EncodeToString(juli.Bytes()))
	}

}

func ExampleBuildLogicID() {
	fmt.Println("---------- ExampleBuildLogicID ----------")

	baseIDStr := "23Qc9uX7bxJi9eyf3NhtVyDqSTE354ENj8U6XYdb9LSD"
	idStr := []string{
		"AfwFmdvpqjXmVpR7MP1vQonLucaf1rQg6v7BJ9W6AirL",
		"23Qc9uX7bxJi9eyf3NhtVyDqSTE354ENj8U6XYdb9LSD",
		"4WWhd7qwfNtNDEJPz6X67E9aSVzvtG7uicJgmph99Ag6",
		"65U2vEhNAq2PNzMT9SCK2d6sZCULSFvD8vvLwRZpJkgx",
	}

	baseID := nodeStore.AddressFromB58String(baseIDStr)
	sourceID := make([]nodeStore.AddressNet, 0, len(idStr))
	for _, one := range idStr {
		oneID := nodeStore.AddressFromB58String(one)
		sourceID = append(sourceID, oneID)
	}

	logicIDs := BuildNeedLogicID(baseID, sourceID)
	for _, one := range logicIDs {
		// idOne := nodeStore.AddressNet(one)
		fmt.Println("--逻辑节点", one.B58String())
	}
}

/*
构建需要的逻辑节点
*/
func BuildNeedLogicID(baseID nodeStore.AddressNet, sourcesID []nodeStore.AddressNet) []nodeStore.AddressNet {
	idsm := nodeStore.NewIds(baseID, 256)
	for _, one := range sourcesID {
		// utils.Log.Info().Msgf("1111:%s", one.B58String())
		idsm.AddId(one)
	}

	is := idsm.GetIds()
	results := make([]nodeStore.AddressNet, 0, len(is))
	for _, one := range is {
		addrOne := nodeStore.AddressNet(one)
		results = append(results, addrOne)
	}
	return results
	// for _, one := range is {

	// 	idOne := nodeStore.AddressNet(one)

	// 	fmt.Println("--逻辑节点", idOne.B58String())
	// }
}

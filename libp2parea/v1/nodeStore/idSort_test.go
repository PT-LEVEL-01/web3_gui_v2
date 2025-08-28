package nodeStore

import (
	"fmt"
	"math/big"
	"sort"
	"testing"
)

func TestIdSort(t *testing.T) {
	idStoreSimple1()
	AddressSort()
}

func idStoreSimple1() {

	desc := new(IdDESC)

	node1, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	*desc = append(*desc, node1)
	node2, _ := new(big.Int).SetString("31622036050853307757176718873676335712993063093791913422933189278586653352673", 10)
	*desc = append(*desc, node2)
	node3, _ := new(big.Int).SetString("38879061860890225964363770808076149471375052911854164467748691902681942298885", 10)
	*desc = append(*desc, node3)
	node4, _ := new(big.Int).SetString("59422813065590763321187925186011450884940934337897117431794152839561407098597", 10)
	*desc = append(*desc, node4)
	// node5, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	// desc = append(desc, node1)
	// node6, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	// desc = append(desc, node1)
	// node7, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	// desc = append(desc, node1)
	// node8, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	// desc = append(desc, node1)

	sort.Sort(desc)

	for i := 0; i < len(*desc); i++ {
		fmt.Println((*desc)[i].String())
	}
}

func AddressSort() {
	var addrs []AddressNet
	addrs = append(addrs, AddressFromB58String("EW8d17xHcCWaeTWoHzi1uxLiUovAm1V8zvUZYWYn1jZo"))
	addrs = append(addrs, AddressFromB58String("4ba8NCmZUvk5LBpP3CUNWD6vF2Wj2mpthESVj54yaNo4"))
	addrs = append(addrs, AddressFromB58String("FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n"))
	addrs = append(addrs, AddressFromB58String("6Wrco6JKC5c7pg8qp5Z5mNnayHqdW1eyLExHboihCY4c"))

	sort.Sort(AddressBytes(addrs))
	for i := 0; i < len(addrs); i++ {
		fmt.Println("AddressBytes ", i, addrs[i].B58String())
	}

}

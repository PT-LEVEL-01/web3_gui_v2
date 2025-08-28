package nodeStore

import (
	"math/big"
	"sort"
	"web3_gui/utils"
)

/*
K桶
*/
type Bucket struct {
	findNode *big.Int            //
	nodes    []*big.Int          //用于排序的节点
	nodeMap  map[string]*big.Int //索引，去重
}

func NewBucket(length int) *Bucket {
	k := &Bucket{
		nodes:   make([]*big.Int, 0, length),
		nodeMap: make(map[string]*big.Int, length),
	}
	return k
}

func (this *Bucket) Len() int {
	return len(this.nodes)
}

func (this *Bucket) Less(i, j int) bool {
	a := new(big.Int).Xor(this.findNode, this.nodes[i])
	b := new(big.Int).Xor(this.findNode, this.nodes[j])
	if a.Cmp(b) > 0 {
		return false
	} else {
		return true
	}
}

func (this *Bucket) Swap(i, j int) {
	this.nodes[i], this.nodes[j] = this.nodes[j], this.nodes[i]
}

/*
添加节点
*/
func (this *Bucket) Add(nodes ...*big.Int) {
	for _, idOne := range nodes {
		//判断重复的
		key := utils.Bytes2string(idOne.Bytes())
		_, ok := this.nodeMap[key]
		if ok {
			continue
		}
		this.nodes = append(this.nodes, idOne)
		this.nodeMap[key] = idOne
	}
}

/*
获得这个节点由近到远距离排序的节点列表
*/
func (this *Bucket) Get(nodeId *big.Int) []*big.Int {
	this.findNode = nodeId
	sort.Sort(this)
	return this.nodes
}

package chain_orders

import "sync"

type GoodsBase struct {
	lock           *sync.RWMutex //
	GoodsId        []byte        //商品id
	Name           string        //商品名称
	Price          uint64        //价格
	SalesTotalMany bool          //虚拟商品，销售数量无限多
	SalesTotal     uint64        //销售总量
	SoldTotal      uint64        //已经卖出量
	LockTotal      uint64        //未付款，锁定数量
}

func NewGoodsBase() *GoodsBase {
	gb := GoodsBase{
		lock:           new(sync.RWMutex),
		GoodsId:        nil,
		Name:           "",
		Price:          0,
		SalesTotalMany: false,
		SalesTotal:     0,
		SoldTotal:      0,
		LockTotal:      0,
	}
	return &gb
}

func (this *GoodsBase) AddSalesTotal(goodsId []byte, salesTotalMany bool, total, price uint64) {
	this.lock.Lock()
	this.SalesTotalMany = salesTotalMany
	this.SalesTotal += total
	this.Price = price
	this.lock.Unlock()
}

func (this *GoodsBase) GetPrice() uint64 {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.Price
}

/*
锁定一个商品
*/
func (this *GoodsBase) LockOne() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.SalesTotalMany {
		this.LockTotal++
		return true
	}
	if this.SalesTotal <= (this.SoldTotal + this.LockTotal) {
		return false
	}
	this.LockTotal++
	return true
}

/*
解锁一个商品
*/
func (this *GoodsBase) UnLockOne() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.LockTotal--
	this.SoldTotal++
	return true
}

/*
回滚一件商品
*/
func (this *GoodsBase) RollBackOne() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.LockTotal--
	return true
}

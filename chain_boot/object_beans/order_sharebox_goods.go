package object_beans

import (
	"github.com/gogo/protobuf/proto"
	"sync"
	"web3_gui/chain_boot/object_beans/object_beans_protos/protos_object_beans"
)

func init() {
	RegisterObjectClass(CLASS_SHAREBOX_goods_file, ParseOrderShareboxGoods)
}

type OrderShareboxGoods struct {
	ObjectBase
	lock           *sync.RWMutex //
	GoodsId        []byte        //商品id
	Name           string        //商品名称
	Price          uint64        //价格
	SalesTotalMany bool          //虚拟商品，销售数量无限多
	SalesTotal     uint64        //销售总量
	SoldTotal      uint64        //已经卖出量
	LockTotal      uint64        //未付款，锁定数量
}

func NewOrderShareboxGoods(goodsId []byte, salesTotalMany bool, total, price uint64) *OrderShareboxGoods {
	osg := OrderShareboxGoods{
		ObjectBase:     ObjectBase{Class: CLASS_SHAREBOX_goods_file},
		lock:           new(sync.RWMutex), //
		GoodsId:        goodsId,           //商品id
		Name:           "",                //商品名称
		Price:          price,             //价格
		SalesTotalMany: salesTotalMany,    //虚拟商品，销售数量无限多
		SalesTotal:     total,             //销售总量
		SoldTotal:      0,                 //已经卖出量
		LockTotal:      0,                 //未付款，锁定数量
	}
	return &osg
}

func (this *OrderShareboxGoods) GetId() []byte {
	return this.GoodsId
}

func (this *OrderShareboxGoods) Proto() (*[]byte, error) {
	proxyBase := this.GetProto()
	base := protos_object_beans.ObjectShareboxGoods{
		Base:           proxyBase,
		GoodsId:        this.GoodsId,        //商品id
		Name:           this.Name,           //商品名称
		Price:          this.Price,          //价格
		SalesTotalMany: this.SalesTotalMany, //虚拟商品，销售数量无限多
		SalesTotal:     this.SalesTotal,     //销售总量
		SoldTotal:      this.SoldTotal,      //已经卖出量
		LockTotal:      this.LockTotal,      //未付款，锁定数量
	}
	bs, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
创建初始化日志记录
初始化指令中必须包含完整的好友列表
*/
func ParseOrderShareboxGoods(bs []byte) (ObjectItr, error) {
	order := protos_object_beans.ObjectShareboxGoods{}
	err := proto.Unmarshal(bs, &order)
	if err != nil {
		return nil, err
	}
	clientBase := ConvertCommonBase(order.Base)
	addFriend := OrderShareboxGoods{
		ObjectBase:     *clientBase,
		GoodsId:        order.GoodsId,        //商品id
		Name:           order.Name,           //商品名称
		Price:          order.Price,          //价格
		SalesTotalMany: order.SalesTotalMany, //虚拟商品，销售数量无限多
		SalesTotal:     order.SalesTotal,     //销售总量
		SoldTotal:      order.SoldTotal,      //已经卖出量
		LockTotal:      order.LockTotal,      //未付款，锁定数量
	}
	return &addFriend, nil
}

func (this *OrderShareboxGoods) AddSalesTotal(salesTotalMany bool, total, price uint64) {
	this.lock.Lock()
	this.SalesTotalMany = salesTotalMany
	this.SalesTotal += total
	this.Price = price
	this.lock.Unlock()
}

/*
锁定一个商品
*/
func (this *OrderShareboxGoods) LockOne() bool {
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
func (this *OrderShareboxGoods) UnLockOne() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.LockTotal--
	this.SoldTotal++
	return true
}

/*
回滚一件商品
*/
func (this *OrderShareboxGoods) RollBackOne() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.LockTotal--
	return true
}
